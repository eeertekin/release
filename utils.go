package release

import (
	"crypto/sha256"
	"fmt"
	"io"
	"math"
	"math/rand"
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

func Debug(format string, args ...any) {
	if !Verbose {
		return
	}
	fmt.Printf(format, args...)
}
