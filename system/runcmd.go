package system

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

func RunScript(script string) error {
	tempfile, err := ioutil.TempFile("", "cloud-init-script")
	if err != nil {
		return fmt.Errorf("could not create runcmd script: %v", err)
	}
	defer os.Remove(tempfile.Name())
	tempfile.Write([]byte(script))
	tempfile.Close()
	output, err := exec.Command("/bin/sh", tempfile.Name()).CombinedOutput()
	if err != nil {
		return fmt.Errorf("error executing runcmd script: %v", err)
	}
	log.Println("Successfully ran runcmd script, output was", string(output))
	return nil
}
