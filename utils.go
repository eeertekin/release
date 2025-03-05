package release

import (
	"crypto/sha256"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"os"
	"time"
)

var Verbose bool = false

func SHA256(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return err.Error()
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err.Error()
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

func randomInt() int {
	s := rand.NewSource(time.Now().UnixNano())
	return rand.New(s).Intn(math.MaxInt64)
}

func verbose(format string, args ...any) {
	if !Verbose {
		return
	}
	fmt.Printf(format, args...)
}

func fetch(URL string) (*http.Response, error) {
	client := http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, err
	}

	req.Header = HTTPHeader.Clone()
	if HTTPAuthorization != "" {
		req.Header.Set("Authorization", HTTPAuthorization)
	}
	return client.Do(req)
}

func GetArtifactURL(repo, file string) string {
	return fmt.Sprintf("%s/%s/%s?%d", Storage, repo, file, randomInt())
}
