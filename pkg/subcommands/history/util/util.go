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

package util

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// historyFilePath is a constant for the filename storing the history of namespaces
const historyFilePath = "$HOME/.kube/.switch_history"

// ReadHistory reads the context history from the state file
func ReadHistory() ([]string, error) {
	fileName := os.ExpandEnv(historyFilePath)
	file, err := os.Open(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no history entries yet - please run `switch` first")
		}
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

// AppendToHistory appends the given context: namespace to the history file
func AppendToHistory(context, namespace string) error {
	filepath := os.ExpandEnv(historyFilePath)
	f, err := os.OpenFile(filepath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	historyEntry := fmt.Sprintf("%s:: %s\n", context, namespace)

	lastHistoryEntry, err := getLastLineWithSeek(filepath)
	if err != nil {
		return err
	}

	// do not entry history entry if previous entry is identical
	if historyEntry == lastHistoryEntry {
		return nil
	}

	if _, err := f.WriteString(historyEntry); err != nil {
		return err
	}

	return nil
}

// ParseHistoryEntry takes a history entry as argument and returns the context as first, and the namespace as seconds
// return parameter
func ParseHistoryEntry(entry string) (*string, *string, error) {
	split := strings.Split(entry, "::")
	if len(split) == 1 {
		// only context is set (compatibility with old context-only history)
		return &split[0], nil, nil
	} else if len(split) == 2 {
		trimWhitespace := strings.ReplaceAll(split[1], " ", "")
		return &split[0], &trimWhitespace, nil
	}
	return nil, nil, fmt.Errorf("history entry with unrecognized format")
}

// taken from: https://newbedev.com/how-to-read-last-lines-from-a-big-file-with-go-every-10-secs
func getLastLineWithSeek(filepath string) (string, error) {
	fileHandle, err := os.Open(filepath)

	if err != nil {
		return "", fmt.Errorf("cannot open file: %v", err)
	}
	defer fileHandle.Close()

	line := ""
	var cursor int64 = 0
	stat, _ := fileHandle.Stat()
	filesize := stat.Size()

	if filesize == 0 {
		return "", nil
	}

	for {
		cursor -= 1
		if _, err := fileHandle.Seek(cursor, io.SeekEnd); err != nil {
			return "", err
		}

		char := make([]byte, 1)
		if _, err := fileHandle.Read(char); err != nil {
			return "", err
		}

		// stop if we find a line
		if cursor != -1 && (char[0] == 10 || char[0] == 13) {
			break
		}

		line = fmt.Sprintf("%s%s", string(char), line)

		// stop if we are at the beginning
		if cursor == -filesize {
			break
		}
	}

	return line, nil
}
