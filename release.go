package release

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"
)

type Version struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	SHA256   string `json:"SHA256"`
	ExecPath string `json:"-"`
}

var Repo string
var Release string = "latest"

var currentVersion Version

func init() {
	path, err := os.Executable()
	if err != nil {
		path = os.Args[0]
	}

	// TODO :: Check SHA256 errors
	currentVersion = Version{
		ExecPath: path,
		SHA256:   SHA256(path),
	}
}

func Check() bool {
	releaseVersion := manifest()
	if currentVersion.SHA256 == releaseVersion.SHA256 {
		Debug("%s v%s (%s) - latest", currentVersion.ExecPath, releaseVersion.Version, releaseVersion.SHA256)
		return false
	}

	if err := deploy(releaseVersion); err != nil {
		Debug("err> %s\n", err)
		return false
	}

	return true
}

func manifest() *Version {
	URL := fmt.Sprintf("%s/%s?%d", Repo, Release, randomInt())
	Debug(fmt.Sprintf("> Fetching %s\n", URL))

	client := http.Client{Timeout: 30 * time.Second}
	res, err := client.Get(URL)
	if err != nil {
		Debug(fmt.Sprintf("err> %s\n", err))
		return nil
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		Debug("> Manifest fetch failed\n")
		return nil
	}

	manifest, err := io.ReadAll(res.Body)
	if err != nil {
		Debug(fmt.Sprintf("err> %s\n", err))
		return nil
	}

	v := Version{
		ExecPath: currentVersion.ExecPath,
	}
	err = json.Unmarshal(manifest, &v)
	if err != nil {
		Debug(fmt.Sprintf("err> %s\n", err))
		return nil
	}

	return &v
}

func deploy(v *Version) error {
	URL := fmt.Sprintf("%s/%s?%d", Repo, v.SHA256, randomInt())
	Debug(fmt.Sprintf("> Downloading ... %s\n", URL))

	newVersionFile, err := os.CreateTemp("/tmp/", v.Name)
	if err != nil {
		return err
	}

	client := http.Client{Timeout: 30 * time.Second}
	repoVersion, err := client.Get(URL)
	if err != nil {
		return err
	}

	_, err = io.Copy(newVersionFile, repoVersion.Body)
	if err != nil {
		return err
	}
	repoVersion.Body.Close()
	newVersionFile.Chmod(0700)
	newVersionFile.Close()

	Debug(fmt.Sprintf("Downloaded to %s\n", newVersionFile.Name()))

	// TODO :: Check SHA256 errors
	repoVersionSHA := SHA256(newVersionFile.Name())
	if repoVersionSHA != v.SHA256 {
		return fmt.Errorf("remote SHA256 not matched (%s != %s)", v.SHA256, repoVersionSHA)
	}

	Debug("Release file checksum ... OK\n")
	if err := os.Rename(newVersionFile.Name(), v.ExecPath); err != nil {
		exec.Command("/bin/mv", newVersionFile.Name(), v.ExecPath).Run()
	}

	Debug("New version deployed ... %s\n", v.ExecPath)
	return nil
}
