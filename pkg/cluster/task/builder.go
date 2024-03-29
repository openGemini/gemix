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

	"github.com/openGemini/gemix/pkg/cluster/spec"
	logprinter "github.com/openGemini/gemix/pkg/logger/printer"
	"github.com/openGemini/gemix/pkg/meta"
)

// Builder is used to build task
type Builder struct {
	tasks  []Task
	Logger *logprinter.Logger
}

// NewBuilder returns a *Builder instance
func NewBuilder(logger *logprinter.Logger) *Builder {
	return &Builder{Logger: logger}
}

// RootSSH appends a RootSSH task to the current task collection
func (b *Builder) RootSSH(
	host string, port int, user, password, keyFile, passphrase string, sshTimeout, exeTimeout uint64) *Builder {
	b.tasks = append(b.tasks, &RootSSH{
		host:       host,
		port:       port,
		user:       user,
		password:   password,
		keyFile:    keyFile,
		passphrase: passphrase,
		timeout:    sshTimeout,
		exeTimeout: exeTimeout,
	})
	return b
}

// NewSimpleUerSSH  append a UserSSH task to the current task collection with operator.Options and SSHConnectionProps
func NewSimpleUerSSH(logger *logprinter.Logger, host string, port int, user string, sshTimeout, exeTimeout uint64) *Builder {
	return NewBuilder(logger).
		UserSSH(
			host,
			port,
			user,
			sshTimeout,
			exeTimeout,
		)
}

// UserSSH append a UserSSH task to the current task collection
func (b *Builder) UserSSH(host string, port int, deployUser string, sshTimeout, exeTimeout uint64) *Builder {
	b.tasks = append(b.tasks, &UserSSH{
		host:       host,
		port:       port,
		deployUser: deployUser,
		timeout:    sshTimeout,
		exeTimeout: exeTimeout,
	})
	return b
}

// Func append a func task.
func (b *Builder) Func(name string, fn func(ctx context.Context) error) *Builder {
	b.tasks = append(b.tasks, &Func{
		name: name,
		fn:   fn,
	})
	return b
}

// ClusterSSH init all UserSSH need for the cluster.
func (b *Builder) ClusterSSH(
	topo spec.Topology,
	deployUser string, sshTimeout, exeTimeout uint64) *Builder {
	var tasks []Task
	topo.IterInstance(func(inst spec.Instance) {
		tasks = append(tasks, &UserSSH{
			host:       inst.GetManageHost(),
			port:       inst.GetSSHPort(),
			deployUser: deployUser,
			timeout:    sshTimeout,
			exeTimeout: exeTimeout,
		})
	})

	b.tasks = append(b.tasks, &Parallel{inner: tasks})
	return b
}

// Download appends a Downloader task to the current task collection
func (b *Builder) Download(component, os, arch string, version string) *Builder {
	b.tasks = append(b.tasks, NewDownloader(component, os, arch, version))
	return b
}

// InitConfig appends a CopyComponent task to the current task collection
func (b *Builder) InitConfig(clusterName, clusterVersion string, specManager *spec.SpecManager, inst spec.Instance, deployUser string, ignoreCheck bool, paths meta.DirPaths) *Builder {

	b.tasks = append(b.tasks, &InitConfig{
		specManager:    specManager,
		clusterName:    clusterName,
		clusterVersion: clusterVersion,
		instance:       inst,
		deployUser:     deployUser,
		ignoreCheck:    ignoreCheck,
		paths:          paths,
	})
	return b
}

// EnvInit appends a EnvInit task to the current task collection
func (b *Builder) EnvInit(host, deployUser string, userGroup string) *Builder {
	b.tasks = append(b.tasks, &EnvInit{
		host:       host,
		deployUser: deployUser,
		userGroup:  userGroup,
	})
	return b
}

// UserAction appends a UserAction task to the current task collection
func (b *Builder) UserAction(host, deployUser string, userGroup string, skipCreateUser bool) *Builder {
	b.tasks = append(b.tasks, &UserAction{
		skipCreateUser: skipCreateUser,
		host:           host,
		userAction:     UserActionAdd,
		name:           deployUser,
		group:          userGroup,
		sudoer:         true,
	})
	return b
}

