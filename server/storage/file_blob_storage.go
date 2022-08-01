package storage

import (
	"fmt"
	"os"
	"path"

	"go.uber.org/zap"
)

type FileBlobStorage struct {
	file    *os.File
	baseDir string

	log *zap.Logger
}

func NewFileBlobStorage(baseDir string, log *zap.Logger) *FileBlobStorage {
	return &FileBlobStorage{baseDir: baseDir, log: log}
}

func (f *FileBlobStorage) Open(blobInfo *FileMetadata, mode FileOpenMode) error {
	var err error
	if mode == Read {
		if f.file, err = os.Open(blobInfo.Path); err != nil {
			return fmt.Errorf("couldn't open %v to read: %v", blobInfo.Path, err)
		}
		return nil
	}

	// If we have to create this file we need to construct a path first.
	path := path.Join(f.baseDir, blobInfo.Id)

	if f.file, err = os.Create(path); err != nil {
		return fmt.Errorf("couldn't open %v to read: %v", path, err)
	}
	blobInfo.Path = path
	return nil
}

func (f *FileBlobStorage) Close() error {
	return f.file.Close()
}

func (f *FileBlobStorage) Write(p []byte) (n int, err error) {
	return f.file.Write(p)
}

func (f *FileBlobStorage) Read(p []byte) (n int, err error) {
	return f.file.Read(p)
}

func (f *FileBlobStorage) GetPath() (string, error) {
	if f.file == nil {
		return "", fmt.Errorf("unable to get path: no file found")
	}
	return f.file.Name(), nil
}

type FileBlobStorageBuilder struct {
	baseDir string
	logger  *zap.Logger
}

func NewFileBlobStorageBuilder(baseDir string, logger *zap.Logger) (*FileBlobStorageBuilder, error) {
	if err := os.MkdirAll(baseDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("couldn't create blob storage directory: %v", err)
	}
	return &FileBlobStorageBuilder{baseDir: baseDir, logger: logger}, nil
}

func (f *FileBlobStorageBuilder) Build() BlobStorage {
	return NewFileBlobStorage(f.baseDir, f.logger)
}
