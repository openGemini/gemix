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

package util

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"time"
)

var DownloadDst string
var InstallScriptPath string
var ClusterInfoDir string

func init() {
	execPath, _ := os.Getwd()
	currentUser, err := user.Current()
	if err != nil {
		fmt.Println(err)
	}
	homeDirectory := currentUser.HomeDir
	DownloadDst = filepath.Join(homeDirectory, ".gemix", "download")
	InstallScriptPath = filepath.Join(execPath, "scripts/install.sh")
	ClusterInfoDir = filepath.Join(homeDirectory, ".gemix", "cluster-info")

	if err = os.MkdirAll(DownloadDst, 0750); err != nil {
		fmt.Println(err)
	}

	if err = os.MkdirAll(ClusterInfoDir, 0750); err != nil {
		fmt.Println(err)
	}
}

// env
const (
	SshEnvUser = "OPENGEMINI_SSH_USER"
	SshEnvKey  = "OPENGEMINI_SSH_KEY"
	SshEnvPW   = "OPENGEMINI_SSH_PW"
)

// downloader
const (
	DownloadWeb       = "https://github.com/openGemini/openGemini/releases/download"
	DownloadFillChar  = "openGemini-"
	DownloadPkgSuffix = ".tar.gz"

	// fixed values
	DownloadTimeout = 2 * time.Minute

	// default values
	DownloadLatestUrl   = "https://github.com/openGemini/openGemini/releases/latest"
	DownloadDefaultOs   = "linux"
	DownloadDefaultArch = "amd64"
)

// local
const (
	LocalBinRelPath = "usr/bin"
	LocalEtcRelPath = "etc"
	LocalConfName   = "openGemini.conf"
)

// config
const (
	RemoteMetaConfName  = "openGemini-meta.conf"
	RemoteSqlConfName   = "openGemini-sql.conf"
	RemoteStoreConfName = "openGemini-store.conf"
)

// file name
const (
	TsMeta        = "ts-meta" // process name & bin file name
	TsSql         = "ts-sql"
	TsStore       = "ts-store"
	InstallScript = "install.sh"
)

// remote
const (
	// gemix
	RemoteBinRelPath = "bin"
	RemoteEtcRelPath = "etc"

	// openGemini
	RemotePidPath   = "pid"
	RemotePidSuffix = ".pid"
	RemoteLogSuffix = ".log"

	MetaExtraLog  = "meta_extra"
	SqlExtraLog   = "sql_extra"
	StoreExtraLog = "store_extra"
)
