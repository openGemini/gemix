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
	"crypto/tls"
	"fmt"
	"time"

	"github.com/openGemini/gemix/pkg/cluster/ctxt"
	"github.com/openGemini/gemix/pkg/cluster/module"
	"github.com/openGemini/gemix/pkg/cluster/spec"
	logprinter "github.com/openGemini/gemix/pkg/logger/printer"
	"github.com/openGemini/gemix/pkg/set"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	actionPrevMsgs = map[string]string{
		"start":   "Starting",
		"stop":    "Stopping",
		"enable":  "Enabling",
		"disable": "Disabling",
	}
	actionPostMsgs = map[string]string{}
)

func init() {
	for action := range actionPrevMsgs {
		actionPostMsgs[action] = cases.Title(language.English).String(action)
	}
}

// Start the cluster.
func Start(
	ctx context.Context,
	cluster spec.Topology,
	options Options,
	tlsCfg *tls.Config,
) error {
	uniqueHosts := set.NewStringSet()
	roleFilter := set.NewStringSet(options.Roles...)
	nodeFilter := set.NewStringSet(options.Nodes...)
	components := cluster.ComponentsByStartOrder()
	components = FilterComponent(components, roleFilter)

	for _, comp := range components {
		insts := FilterInstance(comp.Instances(), nodeFilter)
		err := StartComponent(ctx, insts, options, tlsCfg)
		if err != nil {
			return errors.WithMessagef(err, "failed to start %s", comp.Name())
		}

		errg, _ := errgroup.WithContext(ctx)
		for _, inst := range insts {
			uniqueHosts.Insert(inst.GetManageHost())
		}
		if err = errg.Wait(); err != nil {
			return err
		}
	}

	//hosts := make([]string, 0, len(uniqueHosts))
	//for host := range uniqueHosts {
	//	hosts = append(hosts, host)
	//}
	//return StartMonitored(ctx, hosts, noAgentHosts, monitoredOptions, options.OptTimeout)
	return nil
}

// Stop the cluster.
func Stop(
	ctx context.Context,
	cluster spec.Topology,
	options Options,
) error {
	roleFilter := set.NewStringSet(options.Roles...)
	nodeFilter := set.NewStringSet(options.Nodes...)
	components := cluster.ComponentsByStopOrder()
	components = FilterComponent(components, roleFilter)

	instCount := map[string]int{}
	cluster.IterInstance(func(inst spec.Instance) {
		instCount[inst.GetManageHost()]++
	})

	for _, comp := range components {
		insts := FilterInstance(comp.Instances(), nodeFilter)
		err := StopComponent(
			ctx,
			insts,
			options,
			true,
		)
		if err != nil && !options.Force {
			return errors.WithMessagef(err, "failed to stop %s", comp.Name())
		}
	}
	return nil
}

func enableInstance(ctx context.Context, ins spec.Instance, timeout uint64, isEnable bool) error {
	e := ctxt.GetInner(ctx).Get(ins.GetManageHost())
	logger := ctx.Value(logprinter.ContextKeyLogger).(*zap.Logger)

	action := "disable"
	if isEnable {
		action = "enable"
	}
	logger.Info(fmt.Sprintf("\t%s instance %s", actionPrevMsgs[action], ins.ID()))

	// Enable/Disable by systemd.
	if err := systemctl(ctx, e, ins.ServiceName(), action, timeout); err != nil {
		return toFailedActionError(err, action, ins.GetManageHost(), ins.ServiceName(), ins.LogDir())
	}

	logger.Info(fmt.Sprintf("\t%s instance %s success", actionPostMsgs[action], ins.ID()))

	return nil
}

func startInstance(ctx context.Context, ins spec.Instance, timeout uint64, tlsCfg *tls.Config) error {
	e := ctxt.GetInner(ctx).Get(ins.GetManageHost())
	logger := ctx.Value(logprinter.ContextKeyLogger).(*zap.Logger)
	logger.Info(fmt.Sprintf("\tStarting instance %s", ins.ID()))

	if err := systemctl(ctx, e, ins.ServiceName(), "start", timeout); err != nil {
		return toFailedActionError(err, "start", ins.GetManageHost(), ins.ServiceName(), ins.LogDir())
	}

	// Check ready.
	if err := ins.Ready(ctx, e, timeout, tlsCfg); err != nil {
		return toFailedActionError(err, "start", ins.GetManageHost(), ins.ServiceName(), ins.LogDir())
	}

	logger.Info(fmt.Sprintf("\tStart instance %s success", ins.ID()))

	return nil
}

