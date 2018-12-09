package file_storage

import (
	"errors"
	"io"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const (
	ModeDir    = "nfs"
	ModeBucket = "bucket"
)

type FileStorage struct {
	uploader *s3manager.Uploader

	dirPath string
	mode    string
	region  string
	bucket  string
}

func NewService(mode, dirPath, bucket string, sess *session.Session) *FileStorage {
	return &FileStorage{
		uploader: s3manager.NewUploader(sess),

		dirPath: dirPath,
		mode:    mode,
		bucket:  bucket,
	}
}

func (service *FileStorage) Upload(filename string, file io.Reader) (string, error) {
	if service.mode == ModeDir {
		return filename, service.uploadToDir(filename, file)
	} else if service.mode == ModeBucket {
		return service.uploadToBucket(filename, file)
	}
	return "", errors.New("Unknow mode")
}

func (service *FileStorage) uploadToDir(filename string, file io.Reader) error {
	f, err := os.OpenFile(service.dirPath+filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, file)
	return err
}

func (s *FileStorage) uploadToBucket(filename string, file io.Reader) (string, error) {
	result, err := s.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(filename),
		//Key:    aws.String(filepath.Base(filename)),
		Body: file,
	})
	if err != nil {
		return "", err
	}
	return result.Location, nil
}
