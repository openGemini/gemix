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
	"fmt"
	"strings"

	"golang.org/x/mod/semver"
)

// NightlyVersionAlias represents latest build of master branch.
const NightlyVersionAlias = "nightly"

// LatestVersionAlias represents the latest build (excluding nightly versions).
const LatestVersionAlias = "latest"

// FmtVer converts a version string to SemVer format, if the string is not a valid
// SemVer and fails to parse and convert it, an error is raised.
func FmtVer(ver string) (string, error) {
	v := ver

	// latest version is an alias
	if strings.ToLower(v) == LatestVersionAlias {
		return v, nil
	}

	if !strings.HasPrefix(ver, "v") {
		v = fmt.Sprintf("v%s", ver)
	}
	if !semver.IsValid(v) {
		return v, fmt.Errorf("version %s is not a valid SemVer string", ver)
	}
	return v, nil
}

type (
	// Version represents a version string, like: v3.1.2
	Version string
)

// IsValid checks whether is the version string valid
func (v Version) IsValid() bool {
	return v != "" && semver.IsValid(string(v))
}

// IsEmpty returns true if the `Version` is a empty string
func (v Version) IsEmpty() bool {
	return v == ""
}

// IsNightly returns true if the version is nightly
func (v Version) IsNightly() bool {
	return strings.Contains(string(v), NightlyVersionAlias)
}

// String implements the fmt.Stringer interface
func (v Version) String() string {
	return string(v)
}
