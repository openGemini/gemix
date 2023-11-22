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

package gui

// SSHConnectionProps is SSHConnectionProps
type SSHConnectionProps struct {
	Password               string
	IdentityFile           string
	IdentityFilePassphrase string
}

// ReadIdentityFileOrPassword is ReadIdentityFileOrPassword
func ReadIdentityFileOrPassword(identityFilePath string, usePass bool) (*SSHConnectionProps, error) {
	// If identity file is not specified, prompt to read password
	if usePass {
		password := PromptForPassword("Input SSH password: ")
		return &SSHConnectionProps{
			Password: password,
		}, nil
	}

	return &SSHConnectionProps{}, nil
}
