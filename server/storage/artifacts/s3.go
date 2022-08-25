package artifacts

import (
	"bytes"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/tensorland/modelbox/server/config"

	"go.uber.org/zap"
)

// Minimum number of bytes we buffer before bytes is written to S3
const BUF_SIZE = 5242880

type S3Writer struct {
	buf            bytes.Buffer
	uploadOut      *s3.CreateMultipartUploadOutput
	completedParts []*s3.CompletedPart
	partIndex      uint64

	fileId string
	bucket string
	svc    *s3.S3
	sess   *session.Session
	logger *zap.Logger
}

func NewS3Writer(sess *session.Session, fileMeta *FileMetadata, bucket string, logger *zap.Logger) (*S3Writer, error) {
	svc := s3.New(sess)
	resp, err := svc.CreateMultipartUpload(&s3.CreateMultipartUploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fileMeta.Id),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create object in s3: %v", err)
	}
	return &S3Writer{uploadOut: resp, partIndex: 1, fileId: fileMeta.Id, bucket: bucket, svc: svc, sess: sess, logger: logger}, nil
}

func (s *S3Writer) GetPath() (string, error) {
	return fmt.Sprintf("s3://%s/%s", s.bucket, s.fileId), nil
}

func (s *S3Writer) Write(p []byte) (int, error) {
	n, _ := s.buf.Write(p)
	var err error
	if s.buf.Len() >= BUF_SIZE {
		err = s.uploadPart(s.buf.Bytes())
		s.buf.Reset()
	}
	return n, err
}

func (s *S3Writer) Close() error {
	// Upload remaining parts of the buffer
	if s.buf.Len() > 0 {
		s.uploadPart(s.buf.Bytes())
	}
	completeInput := &s3.CompleteMultipartUploadInput{
		Bucket:   s.uploadOut.Bucket,
		Key:      s.uploadOut.Key,
		UploadId: s.uploadOut.UploadId,
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: s.completedParts,
		},
	}
	if _, err := s.svc.CompleteMultipartUpload(completeInput); err != nil {
		s.logger.Sugar().Errorf("unable to finish upload to s3: %v", err)
		return err
	}
	return nil
}

func (s *S3Writer) uploadPart(buf []byte) error {
	partInput := &s3.UploadPartInput{
		Body:          bytes.NewReader(buf),
		Bucket:        s.uploadOut.Bucket,
		Key:           s.uploadOut.Key,
		PartNumber:    aws.Int64(int64(s.partIndex)),
		UploadId:      s.uploadOut.UploadId,
		ContentLength: aws.Int64(int64(len(buf))),
	}

	uploadResult, err := s.svc.UploadPart(partInput)
	if err != nil {
		return err
	}
	completedPart := &s3.CompletedPart{
		ETag:       uploadResult.ETag,
		PartNumber: aws.Int64(int64(s.partIndex)),
	}
	s.completedParts = append(s.completedParts, completedPart)
	s.partIndex = s.partIndex + 1
	return nil
}

type S3Reader struct {
	svc    *s3.S3
	out    *s3.GetObjectOutput
	logger *zap.Logger
}

func NewS3Reader(sess *session.Session, fileMeta *FileMetadata, bucket string, logger *zap.Logger) (*S3Reader, error) {
	svc := s3.New(sess)
	getInput := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fileMeta.Id),
	}
	resp, err := svc.GetObject(getInput)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve object: %v", err)
	}
	return &S3Reader{
		svc:    svc,
		out:    resp,
		logger: logger,
	}, nil
}

func (s *S3Reader) Read(p []byte) (n int, err error) {
	return s.out.Body.Read(p)
}

func (s *S3Reader) Close() error {
	return s.out.Body.Close()
}

type S3StorageBuilder struct {
	config *config.S3StorageConfig
	logger *zap.Logger
}

func NewS3StorageBuilder(s3Config *config.S3StorageConfig, logger *zap.Logger) *S3StorageBuilder {
	return &S3StorageBuilder{config: s3Config, logger: logger}
}

func (s *S3StorageBuilder) BuildWriter(fileMeta *FileMetadata) (BlobStorageWriter, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(s.config.Region),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create aws session: %v", err)
	}
	return NewS3Writer(sess, fileMeta, s.config.Bucket, s.logger)
}

func (s *S3StorageBuilder) BuildReader(fileMeta *FileMetadata) (BlobStorageReader, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(s.config.Region),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create aws session: %v", err)
	}
	return NewS3Reader(sess, fileMeta, s.config.Bucket, s.logger)
}

func (*S3StorageBuilder) Backend() string {
	return "s3"
}
