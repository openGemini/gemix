package stop

import (
	"fmt"
	"openGemini-UP/pkg/config"
	"openGemini-UP/pkg/exec"
	"openGemini-UP/util"
	"sync"

	"golang.org/x/crypto/ssh"
)

type Stop interface {
	Prepare() error
	Run() error
	Close()
}

type GeminiStop struct {
	// ip -> remotes
	remotes map[string]*config.RemoteHost

	// ip -> actions
	stops map[string]*exec.StopAction

	// ip -> ssh clients
	sshClients map[string]*ssh.Client

	configurator config.Configurator // conf reader
	executor     exec.Executor       // execute commands on remote host

	needDelete bool // whether to delete logs and data
	upDataPath string

	wg sync.WaitGroup
}

func NewGeminiStop(delete bool) Stop {
	new := &GeminiStop{
		remotes:      make(map[string]*config.RemoteHost),
		stops:        make(map[string]*exec.StopAction),
		sshClients:   make(map[string]*ssh.Client),
		configurator: config.NewGeminiConfigurator(util.User_conf_path, util.Conf_gen_script_path),
		needDelete:   delete,
	}
	return new
}

func (s *GeminiStop) Prepare() error {
	var err error
	if err = s.configurator.RunWithoutGen(); err != nil {
		return err
	}
	conf := s.configurator.GetConfig()
	s.upDataPath = conf.SSHConfig.UpDataPath

	if err = s.prepareRemotes(conf); err != nil {
		return err
	}

	s.executor = exec.NewGeminiExecutor(s.sshClients)

	if err = s.prepareStopActions(conf); err != nil {
		return err
	}

	return nil
}

func (s *GeminiStop) prepareRemotes(c *config.Config) error {
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
		s.remotes[ip] = &config.RemoteHost{
			Ip:       ip,
			SSHPort:  sshConfig.Port,
			User:     sshConfig.User,
			Password: sshConfig.Password,
			KeyPath:  sshConfig.KeyPath,
			Typ:      typ,
		}
	}

	if err := s.tryConnect(); err != nil {
		return err
	}

	return nil
}

func (s *GeminiStop) tryConnect() error {
	for ip, r := range s.remotes {
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
		s.sshClients[ip] = sshClient
	}
	return nil
}

func (s *GeminiStop) prepareStopActions(c *config.Config) error {
	hostMap := make(map[string]string)
	for i := 0; i < len(c.HostConfig.HostNames); i++ {
		hostMap[c.HostConfig.HostNames[i]] = c.HostConfig.HostIPs[i]
	}

	// ts-meta
	for _, hostName := range c.CommonConfig.MetaHosts {
		ip := hostMap[hostName]
		if s.stops[ip] == nil {
			s.stops[ip] = &exec.StopAction{
				Remote: s.remotes[ip],
			}
		}
		s.stops[ip].ProcessNames = append(s.stops[ip].ProcessNames, util.TS_META)
	}

	// ts-sql
	for _, hostName := range c.CommonConfig.SqlHosts {
		ip := hostMap[hostName]
		if s.stops[ip] == nil {
			s.stops[ip] = &exec.StopAction{
				Remote: s.remotes[ip],
			}
		}
		s.stops[ip].ProcessNames = append(s.stops[ip].ProcessNames, util.TS_SQL)
	}

	// ts-store
	for _, hostName := range c.CommonConfig.StoreHosts {
		ip := hostMap[hostName]
		if s.stops[ip] == nil {
			s.stops[ip] = &exec.StopAction{
				Remote: s.remotes[ip],
			}
		}
		s.stops[ip].ProcessNames = append(s.stops[ip].ProcessNames, util.TS_STORE)
	}
	return nil
}

func (s *GeminiStop) Run() error {
	if s.executor == nil {
		return util.UnexpectedNil
	}

	s.wg.Add(len(s.stops))
	for _, action := range s.stops {
		go func(action *exec.StopAction) {
			defer s.wg.Done()
			s.executor.ExecStopAction(action)
		}(action)
	}
	s.wg.Wait()

	// need to delete all logs and data on remote hosts
	if s.needDelete {
		s.wg.Add(len(s.stops))
		for ip := range s.stops {
			go func(ip string) {
				defer s.wg.Done()
				command := fmt.Sprintf("rm -rf %s; rm -rf %s;", s.upDataPath, util.OpenGemini_path)
				s.executor.ExecCommand(ip, command)
			}(ip)
		}
		s.wg.Wait()
	}
	return nil
}

func (s *GeminiStop) Close() {
	var err error
	for _, ssh := range s.sshClients {
		if ssh != nil {
			if err = ssh.Close(); err != nil {
				fmt.Println(err)
			}
		}
	}
}
