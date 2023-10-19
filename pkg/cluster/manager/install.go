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
	"errors"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/openGemini/gemix/pkg/cluster/config"
	"github.com/openGemini/gemix/pkg/cluster/operation"
	"github.com/openGemini/gemix/util"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type UploadAction struct {
	uploadInfo []*config.UploadInfo
	remoteHost *config.RemoteHost
}

type Installer interface {
	PrepareForInstall() error
	Install() error
	Close()
}

type GeminiInstaller struct {
	version string
	// ip -> remotes
	remotes map[string]*config.RemoteHost
	uploads map[string]*UploadAction

	// ip -> ssh clients
	sshClients  map[string]*ssh.Client
	sftpClients map[string]*sftp.Client

	configurator config.Configurator // conf reader
	executor     operation.Executor  // execute commands on remote host

	clusterOptions util.ClusterOptions

	wg sync.WaitGroup
}

func NewGeminiInstaller(ops util.ClusterOptions) Installer {
	return &GeminiInstaller{
		remotes:        make(map[string]*config.RemoteHost),
		uploads:        make(map[string]*UploadAction),
		sshClients:     make(map[string]*ssh.Client),
		sftpClients:    make(map[string]*sftp.Client),
		version:        ops.Version,
		configurator:   config.NewGeminiConfigurator(ops.YamlPath, filepath.Join(util.DownloadDst, ops.Version, util.LocalEtcRelPath, util.LocalConfName), filepath.Join(util.DownloadDst, ops.Version, util.LocalEtcRelPath)),
		clusterOptions: ops,
	}
}

func (d *GeminiInstaller) PrepareForInstall() error {
	var err error
	if err = d.configurator.BuildConfig(); err != nil {
		return err
	}
	conf := d.configurator.GetConfig()

	dOps := operation.DownloadOptions{
		Version: d.version,
		Os:      conf.CommonConfig.Os,
		Arch:    conf.CommonConfig.Arch,
	}
	downloader := operation.NewGeminiDownloader(dOps)
	if err = downloader.Run(); err != nil {
		return err
	}

	if err = d.configurator.GenClusterConfs(); err != nil {
		return err
	}

	// check the internet with all the remote servers
	if err = d.prepareRemotes(conf, true); err != nil {
		fmt.Printf("Failed to establish SSH connections with all remote servers. The specific error is: %s\n", err)
		return err
	}
	fmt.Println("Success to establish SSH connections with all remote servers.")

	d.executor = operation.NewGeminiExecutor(d.sshClients)

	if err = d.prepareForUpload(); err != nil {
		return err
	}

	if err = d.prepareUploadActions(conf); err != nil {
		return err
	}

	return nil
}

func (d *GeminiInstaller) prepareRemotes(c *config.Config, needSftp bool) error {
	if c == nil {
		return util.ErrUnexpectedNil
	}

	for ip, ssh := range c.SSHConfig {
		d.remotes[ip] = &config.RemoteHost{
			Ip:         ip,
			SSHPort:    ssh.Port,
			UpDataPath: ssh.UpDataPath,
			LogPath:    ssh.LogPath,
			User:       d.clusterOptions.User,
			Typ:        d.clusterOptions.SshType,
			Password:   d.clusterOptions.Password,
			KeyPath:    d.clusterOptions.Key,
		}
	}

	if err := d.tryConnect(c, needSftp); err != nil {
		return err
	}

	return nil
}

