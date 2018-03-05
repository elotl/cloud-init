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

package system

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	// DirPath is the path of the default SSH directory.
	DirPath = ".ssh"
	// AuthorizedKeysPath is the path of the default authorized_keys file for SSH.
	AuthorizedKeysPath = ".ssh/authorized_keys"
)

func SetupDirectory(dirname string, perms os.FileMode, userID int, groupID int) error {
	if err := os.MkdirAll(dirname, perms); err != nil {
		return err
	}

	if err := os.Chmod(dirname, perms); err != nil {
		return err
	}

	return os.Chown(dirname, userID, groupID)
}

func AuthorizeSSHKeys2(username string, keysName string, keys []string) error {
	u, err := user.Lookup(username)
	if err != nil {
		fmt.Printf("Could not set authorized keys for %s: %v\n", username, err)
		return err
	}
	sshdir := filepath.Join(u.HomeDir, DirPath)
	uid, _ := strconv.Atoi(u.Uid)
	gid, _ := strconv.Atoi(u.Gid)
	if err = SetupDirectory(sshdir, 0700, uid, gid); err != nil {
		fmt.Printf("Could not setup .ssh directory for %s: %v\n", username, err)
		return err
	}
	sshfile := filepath.Join(u.HomeDir, AuthorizedKeysPath)
	addTrailingNewline := false
	if _, err := os.Stat(sshfile); !os.IsNotExist(err) {
		// path to ssh file exists, get the contents
		contents, err := ioutil.ReadFile(sshfile)
		if err != nil {
			fmt.Println("Error reading sshfile at:", sshfile, err)
		} else if contents[len(contents)-1] != byte('\n') {
			addTrailingNewline = true
		}
	}

	// make the authorized keys file if it doesn't exist, chmod it
	f, err := os.OpenFile(sshfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Println("Error opening authorized keys file at", sshfile, err)
	}
	joined := fmt.Sprintf("%s\n", strings.Join(keys, "\n"))
	if addTrailingNewline {
		joined = "\n" + joined
	}
	if _, err := f.WriteString(joined); err != nil {
		return err
	}
	if err := f.Chown(uid, gid); err != nil {
		fmt.Println("Error setting rightful owner and group", sshfile, err)
	}

	return f.Close()
}

// Add the provide SSH public key to the core user's list of
// authorized keys
func AuthorizeSSHKeys(user string, keysName string, keys []string) error {
	for i, key := range keys {
		keys[i] = strings.TrimSpace(key)
	}

	// join all keys with newlines, ensuring the resulting string
	// also ends with a newline
	joined := fmt.Sprintf("%s\n", strings.Join(keys, "\n"))

	cmd := exec.Command("update-ssh-keys", "-u", user, "-a", keysName)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		stdin.Close()
		return err
	}

	_, err = io.WriteString(stdin, joined)
	if err != nil {
		return err
	}

	stdin.Close()
	stdoutBytes, _ := ioutil.ReadAll(stdout)
	stderrBytes, _ := ioutil.ReadAll(stderr)

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("Call to update-ssh-keys failed with %v: %s %s", err, string(stdoutBytes), string(stderrBytes))
	}

	return nil
}
