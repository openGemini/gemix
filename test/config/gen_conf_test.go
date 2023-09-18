package test

import (
	"openGemini-UP/pkg/config"
	"testing"
)

func TestGenConf(t *testing.T) {
	tomlPath := "../example/openGenimi-template.conf"
	yamlPath := "../example/topology.example1.yaml"

	var conf config.Toml
	var y config.Yaml
	var err error
	conf, err = config.ReadFromToml(tomlPath)
	if err != nil {
		t.Fatalf("Read from toml failed: %v", err)
	}

	y, err = config.ReadFromYaml(yamlPath)
	if err != nil {
		t.Fatalf("Read from yaml failed: %v", err)
	}

	config.GenConfs(y, conf, "../../data/etc")
}

func TestGenConfWithGlobalDefaults(t *testing.T) {
	tomlPath := "../example/openGenimi-template.conf"
	yamlPath := "../example/topology.example2.yaml"

	var conf config.Toml
	var y config.Yaml
	var err error
	conf, err = config.ReadFromToml(tomlPath)
	if err != nil {
		t.Fatalf("Read from toml failed: %v", err)
	}

	y, err = config.ReadFromYaml(yamlPath)
	if err != nil {
		t.Fatalf("Read from yaml failed: %v", err)
	}

	config.GenConfs(y, conf, "../../data/etc")
}

func TestGenConfWith3Hosts(t *testing.T) {
	tomlPath := "../example/openGenimi-template.conf"
	yamlPath := "../example/topology.example3.yaml"

	var conf config.Toml
	var y config.Yaml
	var err error
	conf, err = config.ReadFromToml(tomlPath)
	if err != nil {
		t.Fatalf("Read from toml failed: %v", err)
	}

	y, err = config.ReadFromYaml(yamlPath)
	if err != nil {
		t.Fatalf("Read from yaml failed: %v", err)
	}

	config.GenConfs(y, conf, "../../data/etc")
}
