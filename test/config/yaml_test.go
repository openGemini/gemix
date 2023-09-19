package test

import (
	"openGemini-UP/pkg/config"
	"testing"
)

func TestYaml(t *testing.T) {
	yamlPath := "/Users/liujibo/Desktop/openGemini-UP/topology.example.yaml"

	var conf config.Yaml
	var err error
	conf, err = config.ReadFromYaml(yamlPath)

	if err != nil || conf.Global.SSHPort == 0 {
		t.Fatalf("Read from yaml failed: %v", err)
	}
}
