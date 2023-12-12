// Copyright 2023 Huawei Cloud Computing Technologies Co., Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/AstroProfundis/tabby"
	"github.com/fatih/color"
	"golang.org/x/term"
)

// PrintTable accepts a matrix of strings and print them as ASCII table to terminal
func PrintTable(rows [][]string, header bool) {
	// Print the table
	t := tabby.New()
	if header {
		addRow(t, rows[0], header)
		rows = rows[1:]
	}
	for _, row := range rows {
		addRow(t, row, false)
	}
	t.Print()
}

func addRow(t *tabby.Tabby, rawLine []string, header bool) {
	// Convert []string to []any
	row := make([]any, len(rawLine))
	for i, v := range rawLine {
		row[i] = v
	}

	// Add line to the table
	if header {
		t.AddHeader(row...)
	} else {
		t.AddLine(row...)
	}
}

// pre-defined ascii art strings
const (
	ASCIIArtWarning = `
  ██     ██  █████  ██████  ███    ██ ██ ███    ██  ██████
  ██     ██ ██   ██ ██   ██ ████   ██ ██ ████   ██ ██
  ██  █  ██ ███████ ██████  ██ ██  ██ ██ ██ ██  ██ ██   ███
  ██ ███ ██ ██   ██ ██   ██ ██  ██ ██ ██ ██  ██ ██ ██    ██
   ███ ███  ██   ██ ██   ██ ██   ████ ██ ██   ████  ██████
`
)

// Prompt accepts input from console by user
func Prompt(prompt string) string {
	if prompt != "" {
		prompt += " " // append a whitespace
	}
	fmt.Print(prompt)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return ""
	}
	return strings.TrimSuffix(input, "\n")
}

// PromptForConfirmYes accepts yes / no from console by user, default to No and only return true
// if the user input is Yes
func PromptForConfirmYes(format string, a ...any) (bool, string) {
	ans := Prompt(fmt.Sprintf(format, a...) + "(default=N)")
	switch strings.TrimSpace(strings.ToLower(ans)) {
	case "y", "yes":
		return true, ans
	default:
		return false, ans
	}
}

// PromptForConfirmNo accepts yes / no from console by user, default to Yes and only return true
// if the user input is No
func PromptForConfirmNo(format string, a ...any) (bool, string) {
	ans := Prompt(fmt.Sprintf(format, a...) + "(default=Y)")
	switch strings.TrimSpace(strings.ToLower(ans)) {
	case "n", "no":
		return true, ans
	default:
		return false, ans
	}
}

// PromptForConfirmOrAbortError accepts yes / no from console by user, generates AbortError if user does not input yes.
func PromptForConfirmOrAbortError(format string, a ...any) error {
	if pass, ans := PromptForConfirmYes(format, a...); !pass {
		return fmt.Errorf("operation aborted by user (with answer '%s')", ans)
	}
	return nil
}

// PromptForConfirmAnswer accepts string from console by user, default to empty and only return
// true if the user input is exactly the same as pre-defined answer.
func PromptForConfirmAnswer(answer string, format string, a ...any) (bool, string) {
	ans := Prompt(fmt.Sprintf(format, a...) + fmt.Sprintf("\n(Type \"%s\" to continue)\n:", color.CyanString(answer)))
	if ans == answer {
		return true, ans
	}
	return false, ans
}

// PromptForAnswerOrAbortError accepts string from console by user, generates AbortError if user does
// not input the pre-defined answer.
func PromptForAnswerOrAbortError(answer string, format string, a ...any) error {
	if pass, ans := PromptForConfirmAnswer(answer, format, a...); !pass {
		return errOperationAbort.New("Operation aborted by user (with incorrect answer '%s')", ans)
	}
	return nil
}

// PromptForPassword reads a password input from console
func PromptForPassword(format string, a ...any) string {
	defer fmt.Println("")

	fmt.Printf(format, a...)

	input, err := term.ReadPassword(syscall.Stdin)

	if err != nil {
		return ""
	}
	return strings.TrimSpace(strings.Trim(string(input), "\n"))
}
