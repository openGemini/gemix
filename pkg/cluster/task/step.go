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

package task

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/openGemini/gemix/pkg/cluster/ctxt"
	"github.com/openGemini/gemix/pkg/gui/progress"
	"go.uber.org/zap"
)

// StepDisplay is a task that will display a progress bar for inner task.
type StepDisplay struct {
	//hidden   bool
	inner    Task
	prefix   string
	children map[Task]struct{}
	Logger   *zap.Logger

	teaProgram *tea.Program
}

func addChildren(m map[Task]struct{}, task Task) {
	if _, exists := m[task]; exists {
		return
	}
	m[task] = struct{}{}
	if t, ok := task.(*Serial); ok {
		t.hideDetailDisplay = true
		for _, tx := range t.inner {
			if _, exists := m[tx]; !exists {
				addChildren(m, tx)
			}
		}
	} else if t, ok := task.(*Parallel); ok {
		t.hideDetailDisplay = true
		for _, tx := range t.inner {
			if _, exists := m[tx]; !exists {
				addChildren(m, tx)
			}
		}
	}
}

func newStepDisplay(prefix string, inner Task, logger *zap.Logger) *StepDisplay {
	children := make(map[Task]struct{})
	addChildren(children, inner)
	return &StepDisplay{
		inner:    inner,
		prefix:   prefix,
		children: children,
		Logger:   logger,

		teaProgram: progress.NewSpinnerProgram(prefix),
	}
}

// SetLogger set the logger of step
func (s *StepDisplay) SetLogger(logger *zap.Logger) *StepDisplay {
	s.Logger = logger
	return s
}

// Execute implements the Task interface
func (s *StepDisplay) Execute(ctx context.Context) error {
	go s.teaProgram.Run() // nolint:errcheck

	ctxt.GetInner(ctx).Ev.Subscribe(ctxt.EventTaskBegin, s.handleTaskBegin)
	ctxt.GetInner(ctx).Ev.Subscribe(ctxt.EventTaskProgress, s.handleTaskProgress)
	err := s.inner.Execute(ctx)
	ctxt.GetInner(ctx).Ev.Unsubscribe(ctxt.EventTaskProgress, s.handleTaskProgress)
	ctxt.GetInner(ctx).Ev.Unsubscribe(ctxt.EventTaskBegin, s.handleTaskBegin)

	s.teaProgram.Send(progress.FinishedMsg{Finished: true})
	return err
}

// Rollback implements the Task interface
func (s *StepDisplay) Rollback(ctx context.Context) error {
	return s.inner.Rollback(ctx)
}

// String implements the fmt.Stringer interface
func (s *StepDisplay) String() string {
	return s.inner.String()
}

func (s *StepDisplay) handleTaskBegin(task Task) {
	if _, ok := s.children[task]; !ok {
		return
	}
	s.teaProgram.Send(nil)
	fmt.Println(task.String())
}

func (s *StepDisplay) handleTaskProgress(task Task, p string) {
	if _, ok := s.children[task]; !ok {
		return
	}

	s.teaProgram.Send(nil)
	fmt.Println(task.String())
}

// ParallelStepDisplay is a task that will display multiple progress bars in parallel for inner tasks.
// Inner tasks will be executed in parallel.
type ParallelStepDisplay struct {
	inner  *Parallel
	prefix string
	Logger *zap.Logger
}

func newParallelStepDisplay(prefix string, ignoreError bool, sdTasks ...*StepDisplay) *ParallelStepDisplay {
	tasks := make([]Task, 0, len(sdTasks))
	for _, t := range sdTasks {
		tasks = append(tasks, t)
	}
	return &ParallelStepDisplay{
		inner:  &Parallel{inner: tasks, ignoreError: ignoreError},
		prefix: prefix,
	}
}

// SetLogger set the logger of step
func (ps *ParallelStepDisplay) SetLogger(logger *zap.Logger) *ParallelStepDisplay {
	ps.Logger = logger
	return ps
}

// Execute implements the Task interface
func (ps *ParallelStepDisplay) Execute(ctx context.Context) error {
	fmt.Println(ps.prefix)

	// Preserve space for the bar
	//fmt.Print(strings.Repeat("\n", len(ps.inner.inner)+1))

	err := ps.inner.Execute(ctx)
	return err
}

// Rollback implements the Task interface
func (ps *ParallelStepDisplay) Rollback(ctx context.Context) error {
	return ps.inner.Rollback(ctx)
}

// String implements the fmt.Stringer interface
func (ps *ParallelStepDisplay) String() string {
	return ps.inner.String()
}
