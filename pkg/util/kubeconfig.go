package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

const (
	// kubeconfigCurrentContext is a constant for the current context in a kubeconfig file
	kubeconfigCurrentContext = "current-context:"
	// TemporaryKubeconfigDir is a constant for the directory where the switcher stores the temporary kubeconfig files
	TemporaryKubeconfigDir = "$HOME/.kube/.switch_tmp"
)

// GetContextWithoutFolderPrefix returns the real kubeconfig context name
// selectable kubeconfig context names have the folder prefixed like <parent-folder>/<context-name>
func GetContextWithoutFolderPrefix(path string) string {
	split := strings.Split(path, "/")
	return split[len(split)-1]
}

func SetCurrentContextOnTemporaryKubeconfigFile(kubeconfigData []byte, ctxnName string) (string, error) {
	currentContext := fmt.Sprintf("%s %s", kubeconfigCurrentContext, ctxnName)

	lines := strings.Split(string(kubeconfigData), "\n")

	foundCurrentContext := false
	for i, line := range lines {
		if !strings.HasPrefix(line, "#") && strings.Contains(line, kubeconfigCurrentContext) {
			foundCurrentContext = true
			lines[i] = currentContext
		}
	}

	output := strings.Join(lines, "\n")
	tempDir := os.ExpandEnv(TemporaryKubeconfigDir)
	err := os.Mkdir(tempDir, 0700)
	if err != nil && !os.IsExist(err) {
		return "", err
	}

	// write temporary kubeconfig file
	file, err := ioutil.TempFile(tempDir, "config.*.tmp")
	if err != nil {
		return "", err
	}

	_, err = file.Write([]byte(output))
	if err != nil {
		return "", err
	}

	// if written file does not have the current context set,
	// add the context at the last line of the file
	if !foundCurrentContext {
		return file.Name(), appendCurrentContextToTemporaryKubeconfigFile(file.Name(), currentContext)
	}

	return file.Name(), nil
}

func appendCurrentContextToTemporaryKubeconfigFile(kubeconfigPath, currentContext string) error {
	file, err := os.OpenFile(kubeconfigPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.WriteString(currentContext); err != nil {
		return err
	}
	return nil
}
