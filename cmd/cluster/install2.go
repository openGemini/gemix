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
	"github.com/openGemini/gemix/utils"
	"github.com/spf13/cobra"
)

//lint:ignore U1000 keep this
func installCmd2() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "install2",
		Short:        "Install an openGemini cluster for production",
		Long:         `Install an openGemini cluster for production. SSH connection will be used to deploy files, as well as create system users for running the service.`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ops, err := getClusterInstallOptions(cmd)
			if err != nil {
				return err
			}

			err = InstallCluster(ops)
			if err != nil {
				return err
			}

			// save cluster information
			err = utils.SaveClusterOptionsToFile(filepath.Join(utils.ClusterInfoDir, ops.Name), ops)
			return err
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			switch len(args) {
			case 2:
				return nil, cobra.ShellCompDirectiveDefault
			default:
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
		},
	}

	cmd.Flags().StringP("name", "n", "", "cluster name")
	cmd.Flags().StringP("version", "v", "", "component version")
	cmd.Flags().StringP("yaml", "y", "", "The path to cluster topology yaml file")
	cmd.Flags().StringP("user", "u", "", "The user name to login via SSH. The user must has root (or sudo) privilege.")
	cmd.Flags().StringP("key", "k", "", "The path of the SSH identity file. If specified, public key authentication will be used.")
	cmd.Flags().StringP("password", "p", "", "The password of target hosts. If specified, password authentication will be used.")
	// TODO: skip-create-user
	return cmd
}

func InstallCluster(ops utils.ClusterOptions) error {
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

func getClusterInstallOptions(cmd *cobra.Command) (utils.ClusterOptions, error) {
	var ops utils.ClusterOptions
	if name, _ := cmd.Flags().GetString("name"); name == "" {
		return ops, fmt.Errorf("the cluster name is required")
	} else if !utils.CheckClusterNameValid(name) {
		return ops, fmt.Errorf("the cluster name is not valid")
	} else {
		ops.Name = name
	}
	if version, _ := cmd.Flags().GetString("version"); version == "" {
		latestVer, err := utils.GetLatestVerFromCurl()
		if err != nil {
			return ops, err
		} else {
			ops.Version = latestVer
		}
	} else {
		ops.Version = version
	}
	if user, _ := cmd.Flags().GetString("user"); user == "" {
		has, value := GetEnv(utils.SshEnvUser)
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
		hasKey, key := GetEnv(utils.SshEnvKey)
		if hasKey {
			ops.Key = key
			ops.SshType = utils.SSH_KEY
		} else {
			hasPW, pw := GetEnv(utils.SshEnvPW)
			if hasPW {
				ops.Password = pw
				ops.SshType = utils.SSH_PW
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
			ops.SshType = utils.SSH_KEY
		} else {
			ops.SshType = utils.SSH_PW
		}
	}

	// TODO: check path is valid
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

func ReadClusterOptionsByName(cmd *cobra.Command) (utils.ClusterOptions, error) {
	var ops utils.ClusterOptions
	var err error
	if name, _ := cmd.Flags().GetString("name"); name == "" {
		return ops, fmt.Errorf("the cluster name is required")
	} else if !utils.CheckClusterNameExist(name) {
		return ops, fmt.Errorf("the cluster name is not existed, please install the cluster first")
	} else {
		ops, err = utils.LoadClusterOptionsFromFile(filepath.Join(utils.ClusterInfoDir, name))
		if err != nil {
			return ops, err
		}
	}
	return ops, err
}
