// Copyright 2020 PingCAP, Inc.
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

// DefaultGemixHome represents the default home directory for this build of gemix
// If this is left empty, the default will be thee combination of the running
// user's home directory and ProfileDirName
var DefaultGemixHome string

// ProfileDirName is the name of the profile directory to be used
var ProfileDirName = ".gemix"

// Notice: if you try to add a new env name which is notable by the user, shou should
// add it to cmd/env.go:envList so that the command `gemix env` will show that env.
const (
	// ComponentParentDir represent the parent directory of all downloaded components
	ComponentParentDir = "components"

	// ManifestParentDir represent the parent directory of all manifests
	ManifestParentDir = "manifests"

	// KeyInfoParentDir represent the parent directory of all keys
	KeyInfoParentDir = "keys"

	// DefaultPrivateKeyName represents the default private key file stored in ${GEMIX_HOME}/keys
	DefaultPrivateKeyName = "private.json"

	// DataParentDir represent the parent directory of all running instances
	DataParentDir = "data"

	// TelemetryDir represent the parent directory of telemetry info
	TelemetryDir = "telemetry"

	// StorageParentDir represent the parent directory of running component
	StorageParentDir = "storage"

	// EnvNameInstanceDataDir represents the working directory of specific instance
	EnvNameInstanceDataDir = "GEMIX_INSTANCE_DATA_DIR"

	// EnvNameComponentDataDir represents the working directory of specific component
	EnvNameComponentDataDir = "GEMIX_COMPONENT_DATA_DIR"

	// EnvNameComponentInstallDir represents the install directory of specific component
	EnvNameComponentInstallDir = "GEMIX_COMPONENT_INSTALL_DIR"

	// EnvNameWorkDir represents the work directory of Gemix where user type the command `gemix xxx`
	EnvNameWorkDir = "GEMIX_WORK_DIR"

	// EnvNameUserInputVersion represents the version user specified when running a component by `gemix component:version`
	EnvNameUserInputVersion = "GEMIX_USER_INPUT_VERSION"

	// EnvNameGemixVersion represents the version of Gemix itself, not the version of component
	EnvNameGemixVersion = "GEMIX_VERSION"

	// EnvNameHome represents the environment name of gemix home directory
	EnvNameHome = "GEMIX_HOME"

	// EnvNameTelemetryStatus represents the environment name of gemix telemetry status
	EnvNameTelemetryStatus = "GEMIX_TELEMETRY_STATUS"

	// EnvNameTelemetryUUID represents the environment name of gemix telemetry uuid
	EnvNameTelemetryUUID = "GEMIX_TELEMETRY_UUID"

	// EnvNameTelemetryEventUUID represents the environment name of gemix telemetry event uuid
	EnvNameTelemetryEventUUID = "GEMIX_TELEMETRY_EVENT_UUID"

	// EnvNameTelemetrySecret represents the environment name of gemix telemetry secret
	EnvNameTelemetrySecret = "GEMIX_TELEMETRY_SECRET"

	// EnvTag is the tag of the running component
	EnvTag = "GEMIX_TAG"

	// EnvNameSSHPassPrompt is the variable name by which user specific the password prompt for sshpass
	EnvNameSSHPassPrompt = "GEMIX_SSHPASS_PROMPT"

	// EnvNameNativeSSHClient is the variable name by which user can specific use native ssh client or not
	EnvNameNativeSSHClient = "GEMIX_NATIVE_SSH"

	// EnvNameSSHPath is the variable name by which user can specific the executable ssh binary path
	EnvNameSSHPath = "GEMIX_SSH_PATH"

	// EnvNameSCPPath is the variable name by which user can specific the executable scp binary path
	EnvNameSCPPath = "GEMIX_SCP_PATH"

	// EnvNameKeepSourceTarget is the variable name by which user can keep the source target or not
	EnvNameKeepSourceTarget = "GEMIX_KEEP_SOURCE_TARGET"

	// EnvNameMirrorSyncScript make it possible for user to sync mirror commit to other place (eg. CDN)
	EnvNameMirrorSyncScript = "GEMIX_MIRROR_SYNC_SCRIPT"

	// EnvNameLogPath is the variable name by which user can write the log files into
	EnvNameLogPath = "GEMIX_LOG_PATH"

	// EnvNameDebug is the variable name by which user can set gemix runs in debug mode(eg. print panic logs)
	EnvNameDebug = "GEMIX_CLUSTER_DEBUG"

	// MetaFilename represents the process meta file name
	MetaFilename = "gemix_process_meta"
)
