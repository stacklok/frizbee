//
// Copyright 2023 Stacklok, Inc.
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

// Package constants provides constants for the frizbee utilities.
package constants

import (
	"runtime/debug"
	"strings"
	"text/template"
)

const (
	// UserAgent is the user agent string used by frizbee.
	//
	// TODO (jaosorior): Add version information to this.
	UserAgent = "frizbee"
)

var (
	// CLIVersion is the version of the frizbee CLI.
	// nolint: gochecknoglobals
	CLIVersion = "dev"
	// VerboseCLIVersion is the verbose version of the frizbee CLI.
	// nolint: gochecknoglobals
	VerboseCLIVersion = ""
)

type versionInfo struct {
	Version   string
	GoVersion string
	Time      string
	Commit    string
	OS        string
	Arch      string
	Modified  bool
}

const (
	verboseTemplate = `Version: {{ .Version }}
Go Version: {{.GoVersion}}
Git Commit: {{.Commit}}
Commit Date: {{.Time}}
OS/Arch: {{.OS}}/{{.Arch}}
Dirty: {{.Modified}}
`
)

// nolint:gochecknoinits
func init() {
	buildinfo, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}

	var vinfo versionInfo
	vinfo.Version = CLIVersion
	vinfo.GoVersion = buildinfo.GoVersion

	for _, kv := range buildinfo.Settings {
		switch kv.Key {
		case "vcs.time":
			vinfo.Time = kv.Value
		case "vcs.revision":
			vinfo.Commit = kv.Value
		case "vcs.modified":
			vinfo.Modified = kv.Value == "true"
		case "GOOS":
			vinfo.OS = kv.Value
		case "GOARCH":
			vinfo.Arch = kv.Value
		}
	}
	VerboseCLIVersion = vinfo.String()
}

func (vvs *versionInfo) String() string {
	stringBuilder := &strings.Builder{}
	tmpl := template.Must(template.New("version").Parse(verboseTemplate))
	err := tmpl.Execute(stringBuilder, vvs)
	if err != nil {
		panic(err)
	}
	return stringBuilder.String()
}
