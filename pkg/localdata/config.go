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

package localdata

import (
	"os"
	"path"

	"github.com/BurntSushi/toml"
	"github.com/openGemini/gemix/pkg/utils"
)

type configBase struct {
	file string
}

// GemixConfig represent the config file of Gemix
type GemixConfig struct {
	configBase
	Mirror string `toml:"mirror"`
}

// InitConfig returns a GemixConfig struct which can flush config back to disk
func InitConfig(root string) (*GemixConfig, error) {
	config := GemixConfig{configBase{path.Join(root, "gemix.toml")}, ""}
	if utils.IsNotExist(config.file) {
		return &config, nil
	}
	// We can ignore any error at current
	// If we have more configs in the future, we should check the error
	if _, err := toml.DecodeFile(config.file, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// Flush config to disk
func (c *GemixConfig) Flush() error {
	f, err := os.OpenFile(c.file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0664)
	if err != nil {
		return err
	}
	defer f.Close()

	return toml.NewEncoder(f).Encode(c)
}