// CopyComponent appends a CopyComponent task to the current task collection
func (b *Builder) CopyComponent(pkgSrc, component, os, arch string,
	version string,
	srcPath, dstHost, dstDir string,
) *Builder {
	b.tasks = append(b.tasks, &CopyComponent{
		srcPkgName: pkgSrc,
		component:  component,
		os:         os,
		arch:       arch,
		version:    version,
		srcPath:    srcPath,
		host:       dstHost,
		dstDir:     dstDir,
	})
	return b
}

// MonitoredConfig appends a CopyComponent task to the current task collection
func (b *Builder) MonitoredConfig(clusterName, comp, host string, info *spec.MonitorHostInfo, globResCtl meta.ResourceControl, options *spec.TSMonitoredOptions, deployUser string, tlsEnabled bool, paths meta.DirPaths) *Builder {
	b.tasks = append(b.tasks, &MonitoredConfig{
		clusterName: clusterName,
		component:   comp,
		host:        host,
		info:        info,
		globResCtl:  globResCtl,
		options:     options,
		deployUser:  deployUser,
		tlsEnabled:  tlsEnabled,
		paths:       paths,
	})
	return b
}

// SSHKeyGen appends a SSHKeyGen task to the current task collection
func (b *Builder) SSHKeyGen(keypath string) *Builder {
	b.tasks = append(b.tasks, &SSHKeyGen{
		keypath: keypath,
	})
	return b
}

// SSHKeySet appends a SSHKeySet task to the current task collection
func (b *Builder) SSHKeySet(privKeyPath, pubKeyPath string) *Builder {
	b.tasks = append(b.tasks, &SSHKeySet{
		privateKeyPath: privKeyPath,
		publicKeyPath:  pubKeyPath,
	})
	return b
}

// Mkdir appends a Mkdir task to the current task collection
func (b *Builder) Mkdir(user, host string, dirs ...string) *Builder {
	b.tasks = append(b.tasks, &Mkdir{
		user: user,
		host: host,
		dirs: dirs,
	})
	return b
}

//// Shell command on cluster host
//func (b *Builder) Shell(host, command, cmdID string, sudo bool) *Builder {
//	b.tasks = append(b.tasks, &Shell{
//		host:    host,
//		command: command,
//		sudo:    sudo,
//		cmdID:   cmdID,
//	})
//	return b
//}

// Parallel appends a parallel task to the current task collection
func (b *Builder) Parallel(ignoreError bool, tasks ...Task) *Builder {
	if len(tasks) > 0 {
		b.tasks = append(b.tasks, &Parallel{ignoreError: ignoreError, inner: tasks})
	}
	return b
}

// Serial appends the tasks to the tail of queue
func (b *Builder) Serial(tasks ...Task) *Builder {
	if len(tasks) > 0 {
		b.tasks = append(b.tasks, tasks...)
	}
	return b
}

// Build returns a task that contains all tasks appended by previous operation
func (b *Builder) Build() Task {
	// Serial handles event internally. So the following 3 lines are commented out.
	// if len(b.tasks) == 1 {
	//  return b.tasks[0]
	// }
	return &Serial{inner: b.tasks}
}

// Step appends a new StepDisplay task, which will print single line progress for inner tasks.
func (b *Builder) Step(prefix string, inner Task, logger *logprinter.Logger) *Builder {
	b.Serial(newStepDisplay(prefix, inner, logger))
	return b
}

// ParallelStep appends a new ParallelStepDisplay task, which will print multi line progress in parallel
// for inner tasks. Inner tasks must be a StepDisplay task.
func (b *Builder) ParallelStep(prefix string, ignoreError bool, tasks ...*StepDisplay) *Builder {
	b.tasks = append(b.tasks, newParallelStepDisplay(prefix, ignoreError, tasks...).SetLogger(b.Logger))
	return b
}

// BuildAsStep returns a task that is wrapped by a StepDisplay. The task will print single line progress.
func (b *Builder) BuildAsStep(prefix string) *StepDisplay {
	inner := b.Build()
	return newStepDisplay(prefix, inner, b.Logger)
}