func systemctl(ctx context.Context, executor ctxt.Executor, service string, action string, timeout uint64) error {
	logger := ctx.Value(logprinter.ContextKeyLogger).(*zap.Logger)
	c := module.SystemdModuleConfig{
		Unit:         service,
		ReloadDaemon: true,
		Action:       action,
		Timeout:      time.Second * time.Duration(timeout),
	}
	systemd := module.NewSystemdModule(c)
	stdout, stderr, err := systemd.Execute(ctx, executor)

	if len(stdout) > 0 {
		fmt.Println(string(stdout))
	}
	if len(stderr) > 0 && !bytes.Contains(stderr, []byte("Created symlink ")) && !bytes.Contains(stderr, []byte("Removed symlink ")) {
		logger.Error(string(stderr))
	}
	if len(stderr) > 0 && action == "stop" {
		// ignore "unit not loaded" error, as this means the unit is not
		// exist, and that's exactly what we want
		// NOTE: there will be a potential bug if the unit name is set
		// wrong and the real unit still remains started.
		if bytes.Contains(stderr, []byte(" not loaded.")) {
			logger.Warn(string(stderr))
			return nil // reset the error to avoid exiting
		}
		logger.Error(string(stderr))
	}
	return err
}

// EnableComponent enable/disable the instances
func EnableComponent(ctx context.Context, instances []spec.Instance, noAgentHosts set.StringSet, options Options, isEnable bool) error {
	if len(instances) == 0 {
		return nil
	}

	logger := ctx.Value(logprinter.ContextKeyLogger).(*zap.Logger)
	name := instances[0].ComponentName()
	if isEnable {
		logger.Info(fmt.Sprintf("Enabling component %s", name))
	} else {
		logger.Info(fmt.Sprintf("Disabling component %s", name))
	}

	errg, _ := errgroup.WithContext(ctx)

	for _, ins := range instances {
		ins := ins

		errg.Go(func() error {
			err := enableInstance(ctx, ins, options.OptTimeout, isEnable)
			if err != nil {
				return err
			}
			return nil
		})
	}

	return errg.Wait()
}

// StartComponent start the instances.
func StartComponent(ctx context.Context, instances []spec.Instance, options Options, tlsCfg *tls.Config) error {
	if len(instances) == 0 {
		return nil
	}

	logger := ctx.Value(logprinter.ContextKeyLogger).(*zap.Logger)
	name := instances[0].ComponentName()
	logger.Info(fmt.Sprintf("Starting component %s", name))

	errg, _ := errgroup.WithContext(ctx)
	for _, ins := range instances {
		ins := ins

		errg.Go(func() error {
			if err := ins.PrepareStart(ctx, tlsCfg); err != nil {
				return err
			}
			return startInstance(ctx, ins, options.OptTimeout, tlsCfg)
		})
	}

	return errg.Wait()
}

func stopInstance(ctx context.Context, ins spec.Instance, timeout uint64) error {
	e := ctxt.GetInner(ctx).Get(ins.GetManageHost())
	logger := ctx.Value(logprinter.ContextKeyLogger).(*zap.Logger)
	logger.Info(fmt.Sprintf("\tStopping instance %s", ins.GetManageHost()))

	if err := systemctl(ctx, e, ins.ServiceName(), "stop", timeout); err != nil {
		return toFailedActionError(err, "stop", ins.GetManageHost(), ins.ServiceName(), ins.LogDir())
	}

	logger.Info(fmt.Sprintf("\tStop %s %s success", ins.ComponentName(), ins.ID()))
	return nil
}

// StopComponent stop the instances.
func StopComponent(ctx context.Context,
	instances []spec.Instance,
	options Options,
	forceStop bool,
) error {
	if len(instances) == 0 {
		return nil
	}

	logger := ctx.Value(logprinter.ContextKeyLogger).(*zap.Logger)
	name := instances[0].ComponentName()
	logger.Info(fmt.Sprintf("Stopping component %s", name))

	errg, _ := errgroup.WithContext(ctx)

	for _, ins := range instances {
		ins := ins
		// the checkpoint part of context can't be shared between goroutines
		// since it's used to trace the stack, so we must create a new layer
		// of checkpoint context every time put it into a new goroutine.
		errg.Go(func() error {
			err := stopInstance(ctx, ins, options.OptTimeout)
			if err != nil {
				return err
			}
			return nil
		})
	}

	return errg.Wait()
}

// toFailedActionError formats the errror msg for failed action
func toFailedActionError(err error, action string, host, service, logDir string) error {
	return errors.WithMessagef(err,
		"failed to %s: %s %s, please check the instance's log(%s) for more detail.",
		action, host, service, logDir,
	)
}

//lint:ignore U1000 keep this
func executeSSHCommand(ctx context.Context, action, host, command string) error {
	if command == "" {
		return nil
	}
	e, found := ctxt.GetInner(ctx).GetExecutor(host)
	if !found {
		return fmt.Errorf("no executor")
	}
	logger := ctx.Value(logprinter.ContextKeyLogger).(*zap.Logger)
	logger.Info(fmt.Sprintf("\t%s on %s", action, host))
	stdout, stderr, err := e.Execute(ctx, command, false)
	if err != nil {
		return errors.WithMessagef(err, "stderr: %s", string(stderr))
	}
	logger.Info(fmt.Sprintf("\t%s", stdout))
	return nil
}
