package deploy

import (
	"fmt"
	"openGemini-UP/pkg/config"
	"openGemini-UP/pkg/download"
	"openGemini-UP/pkg/exec"
	"openGemini-UP/util"
	"strconv"
	"sync"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type UploadAction struct {
	uploadInfo []*config.UploadInfo
	remoteHost *config.RemoteHost
}

type Deployer interface {
	PrepareForDeploy() error
	Deploy() error
	PrepareForStart() error
	Start() error
	Close()
}

type GeminiDeployer struct {
	version string
	// ip -> remotes
	remotes map[string]*config.RemoteHost
	uploads map[string]*UploadAction
	runs    *exec.RunActions

	// ip -> ssh clients
	sshClients  map[string]*ssh.Client
	sftpClients map[string]*sftp.Client

	configurator config.Configurator // conf reader
	downloader   download.Downloader // download files from internet
	executor     exec.Executor       // execute commands on remote host

	wg sync.WaitGroup
}

func NewGeminiDeployer(v string) Deployer {
	return &GeminiDeployer{
		remotes:      make(map[string]*config.RemoteHost),
		uploads:      make(map[string]*UploadAction),
		sshClients:   make(map[string]*ssh.Client),
		sftpClients:  make(map[string]*sftp.Client),
		version:      v,
		configurator: config.NewGeminiConfigurator(util.User_conf_path, util.Conf_gen_script_path),
		downloader:   download.NewGeminiDownloader(v),
		runs:         &exec.RunActions{},
	}
}

func (d *GeminiDeployer) PrepareForDeploy() error {
	var err error
	if err = d.downloader.Run(); err != nil {
		return err
	}

	if err = d.configurator.Run(); err != nil {
		return err
	}
	conf := d.configurator.GetConfig()

	if err = d.prepareRemotes(conf, true); err != nil {
		return err
	}

	d.executor = exec.NewGeminiExecutor(d.sshClients)

	if err = d.prepareForUpload(); err != nil {
		return err
	}

	if err = d.prepareUploadActions(conf); err != nil {
		return err
	}

	if err = d.prepareRunActions(conf); err != nil {
		return err
	}

	return nil
}

func (d *GeminiDeployer) PrepareForStart() error {
	var err error
	if err = d.configurator.RunWithoutGen(); err != nil {
		return err
	}
	conf := d.configurator.GetConfig()

	if err = d.prepareRemotes(conf, false); err != nil {
		return err
	}

	d.executor = exec.NewGeminiExecutor(d.sshClients)

	if err = d.prepareRunActions(conf); err != nil {
		return err
	}

	return nil
}

func (d *GeminiDeployer) prepareRemotes(c *config.Config, needSftp bool) error {
	if c == nil {
		return util.UnexpectedNil
	}

	sshConfig := c.SSHConfig
	var typ config.SSHType
	switch sshConfig.Typ {
	case util.SSH_KEY:
		typ = config.SSH_KEY
	case util.SSH_PW:
		typ = config.SSH_PW
	default:
		return util.UnknowSSHType
	}

	for _, ip := range c.HostConfig.HostIPs {
		d.remotes[ip] = &config.RemoteHost{
			Ip:         ip,
			SSHPort:    sshConfig.Port,
			User:       sshConfig.User,
			Password:   sshConfig.Password,
			KeyPath:    sshConfig.KeyPath,
			Typ:        typ,
			UpDataPath: sshConfig.UpDataPath,
		}
	}

	if err := d.tryConnect(needSftp); err != nil {
		return err
	}

	return nil
}

func (d *GeminiDeployer) tryConnect(needSftp bool) error {
	for ip, r := range d.remotes {
		var err error
		var sshClient *ssh.Client
		switch r.Typ {
		case config.SSH_PW:
			sshClient, err = util.NewSSH_PW(r.User, r.Password, r.Ip, r.SSHPort)
		case config.SSH_KEY:
			sshClient, err = util.NewSSH_Key(r.User, r.KeyPath, r.Ip, r.SSHPort)

		}
		if err != nil {
			// TODO(Benevor):close all connection and exit
			return err
		}
		d.sshClients[ip] = sshClient

		if needSftp {
			sftpClient, err := util.NewSftpClient(sshClient)
			if err != nil {
				// TODO(Benevor):close all connection and exit
				return err
			}
			d.sftpClients[ip] = sftpClient

			// Convert relative paths to absolute paths.
			if r.UpDataPath[:1] == "~" {
				pwd, _ := sftpClient.Getwd()
				r.UpDataPath = pwd + r.UpDataPath[1:]
			}
		}
	}
	return nil
}

func (d *GeminiDeployer) prepareForUpload() error {
	if d.executor == nil {
		return util.UnexpectedNil
	}
	for ip, r := range d.remotes {
		binPath := r.UpDataPath + d.version + util.Remote_bin_rel_path
		etcPath := r.UpDataPath + d.version + util.Remote_etc_rel_path
		command := fmt.Sprintf("mkdir -p %s; mkdir -p %s;", binPath, etcPath)
		if _, err := d.executor.ExecCommand(ip, command); err != nil {
			return err
		}
	}
	return nil
}

func (d *GeminiDeployer) prepareUploadActions(c *config.Config) error {
	hostMap := make(map[string]string)
	for i := 0; i < len(c.HostConfig.HostNames); i++ {
		hostMap[c.HostConfig.HostNames[i]] = c.HostConfig.HostIPs[i]
	}

	// ts-meta
	for _, hostName := range c.CommonConfig.MetaHosts {
		host := hostMap[hostName]
		if d.uploads[host] == nil {
			d.uploads[host] = &UploadAction{
				remoteHost: d.remotes[host],
			}
		}
		d.uploads[host].uploadInfo = append(d.uploads[host].uploadInfo, &config.UploadInfo{
			LocalPath:  util.Download_dst + "/" + d.version + util.Local_bin_rel_path + util.TS_META,
			RemotePath: d.remotes[host].UpDataPath + d.version + util.Remote_bin_rel_path,
			FileName:   util.TS_META,
		})
	}

	// ts-sql
	for _, hostName := range c.CommonConfig.SqlHosts {
		host := hostMap[hostName]
		if d.uploads[host] == nil {
			d.uploads[host] = &UploadAction{
				remoteHost: d.remotes[host],
			}
		}
		d.uploads[host].uploadInfo = append(d.uploads[host].uploadInfo, &config.UploadInfo{
			LocalPath:  util.Download_dst + "/" + d.version + util.Local_bin_rel_path + util.TS_SQL,
			RemotePath: d.remotes[host].UpDataPath + d.version + util.Remote_bin_rel_path,
			FileName:   util.TS_SQL,
		})
	}

	// ts-store
	for _, hostName := range c.CommonConfig.StoreHosts {
		host := hostMap[hostName]
		if d.uploads[host] == nil {
			d.uploads[host] = &UploadAction{
				remoteHost: d.remotes[host],
			}
		}
		d.uploads[host].uploadInfo = append(d.uploads[host].uploadInfo, &config.UploadInfo{
			LocalPath:  util.Download_dst + "/" + d.version + util.Local_bin_rel_path + util.TS_STORE,
			RemotePath: d.remotes[host].UpDataPath + d.version + util.Remote_bin_rel_path,
			FileName:   util.TS_STORE,
		})
	}

	// conf and script
	for _, host := range c.HostConfig.HostIPs {
		if d.uploads[host] == nil {
			d.uploads[host] = &UploadAction{
				remoteHost: d.remotes[host],
			}
		}
		d.uploads[host].uploadInfo = append(d.uploads[host].uploadInfo, &config.UploadInfo{
			LocalPath:  util.Download_dst + util.Local_etc_rel_path + host + util.Remote_conf_suffix,
			RemotePath: d.remotes[host].UpDataPath + d.version + util.Remote_etc_rel_path,
			FileName:   host + util.Remote_conf_suffix,
		})

		d.uploads[host].uploadInfo = append(d.uploads[host].uploadInfo, &config.UploadInfo{
			LocalPath:  util.Install_script_path,
			RemotePath: d.remotes[host].UpDataPath + d.version + util.Remote_etc_rel_path,
			FileName:   util.Install_Script,
		})
	}

	return nil
}

func (d *GeminiDeployer) prepareRunActions(c *config.Config) error {
	hostMap := make(map[string]string)
	for i := 0; i < len(c.HostConfig.HostNames); i++ {
		hostMap[c.HostConfig.HostNames[i]] = c.HostConfig.HostIPs[i]
	}

	// ts-meta
	i := 1
	for _, hostName := range c.CommonConfig.MetaHosts {
		host := hostMap[hostName]

		d.runs.MetaAction = append(d.runs.MetaAction, &exec.RunAction{
			Info: &exec.RunInfo{
				ScriptPath: d.remotes[host].UpDataPath + d.version + util.Remote_etc_rel_path + util.Install_Script,
				Args: []string{util.TS_META, util.OpenGemini_path,
					d.remotes[host].UpDataPath + d.version + util.Remote_bin_rel_path + util.TS_META,
					d.remotes[host].UpDataPath + d.version + util.Remote_etc_rel_path + host + util.Remote_conf_suffix,
					util.OpenGemini_path + util.Remote_pid_path + util.META + strconv.Itoa(i) + util.Remote_pid_suffix,
					util.OpenGemini_path + util.Remote_log_path + strconv.Itoa(i) + util.META_extra_log + strconv.Itoa(i) + util.Remote_log_suffix,
					strconv.Itoa(i)},
			},
			Remote: d.remotes[host],
		})
		i++
	}

	// ts-sql
	i = 1
	for _, hostName := range c.CommonConfig.SqlHosts {
		host := hostMap[hostName]

		d.runs.SqlAction = append(d.runs.SqlAction, &exec.RunAction{
			Info: &exec.RunInfo{
				ScriptPath: d.remotes[host].UpDataPath + d.version + util.Remote_etc_rel_path + util.Install_Script,
				Args: []string{util.TS_SQL, util.OpenGemini_path,
					d.remotes[host].UpDataPath + d.version + util.Remote_bin_rel_path + util.TS_SQL,
					d.remotes[host].UpDataPath + d.version + util.Remote_etc_rel_path + host + util.Remote_conf_suffix,
					util.OpenGemini_path + util.Remote_pid_path + util.SQL + strconv.Itoa(i) + util.Remote_pid_suffix,
					util.OpenGemini_path + util.Remote_log_path + strconv.Itoa(i) + util.SQL_extra_log + strconv.Itoa(i) + util.Remote_log_suffix,
					strconv.Itoa(i)},
			},
			Remote: d.remotes[host],
		})
		i++
	}

	// ts-store
	i = 1
	for _, hostName := range c.CommonConfig.StoreHosts {
		host := hostMap[hostName]

		d.runs.StoreAction = append(d.runs.StoreAction, &exec.RunAction{
			Info: &exec.RunInfo{
				ScriptPath: d.remotes[host].UpDataPath + d.version + util.Remote_etc_rel_path + util.Install_Script,
				Args: []string{util.TS_STORE, util.OpenGemini_path,
					d.remotes[host].UpDataPath + d.version + util.Remote_bin_rel_path + util.TS_STORE,
					d.remotes[host].UpDataPath + d.version + util.Remote_etc_rel_path + host + util.Remote_conf_suffix,
					util.OpenGemini_path + util.Remote_pid_path + util.STORE + strconv.Itoa(i) + util.Remote_pid_suffix,
					util.OpenGemini_path + util.Remote_log_path + strconv.Itoa(i) + util.STORE_extra_log + strconv.Itoa(i) + util.Remote_log_suffix,
					strconv.Itoa(i)},
			},
			Remote: d.remotes[host],
		})
		i++
	}

	return nil
}

func (d *GeminiDeployer) Deploy() error {
	d.uploadFiles()
	d.startCluster()
	return nil
}

func (d *GeminiDeployer) Start() error {
	d.startCluster()
	return nil
}

func (d *GeminiDeployer) uploadFiles() {
	d.wg.Add(len(d.uploads))
	for ip, action := range d.uploads {
		go func(ip string, action *UploadAction) {
			defer d.wg.Done()
			for _, c := range action.uploadInfo {
				// check whether need to upload the file
				// only support Linux
				cmd := fmt.Sprintf("if [ -f %s ]; then echo 'File exists'; else echo 'File not found'; fi", c.RemotePath+c.FileName)
				output, err := d.executor.ExecCommand(ip, cmd)
				if string(output) == "File exists\n" && err == nil {
					fmt.Printf("%s exists on %s.\n", c.FileName, c.RemotePath)
				} else {
					util.UploadFile(action.remoteHost.Ip, c.LocalPath, c.RemotePath, d.sftpClients[action.remoteHost.Ip])
				}
			}
		}(ip, action)
	}
	d.wg.Wait()
}

func (d *GeminiDeployer) startCluster() {
	if d.executor == nil {
		return
	}

	// start all ts-meta concurrently
	d.wg.Add(len(d.runs.MetaAction))
	for _, action := range d.runs.MetaAction {
		go func(action *exec.RunAction) {
			defer d.wg.Done()
			d.executor.ExecRunAction(action)
		}(action)
	}
	d.wg.Wait()

	// time for campaign
	time.Sleep(time.Second)

	// start all ts-store and ts-sql concurrently
	d.wg.Add(len(d.runs.SqlAction) + len(d.runs.StoreAction))
	for _, action := range d.runs.StoreAction {
		go func(action *exec.RunAction) {
			defer d.wg.Done()
			d.executor.ExecRunAction(action)
		}(action)
	}
	for _, action := range d.runs.SqlAction {
		go func(action *exec.RunAction) {
			defer d.wg.Done()
			d.executor.ExecRunAction(action)
		}(action)
	}
	d.wg.Wait()
}

func (d *GeminiDeployer) Close() {
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
