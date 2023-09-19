package deploy

import (
	"fmt"
	"io/ioutil"
	"openGemini-UP/pkg/config"
	"openGemini-UP/pkg/download"
	"openGemini-UP/pkg/exec"
	"openGemini-UP/util"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type ClusterOptions struct {
	Version  string
	User     string
	Key      string
	Password string
	SshType  config.SSHType
	YamlPath string
}

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
	executor     exec.Executor       // execute commands on remote host

	clusterOptions ClusterOptions

	wg sync.WaitGroup
}

func NewGeminiDeployer(ops ClusterOptions) Deployer {
	return &GeminiDeployer{
		remotes:        make(map[string]*config.RemoteHost),
		uploads:        make(map[string]*UploadAction),
		sshClients:     make(map[string]*ssh.Client),
		sftpClients:    make(map[string]*sftp.Client),
		version:        ops.Version,
		configurator:   config.NewGeminiConfigurator(ops.YamlPath, filepath.Join(util.Download_dst, ops.Version, util.Local_etc_rel_path, util.Local_conf_name), filepath.Join(util.Download_dst, util.Local_etc_rel_path), ops.Version),
		runs:           &exec.RunActions{},
		clusterOptions: ops,
	}
}

func (d *GeminiDeployer) PrepareForDeploy() error {
	var err error
	if err = d.configurator.Run(); err != nil {
		return err
	}
	conf := d.configurator.GetConfig()

	dOps := download.DownloadOptions{
		Version: d.version,
		Os:      conf.CommonConfig.Os,
		Arch:    conf.CommonConfig.Arch,
	}
	downloader := download.NewGeminiDownloader(dOps)
	if err = downloader.Run(); err != nil {
		return err
	}

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
	var version string
	if version, err = d.getVersion(); err != nil {
		return err
	}
	d.version = version

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

			pwd, _ := sftpClient.Getwd()
			// Convert relative paths to absolute paths.
			if len(r.UpDataPath) > 1 && r.UpDataPath[:1] == "~" {
				r.UpDataPath = filepath.Join(pwd, r.UpDataPath[1:])
			}

			// Convert relative paths in openGemini.conf to absolute paths.
			confPath := filepath.Join(util.Download_dst, util.Local_etc_rel_path, r.Ip+util.Remote_conf_suffix)
			hostToml, _ := config.ReadFromToml(confPath)
			if len(hostToml.Meta.Dir) > 1 && hostToml.Meta.Dir[:1] == "~" {
				hostToml.Meta.Dir = filepath.Join(pwd, hostToml.Meta.Dir[1:])
			}
			if len(hostToml.Data.StoreDataDir) > 1 && hostToml.Data.StoreDataDir[:1] == "~" {
				hostToml.Data.StoreDataDir = filepath.Join(pwd, hostToml.Data.StoreDataDir[1:])
			}
			if len(hostToml.Data.StoreWalDir) > 1 && hostToml.Data.StoreWalDir[:1] == "~" {
				hostToml.Data.StoreWalDir = filepath.Join(pwd, hostToml.Data.StoreWalDir[1:])
			}
			if len(hostToml.Data.StoreMetaDir) > 1 && hostToml.Data.StoreMetaDir[:1] == "~" {
				hostToml.Data.StoreMetaDir = filepath.Join(pwd, hostToml.Data.StoreMetaDir[1:])
			}
			if len(hostToml.Logging.Path) > 1 && hostToml.Logging.Path[:1] == "~" {
				hostToml.Logging.Path = filepath.Join(pwd, hostToml.Logging.Path[1:])
			}
			config.GenNewToml(hostToml, confPath)
		}
	}
	return nil
}

func (d *GeminiDeployer) prepareForUpload() error {
	if d.executor == nil {
		return util.UnexpectedNil
	}
	for ip, r := range d.remotes {
		binPath := filepath.Join(r.UpDataPath, d.version, util.Remote_bin_rel_path)
		etcPath := filepath.Join(r.UpDataPath, d.version, util.Remote_etc_rel_path)
		command := fmt.Sprintf("mkdir -p %s; mkdir -p %s;", binPath, etcPath)
		if _, err := d.executor.ExecCommand(ip, command); err != nil {
			return err
		}
	}
	return nil
}

func (d *GeminiDeployer) prepareUploadActions(c *config.Config) error {
	// ts-meta
	for _, host := range c.CommonConfig.MetaHosts {
		if d.uploads[host] == nil {
			d.uploads[host] = &UploadAction{
				remoteHost: d.remotes[host],
			}
		}
		d.uploads[host].uploadInfo = append(d.uploads[host].uploadInfo, &config.UploadInfo{
			LocalPath:  filepath.Join(util.Download_dst, d.version, util.Local_bin_rel_path, util.TS_META),
			RemotePath: filepath.Join(d.remotes[host].UpDataPath, d.version, util.Remote_bin_rel_path),
			FileName:   util.TS_META,
		})
	}

	// ts-sql
	for _, host := range c.CommonConfig.SqlHosts {
		if d.uploads[host] == nil {
			d.uploads[host] = &UploadAction{
				remoteHost: d.remotes[host],
			}
		}
		d.uploads[host].uploadInfo = append(d.uploads[host].uploadInfo, &config.UploadInfo{
			LocalPath:  filepath.Join(util.Download_dst, d.version, util.Local_bin_rel_path, util.TS_SQL),
			RemotePath: filepath.Join(d.remotes[host].UpDataPath, d.version, util.Remote_bin_rel_path),
			FileName:   util.TS_SQL,
		})
	}

	// ts-store
	for _, host := range c.CommonConfig.StoreHosts {
		if d.uploads[host] == nil {
			d.uploads[host] = &UploadAction{
				remoteHost: d.remotes[host],
			}
		}
		d.uploads[host].uploadInfo = append(d.uploads[host].uploadInfo, &config.UploadInfo{
			LocalPath:  filepath.Join(util.Download_dst, d.version, util.Local_bin_rel_path, util.TS_STORE),
			RemotePath: filepath.Join(d.remotes[host].UpDataPath, d.version, util.Remote_bin_rel_path),
			FileName:   util.TS_STORE,
		})
	}

	// conf and script
	for host := range c.SSHConfig {
		if d.uploads[host] == nil {
			d.uploads[host] = &UploadAction{
				remoteHost: d.remotes[host],
			}
		}
		d.uploads[host].uploadInfo = append(d.uploads[host].uploadInfo, &config.UploadInfo{
			LocalPath:  filepath.Join(util.Download_dst, util.Local_etc_rel_path, host+util.Remote_conf_suffix),
			RemotePath: filepath.Join(d.remotes[host].UpDataPath, d.version, util.Remote_etc_rel_path),
			FileName:   host + util.Remote_conf_suffix,
		})

		d.uploads[host].uploadInfo = append(d.uploads[host].uploadInfo, &config.UploadInfo{
			LocalPath:  util.Install_script_path,
			RemotePath: filepath.Join(d.remotes[host].UpDataPath, d.version, util.Remote_etc_rel_path),
			FileName:   util.Install_Script,
		})
	}

	return nil
}

