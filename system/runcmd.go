package system

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

func RunScript(script string) error {
	tempfile, err := ioutil.TempFile("", "cloud-init-script")
	if err != nil {
		return fmt.Errorf("")
	}
	defer os.Remove(tempfile.Name())
	tempfile.Write([]byte(script))
	tempfile.Close()
	output, err := exec.Command("/bin/ash", tempfile.Name()).CombinedOutput()
	if err != nil {
		return fmt.Errorf("")
	}
	fmt.Println("Ran script, the output was", string(output))
	return nil
}
