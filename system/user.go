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
	"log"
	"os/exec"
	"os/user"
	"strings"

	"github.com/elotl/cloud-init/config"
)

func UserExists(u *config.User) bool {
	_, err := user.Lookup(u.Name)
	return err == nil
}

func CreateUser(u *config.User) error {
	args := []string{}

	if u.GECOS != "" {
		args = append(args, "-g", fmt.Sprintf("%q", u.GECOS))
	}

	if u.Homedir != "" {
		args = append(args, "-h", u.Homedir)
	}

	if u.NoCreateHome {
		args = append(args, "-H")
	}

	if u.PrimaryGroup != "" {
		args = append(args, "-G", u.PrimaryGroup)
	}

	if u.System {
		args = append(args, "-S")
	}
	// if u.PasswordHash != "" {
	// 	args = append(args, "--password", u.PasswordHash)
	// } else {
	// 	args = append(args, "--password", "*")
	// }
	// if u.NoUserGroup {
	// 	args = append(args, "--no-user-group")
	// }
	// if u.NoLogInit {
	// 	args = append(args, "--no-log-init")
	// }

	if u.Shell != "" {
		args = append(args, "-s", u.Shell)
	}

	args = append(args, "-D")
	args = append(args, u.Name)

	fmt.Println("adduser", args)
	output, err := exec.Command("adduser", args...).CombinedOutput()
	if err != nil {
		log.Printf("Command 'useradd %s' failed: %v\n%s", strings.Join(args, " "), err, output)
	}
	if len(u.Groups) > 0 {
		fmt.Println("got groups: ", u.Groups)
		for _, group := range u.Groups {
			args := []string{u.Name, group}
			output, err := exec.Command("adduser", args...).CombinedOutput()
			if err != nil {
				log.Printf("Command 'adduser %s' failed: %v\n%s", strings.Join(args, " "), err, output)
			}
		}
	}
	if u.PasswordHash != "" {
		err := SetUserPassword(u.Name, u.PasswordHash)
		if err != nil {
			log.Printf("Error setting password for %s: %v\n", u.Name, err)
		}
	}
	return err
}

func SetUserPassword(user, hash string) error {
	cmd := exec.Command("/usr/sbin/chpasswd", "-e")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		fmt.Println("Error in start")
		return err
	}

	arg := fmt.Sprintf("%s:%s", user, hash)
	_, err = stdin.Write([]byte(arg))
	if err != nil {
		fmt.Println("Error writing to pipe")
		return err
	}
	err = stdin.Close()
	if err != nil {
		fmt.Println("Error closing")
	}

	err = cmd.Wait()
	if err != nil {
		fmt.Println("Error writing to pipe")
		return err
	}

	return nil
}
