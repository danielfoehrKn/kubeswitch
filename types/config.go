package types

import "time"

type HookType string

const (
	HookTypeExecutable    HookType = "Executable"
	HookTypeCommand HookType = "Command"
)


type Config struct {
	Kind string `yaml:"kind"`
	Hooks []Hook `yaml:"hooks"`
}

type Hook struct {
	Name    string `yaml:"name"`
	Type    HookType `yaml:"type"`
	Path    *string `yaml:"path"`
	Arguments []string `yaml:"arguments"`
	Execution *struct {
		Interval    *time.Duration `yaml:"interval"`
	} `yaml:"execution"`
}

type HookState struct {
	HookName    string `yaml:"hookName"`
	LastExecutionTime  time.Time `yaml:"lastExecutionTime"`
}