package pkg

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/danielfoehrkn/kubectlSwitch/types"
	"github.com/sirupsen/logrus"
)

var (
	logger = logrus.New()
)

func Hooks(configPath string, stateDirectory string, flagHookName string, runImmediately bool) error {
	config, err := LoadConfigFromFile(configPath)
	if err != nil {
		return err
	}

	// only log if explicitly requested to run hooks
	// otherwise silently fail (for normal execution with switcher)
	if config == nil && runImmediately {
		logger.Infof("Configuration file not found under path: %q", configPath)
	}

	if config == nil || len(config.Hooks) == 0 {
		return nil
	}

	// create hook state directory
	err = os.Mkdir(stateDirectory, 0700)
	if err != nil && !os.IsExist(err) {
		return err
	}

	var hooksToBeExecuted []types.Hook
	if len(flagHookName) > 0 {
		hook := getHookForName(config, flagHookName)
		if hook == nil {
			return fmt.Errorf("no hook with name %q found", flagHookName)
		}
		hooksToBeExecuted = append(hooksToBeExecuted, *hook)
	} else if runImmediately {
		hooksToBeExecuted = config.Hooks
	} else {
		hooksToBeExecuted = getHooksToBeExecuted(config.Hooks, stateDirectory)
	}

	if len(hooksToBeExecuted) == 0 {
		logger.Debug("No hooks need to be executed.")
		return nil
	}

	for _, hook := range hooksToBeExecuted {
		stateFileName := getHookStateFileName(hook.Name, stateDirectory)
		if err := updateHookState(hook.Name, stateFileName); err != nil {
			return err
		}

		if err := executeHook(hook); err != nil {
			logger.Error(err)
		}
	}

	return nil
}

func getHookForName(c *types.Config, name string) *types.Hook {
	for _, hook := range c.Hooks {
		if hook.Name == name {
			return &hook
		}
	}
	return nil
}

func getHooksToBeExecuted(hooks []types.Hook, stateDir string) []types.Hook {
	var hooksToBeExecuted []types.Hook
	for _, hook := range hooks {
		if hook.Type != types.HookTypeExecutable && hook.Type != types.HookTypeInlineCommand {
			continue
		}

		if hook.Execution == nil || hook.Execution.Interval == nil {
			hooksToBeExecuted = append(hooksToBeExecuted, hook)
			continue
		}

		stateFileName := getHookStateFileName(hook.Name, stateDir)
		// check by reading the hook state
		hookState, err := getHookState(stateFileName)
		if err != nil {
			logger.Warnf("failed to get hook state for %q", hook.Name)
			continue
		}

		// first hook invocation or state deleted
		if hookState == nil {
			hooksToBeExecuted = append(hooksToBeExecuted, hook)
			continue
		}

		if time.Now().UTC().After(hookState.LastExecutionTime.UTC().Add(*hook.Execution.Interval)) {
			logger.Infof("Hook %q has not been run in %s.", hook.Name, hook.Execution.Interval.String())
			hooksToBeExecuted = append(hooksToBeExecuted, hook)
		}
	}
	return hooksToBeExecuted
}

func getHookStateFileName(hookName string, stateDir string) string {
	stateFileName := fmt.Sprintf("%s/hookstate-%s.yaml", stateDir, hookName)
	return stateFileName
}

func executeHook(hook types.Hook) error {
	logger.Infof("Executing hook %q...", hook.Name)

	var cmd *exec.Cmd
	if hook.Type == types.HookTypeInlineCommand {
		arguments := []string{"-c"}
		arguments = append(arguments, hook.Arguments...)
		cmd = exec.Command("bash", arguments...)
	} else {
		// HookTypeExecutable
		if hook.Path == nil || len(*hook.Path) == 0 {
			return fmt.Errorf("cannot execute hook %q - no executable path set", hook.Name)
		}

		if _, err := os.Stat(*hook.Path); err != nil {
			return fmt.Errorf("cannot find executable for hook with name %q. File does not exist: %q", hook.Name, *hook.Path)
		}
		cmd = exec.Command(*hook.Path, hook.Arguments...)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println(err)
	}

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("error running hook %q: %+v", hook.Name, err)
	}

	// print the output of the subprocess
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		m := scanner.Text()
		logger.Info(m)
	}
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("error waiting for hook %q: %+v", hook.Name, err)
	}
	return nil
}
