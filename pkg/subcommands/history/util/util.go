// Copyright 2021 Daniel Foehr
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

package util

import (
	"bufio"
	"fmt"
	"os"
)

const (
	// historyFileName is a constant for the filename storing the history of contexts
	historyFileName = "$HOME/.kube/.switch_history"
)

// AppendContextToHistory appends the given context (should include the parent folder name for uniqueness)
// to the history state file
func AppendContextToHistory(context string) error {
	fileName := os.ExpandEnv(historyFileName)
	f, err := os.OpenFile(fileName,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.WriteString(fmt.Sprintf("%s\n", context)); err != nil {
		return err
	}

	return nil
}

// ReadHistory reads the context history from the state file
func ReadHistory() ([]string, error) {
	fileName := os.ExpandEnv(historyFileName)
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append([]string{scanner.Text()}, lines...)
	}
	return lines, scanner.Err()
}
