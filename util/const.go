package util

import "time"

// TODO:这里的大多数后面要改成从配置文件中读取，这里仅仅存储一个默认值

// downloader
const (
	Download_web        = "https://github.com/openGemini/openGemini/releases/download"
	Download_fill_char  = "openGemini-"
	Download_pkg_suffix = ".tar.gz"

	// fixed values
	Download_dst     = "./data"
	Download_timeout = 2 * time.Minute

	// default values
	Download_default_version = "v1.0.0"
	Download_default_os      = "linux"
	Download_default_arch    = "amd64"
)

// local
const (
	Local_bin_rel_path = "usr/bin"
	Local_etc_rel_path = "etc"
	Local_conf_name    = "openGemini.conf"
)

// config
const (
	Install_script_path = "./scripts/install.sh"
	Remote_conf_suffix  = "-openGemini.conf"
)

// file name
const (
	TS_META        = "ts-meta" // process name & bin file name
	TS_SQL         = "ts-sql"
	TS_STORE       = "ts-store"
	Install_Script = "install.sh"
)

// remote
const (
	// openGemini-UP
	Remote_bin_rel_path = "bin"
	Remote_etc_rel_path = "etc"

	// openGemini
	Remote_pid_path   = "pid"
	Remote_pid_suffix = ".pid"
	Remote_log_suffix = ".log"

	META_extra_log  = "meta_extra"
	SQL_extra_log   = "sql_extra"
	STORE_extra_log = "store_extra"
	META            = "meta"
	SQL             = "sql"
	STORE           = "store"
)

// version
const (
	VersionFile = "version"
)
