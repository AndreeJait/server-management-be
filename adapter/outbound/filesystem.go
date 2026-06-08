package outbound

import (
	"os"

	"github.com/AndreeJait/server-management-be/port/outbound"
)

type osFilesystem struct{}

func NewFilesystem() outbound.Filesystem {
	return &osFilesystem{}
}

func (f *osFilesystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (f *osFilesystem) WriteFile(path string, data []byte, perm os.FileMode) error {
	return os.WriteFile(path, data, perm)
}

func (f *osFilesystem) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (f *osFilesystem) RemoveFile(path string) error {
	return os.Remove(path)
}

func (f *osFilesystem) RemoveAll(path string) error {
	return os.RemoveAll(path)
}