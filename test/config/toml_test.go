package test

import (
	"openGemini-UP/pkg/config"
	"testing"
)

func TestToml(t *testing.T) {
	tomlPath := "/Users/liujibo/Desktop/openGemini-UP/data/v1.0.0/etc/openGemini.conf"
	newPath := "/Users/liujibo/Desktop/openGemini-UP/test.conf"

	var conf config.Toml
	var err error
	conf, err = config.ReadFromToml(tomlPath)

	if err != nil {
		t.Fatalf("Read from toml failed: %v", err)
	}

	config.GenNewToml(conf, newPath)
}
