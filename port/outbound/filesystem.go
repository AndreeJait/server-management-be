package outbound

import "os"

type Filesystem interface {
	MkdirAll(path string, perm os.FileMode) error
	WriteFile(path string, data []byte, perm os.FileMode) error
	ReadFile(path string) ([]byte, error)
	RemoveFile(path string) error
	RemoveAll(path string) error
}