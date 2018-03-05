package system

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var randLetters = []rune("abcdefghijklmnopqrstuvwxyz")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = randLetters[rand.Intn(len(randLetters))]
	}
	return string(b)
}

func makeTempFile() (string, func()) {
	tempFile, err := ioutil.TempFile("", "cloud-init-tf")
	if err != nil {
		panic(err)
	}
	return tempFile.Name(), func() { os.Remove(tempFile.Name()) }
}

// test out getting the contents of a file
// file doesn't exist
// file exists
// file exists and lacks newline

func TestAuthorizedKeyContentsExists(t *testing.T) {
	filename, closer := makeTempFile()
	defer closer()
	realContents := "key1\nkey2\n"
	err := ioutil.WriteFile(filename, []byte(realContents), 0660)
	assert.NoError(t, err)
	contents, err := GetAuthorizedKeysContents(filename)
	assert.NoError(t, err)
	assert.Equal(t, realContents, contents)
}

func TestAuthorizedKeyContentsExistsNoNewline(t *testing.T) {
	filename, closer := makeTempFile()
	defer closer()
	realContents := "key1\nkey2"
	err := ioutil.WriteFile(filename, []byte(realContents), 0660)
	assert.NoError(t, err)
	contents, err := GetAuthorizedKeysContents(filename)
	assert.NoError(t, err)
	assert.Equal(t, realContents+"\n", contents)
}

func TestAuthorizedKeyContentsNoExists(t *testing.T) {
	filepath := "/tmp/this_file_doesnt_exist"
	contents, err := GetAuthorizedKeysContents(filepath)
	assert.NoError(t, err)
	assert.Empty(t, contents)
}

func cleanupHomedir(t *testing.T, homedir string) {
	sshdir := filepath.Join(homedir, ".ssh")
	authFile := filepath.Join(homedir, ".ssh/authorized_keys")
	err := os.Remove(authFile)
	assert.NoError(t, err)
	err = os.Remove(sshdir)
	assert.NoError(t, err)
	err = os.Remove(homedir)
	assert.NoError(t, err)
}

func TestSSHAuhorizer(t *testing.T) {
	// create a unique username
	u, err := user.Current()
	assert.NoError(t, err)
	uid, err := strconv.Atoi(u.Uid)
	assert.NoError(t, err)
	gid, err := strconv.Atoi(u.Gid)
	assert.NoError(t, err)
	keys := []string{
		`ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDULTftpWMj4nD+7Ps
B8itam2T6Aqm9Z+ursQG1SRiK4ie5rHGJoteGnbH91Uix/HDE5GC3Hz
ICQVOnQay4hwJUKRfEUEWj1Sncer/BL2igDquABlcXNl2dgOlfJ8a3q
6IZnQpdEe6Vrqg/Ui082UxuZ08pNV94M/5IhR2fx0EbY66PQ97o+ywH
sB7oXDO8p/+mGL+h7cxFY7hILXTa5/3TGBEgcA65Rrmq22eiRt97RGh
DjfzIqTqb8gwuhTSNN7FWDLrEyRwJMbaTgDSoMIZdLtndVrGEqFHUO+
WzinSiEQCs2MDDnTk29bleHAEktu1x68GYhg9S7O/gZq8/swAV
core@valid1`,
		`ssh-dss AAAAB3NzaC1kc3MAAACBAJA94Sqw80BSKjVTNZD6570nXIN
hP8R2UhbBuydT+GI6CfA9Dw7O0udJQUfrqARFcRQR/syc72CO6jaKNE
3/A5E+8uVmRZt7s9VtA47s1qxqHswth74m1Nb86n2OTB0HcW63FsXo2
cJF+r+l6F3IcRPi4z/eaEKG7uhAS59TjH2tAAAAFQC0I9kL3oceMT1O
44WPe6NZ8w8CMwAAAIABGm2Yg8nGFZbo/W8njuM79w0W2P1NBVNWzBH
WQqVbr4i1bWTSSc9X+itQUpeF6zAUDsUoprhNise2NLrMYCLFo9JxhE
iYAcEJ/YbKEnjtJzaAmQNpyh3rCWuOcGPTevjAZIkl+zEc+/N7tCW1e
uDYm6IXZ8LEQyTUQUdU4pZ2OgAAAIABk1ZA3+TiCMaoAafNVUZ7zwqk
888yVOgsJ7HGGDGRMo5ytr2SUJB7QWsLX6Un/Zbu32nXsAqtqagxd6F
Ies98TSekMh/hAv9uK92mEsXSINXOeIMKRedqOyPgk5IEOsFpxAUO4T
xpYToeuM8HRemecxw2eIFHnax+mQqCsi7FgQ== core@valid2`,
	}
	homedir := fmt.Sprintf("/tmp/%s", randSeq(10))
	defer cleanupHomedir(t, homedir)
	authorizer := SSHAuthorizer{
		HomeDir: homedir,
		Uid:     uid,
		Gid:     gid,
	}
	err = authorizer.Authorize(keys)
	assert.NoError(t, err)

	sshdir := filepath.Join(authorizer.HomeDir, ".ssh")
	assert.DirExists(t, sshdir)
	authFile := filepath.Join(sshdir, "authorized_keys")
	assert.FileExists(t, authFile)
	// assert has permissions
	contents, err := GetAuthorizedKeysContents(authFile)
	assert.NoError(t, err)
	assert.Equal(t, strings.Join(keys, "\n")+"\n", contents)
}
