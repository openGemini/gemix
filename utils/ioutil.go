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

package utils

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

var (
	fileLocks = make(map[string]*sync.Mutex)
	filesLock = sync.Mutex{}
)

// IsExist check whether a path is existed
func IsExist(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// IsNotExist check whether a path is not existed
func IsNotExist(path string) bool {
	_, err := os.Stat(path)
	return os.IsNotExist(err)
}

func fileLock(path string) *sync.Mutex {
	filesLock.Lock()
	defer filesLock.Unlock()

	if _, ok := fileLocks[path]; !ok {
		fileLocks[path] = &sync.Mutex{}
	}

	return fileLocks[path]
}

// SaveFileWithBackup will backup the file before save it.
// e.g., backup meta.yaml as meta-2006-01-02T15:04:05Z07:00.yaml
// backup the files in the same dir of path if backupDir is empty.
func SaveFileWithBackup(path string, data []byte, backupDir string) error {
	fileLock(path).Lock()
	defer fileLock(path).Unlock()

	info, err := os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		return errors.WithStack(err)
	}

	if info != nil && info.IsDir() {
		return errors.Errorf("%s is directory", path)
	}

	// backup file
	if !os.IsNotExist(err) {
		base := filepath.Base(path)
		dir := filepath.Dir(path)

		var backupName string
		timestr := time.Now().Format(time.RFC3339Nano)
		p := strings.Split(base, ".")
		if len(p) == 1 {
			backupName = base + "-" + timestr
		} else {
			backupName = strings.Join(p[0:len(p)-1], ".") + "-" + timestr + "." + p[len(p)-1]
		}

		backupData, err := os.ReadFile(path)
		if err != nil {
			return errors.WithStack(err)
		}

		var backupPath string
		if backupDir != "" {
			backupPath = filepath.Join(backupDir, backupName)
		} else {
			backupPath = filepath.Join(dir, backupName)
		}
		err = os.WriteFile(backupPath, backupData, 0640)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	err = os.WriteFile(path, data, 0640)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
