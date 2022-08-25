package artifacts

import (
	"fmt"
	"os"
	"path"

	"go.uber.org/zap"
)

type FileWriter struct {
	file    *os.File
	baseDir string
	logger  *zap.Logger
}

func NewFileWriter(baseDir string, fileMeta *FileMetadata, logger *zap.Logger) (*FileWriter, error) {
	path := path.Join(baseDir, fileMeta.Id)
	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("couldn't open %v to read: %v", path, err)
	}
	fileMeta.Path = path
	return &FileWriter{
		file:    file,
		baseDir: baseDir,
		logger:  logger,
	}, nil
}

func (f *FileWriter) Write(p []byte) (n int, err error) {
	return f.file.Write(p)
}

func (f *FileWriter) Close() error {
	return f.file.Close()
}

func (f *FileWriter) GetPath() (string, error) {
	if f.file == nil {
		return "", fmt.Errorf("unable to get path: no file found")
	}
	return f.file.Name(), nil
}

type FileReader struct {
	file    *os.File
	baseDir string
	logger  *zap.Logger
}

func NewFileReader(baseDir string, fileMeta *FileMetadata, logger *zap.Logger) (*FileReader, error) {
	file, err := os.Open(fileMeta.Path)
	if err != nil {
		return nil, fmt.Errorf("couldn't open %v to read: %v", fileMeta.Path, err)
	}
	return &FileReader{
		file:    file,
		baseDir: baseDir,
		logger:  logger,
	}, nil
}

func (f *FileReader) Read(p []byte) (n int, err error) {
	return f.file.Read(p)
}

func (f *FileReader) Close() error {
	return f.file.Close()
}

func (f *FileReader) GetPath() (string, error) {
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

func (f *FileBlobStorageBuilder) BuildWriter(fileMeta *FileMetadata) (BlobStorageWriter, error) {
	return NewFileWriter(f.baseDir, fileMeta, f.logger)
}

func (f *FileBlobStorageBuilder) BuildReader(fileMeta *FileMetadata) (BlobStorageReader, error) {
	return NewFileReader(f.baseDir, fileMeta, f.logger)
}

func (*FileBlobStorageBuilder) Backend() string {
	return "filesystem"
}
