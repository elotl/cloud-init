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

package file

import (
	"io/ioutil"
	"os"

	"github.com/elotl/cloud-init/datasource"
)

type localFile struct {
	path string
}

func NewDatasource(path string) *localFile {
	return &localFile{path}
}

func (f *localFile) IsAvailable() bool {
	_, err := os.Stat(f.path)
	return !os.IsNotExist(err)
}

func (f *localFile) AvailabilityChanges() bool {
	return true
}

func (f *localFile) ConfigRoot() string {
	return ""
}

func (f *localFile) FetchMetadata() (datasource.Metadata, error) {
	return datasource.Metadata{}, nil
}

func (f *localFile) FetchUserdata() ([]byte, error) {
	return ioutil.ReadFile(f.path)
}

func (f *localFile) Type() string {
	return "local-file"
}
