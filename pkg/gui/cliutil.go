// Copyright 2020 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package gui

import (
	"fmt"
	"strings"

	"github.com/joomcode/errorx"
	"github.com/openGemini/gemix/pkg/utils"
	"github.com/spf13/cobra"
)

var (
	errNS             = errorx.NewNamespace("gui")
	errMismatchArgs   = errNS.NewType("mismatch_args", utils.ErrTraitPreCheck)
	errOperationAbort = errNS.NewType("operation_aborted", utils.ErrTraitPreCheck)
)

// CheckCommandArgsAndMayPrintHelp checks whether user passes enough number of arguments.
// If insufficient number of arguments are passed, an error with proper suggestion will be raised.
// When no argument is passed, command help will be printed and no error will be raised.
func CheckCommandArgsAndMayPrintHelp(cmd *cobra.Command, args []string, minArgs int) (shouldContinue bool, err error) {
	if minArgs == 0 {
		return true, nil
	}
	lenArgs := len(args)
	if lenArgs == 0 {
		return false, cmd.Help()
	}
	if lenArgs < minArgs {
		return false, errMismatchArgs.
			New("Expect at least %d arguments, but received %d arguments", minArgs, lenArgs).
			WithProperty(SuggestionFromString(cmd.UsageString()))
	}
	return true, nil
}

// SuggestionFromString creates a suggestion from string.
// Usage: SomeErrorX.WithProperty(SuggestionFromString(..))
func SuggestionFromString(str string) (errorx.Property, string) {
	return utils.ErrPropSuggestion, strings.TrimSpace(str)
}

// SuggestionFromFormat creates a suggestion from a format.
// Usage: SomeErrorX.WithProperty(SuggestionFromFormat(..))
func SuggestionFromFormat(format string, a ...any) (errorx.Property, string) {
	s := fmt.Sprintf(format, a...)
	return SuggestionFromString(s)
}

// BeautifyCobraUsageAndHelp beautifies cobra usages and help.
func BeautifyCobraUsageAndHelp(rootCmd *cobra.Command) {
	s := `Usage:{{if .Runnable}}
  {{ColorCommand}}{{gemixCmdLine .UseLine}}{{ColorReset}}{{end}}{{if .HasAvailableSubCommands}}
  {{ColorCommand}}{{gemixCmdPath .Use}} [command]{{ColorReset}}{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{ColorCommand}}{{.NameAndAliases}}{{ColorReset}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{ColorCommand}}{{gemixCmdPath .Use}} help [command]{{ColorReset}}" for more information about a command.{{end}}
`
	rootCmd.SetUsageTemplate(s)
}
