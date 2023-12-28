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

//lint:ignore U1000 keep this
type downloadSpinnerModel struct {
	spinnerModel

	pw *progressWriter

	percent float64 // 0 - 1.0
}

//lint:ignore U1000 keep this
func (m *downloadSpinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

//lint:ignore U1000 keep this
func (m *downloadSpinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		default:
		}
	case ErrMsg:
		m.err = msg.Err
		return m, tea.Quit
	case progressMsg:
		if msg >= 1.0 {
			m.percent = 1.0
			return m, tea.Quit
		}
		m.percent = float64(msg)
		return m, m.spinner.Tick
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
	return m, nil
}

//lint:ignore U1000 keep this
func (m *downloadSpinnerModel) resetSpinner() {
	m.spinner = spinner.New()
	m.spinner.Style = spinnerStyle
	m.spinner.Spinner = spinner.Dot
}

//lint:ignore U1000 keep this
func (m *downloadSpinnerModel) View() (s string) {
	if m.err != nil {
		s += fmt.Sprintf("\n%s %s %s\n", m.prefix, "...", errorStyle(m.err.Error()))
	} else if m.percent >= 1.0 {
		s += fmt.Sprintf("\n%s %s %s\n", m.prefix, "...", greenStyle("Done"))
	} else {
		s += fmt.Sprintf("\n%s %s %.0f%s %s\n", m.prefix, m.spinner.View(), m.percent*100, "%", textStyle("Downloading..."))
	}
	return
}