func (d *GeminiInstaller) tryConnect(c *config.Config, needSftp bool) error {
	for ip, r := range d.remotes {
		var err error
		var sshClient *ssh.Client
		switch r.Typ {
		case util.SSH_PW:
			sshClient, err = util.NewSSH_PW(r.User, r.Password, r.Ip, r.SSHPort)
		case util.SSH_KEY:
			sshClient, err = util.NewSSH_Key(r.User, r.KeyPath, r.Ip, r.SSHPort)

		}
		if err != nil {
			return err
		}
		d.sshClients[ip] = sshClient

		if needSftp {
			sftpClient, err := util.NewSftpClient(sshClient)
			if err != nil {
				return err
			}
			d.sftpClients[ip] = sftpClient

			pwd, _ := sftpClient.Getwd()
			// Convert relative paths to absolute paths.
			if len(r.UpDataPath) > 1 && r.UpDataPath[:1] == "~" {
				r.UpDataPath = filepath.Join(pwd, r.UpDataPath[1:])
			}
		}
	}
	if needSftp {
		for _, host := range c.CommonConfig.MetaHosts {
			pwd, _ := d.sftpClients[host].Getwd()
			confPath := filepath.Join(util.DownloadDst, d.version, util.LocalEtcRelPath, host, util.RemoteMetaConfName)
			hostToml, _ := config.ReadFromToml(confPath)
			// Convert relative paths in openGemini.conf to absolute paths.
			hostToml = config.ConvertToml(hostToml, pwd)
			if err := config.GenNewToml(hostToml, confPath); err != nil {
				return err
			}
		}
		for _, host := range c.CommonConfig.SqlHosts {
			pwd, _ := d.sftpClients[host].Getwd()
			confPath := filepath.Join(util.DownloadDst, d.version, util.LocalEtcRelPath, host, util.RemoteSqlConfName)
			hostToml, _ := config.ReadFromToml(confPath)
			// Convert relative paths in openGemini.conf to absolute paths.
			hostToml = config.ConvertToml(hostToml, pwd)
			if err := config.GenNewToml(hostToml, confPath); err != nil {
				return err
			}
		}
		for _, host := range c.CommonConfig.StoreHosts {
			pwd, _ := d.sftpClients[host].Getwd()
			confPath := filepath.Join(util.DownloadDst, d.version, util.LocalEtcRelPath, host, util.RemoteStoreConfName)
			hostToml, _ := config.ReadFromToml(confPath)
			// Convert relative paths in openGemini.conf to absolute paths.
			hostToml = config.ConvertToml(hostToml, pwd)
			if err := config.GenNewToml(hostToml, confPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *GeminiInstaller) prepareForUpload() error {
	if d.executor == nil {
		return util.ErrUnexpectedNil
	}
	for ip, r := range d.remotes {
		binPath := filepath.Join(r.UpDataPath, d.version, util.RemoteBinRelPath)
		etcPath := filepath.Join(r.UpDataPath, d.version, util.RemoteEtcRelPath)
		command := fmt.Sprintf("mkdir -p %s; mkdir -p %s;", binPath, etcPath)
		if _, err := d.executor.ExecCommand(ip, command); err != nil {
			return err
		}
	}
	return nil
}

func (d *GeminiInstaller) prepareUploadActions(c *config.Config) error {
	// ts-meta(bin and config files)
	for _, host := range c.CommonConfig.MetaHosts {
		if d.uploads[host] == nil {
			d.uploads[host] = &UploadAction{
				remoteHost: d.remotes[host],
			}
		}
		d.uploads[host].uploadInfo = append(d.uploads[host].uploadInfo, &config.UploadInfo{
			LocalPath:  filepath.Join(util.DownloadDst, d.version, util.LocalBinRelPath, util.TsMeta),
			RemotePath: filepath.Join(d.remotes[host].UpDataPath, d.version, util.RemoteBinRelPath),
			FileName:   util.TsMeta,
		})
		d.uploads[host].uploadInfo = append(d.uploads[host].uploadInfo, &config.UploadInfo{
			LocalPath:  filepath.Join(util.DownloadDst, d.version, util.LocalEtcRelPath, host, util.RemoteMetaConfName),
			RemotePath: filepath.Join(d.remotes[host].UpDataPath, d.version, util.RemoteEtcRelPath),
			FileName:   util.RemoteMetaConfName,
		})
	}

	// ts-sql(bin and config files)
	for _, host := range c.CommonConfig.SqlHosts {
		if d.uploads[host] == nil {
			d.uploads[host] = &UploadAction{
				remoteHost: d.remotes[host],
			}
		}
		d.uploads[host].uploadInfo = append(d.uploads[host].uploadInfo, &config.UploadInfo{
			LocalPath:  filepath.Join(util.DownloadDst, d.version, util.LocalBinRelPath, util.TsSql),
			RemotePath: filepath.Join(d.remotes[host].UpDataPath, d.version, util.RemoteBinRelPath),
			FileName:   util.TsSql,
		})
		d.uploads[host].uploadInfo = append(d.uploads[host].uploadInfo, &config.UploadInfo{
			LocalPath:  filepath.Join(util.DownloadDst, d.version, util.LocalEtcRelPath, host, util.RemoteSqlConfName),
			RemotePath: filepath.Join(d.remotes[host].UpDataPath, d.version, util.RemoteEtcRelPath),
			FileName:   util.RemoteSqlConfName,
		})
	}

	// ts-store(bin and config files)
	for _, host := range c.CommonConfig.StoreHosts {
		if d.uploads[host] == nil {
			d.uploads[host] = &UploadAction{
				remoteHost: d.remotes[host],
			}
		}
		d.uploads[host].uploadInfo = append(d.uploads[host].uploadInfo, &config.UploadInfo{
			LocalPath:  filepath.Join(util.DownloadDst, d.version, util.LocalBinRelPath, util.TsStore),
			RemotePath: filepath.Join(d.remotes[host].UpDataPath, d.version, util.RemoteBinRelPath),
			FileName:   util.TsStore,
		})
		d.uploads[host].uploadInfo = append(d.uploads[host].uploadInfo, &config.UploadInfo{
			LocalPath:  filepath.Join(util.DownloadDst, d.version, util.LocalEtcRelPath, host, util.RemoteStoreConfName),
			RemotePath: filepath.Join(d.remotes[host].UpDataPath, d.version, util.RemoteEtcRelPath),
			FileName:   util.RemoteStoreConfName,
		})
	}

	// script
	for host := range c.SSHConfig {
		if d.uploads[host] == nil {
			d.uploads[host] = &UploadAction{
				remoteHost: d.remotes[host],
			}
		}
		d.uploads[host].uploadInfo = append(d.uploads[host].uploadInfo, &config.UploadInfo{
			LocalPath:  util.InstallScriptPath,
			RemotePath: filepath.Join(d.remotes[host].UpDataPath, d.version, util.RemoteEtcRelPath),
			FileName:   util.InstallScript,
		})
	}

	return nil
}

func (d *GeminiInstaller) Install() error {
	fmt.Println("Start to install openGemini...")
	errChan := make(chan error, len(d.uploads))
	var wgp sync.WaitGroup
	wgp.Add(2)

	go func() {
		defer wgp.Done()
		d.wg.Add(len(d.uploads))
		for ip, action := range d.uploads {
			go func(ip string, action *UploadAction, errChan chan error) {
				defer d.wg.Done()
				for _, c := range action.uploadInfo {
					// check whether need to upload the file
					// only support Linux
					cmd := fmt.Sprintf("if [ -f %s ]; then echo 'File exists'; else echo 'File not found'; fi", filepath.Join(c.RemotePath, c.FileName))
					output, err := d.executor.ExecCommand(ip, cmd)
					if string(output) == "File exists\n" && err == nil {
						fmt.Printf("%s exists on %s.\n", c.FileName, c.RemotePath)
					} else {
						if err := util.UploadFile(action.remoteHost.Ip, c.LocalPath, c.RemotePath, d.sftpClients[action.remoteHost.Ip]); err != nil {
							fmt.Printf("upload %s to %s error: %v\n", c.LocalPath, action.remoteHost.Ip, err)
							errChan <- err
						}
					}
				}

			}(ip, action, errChan)
		}
		d.wg.Wait()
		close(errChan)
	}()

	var has_err = false
	go func() {
		defer wgp.Done()
		for {
			err, ok := <-errChan
			if !ok {
				break
			}
			fmt.Println(err)
			has_err = true
		}
	}()

	wgp.Wait()
	if has_err {
		return errors.New("install cluster failed")
	} else {
		return nil
	}
}

func (d *GeminiInstaller) Close() {
	var err error
	for _, sftp := range d.sftpClients {
		if sftp != nil {
			if err = sftp.Close(); err != nil {
				fmt.Println(err)
			}
		}
	}

	for _, ssh := range d.sshClients {
		if err = ssh.Close(); err != nil {
			fmt.Println(err)
		}
	}
}
