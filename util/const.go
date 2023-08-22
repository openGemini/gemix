package util

import "time"

// TODO:这里的大多数后面要改成从配置文件中读取，这里仅仅存储一个默认值

// downloader
const (
	Download_web       = "https://github.com/openGemini/openGemini/releases/download"
	Download_version   = "v1.0.0"
	Download_fill_char = "openGemini-"
	Download_type      = "-linux-amd64.tar.gz"
	Download_dst       = "./data"
	Download_timeout   = 2 * time.Minute
)

// local
const (
	Local_bin_rel_path = "/usr/bin/"
	Local_etc_rel_path = "/etc/"
)

// config
const (
	User_conf_path       = "./up.conf"
	Conf_gen_script_path = "./scripts/conf_gen.sh"
	Install_script_path  = "./scripts/install.sh"
	Remote_conf_suffix   = "-openGemini.conf"
	SSH_KEY              = "SSH_KEY"
	SSH_PW               = "SSH_PW"
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
	Remote_bin_rel_path = "bin/"
	Remote_etc_rel_path = "etc/"

	// openGemini
	OpenGemini_path   = "/tmp/openGemini"
	Remote_pid_path   = "/pid/"
	Remote_log_path   = "/logs/"
	Remote_pid_suffix = ".pid"
	Remote_log_suffix = ".log"

	META_extra_log  = "/meta_extra"
	SQL_extra_log   = "/sql_extra"
	STORE_extra_log = "/store_extra"
	META            = "meta"
	SQL             = "sql"
	STORE           = "store"
)
