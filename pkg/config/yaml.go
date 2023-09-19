package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Yaml struct {
	Global  GlobalYaml  `yaml:"global"`
	TsMeta  []MetaYaml  `yaml:"ts-meta"`
	TsSql   []SqlYaml   `yaml:"ts-sql"`
	TsStore []StoreYaml `yaml:"ts-store"`
}

type GlobalYaml struct {
	SSHPort   int    `yaml:"ssh_port"`
	LogDir    string `yaml:"log_dir"`
	DeployDir string `yaml:"deploy_dir"`

	OS   string `yaml:"os"`
	Arch string `yaml:"arch"`
}

type MetaYaml struct {
	Host string `yaml:"host"`
	// default value in GlobalYaml
	SSHPort   int    `yaml:"ssh_port"`
	LogDir    string `yaml:"log_dir"`
	DeployDir string `yaml:"deploy_dir"`

	// do not need default value
	ClientPort int    `yaml:"client_port"`
	PeerPort   int    `yaml:"peer_port"`
	RaftPort   int    `yaml:"raft_port"`
	GossipPort int    `yaml:"gossip_port"`
	DataDir    string `yaml:"data_dir"`
}

type SqlYaml struct {
	Host string `yaml:"host"`
	// default value in GlobalYaml
	SSHPort   int    `yaml:"ssh_port"`
	LogDir    string `yaml:"log_dir"`
	DeployDir string `yaml:"deploy_dir"`

	// do not need default value
	Port       int `yaml:"port"`
	FlightPort int `yaml:"flight_port"`
}

type StoreYaml struct {
	Host string `yaml:"host"`
	// default value in GlobalYaml
	SSHPort   int    `yaml:"ssh_port"`
	LogDir    string `yaml:"log_dir"`
	DeployDir string `yaml:"deploy_dir"`

	// do not need default value
	IngestPort int    `yaml:"ingest_port"`
	SelectPort int    `yaml:"select_port"`
	GossipPort int    `yaml:"gossip_port"`
	DataDir    string `yaml:"data_dir"`
	MetaDir    string `yaml:"meta_dir"`
}

func updataWithGlobalDefaults(y *Yaml) {
	for i := range y.TsMeta {
		if y.TsMeta[i].SSHPort == 0 {
			y.TsMeta[i].SSHPort = y.Global.SSHPort
		}
		if y.TsMeta[i].LogDir == "" {
			y.TsMeta[i].LogDir = y.Global.LogDir
		}
		if y.TsMeta[i].DeployDir == "" {
			y.TsMeta[i].DeployDir = y.Global.DeployDir
		}
	}
	for i := range y.TsSql {
		if y.TsSql[i].SSHPort == 0 {
			y.TsSql[i].SSHPort = y.Global.SSHPort
		}
		if y.TsSql[i].LogDir == "" {
			y.TsSql[i].LogDir = y.Global.LogDir
		}
		if y.TsSql[i].DeployDir == "" {
			y.TsSql[i].DeployDir = y.Global.DeployDir
		}
	}
	for i := range y.TsStore {
		if y.TsStore[i].SSHPort == 0 {
			y.TsStore[i].SSHPort = y.Global.SSHPort
		}
		if y.TsStore[i].LogDir == "" {
			y.TsStore[i].LogDir = y.Global.LogDir
		}
		if y.TsStore[i].DeployDir == "" {
			y.TsStore[i].DeployDir = y.Global.DeployDir
		}
	}
}

func ReadFromYaml(yamlPath string) (Yaml, error) {
	var err error
	yamlFile, err := ioutil.ReadFile(yamlPath)
	if err != nil {
		return Yaml{}, err
	}
	var y Yaml
	if err = yaml.Unmarshal(yamlFile, &y); err != nil {
		return Yaml{}, err
	}

	// Update with default values
	updataWithGlobalDefaults(&y)

	return y, nil
}
