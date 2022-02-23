package testutils

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/jimsnab/go-simpleutils"
)

type FileIo interface {
	ReadFile(filePath string) ([]byte, error)
	WriteFile(filePath string, data []byte, perm fs.FileMode) error
	Stat(name string) (os.FileInfo, error)
	MkdirAll(name string, perm fs.FileMode) error
	RemoveAll(name string) error
	IsDirectory(name string) (bool, error)
	CopyFile(src, dest string) (int64, error)
	ReadDir(src string) ([]fs.DirEntry, error)
	FileExists(name string) (bool, error)
	Walk(root string, fn filepath.WalkFunc) error
}

type realFileIo struct {
}

func NewFileIo() FileIo {
	return &realFileIo{}
}

func (rfi *realFileIo) ReadFile(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}

func (rfi *realFileIo) WriteFile(filePath string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(filePath, data, perm)
}

func (rfi *realFileIo) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (rfi *realFileIo) MkdirAll(name string, perm fs.FileMode) error {
	return os.MkdirAll(name, perm)
}

func (rfi *realFileIo) RemoveAll(name string) error {
	err := os.RemoveAll(name)
	if err == nil {
		var isDir bool
		if isDir, err = simpleutils.IsDirectory(name); err != nil {
			return err
		}
		if isDir {
			err = os.ErrExist
		}
	}
	return err
}

func (rfi *realFileIo) IsDirectory(name string) (bool, error) {
	return simpleutils.IsDirectory(name)
}

func (rfi *realFileIo) CopyFile(src, dest string) (int64, error) {
	return simpleutils.CopyFile(src, dest)
}

func (rfi *realFileIo) ReadDir(src string) (files []fs.DirEntry, err error) {
	return os.ReadDir(src)
}

func (rfi *realFileIo) FileExists(name string) (bool, error) {
	return simpleutils.FileExists(name)
}

func (rfi *realFileIo) Walk(root string, fn filepath.WalkFunc) error {
	return filepath.Walk(root, fn)
}