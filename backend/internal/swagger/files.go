package swagger

import (
	"context"
	"errors"
	"io/fs"
	"net/http"
	"os"

	swaggerFiles "github.com/swaggo/files/v2"
	"golang.org/x/net/webdav"
)

type readOnlySwaggerFS struct{}

func (readOnlySwaggerFS) Mkdir(context.Context, string, os.FileMode) error { return fs.ErrPermission }

func (readOnlySwaggerFS) RemoveAll(context.Context, string) error { return fs.ErrPermission }

func (readOnlySwaggerFS) Rename(context.Context, string, string) error { return fs.ErrPermission }

func (readOnlySwaggerFS) OpenFile(_ context.Context, name string, _ int, _ os.FileMode) (webdav.File, error) {
	file, err := http.FS(swaggerFiles.FS).Open(name)
	if err != nil {
		return nil, err
	}
	return readOnlySwaggerFile{File: file}, nil
}

func (readOnlySwaggerFS) Stat(_ context.Context, name string) (os.FileInfo, error) {
	return fs.Stat(swaggerFiles.FS, name)
}

type readOnlySwaggerFile struct{ http.File }

func (readOnlySwaggerFile) Write([]byte) (int, error) {
	return 0, errors.New("swagger UI files are read-only")
}

func FilesHandler() *webdav.Handler {
	return &webdav.Handler{FileSystem: readOnlySwaggerFS{}, LockSystem: webdav.NewMemLS()}
}
