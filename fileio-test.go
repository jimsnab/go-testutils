package testutils

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
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

type TestFileHook func(args ...string) error

type TestFileIo struct {
	files      map[string]*TestFile
	ForceError map[string]error
	Hooks      map[string]TestFileHook
}

func NewTestFileIo() *TestFileIo {
	return &TestFileIo{
		files:      make(map[string]*TestFile),
		ForceError: make(map[string]error),
		Hooks:      make(map[string]TestFileHook),
	}
}

func (tfi *TestFileIo) getError(api string, path ...string) error {
	hook, exists := tfi.Hooks[api]
	if exists {
		err := hook(path...)
		if err != nil {
			return err
		}
	}

	err, exists := tfi.ForceError[api]
	if !exists {
		for _, p := range path {
			err, exists = tfi.ForceError[p]
			if exists {
				break
			}
		}
	}
	if exists {
		delete(tfi.ForceError, api)
		return err
	}
	return nil
}

func (tfi *TestFileIo) ReadFile(filePath string) ([]byte, error) {
	err := tfi.getError("ReadFile", filePath)
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
	err := tfi.getError("WriteFile", filePath)
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
	err := tfi.getError("Stat", name)
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
	err := tfi.getError("MkdirAll", name)
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
	err := tfi.getError("RemoveAll", name)
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
	err := tfi.getError("IsDirectory", name)
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
	err := tfi.getError("CopyFile", src, dest)
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
	err = tfi.getError("ReadDir", src)
	if err != nil {
		return
	}

	dirSlash := path.Clean(src) + "/"

	for file, tf := range tfi.files {
		if strings.HasPrefix(file, dirSlash) {
			files = append(files, fs.FileInfoToDirEntry(tf.fi))
		}
	}

	return
}

func (tfi *TestFileIo) FileExists(name string) (bool, error) {
	err := tfi.getError("FileExists", name)
	if err != nil {
		return false, err
	}

	tf, exists := tfi.files[name]
	if !exists || tf.fi.IsDir() {
		return false, nil
	}
	return true, nil
}

func (tfi *TestFileIo) Walk(root string, fn filepath.WalkFunc) (err error) {
	err = tfi.getError("Walk", root)
	if err != nil {
		return
	}

	cleanRoot := path.Clean(root)

	tf, exists := tfi.files[cleanRoot]
	if exists {
		err = fn(cleanRoot, tf.fi, nil)
		if err != nil {
			if errors.Is(err, filepath.SkipDir) {
				err = nil
			}
			return
		}

		dirSlash := cleanRoot + "/"
		for file, tf := range tfi.files {
			if strings.HasPrefix(file, dirSlash) {
				if !strings.Contains(file[len(dirSlash):], "/") {
					// found a file or subdirectory within root (but not a subpath)
					err = tfi.getError("Walk", file)
					if err != nil {
						return
					}
					if tf.fi.IsDir() {
						// subdirectory - recurse
						if err = tfi.Walk(file, fn); err != nil {
							return
						}
					} else {
						// file
						err = fn(file, tf.fi, nil)
						if err != nil {
							if errors.Is(err, filepath.SkipDir) {
								err = nil
							}
							return
						}
					}
				}
			}
		}	
	}
	return
}

func (tfi *TestFileIo) Chtimes(name string, atime, mtime time.Time) error {
	err := tfi.getError("Chtimes", name)
	if err != nil {
		return err
	}

	tf, exist := tfi.files[name]
	if !exist {
		return os.ErrNotExist
	}

	tf.fi = &TestFileInfo{
		name:    tf.fi.Name(),
		size:    tf.fi.Size(),
		mode:    tf.fi.Mode(),
		modTime: mtime,
		isDir:   tf.fi.IsDir(),
	}

	return nil
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

