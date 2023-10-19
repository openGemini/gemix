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

package cluster

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/openGemini/gemix/pkg/cluster/manager"
	"github.com/openGemini/gemix/util"
	"github.com/spf13/cobra"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install an openGemini cluster for production",
	Long:  `Install an openGemini cluster for production. SSH connection will be used to deploy files, as well as create system users for running the service.`,
	Run: func(cmd *cobra.Command, args []string) {
		ops, err := getClusterInstallOptions(cmd)
		if err != nil {
			fmt.Println(err)
			fmt.Println(cmd.UsageString())
			return
		}

		err = InstallCluster(ops)
		if err != nil {
			fmt.Println(err)
			return
		}

		// save cluster infomation
		err = util.SaveClusterOptionsToFile(filepath.Join(util.ClusterInfoDir, ops.Name), ops)
		if err != nil {
			fmt.Println(err)
		}
	},
}

func InstallCluster(ops util.ClusterOptions) error {
	installer := manager.NewGeminiInstaller(ops)
	defer installer.Close()

	if err := installer.PrepareForInstall(); err != nil {
		return err
	}
	if err := installer.Install(); err != nil {
		return err
	}
	fmt.Printf("Successfully installed the openGemini cluster with version : %s\n", ops.Version)
	return nil
}

func init() {
	ClusterCmd.AddCommand(installCmd)
	installCmd.Flags().StringP("name", "n", "", "cluster name")
	installCmd.Flags().StringP("version", "v", "", "component version")
	installCmd.Flags().StringP("yaml", "y", "", "The path to cluster topology yaml file")
	installCmd.Flags().StringP("user", "u", "", "The user name to login via SSH. The user must has root (or sudo) privilege.")
	installCmd.Flags().StringP("key", "k", "", "The path of the SSH identity file. If specified, public key authentication will be used.")
	installCmd.Flags().StringP("password", "p", "", "The password of target hosts. If specified, password authentication will be used.")
}

func getClusterInstallOptions(cmd *cobra.Command) (util.ClusterOptions, error) {
	var ops util.ClusterOptions
	if name, _ := cmd.Flags().GetString("name"); name == "" {
		return ops, fmt.Errorf("the cluster name is required")
	} else if !util.CheckClusterNameValid(name) {
		return ops, fmt.Errorf("the cluster name is not valid")
	} else {
		ops.Name = name
	}
	if version, _ := cmd.Flags().GetString("version"); version == "" {
		latestVer, err := util.GetLatestVerFromCurl()
		if err != nil {
			return ops, err
		} else {
			ops.Version = latestVer
		}
	} else {
		ops.Version = version
	}
	if user, _ := cmd.Flags().GetString("user"); user == "" {
		has, value := GetEnv(util.SshEnvUser)
		if has {
			ops.User = value
		} else {
			return ops, fmt.Errorf("the user is required")
		}
	} else {
		ops.User = user
	}
	password, _ := cmd.Flags().GetString("password")
	key, _ := cmd.Flags().GetString("key")
	if password == "" && key == "" {
		hasKey, key := GetEnv(util.SshEnvKey)
		if hasKey {
			ops.Key = key
			ops.SshType = util.SSH_KEY
		} else {
			hasPW, pw := GetEnv(util.SshEnvPW)
			if hasPW {
				ops.Password = pw
				ops.SshType = util.SSH_PW
			} else {
				return ops, fmt.Errorf("the password and key need at least one")
			}
		}

	} else if password != "" && key != "" {
		return ops, fmt.Errorf("the password and key need only one")
	} else {
		ops.Key = key
		ops.Password = password
		if key != "" {
			ops.SshType = util.SSH_KEY
		} else {
			ops.SshType = util.SSH_PW
		}
	}

	if yPath, _ := cmd.Flags().GetString("yaml"); yPath == "" {
		return ops, fmt.Errorf("the path of cluster configuration file must be specified")
	} else {
		ops.YamlPath = yPath
	}
	return ops, nil
}

func GetEnv(envVar string) (bool, string) {
	value := os.Getenv(envVar)
	if value == "" {
		return false, value
	} else {
		return true, value
	}
}
