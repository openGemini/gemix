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
	"os"
	"regexp"
	"runtime"

	"github.com/fatih/color"
	"github.com/joomcode/errorx"
	"github.com/openGemini/gemix/pkg/cluster/ctxt"
	operator "github.com/openGemini/gemix/pkg/cluster/operation"
	"github.com/openGemini/gemix/pkg/cluster/spec"
	"github.com/openGemini/gemix/pkg/cluster/task"
	"github.com/openGemini/gemix/pkg/gui"
	"github.com/pkg/errors"
)

var (
	clusterNameRegexp = regexp.MustCompile(`^[a-zA-Z0-9\-_\.]+$`)
)

// InstallOptions contains the options for install.
type InstallOptions struct {
	User           string // username of login to the SSH server
	SkipCreateUser bool   // don't create the user
	IdentityFile   string // path to the private key file
	UsePassword    bool   // use password instead of identity file for ssh connection
}

// TODO
// DeployerInstance is an instance can deploy to a target deploy directory.
type DeployerInstance interface {
	Deploy(b *task.Builder, srcPath string, deployDir string, version string, name string, clusterVersion string)
}

// ValidateClusterNameOrError validates a cluster name and returns error if the name is invalid.
func ValidateClusterNameOrError(n string) error {
	if len(n) == 0 {
		return fmt.Errorf("cluster name must not be empty")
	}
	if !clusterNameRegexp.MatchString(n) {
		return fmt.Errorf("cluster name '%s' is invalid. The cluster name should only contain alphabets, numbers, hyphen (-), underscore (_), and dot (.)", n)
	}
	return nil
}

// Install a cluster.
func (m *Manager) Install(
	clusterName string,
	clusterVersion string,
	topoFile string,
	opt InstallOptions,
	skipConfirm bool,
	gOpt operator.Options,
) error {
	if err := ValidateClusterNameOrError(clusterName); err != nil {
		return errors.WithStack(err)
	}

	exist, err := m.specManager.Exist(clusterName)
	if err != nil {
		return errors.WithStack(err)
	}

	if exist {
		return errors.Errorf("cluster name '%s' is duplicated. Please specify another cluster name", clusterName)
	}

	metadata := m.specManager.NewMetadata()
	topo := metadata.GetTopology()
	base := topo.BaseTopo()

	if err = spec.ParseTopologyYaml(topoFile, topo); err != nil {
		return errors.WithStack(err)
	}

	spec.ExpandRelativeDir(topo)

	if err = checkConflict(m, clusterName, topo); err != nil {
		return errors.WithStack(err)
	}

	var (
		sshConnProps *gui.SSHConnectionProps
	)
	if sshConnProps, err = gui.ReadIdentityFileOrPassword(opt.IdentityFile, opt.UsePassword); err != nil {
		return errors.WithStack(err)
	}

	// TODO: Detect CPU Arch Name
	if err = m.fillHost(sshConnProps, topo, opt.User); err != nil {
		return errors.WithStack(err)
	}

	if !skipConfirm {
		if err = m.confirmTopology(clusterName, clusterVersion, topo); err != nil {
			return errors.WithStack(err)
		}
	}

	if err = os.MkdirAll(m.specManager.Path(clusterName), 0750); err != nil {
		return errorx.InitializationFailed.
			Wrap(err, "Failed to create cluster metadata directory '%s'", m.specManager.Path(clusterName)).
			WithProperty(gui.SuggestionFromString("Please check file system permissions and try again."))
	}

	// Initialize environment

	globalOptions := base.GlobalOptions
	metadata.SetUser(globalOptions.User)
	metadata.SetVersion(clusterVersion)

	// generate CA and client cert for TLS enabled cluster
	//_, err = m.genAndSaveCertificate(clusterName, globalOptions)
	//if err != nil {
	//	return err
	//}

	// tasks which are used to download components, download missing component
	downloadCompTasks := buildDownloadCompTasks(clusterVersion, topo, m.logger)

	// tasks which are used to initialize environment
	envInitTasks := buildEnvInitTasks(topo, &opt, &gOpt, sshConnProps, m.logger)

	// tasks which are used to mkdir at remote target host
	mkdirTasks := buildMkdirTasks(topo, &gOpt, sshConnProps, m.logger)

	// tasks which are used to copy components to remote host
	deployCompTasks := buildDeployTasks(clusterName, clusterVersion, topo, &gOpt, sshConnProps, m.logger)

	// generates certificate for instance and transfers it to the server
	//certificateTasks, err := buildCertificateTasks(m, name, topo, metadata.GetBaseMeta(), gOpt, sshProxyProps)
	//if err != nil {
	//	return err
	//}

	refreshConfigTasks := buildInitConfigTasks(m, clusterName, topo, metadata.GetBaseMeta(), gOpt)

	builder := task.NewBuilder(m.logger).
		Step("+ Generate SSH keys",
			task.NewBuilder(m.logger).
				SSHKeyGen(m.specManager.Path(clusterName, "ssh", "id_rsa")).
				Build(),
			m.logger).
		ParallelStep("+ Download openGemini components", false, downloadCompTasks...).
		ParallelStep("+ Initialize target host environments", false, envInitTasks...).
		ParallelStep("+ Mkdir at target hosts", false, mkdirTasks...).
		ParallelStep("+ Deploy openGemini instance", false, deployCompTasks...).
		//ParallelStep("+ Copy certificate to remote host", gOpt.Force, certificateTasks...).
		ParallelStep("+ Init instance configs", gOpt.Force, refreshConfigTasks...)
	//ParallelStep("+ Init monitor configs", gOpt.Force, monitorConfigTasks...)

	t := builder.Build()

	ctx := ctxt.New(
		context.Background(),
		runtime.NumCPU(),
		m.logger,
	)
	if err = t.Execute(ctx); err != nil {
		if errorx.Cast(err) != nil {
			// FIXME: Map possible task errors and give suggestions.
			return errors.WithStack(err)
		}
		return errors.WithStack(err)
	}

	// FIXME: remove me if you finish
	err = m.specManager.SaveMeta(clusterName, metadata)
	if err != nil {
		return err
	}

	hint := color.New(color.FgBlue).Sprintf("%s start %s", "gemix cluster", clusterName)
	m.logger.Infof("Cluster `%s` installed successfully, you can start it with command: `%s`\n", clusterName, hint)
	return nil
}
