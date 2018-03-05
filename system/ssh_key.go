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
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	// DirPath is the path of the default SSH directory.
	//DirPath = ".ssh"
	// AuthorizedKeysPath is the path of the default authorized_keys file for SSH.
	AuthorizedKeysPath = ".ssh/authorized_keys"
)

type SSHAuthorizer struct {
	HomeDir string
	Uid     int
	Gid     int
	//KeysName string
	Keys []string
}

func (ssh *SSHAuthorizer) SetupSSHDirectory() error {
	sshdir := filepath.Join(ssh.HomeDir, ".ssh")
	perms := os.FileMode(0700)
	if err := os.MkdirAll(sshdir, perms); err != nil {
		return err
	}
	if err := os.Chmod(sshdir, perms); err != nil {
		return err
	}

	return os.Chown(sshdir, ssh.Uid, ssh.Gid)
}

func GetAuthorizedKeysContents(sshfile string) (string, error) {
	var contents string
	byteContents, err := ioutil.ReadFile(sshfile)
	if os.IsNotExist(err) {
		return "", nil
	} else if err != nil {
		err := fmt.Errorf("Error reading sshfile %s: %v", sshfile, err)
		return "", err
	}
	// add a trailing slash
	contents = string(byteContents)
	if contents[len(contents)-1] != '\n' {
		contents += "\n"
	}
	return contents, nil
}

func (ssh *SSHAuthorizer) Authorize(keys []string) error {
	if err := ssh.SetupSSHDirectory(); err != nil {
		return fmt.Errorf("Could not setup .ssh directory for uid %d: %v\n",
			ssh.Uid, err)
	}

	sshfile := filepath.Join(ssh.HomeDir, AuthorizedKeysPath)
	contents, err := GetAuthorizedKeysContents(sshfile)
	if err != nil {
		return fmt.Errorf("Could not get contents of authorized_keys for uid %d: %v\n", ssh.Uid, err)
	}

	joined := fmt.Sprintf("%s\n", strings.Join(keys, "\n"))
	contents += joined
	err = ioutil.WriteFile(sshfile, []byte(contents), 0600)
	if err != nil {
		return fmt.Errorf("Could not write authorized_keys for uid %d: %v\n", ssh.Uid, err)
	}
	if err := os.Chown(sshfile, ssh.Uid, ssh.Gid); err != nil {
		return fmt.Errorf("Error setting rightful owner and group of %s: %v",
			sshfile, err)
	}
	return nil
}

func AuthorizeSSHKeys(username string, keys []string) error {
	u, err := user.Lookup(username)
	if err != nil {

		fmt.Printf("Could not set authorized keys for %s: %v\n", username, err)
		return err
	}
	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		return fmt.Errorf("Invalid uid returned from user.Lookup(%s): %v", username, err)
	}
	gid, err := strconv.Atoi(u.Gid)
	if err != nil {
		return fmt.Errorf("Invalid uid returned from user.Lookup(%s): %v", username, err)
	}
	authorizer := SSHAuthorizer{
		HomeDir: u.HomeDir,
		Uid:     uid,
		Gid:     gid,
		//Keys:    keys,
	}
	err = authorizer.Authorize(keys)
	if err != nil {
		return fmt.Errorf("Error setting up ssh authorized keys for %s: %v",
			username, err)
	}
	return nil

	// sshdir := filepath.Join(u.HomeDir, DirPath)
	// if err = SetupSSHDirectory(sshdir, 0700, uid, gid); err != nil {
	// 	fmt.Printf("Could not setup .ssh directory for %s: %v\n", username, err)
	// 	return err
	// }
	// sshfile := filepath.Join(u.HomeDir, AuthorizedKeysPath)
	// addTrailingNewline := false
	// if _, err := os.Stat(sshfile); !os.IsNotExist(err) {
	// 	// path to ssh file exists, get the contents
	// 	contents, err := ioutil.ReadFile(sshfile)
	// 	if err != nil {
	// 		fmt.Println("Error reading sshfile at:", sshfile, err)
	// 	} else if contents[len(contents)-1] != byte('\n') {
	// 		addTrailingNewline = true
	// 	}
	// }

	// // make the authorized keys file if it doesn't exist, chmod it
	// f, err := os.OpenFile(sshfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	// if err != nil {
	// 	fmt.Println("Error opening authorized keys file at", sshfile, err)
	// }
	// joined := fmt.Sprintf("%s\n", strings.Join(keys, "\n"))
	// if addTrailingNewline {
	// 	joined = "\n" + joined
	// }
	// if _, err := f.WriteString(joined); err != nil {
	// 	return err
	// }

	//return f.Close()
}

// // Add the provide SSH public key to the core user's list of
// // authorized keys
// func AuthorizeSSHKeys(user string, keysName string, keys []string) error {
// 	for i, key := range keys {
// 		keys[i] = strings.TrimSpace(key)
// 	}

// 	// join all keys with newlines, ensuring the resulting string
// 	// also ends with a newline
// 	joined := fmt.Sprintf("%s\n", strings.Join(keys, "\n"))

// 	cmd := exec.Command("update-ssh-keys", "-u", user, "-a", keysName)
// 	stdin, err := cmd.StdinPipe()
// 	if err != nil {
// 		return err
// 	}

// 	stdout, err := cmd.StdoutPipe()
// 	if err != nil {
// 		return err
// 	}

// 	stderr, err := cmd.StderrPipe()
// 	if err != nil {
// 		return err
// 	}

// 	err = cmd.Start()
// 	if err != nil {
// 		stdin.Close()
// 		return err
// 	}

// 	_, err = io.WriteString(stdin, joined)
// 	if err != nil {
// 		return err
// 	}

// 	stdin.Close()
// 	stdoutBytes, _ := ioutil.ReadAll(stdout)
// 	stderrBytes, _ := ioutil.ReadAll(stderr)

// 	err = cmd.Wait()
// 	if err != nil {
// 		return fmt.Errorf("Call to update-ssh-keys failed with %v: %s %s", err, string(stdoutBytes), string(stderrBytes))
// 	}

// 	return nil
// }
