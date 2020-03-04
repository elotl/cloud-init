// Copyright 2015 CoreOS, Inc.
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

package validate

import (
	"reflect"
	"testing"
)

func TestCheckDiscoveryUrl(t *testing.T) {
	tests := []struct {
		config string

		entries []Entry
	}{
		{},
		{
			config: "coreos:\n  etcd:\n    discovery: https://discovery.etcd.io/00000000000000000000000000000000",
		},
		{
			config: "coreos:\n  etcd:\n    discovery: http://custom.domain/mytoken",
		},
		{
			config:  "coreos:\n  etcd:\n    discovery: disco",
			entries: []Entry{{entryWarning, "discovery URL is not valid", 3}},
		},
	}

	for i, tt := range tests {
		r := Report{}
		n, err := parseCloudConfig([]byte(tt.config), &r)
		if err != nil {
			panic(err)
		}
		checkDiscoveryUrl(n, &r)

		if e := r.Entries(); !reflect.DeepEqual(tt.entries, e) {
			t.Errorf("bad report (%d, %q): want %#v, got %#v", i, tt.config, tt.entries, e)
		}
	}
}

func TestCheckEncoding(t *testing.T) {
	tests := []struct {
		config string

		entries []Entry
	}{
		{},
		{
			config: "write_files:\n  - encoding: base64\n    content: aGVsbG8K",
		},
		{
			config: "write_files:\n  - content: !!binary aGVsbG8K",
		},
		{
			config:  "write_files:\n  - encoding: base64\n    content: !!binary aGVsbG8K",
			entries: []Entry{{entryError, `content cannot be decoded as "base64"`, 3}},
		},
		{
			config: "write_files:\n  - encoding: base64\n    content: !!binary YUdWc2JHOEsK",
		},
		{
			config: "write_files:\n  - encoding: gzip\n    content: !!binary H4sIAOC3tVQAA8tIzcnJ5wIAIDA6NgYAAAA=",
		},
		{
			config: "write_files:\n  - encoding: gzip+base64\n    content: H4sIAOC3tVQAA8tIzcnJ5wIAIDA6NgYAAAA=",
		},
		{
			config:  "write_files:\n  - encoding: custom\n    content: hello",
			entries: []Entry{{entryError, `content cannot be decoded as "custom"`, 3}},
		},
	}

	for i, tt := range tests {
		r := Report{}
		n, err := parseCloudConfig([]byte(tt.config), &r)
		if err != nil {
			panic(err)
		}
		checkEncoding(n, &r)

		if e := r.Entries(); !reflect.DeepEqual(tt.entries, e) {
			t.Errorf("bad report (%d, %q): want %#v, got %#v", i, tt.config, tt.entries, e)
		}
	}
}

