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

// ref: https://github.com/charmbracelet/bubbletea/blob/master/examples/spinners/main.go

package progress

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type FinishedMsg struct{ Finished bool }

func NewSpinnerProgram(prefix string) *tea.Program {
	m := spinnerModel{
		prefix: prefix,
	}
	m.resetSpinner()
	return tea.NewProgram(&m)
}

type spinnerModel struct {
	spinner spinner.Model

	finished bool
	prefix   string
	err      error
}

func (m *spinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m *spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		default:
			return m, nil
		}
	case ErrMsg:
		m.err = msg.Err
		return m, tea.Quit
	case FinishedMsg:
		m.finished = msg.Finished
		return m, tea.Quit
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		if m.err != nil {
			return m, tea.Quit
		}
		return m, cmd
	default:
		return m, nil
	}
}

func (m *spinnerModel) resetSpinner() {
	m.spinner = spinner.New()
	m.spinner.Style = spinnerStyle
	m.spinner.Spinner = spinner.Dot
}

func (m *spinnerModel) View() (s string) {
	if m.err != nil {
		s += fmt.Sprintf("%s %s %s\n", m.prefix, "...", errorStyle(m.err.Error()))
	} else if m.finished {
		s += fmt.Sprintf("%s %s %s\n", m.prefix, "...", greenStyle("Done"))
	} else {
		s += fmt.Sprintf("%s %s %s\n", m.prefix, m.spinner.View(), textStyle("Doing..."))
	}
	return
}
