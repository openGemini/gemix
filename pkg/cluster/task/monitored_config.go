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
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/openGemini/gemix/pkg/cluster/ctxt"
	"github.com/openGemini/gemix/pkg/cluster/spec"
	"github.com/openGemini/gemix/pkg/cluster/template"
	"github.com/openGemini/gemix/pkg/cluster/template/config"
	"github.com/openGemini/gemix/pkg/cluster/template/scripts"
	system "github.com/openGemini/gemix/pkg/cluster/template/systemd"
	"github.com/openGemini/gemix/pkg/meta"
	"github.com/openGemini/gemix/pkg/utils"
	"github.com/pkg/errors"
)

// MonitoredConfig is used to generate the monitor node configuration
type MonitoredConfig struct {
	clusterName string
	component   string
	host        string
	info        *spec.MonitorHostInfo
	globResCtl  meta.ResourceControl
	options     *spec.TSMonitoredOptions
	deployUser  string
	tlsEnabled  bool
	paths       meta.DirPaths
}

// Execute implements the Task interface
func (m *MonitoredConfig) Execute(ctx context.Context) error {
	// Copy to remote server
	exec, found := ctxt.GetInner(ctx).GetExecutor(m.host)
	if !found {
		return ErrNoExecutor
	}

	if err := utils.MkdirAll(m.paths.Cache, 0755); err != nil {
		return err
	}

	if err := m.syncMonitoredSystemConfig(ctx, exec, m.component); err != nil {
		return err
	}

	var cfg template.ConfigGenerator
	switch m.component {
	case spec.ComponentTSMonitor:
		monitorConfig := &config.TsMonitorConfig{
			Host:         m.host,
			MetricPath:   m.info.MetricPath,
			ErrorLogPath: m.info.ErrorLogPath,
			DataPath:     m.info.DataPath,
			WALPath:      m.info.WALPath,
			ProcessName:  strings.Join(m.info.ProcessName.Slice(), ","),
			MonitorAddr:  m.info.MonitorAddr,
			MonitorDB:    strings.Replace(m.clusterName, "-", "_", -1),
			TLSEnabled:   false,
			LoggingPath:  m.paths.Log,
		}
		if err := m.syncTsMonitorConfig(ctx, exec, monitorConfig); err != nil {
			return err
		}
		cfg = &scripts.TSMonitorScript{
			DeployDir: m.paths.Deploy,
			LogDir:    m.paths.Log,
		}
	default:
		return fmt.Errorf("unknown monitored component %s", m.component)
	}

	return m.syncMonitoredScript(ctx, exec, m.component, cfg)
}

func (m *MonitoredConfig) syncMonitoredSystemConfig(ctx context.Context, exec ctxt.Executor, comp string) (err error) {
	sysCfg := filepath.Join(m.paths.Cache, fmt.Sprintf("%s-%s.service", comp, m.host))

	//// insert checkpoint
	//point := checkpoint.Acquire(ctx, spec.CopyConfigFile, map[string]any{"config-file": sysCfg})
	//defer func() {
	//	point.Release(err, zap.String("config-file", sysCfg))
	//}()
	//if point.Hit() != nil {
	//	return nil
	//}

	user := "root" // TODO: use real user

	resource := spec.MergeResourceControl(m.globResCtl, m.globResCtl)
	systemCfg := system.NewConfig(comp, user, m.paths.Deploy).
		WithMemoryLimit(resource.MemoryLimit).
		WithCPUQuota(resource.CPUQuota).
		WithLimitCORE(resource.LimitCORE).
		WithIOReadBandwidthMax(resource.IOReadBandwidthMax).
		WithIOWriteBandwidthMax(resource.IOWriteBandwidthMax)

	if err = systemCfg.ConfigToFile(sysCfg); err != nil {
		return errors.WithStack(err)
	}
	tgt := filepath.Join("/tmp", comp+"_"+uuid.New().String()+".service")
	if err = exec.Transfer(ctx, sysCfg, tgt, false, 0, false); err != nil {
		return errors.WithMessagef(err, "transfer from %s to %s failed", sysCfg, tgt)
	}
	serviceFile := fmt.Sprintf("%s.service", comp)
	cmd := fmt.Sprintf("mv %s /etc/systemd/system/%s", tgt, serviceFile)
	if _, _, err := exec.Execute(ctx, cmd, true); err != nil {
		return errors.WithMessagef(err, "execute: %s", cmd)
	}
	return nil
}

func (m *MonitoredConfig) syncMonitoredScript(ctx context.Context, exec ctxt.Executor, comp string, cfg template.ConfigGenerator) error {
	fp := filepath.Join(m.paths.Cache, fmt.Sprintf("run_%s_%s.sh", comp, m.host))
	if err := cfg.ConfigToFile(fp); err != nil {
		return err
	}
	dst := filepath.Join(m.paths.Deploy, "scripts", fmt.Sprintf("run_%s.sh", comp))
	if err := exec.Transfer(ctx, fp, dst, false, 0, false); err != nil {
		return err
	}
	if _, _, err := exec.Execute(ctx, "chmod +x "+dst, false); err != nil {
		return err
	}

	return nil
}

func (m *MonitoredConfig) syncTsMonitorConfig(ctx context.Context, exec ctxt.Executor, cfg template.ConfigGenerator) error {
	fp := filepath.Join(m.paths.Cache, fmt.Sprintf("ts-monitor_%s.toml", m.host))
	if err := cfg.ConfigToFile(fp); err != nil {
		return err
	}
	dst := filepath.Join(m.paths.Deploy, "conf", "ts-monitor.toml")
	return exec.Transfer(ctx, fp, dst, false, 0, false)
}

// Rollback implements the Task interface
func (m *MonitoredConfig) Rollback(ctx context.Context) error {
	return ErrUnsupportedRollback
}

// String implements the fmt.Stringer interface
func (m *MonitoredConfig) String() string {
	return fmt.Sprintf("MonitoredConfig: cluster=%s, user=%s, %v", m.clusterName, m.deployUser, m.paths)
}
