package server

import (
	"time"

	"github.com/tensorland/modelbox/sdk-go/proto"
	"github.com/tensorland/modelbox/server/storage"
	"github.com/tensorland/modelbox/server/storage/artifacts"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func NewFileSetFromProto(pb []*proto.FileMetadata) artifacts.FileSet {
	files := make([]*artifacts.FileMetadata, len(pb))
	for i, b := range pb {
		files[i] = artifacts.NewFileMetadata(
			b.ParentId,
			b.Path,
			b.Checksum,
			artifacts.FileMIMEType(b.FileType),
			b.CreatedAt.AsTime().Unix(),
			b.UpdatedAt.AsTime().Unix(),
		)
	}
	return files
}

func FileSetToProto(f artifacts.FileSet) []*proto.FileMetadata {
	files := make([]*proto.FileMetadata, len(f))
	for i, f := range f {
		files[i] = &proto.FileMetadata{
			Id:        f.Id,
			ParentId:  f.ParentId,
			FileType:  FileTypeToProto(f.Type),
			Checksum:  f.Checksum,
			Path:      f.Path,
			CreatedAt: timestamppb.New(time.Unix(f.CreatedAt, 0)),
			UpdatedAt: timestamppb.New(time.Unix(f.UpdatedAt, 0)),
		}
	}
	return files
}

func MLFrameworkFromProto(fwk proto.MLFramework) storage.MLFramework {
	switch fwk {
	case proto.MLFramework_PYTORCH:
		return storage.Pytorch
	case proto.MLFramework_KERAS:
		return storage.Keras
	}
	return storage.Unknown
}

func MLFrameworkToProto(fwk storage.MLFramework) proto.MLFramework {
	switch fwk {
	case storage.Pytorch:
		return proto.MLFramework_PYTORCH
	case storage.Keras:
		return proto.MLFramework_KERAS
	}
	return proto.MLFramework_UNKNOWN
}

func FileTypeFromProto(t proto.FileType) artifacts.FileMIMEType {
	switch t {
	case proto.FileType_CHECKPOINT:
		return artifacts.CheckpointFile
	case proto.FileType_MODEL:
		return artifacts.ModelFile
	}
	return artifacts.UnknownFile
}

func FileTypeToProto(t artifacts.FileMIMEType) proto.FileType {
	switch t {
	case artifacts.CheckpointFile:
		return proto.FileType_CHECKPOINT
	case artifacts.ModelFile:
		return proto.FileType_MODEL
	case artifacts.TextFile:
		return proto.FileType_TEXT
	case artifacts.ImageFile:
		return proto.FileType_IMAGE
	case artifacts.AudioFile:
		return proto.FileType_AUDIO
	case artifacts.VideoFile:
		return proto.FileType_VIDEO
	}
	return proto.FileType_UNDEFINED
}

func getMetadataOrDefault(meta *proto.Metadata) map[string]*structpb.Value {
	if meta != nil {
		return meta.Metadata
	}
	return nil
}
