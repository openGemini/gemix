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

package operation

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/openGemini/gemix/pkg/cluster/ctxt"
	"github.com/openGemini/gemix/pkg/cluster/executor"
	"github.com/openGemini/gemix/pkg/cluster/module"
	"github.com/openGemini/gemix/pkg/cluster/spec"
	logprinter "github.com/openGemini/gemix/pkg/logger/printer"
	"github.com/openGemini/gemix/pkg/set"
	"github.com/pkg/errors"
)

// Destroy the cluster.
func Destroy(
	ctx context.Context,
	cluster spec.Topology,
	options Options,
) error {
	coms := cluster.ComponentsByStopOrder()

	instCount := map[string]int{}
	cluster.IterInstance(func(inst spec.Instance) {
		instCount[inst.GetManageHost()]++
	})

	for _, com := range coms {
		insts := com.Instances()
		err := DestroyComponent(ctx, insts, cluster, options)
		if err != nil && !options.Force {
			return errors.WithMessagef(err, "failed to destroy %s", com.Name())
		}
	}

	gOpts := cluster.BaseTopo().GlobalOptions

	// Delete all global deploy directory
	for host := range instCount {
		if err := DeleteGlobalDirs(ctx, host, gOpts); err != nil {
			return nil
		}
	}

	// after all things done, try to remove SSH public key
	for host := range instCount {
		if err := DeletePublicKey(ctx, host); err != nil {
			return nil
		}
	}

	return nil
}

// DeleteGlobalDirs deletes all global directories if they are empty
func DeleteGlobalDirs(ctx context.Context, host string, options *spec.GlobalOptions) error {
	if options == nil {
		return nil
	}

	e := ctxt.GetInner(ctx).Get(host)
	logger := ctx.Value(logprinter.ContextKeyLogger).(*logprinter.Logger)
	logger.Infof("Clean global directories %s", host)
	for _, dir := range []string{options.LogDir, options.DeployDir, options.DataDir} {
		if dir == "" {
			continue
		}
		dir = spec.Abs(options.User, dir)

		logger.Infof("\tClean directory %s on instance %s", dir, host)

		c := module.ShellModuleConfig{
			Command:  fmt.Sprintf("rmdir %s > /dev/null 2>&1 || true", dir),
			Chdir:    "",
			UseShell: false,
		}
		shell := module.NewShellModule(c)
		stdout, stderr, err := shell.Execute(ctx, e)

		if len(stdout) > 0 {
			fmt.Println(string(stdout))
		}
		if len(stderr) > 0 {
			logger.Errorf(string(stderr))
		}

		if err != nil {
			return errors.WithMessagef(err, "failed to clean directory %s on: %s", dir, host)
		}
	}

	logger.Infof("Clean global directories %s success", host)
	return nil
}

// DeletePublicKey deletes the SSH public key from host
func DeletePublicKey(ctx context.Context, host string) error {
	e := ctxt.GetInner(ctx).Get(host)
	logger := ctx.Value(logprinter.ContextKeyLogger).(*logprinter.Logger)
	logger.Infof("Delete public key %s", host)
	_, pubKeyPath := ctxt.GetInner(ctx).GetSSHKeySet()
	publicKey, err := os.ReadFile(pubKeyPath)
	if err != nil {
		return errors.WithStack(err)
	}

	pubKey := string(bytes.TrimSpace(publicKey))
	pubKey = strings.ReplaceAll(pubKey, "/", "\\/")
	pubKeysFile := executor.FindSSHAuthorizedKeysFile(ctx, e)

	// delete the public key with Linux `sed` toolkit
	c := module.ShellModuleConfig{
		Command:  fmt.Sprintf("sed -i '/%s/d' %s", pubKey, pubKeysFile),
		UseShell: false,
	}
	shell := module.NewShellModule(c)
	stdout, stderr, err := shell.Execute(ctx, e)

	if len(stdout) > 0 {
		fmt.Println(string(stdout))
	}
	if len(stderr) > 0 {
		logger.Errorf(string(stderr))
	}

	if err != nil {
		return errors.WithMessagef(err, "failed to delete pulblic key on: %s", host)
	}

	logger.Infof("Delete public key %s success", host)
	return nil
}

// CleanupComponent cleanup the instances
func CleanupComponent(ctx context.Context, delFileMaps map[string]set.StringSet) error {
	logger := ctx.Value(logprinter.ContextKeyLogger).(*logprinter.Logger)
	for host, delFiles := range delFileMaps {
		e := ctxt.GetInner(ctx).Get(host)
		logger.Infof("Cleanup instance %s", host)
		logger.Debugf("Deleting paths on %s: %s", host, strings.Join(delFiles.Slice(), " "))
		c := module.ShellModuleConfig{
			Command:  fmt.Sprintf("rm -rf %s;", strings.Join(delFiles.Slice(), " ")),
			Sudo:     true, // the .service files are in a directory owned by root
			Chdir:    "",
			UseShell: true,
		}
		shell := module.NewShellModule(c)
		stdout, stderr, err := shell.Execute(ctx, e)

		if len(stdout) > 0 {
			fmt.Println(string(stdout))
		}
		if len(stderr) > 0 {
			logger.Errorf(string(stderr))
		}

		if err != nil {
			return errors.WithMessagef(err, "failed to cleanup: %s", host)
		}

		logger.Infof("Cleanup %s success", host)
	}

	return nil
}

