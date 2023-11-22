package operation

// Options represents the operation options
type Options struct {
	Roles               []string
	Nodes               []string
	Force               bool   // Option for upgrade/tls subcommand
	SSHTimeout          uint64 // timeout in seconds when connecting an SSH server
	OptTimeout          uint64 // timeout in seconds for operations that support it, not to confuse with SSH timeout
	APITimeout          uint64 // timeout in seconds for API operations that support it, like transferring store leader
	IgnoreConfigCheck   bool   // should we ignore the config check result after init config
	Concurrency         int    // max number of parallel tasks to run
	SSHProxyHost        string // the ssh proxy host
	SSHProxyPort        int    // the ssh proxy port
	SSHProxyUser        string // the ssh proxy user
	SSHProxyIdentity    string // the ssh proxy identity file
	SSHProxyUsePassword bool   // use password instead of identity file for ssh proxy connection
	SSHProxyTimeout     uint64 // timeout in seconds when connecting the proxy host
	// SSHCustomScripts    SSHCustomScripts // custom scripts to be executed during the operation

	// What type of things should we cleanup in clean command
	CleanupData     bool // should we cleanup data
	CleanupLog      bool // should we clenaup log
	CleanupAuditLog bool // should we clenaup tidb server auit log

	// Some data will be retained when destroying instances
	RetainDataRoles []string
	RetainDataNodes []string

	DisplayMode string // the output format
	// Operation   Operation
}
