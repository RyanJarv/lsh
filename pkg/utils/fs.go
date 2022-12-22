package utils

import (
	"bufio"
	"fmt"
	L "github.com/ryanjarv/msh/pkg/logger"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

var DevNull, _ = os.Open(os.DevNull)

func NewTmpDir(id string) (TmpDir, error) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "pkg-"+id)
	return TmpDir{Path: tmpDir}, Wrap(err, "NewTmpDir: %e", err)
}

type TmpDir struct {
	Path string
}

func (t TmpDir) AddFile(src, dst string) {
	if err := os.Link(src, filepath.Join(t.Path, dst)); err != nil {
		panic(err)
	}
}

func (t TmpDir) AddContent(content []byte, dst string) {
	if err := ioutil.WriteFile(filepath.Join(t.Path, dst), content, os.FileMode(0400)); err != nil {
		panic(err)
	}
}

func (t TmpDir) Remove() error {
	err := os.RemoveAll(t.Path)
	return Wrap(err, "pagh %s", t.Path)
}

func IsTTY(file *os.File) bool {
	if os.Getenv("MSH_TTY") == "false" {
		return false
	}

	envName := fmt.Sprintf("MSH_TTY_%d", int(file.Fd()))
	if env := os.Getenv(envName); env == "false" {
		return false
	} else if env == "true" {
		return true
	}

	stat, err := file.Stat()
	if err != nil {
		L.Debug.Printf("fd %d: stat failed, assuming we're not using a tty", file.Fd())
		return false
	}

	if stat.Mode()&os.ModeCharDevice == 0 {
		L.Debug.Printf("fd %d: pipe\n", file.Fd())
		return false
	} else {
		L.Debug.Printf("fd %d: tty\n", file.Fd())
		return true
	}
}

func ReadLine(r io.Reader) string {
	scan := bufio.NewScanner(r)
	scan.Scan()
	return scan.Text()
}
