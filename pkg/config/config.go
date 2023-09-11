package config

import (
	"io/ioutil"
	"openGemini-UP/util"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

type CommonConfig struct {
	MetaHosts  []string `toml:"meta"`
	StoreHosts []string `toml:"store"`
	SqlHosts   []string `toml:"sql"`
}

type HostConfig struct {
	HostNames []string `toml:"name"`
	HostIPs   []string `toml:"ip"`
}

type SSHConfig struct {
	Port       int    `toml:"port"`
	User       string `toml:"user"`
	Typ        string `toml:"type"`
	Password   string `toml:"password"`
	KeyPath    string `toml:"key-path"`
	UpDataPath string `toml:"up-data-path"`
}

type Config struct {
	CommonConfig *CommonConfig `toml:"common"`
	HostConfig   *HostConfig   `toml:"host"`
	SSHConfig    *SSHConfig    `toml:"ssh"`
}

type Configurator interface {
	Run() error
	RunWithoutGen() error
	GetConfig() *Config
}

type GeminiConfigurator struct {
	confPath   string
	scriptPath string
	conf       *Config
}

func NewGeminiConfigurator(cPath, sPath string) Configurator {
	return &GeminiConfigurator{
		confPath:   cPath,
		scriptPath: sPath,
		conf: &Config{
			CommonConfig: &CommonConfig{},
			HostConfig:   &HostConfig{},
			SSHConfig:    &SSHConfig{},
		},
	}
}

func (c *GeminiConfigurator) Run() error {
	var err error
	if err = c.fromTomlFile(); err != nil {
		return err
	}
	if err = c.generateConf(); err != nil {
		return err
	}
	return nil
}

func (c *GeminiConfigurator) RunWithoutGen() error {
	var err error
	if err = c.fromTomlFile(); err != nil {
		return err
	}
	return nil
}

func (c *GeminiConfigurator) generateConf() error {
	confPah := filepath.Join(util.Download_dst, util.Local_etc_rel_path)
	if _, err := os.Stat(confPah); os.IsNotExist(err) {
		errDir := os.MkdirAll(confPah, 0755)
		if errDir != nil {
			return errDir
		}
	}

	cmd := exec.Command("/bin/bash", c.scriptPath)
	cmd.Args = append(cmd.Args, util.Download_dst)
	cmd.Args = append(cmd.Args, c.conf.HostConfig.HostIPs...)

	_, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	return nil
}

func (c *GeminiConfigurator) fromTomlFile() error {
	content, err := ioutil.ReadFile(path.Clean(c.confPath))
	if err != nil {
		return err
	}

	dec := unicode.BOMOverride(transform.Nop)
	content, _, err = transform.Bytes(dec, content)
	if err != nil {
		return err
	}
	return fromToml(c.conf, string(content))
}

func fromToml(c *Config, input string) error {
	_, err := toml.Decode(input, c)
	return err
}

func (c *GeminiConfigurator) GetConfig() *Config {
	return c.conf
}
