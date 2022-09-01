// Copyright 2021 The Kubeswitch authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import (
	"time"

	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	// HookTypeExecutable defines a hook that contains an external binary executable to be called
	HookTypeExecutable HookType = "Executable"
	// HookTypeInlineCommand defines a hook that directly contains bourne shell
	// commands to be executed
	HookTypeInlineCommand HookType = "InlineCommand"
)

// ValidHookTypes contains all valid hook types
var ValidHookTypes = sets.NewString(string(HookTypeInlineCommand), string(HookTypeExecutable))

// HookType is the type of hook (either "Executable" or "InlineCommand")
type HookType string

// Hook contains configurations for a Hook
type Hook struct {
	// Name is the name of the Hook
	Name string `yaml:"name"`
	// Type is the type of the Hook (either "Executable" or "InlineCommand")
	Type HookType `yaml:"type"`
	// Path defines the path to the executable that shall be called when the
	// Type is "Executable"
	Path *string `yaml:"path"`
	// Arguments are the arguments with which the external executable of a Hook of type
	// "Executable" will be called
	Arguments []string `yaml:"arguments"`
	// Execution contains configuration regarding the execution of the Hook
	Execution *struct {
		// Interval defines the interval how often the Hook is executed
		// if this field is not set, it can only be executed on demand with "switch hooks --hook-name <name>"
		Interval *time.Duration `yaml:"interval"`
	} `yaml:"execution"`
}

// HookState contains the definition for the hook state
type HookState struct {
	// HookName is the name of the Hooks
	HookName string `yaml:"hookName"`
	// LastExecutionTime is the last execution time of the hook
	// used to check if the Hook has to be executed again
	LastExecutionTime time.Time `yaml:"lastExecutionTime"`
}
