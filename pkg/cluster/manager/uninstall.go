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

package manager

import (
	"context"
	"fmt"

	"github.com/fatih/color"
	"github.com/openGemini/gemix/pkg/cluster/ctxt"
	operator "github.com/openGemini/gemix/pkg/cluster/operation"
	"github.com/openGemini/gemix/pkg/gui"
	"github.com/pkg/errors"
)

// UninstallCluster destroy the cluster.
func (m *Manager) UninstallCluster(name string, gOpt operator.Options, destroyOpt operator.Options, skipConfirm bool) error {
	if err := ValidateClusterNameOrError(name); err != nil {
		return err
	}

	metadata, err := m.meta(name)
	if err != nil {
		return err
	}

	topo := metadata.GetTopology()
	base := metadata.GetBaseMeta()

	if !skipConfirm {
		fmt.Println(color.HiRedString(gui.ASCIIArtWarning))
		if err := gui.PromptForAnswerOrAbortError(
			"Yes, I know my cluster and data will be deleted.",
			fmt.Sprintf("This operation will destroy %s %s cluster %s and its data.",
				m.sysName,
				color.HiYellowString(base.Version),
				color.HiYellowString(name),
			)+"\nAre you sure to continue?",
		); err != nil {
			return err
		}
		m.logger.Infof("Destroying cluster...")
	}

	b, err := m.sshTaskBuilder(name, topo, base.User, gOpt)
	if err != nil {
		return err
	}
	t := b.
		Func("StopCluster", func(ctx context.Context) error {
			return operator.Stop(
				ctx,
				topo,
				operator.Options{Force: destroyOpt.Force},
			)
		}).
		Func("UninstallCluster", func(ctx context.Context) error {
			return operator.Uninstall(ctx, topo, destroyOpt)
		}).
		Build()

	ctx := ctxt.New(
		context.Background(),
		gOpt.Concurrency,
		m.logger,
	)
	if err = t.Execute(ctx); err != nil {
		return errors.WithStack(err)
	}

	if err = m.specManager.Remove(name); err != nil {
		return errors.WithStack(err)
	}

	m.logger.Infof("Uninstall cluster `%s` successfully", name)
	return nil
}
