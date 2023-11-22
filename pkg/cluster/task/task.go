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
	stderrors "errors"
	"fmt"
	"runtime"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

var (
	// ErrUnsupportedRollback means the task do not support rollback.
	ErrUnsupportedRollback = stderrors.New("unsupported rollback")
	// ErrNoExecutor means not being able to get the executor.
	ErrNoExecutor = stderrors.New("no executor")
	// ErrNoOutput means not being able to get the output of host.
	ErrNoOutput = stderrors.New("no outputs available")
)

type (
	// Task represents a operation while gemix execution
	Task interface {
		fmt.Stringer
		Execute(ctx context.Context) error
		Rollback(ctx context.Context) error
	}

	// Serial will execute a bundle of task in serialized way
	Serial struct {
		ignoreError       bool
		hideDetailDisplay bool
		inner             []Task
	}

	// Parallel will execute a bundle of task in parallelism way
	Parallel struct {
		ignoreError       bool
		hideDetailDisplay bool
		inner             []Task
	}
)

func isDisplayTask(t Task) bool {
	if _, ok := t.(*Serial); ok {
		return true
	}
	if _, ok := t.(*Parallel); ok {
		return true
	}
	if _, ok := t.(*StepDisplay); ok {
		return true
	}
	if _, ok := t.(*ParallelStepDisplay); ok {
		return true
	}
	return false
}

// Execute implements the Task interface
func (s *Serial) Execute(ctx context.Context) error {
	for _, t := range s.inner {
		if !isDisplayTask(t) {
			if !s.hideDetailDisplay {
				fmt.Printf("+ [ Serial ] - %s\n", t.String())
			}
		}
		err := t.Execute(ctx)
		if err != nil && !s.ignoreError {
			return errors.WithStack(err)
		}
	}
	return nil
}

// Rollback implements the Task interface
func (s *Serial) Rollback(ctx context.Context) error {
	// Rollback in reverse order
	for i := len(s.inner) - 1; i >= 0; i-- {
		err := s.inner[i].Rollback(ctx)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

// String implements the fmt.Stringer interface
func (s *Serial) String() string {
	var ss []string
	for _, t := range s.inner {
		ss = append(ss, t.String())
	}
	return strings.Join(ss, "\n")
}

// Execute implements the Task interface
func (pt *Parallel) Execute(ctx context.Context) error {
	var firstError error
	var mu sync.Mutex
	wg := sync.WaitGroup{}

	maxWorkers := runtime.NumCPU()
	workerPool := make(chan struct{}, maxWorkers)

	for _, t := range pt.inner {
		wg.Add(1)
		workerPool <- struct{}{}

		// the checkpoint part of context can't be shared between goroutines
		// since it's used to trace the stack, so we must create a new layer
		// of checkpoint context every time put it into a new goroutine.
		go func(ctx context.Context, t Task) {
			defer func() {
				<-workerPool
				wg.Done()
			}()
			if !isDisplayTask(t) {
				if !pt.hideDetailDisplay {
					fmt.Printf("+ [Parallel] - %s\n", t.String())
				}
			}
			err := t.Execute(ctx)
			if err != nil {
				mu.Lock()
				if firstError == nil {
					firstError = err
				}
				mu.Unlock()
			}
		}(ctx, t)
	}
	wg.Wait()
	if pt.ignoreError {
		return nil
	}
	return errors.WithStack(firstError)
}

// Rollback implements the Task interface
func (pt *Parallel) Rollback(ctx context.Context) error {
	var firstError error
	var mu sync.Mutex
	wg := sync.WaitGroup{}
	for _, t := range pt.inner {
		wg.Add(1)

		// the checkpoint part of context can't be shared between goroutines
		// since it's used to trace the stack, so we must create a new layer
		// of checkpoint context every time put it into a new goroutine.
		go func(ctx context.Context, t Task) {
			defer wg.Done()
			err := t.Rollback(ctx)
			if err != nil {
				mu.Lock()
				if firstError == nil {
					firstError = err
				}
				mu.Unlock()
			}
		}(ctx, t)
	}
	wg.Wait()
	return errors.WithStack(firstError)
}

// String implements the fmt.Stringer interface
func (pt *Parallel) String() string {
	var ss []string
	for _, t := range pt.inner {
		ss = append(ss, t.String())
	}
	return strings.Join(ss, "\n")
}
