package release

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

type Version struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	SHA256   string `json:"SHA256"`
	ExecPath string `json:"-"`
	Repo     string `json:"-"`
}

func (v *Version) Deploy() error {
	URL := GetArtifactURL(v.Repo, v.SHA256)
	verbose("deploy> Downloading ... %s\n", URL)

	tmp, err := os.CreateTemp("/tmp/", v.Name)
	if err != nil {
		return err
	}

	repo_bin, err := fetch(URL)
	if err != nil {
		return err
	}

	_, err = io.Copy(tmp, repo_bin.Body)
	if err != nil {
		return err
	}
	repo_bin.Body.Close()
	tmp.Chmod(0700)
	tmp.Close()

	verbose("deploy> Downloaded to %s\n", tmp.Name())

	// TODO :: Check SHA256 errors
	tmpSHA := SHA256(tmp.Name())
	if tmpSHA != v.SHA256 {
		return fmt.Errorf("deploy> remote SHA256 not matched (%s != %s)", v.SHA256, tmpSHA)
	}

	verbose("deploy> Release file checksum ... OK\n")
	if err := os.Rename(tmp.Name(), v.ExecPath); err != nil {
		exec.Command("/bin/mv", tmp.Name(), v.ExecPath).Run()
	}

	verbose("deploy> New version installed ... %s\n", v.ExecPath)
	return nil
}
