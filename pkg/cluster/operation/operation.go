package operation

import (
	"github.com/openGemini/gemix/pkg/cluster/spec"
	"github.com/openGemini/gemix/pkg/set"
)

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
	CleanupAuditLog bool // should we clenaup openGemini server auit log

	// Some data will be retained when destroying instances
	RetainDataRoles []string
	RetainDataNodes []string

	DisplayMode string // the output format
	// Operation   Operation
}

// FilterComponent filter components by set
func FilterComponent(comps []spec.Component, components set.StringSet) (res []spec.Component) {
	if len(components) == 0 {
		res = comps
		return
	}

	for _, c := range comps {
		role := c.Name()
		if !components.Exist(role) {
			continue
		}

		res = append(res, c)
	}
	return
}

// FilterInstance filter instances by set
func FilterInstance(instances []spec.Instance, nodes set.StringSet) (res []spec.Instance) {
	if len(nodes) == 0 {
		res = instances
		return
	}

	for _, c := range instances {
		if !nodes.Exist(c.ID()) {
			continue
		}
		res = append(res, c)
	}

	return
}
