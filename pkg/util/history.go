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