// DestroyComponent destroy the instances.
func DestroyComponent(ctx context.Context, instances []spec.Instance, cls spec.Topology, options Options) error {
	if len(instances) == 0 {
		return nil
	}

	logger := ctx.Value(logprinter.ContextKeyLogger).(*logprinter.Logger)
	name := instances[0].ComponentName()
	fmt.Printf("Destroying component %s\n", name)
	logger.Infof("Destroying component %s\n", name)

	retainDataRoles := set.NewStringSet(options.RetainDataRoles...)
	retainDataNodes := set.NewStringSet(options.RetainDataNodes...)

	for _, ins := range instances {
		// Some data of instances will be retained
		dataRetained := retainDataRoles.Exist(ins.ComponentName()) ||
			retainDataNodes.Exist(ins.ID()) || retainDataNodes.Exist(ins.GetHost()) || retainDataRoles.Exist(ins.GetManageHost())

		e := ctxt.GetInner(ctx).Get(ins.GetManageHost())
		logger.Infof("\tDestroying instance %s\n", ins.GetManageHost())

		var dataDirs []string
		if len(ins.DataDir()) > 0 {
			dataDirs = strings.Split(ins.DataDir(), ",")
		}

		deployDir := ins.DeployDir()
		delPaths := set.NewStringSet()

		// Retain the deploy directory if the users want to retain the data directory
		// and the data directory is a sub-directory of deploy directory
		keepDeployDir := false

		for _, dataDir := range dataDirs {
			// Don't delete the parent directory if any sub-directory retained
			keepDeployDir = (dataRetained && strings.HasPrefix(dataDir, deployDir)) || keepDeployDir
			if !dataRetained && cls.CountDir(ins.GetManageHost(), dataDir) == 1 {
				// only delete path if it is not used by any other instance in the cluster
				delPaths.Insert(dataDir)
			}
		}

		logDir := ins.LogDir()

		if keepDeployDir {
			delPaths.Insert(filepath.Join(deployDir, "conf"))
			delPaths.Insert(filepath.Join(deployDir, "bin"))
			delPaths.Insert(filepath.Join(deployDir, "scripts"))
			//if cls.BaseTopo().GlobalOptions.TLSEnabled {
			//	delPaths.Insert(filepath.Join(deployDir, spec.TLSCertKeyDir))
			//}
			// only delete path if it is not used by any other instance in the cluster
			if strings.HasPrefix(logDir, deployDir) && cls.CountDir(ins.GetManageHost(), logDir) == 1 {
				delPaths.Insert(logDir)
			}
		} else {
			// only delete path if it is not used by any other instance in the cluster
			if cls.CountDir(ins.GetManageHost(), logDir) == 1 {
				delPaths.Insert(logDir)
			}
			if cls.CountDir(ins.GetManageHost(), ins.DeployDir()) == 1 {
				delPaths.Insert(ins.DeployDir())
			}
		}

		// check for deploy dir again, to avoid unused files being left on disk
		dpCnt := 0
		for _, dir := range delPaths.Slice() {
			if strings.HasPrefix(dir, deployDir+"/") { // only check subdir of deploy dir
				dpCnt++
			}
		}
		if cls.CountDir(ins.GetManageHost(), deployDir)-dpCnt == 1 {
			delPaths.Insert(deployDir)
		}

		if svc := ins.ServiceName(); svc != "" {
			delPaths.Insert(fmt.Sprintf("/etc/systemd/system/%s", svc))
		}
		logger.Debugf("Deleting paths on %s: %s\n", ins.GetManageHost(), strings.Join(delPaths.Slice(), " "))
		for _, delPath := range delPaths.Slice() {
			c := module.ShellModuleConfig{
				Command:  fmt.Sprintf("rm -rf %s;", delPath),
				Sudo:     true, // the .service files are in a directory owned by root
				Chdir:    "",
				UseShell: false,
			}
			shell := module.NewShellModule(c)
			_, _, err := shell.Execute(ctx, e)

			if err != nil {
				// Ignore error and continue.For example, deleting a mount point will result in a "Device or resource busy" error.
				logger.Warnf(color.YellowString("Warn: failed to delete path \"%s\" on %s.Please check this error message and manually delete if necessary.\nerrmsg: %s", delPath, ins.GetManageHost(), err))
			}
		}

		logger.Infof("Destroy %s finished\n", ins.GetManageHost())
		logger.Infof("- Destroy %s paths: %v\n", ins.ComponentName(), delPaths.Slice())
	}

	return nil
}
