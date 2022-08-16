package artifacts

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/diptanu/modelbox/server/config"
	"github.com/diptanu/modelbox/server/utils"
	"go.uber.org/zap"
)

type FileMIMEType uint8

const (
	UnknownFile FileMIMEType = iota
	CheckpointFile
	ModelFile
	TextFile
	ImageFile
	AudioFile
	VideoFile
)

/*
 * FileMetadata are metadata about files and other blobs such as models.
 * They can be associated with any modelbox object.
 */
type FileMetadata struct {
	Id        string
	ParentId  string
	Type      FileMIMEType
	Path      string
	Checksum  string
	CreatedAt int64
	UpdatedAt int64
}

func NewFileMetadata(
	parent, path, checksum string,
	blobType FileMIMEType,
	createdAt, updatedAt int64,
) *FileMetadata {
	currentTime := time.Now().Unix()
	if createdAt == 0 {
		createdAt = currentTime
	}
	if updatedAt == 0 {
		updatedAt = currentTime
	}
	blob := &FileMetadata{
		ParentId:  parent,
		Path:      path,
		Checksum:  checksum,
		Type:      blobType,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
	blob.CreateId()
	return blob
}

func (b *FileMetadata) CreateId() {
	h := sha1.New()
	utils.HashString(h, b.ParentId)
	utils.HashInt(h, int(b.Type))
	utils.HashString(h, b.Checksum)
	b.Id = fmt.Sprintf("%x", h.Sum(nil))
}

func (b *FileMetadata) ToJson() ([]byte, error) {
	bytes, err := json.Marshal(b)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

type FileSet []*FileMetadata

type FileOpenMode uint8

const (
	Read FileOpenMode = iota
	Write
)

type BlobStorage interface {
	Open(blob *FileMetadata, mode FileOpenMode) error

	GetPath() (string, error)

	io.ReadWriteCloser
}

type BlobStorageBuilder interface {
	Build() BlobStorage
}

func NewBlobStorageBuilder(
	svrConfig *config.ServerConfig,
	logger *zap.Logger,
) (BlobStorageBuilder, error) {
	switch svrConfig.StorageBackend {
	case config.BLOB_STORAGE_BACKEND_FS:
		return NewFileBlobStorageBuilder(svrConfig.FileStorage.BaseDir, logger)
	}
	return nil, fmt.Errorf("unknown blob storage backend: %v", svrConfig.StorageBackend)
}
