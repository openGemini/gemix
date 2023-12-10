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

package manager

import (
	"context"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/joomcode/errorx"
	"github.com/openGemini/gemix/pkg/cluster/ctxt"
	operator "github.com/openGemini/gemix/pkg/cluster/operation"
	"github.com/openGemini/gemix/pkg/gui"
	"github.com/pkg/errors"
)

// StopCluster stop the cluster.
func (m *Manager) StopCluster(
	name string,
	gOpt operator.Options,
	skipConfirm bool,
) error {
	//// check locked
	//if err := m.specManager.ScaleOutLockedErr(name); err != nil {
	//	return err
	//}

	metadata, err := m.meta(name)
	if err != nil {
		return err
	}

	topo := metadata.GetTopology()
	base := metadata.GetBaseMeta()
	//
	//tlsCfg, err := topo.TLSConfig(m.specManager.Path(name, spec.TLSCertKeyDir))
	//if err != nil {
	//	return err
	//}

	if !skipConfirm {
		if err = gui.PromptForConfirmOrAbortError(
			fmt.Sprintf("Will stop the cluster %s with nodes: %s, roles: %s.\nDo you want to continue? [y/N]:",
				color.HiYellowString(name),
				color.HiRedString(strings.Join(gOpt.Nodes, ",")),
				color.HiRedString(strings.Join(gOpt.Roles, ",")),
			),
		); err != nil {
			return err
		}
	}

	b, err := m.sshTaskBuilder(name, topo, base.User, gOpt)
	if err != nil {
		return err
	}

	t := b.
		Func("StopCluster", func(ctx context.Context) error {
			return operator.Stop(ctx, topo, gOpt)
		}).
		Build()

	ctx := ctxt.New(
		context.Background(),
		gOpt.Concurrency,
		m.logger,
	)
	if err := t.Execute(ctx); err != nil {
		if errorx.Cast(err) != nil {
			// FIXME: Map possible task errors and give suggestions.
			return err
		}
		return errors.WithStack(err)
	}

	m.logger.Infof(fmt.Sprintf("Stopped cluster `%s` successfully", name))
	return nil
}
