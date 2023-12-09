package release

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

var Repo string
var Release string = "latest"
var Storage string

var local_version Version

func init() {
	path, err := os.Executable()
	if err != nil {
		path = os.Args[0]
	}

	// TODO :: Check SHA256 errors
	local_version = Version{
		ExecPath: path,
		SHA256:   SHA256(path),
	}
}

func Update() bool {
	repo_version := GetHead(Repo, Release)
	if repo_version == nil {
		// JSON fetch failed - 404?
		return false
	}
	repo_version.Repo = Repo

	if local_version.SHA256 == repo_version.SHA256 {
		verbose("%s v%s (%s) - latest", local_version.ExecPath, repo_version.Version, repo_version.SHA256)
		return false
	}

	if err := repo_version.Deploy(); err != nil {
		verbose("%s\n", err)
		return false
	}

	return true
}

func GetHead(repo, release string) *Version {
	URL := GetArtifactURL(repo, release)
	verbose("head> Fetching %s\n", URL)

	res, err := fetch(URL)
	if err != nil {
		verbose("head> %s\n", err)
		return nil
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		verbose("head> fetch failed\n")
		return nil
	}

	manifest, err := io.ReadAll(res.Body)
	if err != nil {
		verbose("head> %s\n", err)
		return nil
	}

	v := Version{
		ExecPath: local_version.ExecPath,
	}
	err = json.Unmarshal(manifest, &v)
	if err != nil {
		verbose("head> %s\n", err)
		return nil
	}

	return &v
}

func Install(repo, release string) {
	v := GetHead(repo, release)
	if v == nil {
		fmt.Printf("[%s] Manifest not found\n", repo)
		return
	}
	v.ExecPath = "/bin/" + v.Name
	v.Repo = repo

	if err := v.Deploy(); err != nil {
		fmt.Printf("[%s] %s\n", v.Name, err)
		return
	}

	fmt.Printf("[%s] Install completed\n", v.Name)
}
