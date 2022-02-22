package testutils

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"sort"
	"strings"
	"time"
)

type TestFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (tfi *TestFileInfo) Name() string       { return tfi.name }
func (tfi *TestFileInfo) Size() int64        { return tfi.size }
func (tfi *TestFileInfo) Mode() os.FileMode  { return tfi.mode }
func (tfi *TestFileInfo) ModTime() time.Time { return tfi.modTime }
func (tfi *TestFileInfo) IsDir() bool        { return tfi.isDir }
func (tfi *TestFileInfo) Sys() interface{}   { return nil }

type TestFile struct {
	data []byte
	perm os.FileMode
	fi   os.FileInfo
}
type TestFileIo struct {
	files map[string]*TestFile
	ForceError map[string]error
}

func NewTestFileIo() *TestFileIo {
	return &TestFileIo{files: make(map[string]*TestFile), ForceError: make(map[string]error)}
}

func (tfi *TestFileIo) getError(api string) error {
	err, exists := tfi.ForceError[api]
	if exists {
		delete(tfi.ForceError, api)
		return err
	}
	return nil
}

func (tfi *TestFileIo) ReadFile(filePath string) ([]byte, error) {
	err := tfi.getError("ReadFile")
	if err != nil {
		return nil, err
	}

	tf, exist := tfi.files[filePath]
	if !exist || tf.fi.IsDir() {
		return nil, os.ErrNotExist
	}
	return tf.data, nil
}

func (tfi *TestFileIo) WriteFile(filePath string, data []byte, perm fs.FileMode) error {
	err := tfi.getError("WriteFile")
	if err != nil {
		return err
	}

	parent := path.Dir(filePath)
	if parent == "." {
		panic("relative path is not mocked")
	}

	parentTfi, exists := tfi.files[parent]
	if !exists || !parentTfi.fi.IsDir() {
		return os.ErrNotExist
	}

	existingTfi, exists := tfi.files[filePath]
	if exists {
		if existingTfi.fi.IsDir() {
			return os.ErrExist
		}
	}

	dirPart, fileName := path.Split(filePath)
	dirPart = strings.TrimSuffix(dirPart, "/")

	dfi, exist := tfi.files[dirPart]
	if !exist || !dfi.fi.IsDir() {
		return os.ErrNotExist
	}

	fi := &TestFileInfo{
		name:    fileName,
		size:    int64(len(data)),
		mode:    0,
		modTime: time.Now(),
		isDir:   false,
	}
	tf := &TestFile{
		data: data,
		perm: perm,
		fi:   fi,
	}
	tfi.files[filePath] = tf
	return nil
}

func (tfi *TestFileIo) Stat(name string) (os.FileInfo, error) {
	err := tfi.getError("Stat")
	if err != nil {
		return nil, err
	}

	tf, exist := tfi.files[name]
	if !exist {
		return nil, os.ErrNotExist
	}
	return tf.fi, nil
}

func (tfi *TestFileIo) MkdirAll(name string, perm fs.FileMode) error {
	err := tfi.getError("MkdirAll")
	if err != nil {
		return err
	}

	existingTfi, exists := tfi.files[name]
	if exists {
		if !existingTfi.fi.IsDir() {
			return os.ErrExist
		}
		return nil
	}

	parent := path.Dir(name)
	if parent != "." && parent != "/" {
		err := tfi.MkdirAll(parent, perm)
		if err != nil {
			return err
		}
	}

	_, fileName := path.Split(name)
	fi := &TestFileInfo{
		name:    fileName,
		size:    0,
		mode:    0,
		modTime: time.Now(),
		isDir:   true,
	}
	tf := &TestFile{
		data: nil,
		perm: perm,
		fi:   fi,
	}
	tfi.files[name] = tf

	return nil
}

func (tfi *TestFileIo) IsEmptyDir(name string) bool {
	dirTfi, exists := tfi.files[name]
	if !exists || !dirTfi.fi.IsDir() {
		return false
	}

	dirSlash := name + "/"
	for file := range tfi.files {
		if strings.HasPrefix(file, dirSlash) {
			return false
		}
	}

	return true
}

func (tfi *TestFileIo) RemoveAll(name string) error {
	err := tfi.getError("RemoveAll")
	if err != nil {
		return err
	}

	nameSlash := name + "/"
	removeList := []string{}
	for file := range tfi.files {
		if file == name || strings.HasPrefix(file, nameSlash) {
			removeList = append(removeList, file)
		}
	}

	for _, file := range removeList {
		delete(tfi.files, file)
	}
	return nil
}

func (tfi *TestFileIo) IsDirectory(name string) (bool, error) {
	err := tfi.getError("IsDirectory")
	if err != nil {
		return false, err
	}

	fi, err := tfi.Stat(name)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return fi.IsDir(), nil
}

func (tfi *TestFileIo) CopyFile(src, dest string) (int64, error) {
	err := tfi.getError("CopyFile")
	if err != nil {
		return 0, err
	}

	existingTfi, exists := tfi.files[src]
	if !exists || existingTfi.fi.IsDir() {
		return 0, os.ErrNotExist
	}

	parent := path.Dir(dest)
	parentTfi, exists := tfi.files[parent]
	if !exists || !parentTfi.fi.IsDir() {
		return 0, os.ErrNotExist
	}

	_, fileName := path.Split(dest)
	fi := &TestFileInfo{
		name:    fileName,
		size:    existingTfi.fi.Size(),
		mode:    existingTfi.fi.Mode(),
		modTime: existingTfi.fi.ModTime(),
		isDir:   false,
	}
	tf := &TestFile{
		data: make([]byte, len(existingTfi.data)),
		perm: existingTfi.perm,
		fi:   fi,
	}
	copy(tf.data, existingTfi.data)
	tfi.files[dest] = tf
	return fi.size, nil
}

func (tfi *TestFileIo) ReadDir(src string) (files []fs.DirEntry, err error) {
	err = tfi.getError("ReadDir")
	if err != nil {
		return
	}

	dirSlash := path.Clean(src) + "/"

	for file,tf := range tfi.files {
		if strings.HasPrefix(file, dirSlash) {
			files = append(files, fs.FileInfoToDirEntry(tf.fi))
		}
	}

	return
}

func (tfi *TestFileIo) FileExists(name string) (bool, error) {
	err := tfi.getError("FileExists")
	if err != nil {
		return false, err
	}

	tf, exists := tfi.files[name]
	if !exists || tf.fi.IsDir() {
		return false, nil
	}
	return true, nil
}

func (tfi *TestFileIo) Dump() {
	keys := make([]string, 0, len(tfi.files))
	for k := range tfi.files {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		v := tfi.files[k]
		if !v.fi.IsDir() {
			fmt.Printf("%s - %d bytes\n", k, len(v.data))
		} else {
			fmt.Printf("%s (dir)\n", k)
		}
	}
}

