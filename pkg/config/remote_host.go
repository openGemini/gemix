package config

type SSHType int32

const (
	SSH_UNKNOW SSHType = 0
	SSH_PW     SSHType = 1
	SSH_KEY    SSHType = 2
)

// used by deploy, exe, stop .etc
type RemoteHost struct {
	Ip         string
	SSHPort    int
	User       string
	Password   string
	KeyPath    string
	Typ        SSHType
	UpDataPath string
	LogPath    string
}

type UploadInfo struct {
	LocalPath  string
	RemotePath string
	FileName   string
}