func TestCheckStructure(t *testing.T) {
	tests := []struct {
		config string

		entries []Entry
	}{
		{},

		// Test for unrecognized keys
		{
			config:  "test:",
			entries: []Entry{{entryWarning, "unrecognized key \"test\"", 1}},
		},

		// Test for error on list of nodes
		// Want string
		{
			config: "hostname: true",
		},
		{
			config: "hostname: 4",
		},
		{
			config: "hostname: host",
		},
		{
			config:  "hostname:\n  name:",
			entries: []Entry{{entryWarning, "incorrect type for \"hostname\" (want string)", 1}},
		},
		{
			config:  "hostname:\n  - name",
			entries: []Entry{{entryWarning, "incorrect type for \"hostname\" (want string)", 1}},
		},
		{
			config:  "ssh_authorized_keys: true",
			entries: []Entry{{entryWarning, "incorrect type for \"ssh_authorized_keys\" (want []string)", 1}},
		},
		{
			config:  "ssh_authorized_keys: 4",
			entries: []Entry{{entryWarning, "incorrect type for \"ssh_authorized_keys\" (want []string)", 1}},
		},
		{
			config:  "ssh_authorized_keys: key",
			entries: []Entry{{entryWarning, "incorrect type for \"ssh_authorized_keys\" (want []string)", 1}},
		},
		{
			config:  "ssh_authorized_keys:\n  key: value",
			entries: []Entry{{entryWarning, "incorrect type for \"ssh_authorized_keys\" (want []string)", 1}},
		},
		{
			config: "ssh_authorized_keys:\n  - key",
		},
		{
			config:  "ssh_authorized_keys:\n  - key: value",
			entries: []Entry{{entryWarning, "incorrect type for \"ssh_authorized_keys[0]\" (want string)", 2}},
		},
		// Want []struct
		{
			config:  "users:\n  true",
			entries: []Entry{{entryWarning, "incorrect type for \"users\" (want []struct)", 1}},
		},
		{
			config:  "users:\n  4",
			entries: []Entry{{entryWarning, "incorrect type for \"users\" (want []struct)", 1}},
		},
		{
			config:  "users:\n  bad",
			entries: []Entry{{entryWarning, "incorrect type for \"users\" (want []struct)", 1}},
		},
		{
			config:  "users:\n  bad:",
			entries: []Entry{{entryWarning, "incorrect type for \"users\" (want []struct)", 1}},
		},
		{
			config: "users:\n  - name: good",
		},
		// Want struct within array
		{
			config:  "users:\n  - true",
			entries: []Entry{{entryWarning, "incorrect type for \"users[0]\" (want struct)", 2}},
		},
		{
			config:  "users:\n  - name: hi\n  - true",
			entries: []Entry{{entryWarning, "incorrect type for \"users[1]\" (want struct)", 3}},
		},
		{
			config:  "users:\n  - 4",
			entries: []Entry{{entryWarning, "incorrect type for \"users[0]\" (want struct)", 2}},
		},
		{
			config:  "users:\n  - bad",
			entries: []Entry{{entryWarning, "incorrect type for \"users[0]\" (want struct)", 2}},
		},
		{
			config:  "users:\n  - - bad",
			entries: []Entry{{entryWarning, "incorrect type for \"users[0]\" (want struct)", 2}},
		},
	}

	for i, tt := range tests {
		r := Report{}
		n, err := parseCloudConfig([]byte(tt.config), &r)
		if err != nil {
			panic(err)
		}
		checkStructure(n, &r)

		if e := r.Entries(); !reflect.DeepEqual(tt.entries, e) {
			t.Errorf("bad report (%d, %q): want %#v, got %#v", i, tt.config, tt.entries, e)
		}
	}
}

func TestCheckValidity(t *testing.T) {
	tests := []struct {
		config string

		entries []Entry
	}{
		// string
		{
			config: "hostname: test",
		},

		// unknown
		{
			config: "unknown: hi",
		},
	}

	for i, tt := range tests {
		r := Report{}
		n, err := parseCloudConfig([]byte(tt.config), &r)
		if err != nil {
			panic(err)
		}
		checkValidity(n, &r)

		if e := r.Entries(); !reflect.DeepEqual(tt.entries, e) {
			t.Errorf("bad report (%d, %q): want %#v, got %#v", i, tt.config, tt.entries, e)
		}
	}
}

func TestCheckWriteFiles(t *testing.T) {
	tests := []struct {
		config string

		entries []Entry
	}{
		{},
		{
			config: "write_files:\n  - path: /valid",
		},
		{
			config: "write_files:\n  - path: /tmp/usr/valid",
		},
	}

	for i, tt := range tests {
		r := Report{}
		n, err := parseCloudConfig([]byte(tt.config), &r)
		if err != nil {
			panic(err)
		}
		checkWriteFiles(n, &r)

		if e := r.Entries(); !reflect.DeepEqual(tt.entries, e) {
			t.Errorf("bad report (%d, %q): want %#v, got %#v", i, tt.config, tt.entries, e)
		}
	}
}

func TestCheckWriteFilesUnderCoreos(t *testing.T) {
	tests := []struct {
		config string

		entries []Entry
	}{
		{},
		{
			config: "write_files:\n  - path: /hi",
		},
		{
			config:  "coreos:\n  write_files:\n    - path: /hi",
			entries: []Entry{{entryInfo, "write_files doesn't belong under coreos", 2}},
		},
		{
			config:  "coreos:\n  write-files:\n    - path: /hyphen",
			entries: []Entry{{entryInfo, "write_files doesn't belong under coreos", 2}},
		},
	}

	for i, tt := range tests {
		r := Report{}
		n, err := parseCloudConfig([]byte(tt.config), &r)
		if err != nil {
			panic(err)
		}
		checkWriteFilesUnderCoreos(n, &r)

		if e := r.Entries(); !reflect.DeepEqual(tt.entries, e) {
			t.Errorf("bad report (%d, %q): want %#v, got %#v", i, tt.config, tt.entries, e)
		}
	}
}
