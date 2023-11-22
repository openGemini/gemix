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

package task

import (
	"context"
	"fmt"

	"github.com/openGemini/gemix/pkg/cluster/ctxt"
	"github.com/pkg/errors"
)

type UserActionOpt string

const (
	UserActionAdd UserActionOpt = "add"
	UserActionDel UserActionOpt = "del"
)

const (
	defaultShell = "/bin/bash"

	// TODO: in RHEL/CentOS, the commands are in /usr/sbin, but in some
	// other distros they may be in other location such as /usr/bin, we'll
	// need to check and find the proper path of commands in the future.
	userAddCmd  = "/usr/sbin/useradd"
	userDelCmd  = "/usr/sbin/userdel"
	groupAddCmd = "/usr/sbin/groupadd"
)

// UserAction is used to create user or del user on the target host
type UserAction struct {
	skipCreateUser bool
	host           string // target host
	userAction     UserActionOpt
	name           string // username
	group          string // group name
	home           string // home directory of user
	shell          string
	sudoer         bool // when true, the user will be added to sudoers list
}

// Execute implements the Task interface
func (u *UserAction) Execute(ctx context.Context) error {
	if u.skipCreateUser && u.userAction == UserActionAdd {
		return nil
	}
	exec, found := ctxt.GetInner(ctx).GetExecutor(u.host)
	if !found {
		panic(ErrNoExecutor)
	}

	cmd := ""
	switch u.userAction {
	case UserActionAdd:
		cmd = userAddCmd
		// You have to use -m, otherwise no home directory will be created. If you want to specify the path of the home directory, use -d and specify the path
		// useradd -m -d /PATH/TO/FOLDER
		cmd += " -m"
		if u.home != "" {
			cmd += " -d" + u.home
		}

		// set user's login shell
		if u.shell != "" {
			cmd = fmt.Sprintf("%s -s %s", cmd, u.shell)
		} else {
			cmd = fmt.Sprintf("%s -s %s", cmd, defaultShell)
		}

		// set user's group
		if u.group == "" {
			u.group = u.name
		}

		// groupadd -f <group-name>
		groupAdd := fmt.Sprintf("%s -f %s", groupAddCmd, u.group)

		// useradd -g <group-name> <user-name>
		cmd = fmt.Sprintf("%s -g %s %s", cmd, u.group, u.name)

		// prevent errors when username already in use
		cmd = fmt.Sprintf("id -u %s > /dev/null 2>&1 || (%s && %s)", u.name, groupAdd, cmd)

		// add user to sudoers list
		if u.sudoer {
			sudoLine := fmt.Sprintf("%s ALL=(ALL) NOPASSWD:ALL",
				u.name)
			cmd = fmt.Sprintf("%s && %s",
				cmd,
				fmt.Sprintf("echo '%s' > /etc/sudoers.d/%s", sudoLine, u.name))
		}

	case UserActionDel:
		cmd = fmt.Sprintf("%s -r %s", userDelCmd, u.name)
		// prevent errors when user does not exist
		cmd = fmt.Sprintf("%s || [ $? -eq 6 ]", cmd)
		//	case UserActionModify:
		//		cmd = usermodCmd
	}
	_, _, err := exec.Execute(ctx, cmd, true) // use root to create the dir
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Rollback implements the Task interface
func (u *UserAction) Rollback(ctx context.Context) error {
	return nil
}

// String implements the fmt.Stringer interface
func (u *UserAction) String() string {
	return fmt.Sprintf("UserAction: host=%s, user='%s', group='%s', opt='%s'", u.host, u.name, u.group, u.userAction)
}