func (d *GeminiDeployer) prepareRunActions(c *config.Config) error {
	// ts-meta
	i := 1
	for _, host := range c.CommonConfig.MetaHosts {
		d.runs.MetaAction = append(d.runs.MetaAction, &exec.RunAction{
			Info: &exec.RunInfo{
				ScriptPath: filepath.Join(d.remotes[host].UpDataPath, d.version, util.Remote_etc_rel_path, util.Install_Script),
				Args: []string{util.TS_META, d.remotes[host].LogPath,
					filepath.Join(d.remotes[host].UpDataPath, d.version, util.Remote_bin_rel_path, util.TS_META),
					filepath.Join(d.remotes[host].UpDataPath, d.version, util.Remote_etc_rel_path, host+util.Remote_conf_suffix),
					filepath.Join(d.remotes[host].LogPath, util.Remote_pid_path, util.META+strconv.Itoa(i)+util.Remote_pid_suffix),
					filepath.Join(d.remotes[host].LogPath, strconv.Itoa(i), util.META_extra_log+strconv.Itoa(i)+util.Remote_log_suffix),
					strconv.Itoa(i)},
			},
			Remote: d.remotes[host],
		})
		i++
	}

	// ts-sql
	i = 1
	for _, host := range c.CommonConfig.SqlHosts {
		d.runs.SqlAction = append(d.runs.SqlAction, &exec.RunAction{
			Info: &exec.RunInfo{
				ScriptPath: filepath.Join(d.remotes[host].UpDataPath, d.version, util.Remote_etc_rel_path, util.Install_Script),
				Args: []string{util.TS_SQL, d.remotes[host].LogPath,
					filepath.Join(d.remotes[host].UpDataPath, d.version, util.Remote_bin_rel_path, util.TS_SQL),
					filepath.Join(d.remotes[host].UpDataPath, d.version, util.Remote_etc_rel_path, host+util.Remote_conf_suffix),
					filepath.Join(d.remotes[host].LogPath, util.Remote_pid_path, util.SQL+strconv.Itoa(i)+util.Remote_pid_suffix),
					filepath.Join(d.remotes[host].LogPath, strconv.Itoa(i), util.SQL_extra_log+strconv.Itoa(i)+util.Remote_log_suffix),
					strconv.Itoa(i)},
			},
			Remote: d.remotes[host],
		})
		i++
	}

	// ts-store
	i = 1
	for _, host := range c.CommonConfig.StoreHosts {
		d.runs.StoreAction = append(d.runs.StoreAction, &exec.RunAction{
			Info: &exec.RunInfo{
				ScriptPath: filepath.Join(d.remotes[host].UpDataPath, d.version, util.Remote_etc_rel_path, util.Install_Script),
				Args: []string{util.TS_STORE, d.remotes[host].LogPath,
					filepath.Join(d.remotes[host].UpDataPath, d.version, util.Remote_bin_rel_path, util.TS_STORE),
					filepath.Join(d.remotes[host].UpDataPath, d.version, util.Remote_etc_rel_path, host+util.Remote_conf_suffix),
					filepath.Join(d.remotes[host].LogPath, util.Remote_pid_path, util.STORE+strconv.Itoa(i)+util.Remote_pid_suffix),
					filepath.Join(d.remotes[host].LogPath, strconv.Itoa(i), util.STORE_extra_log+strconv.Itoa(i)+util.Remote_log_suffix),
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
	d.saveVersion()
	return nil
}

func (d *GeminiDeployer) Start() error {
	d.startCluster()
	return nil
}

func (d *GeminiDeployer) saveVersion() error {
	filePath := filepath.Join(util.Download_dst, util.VersionFile)
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(d.version)
	if err != nil {
		return err
	}
	return nil
}

func (d *GeminiDeployer) getVersion() (string, error) {
	filePath := filepath.Join(util.Download_dst, util.VersionFile)
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (d *GeminiDeployer) uploadFiles() {
	d.wg.Add(len(d.uploads))
	for ip, action := range d.uploads {
		go func(ip string, action *UploadAction) {
			defer d.wg.Done()
			for _, c := range action.uploadInfo {
				// check whether need to upload the file
				// only support Linux
				cmd := fmt.Sprintf("if [ -f %s ]; then echo 'File exists'; else echo 'File not found'; fi", filepath.Join(c.RemotePath, c.FileName))
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
