package file_storage_service

import (
	"errors"
	"fmt"
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

type FileStorageService struct {
	dirPath string
	mode    string
	region  string
	bucket  string
}

func NewService(mode, dirPath, region, bucket string) *FileStorageService {
	return &FileStorageService{
		dirPath: dirPath,
		mode:    mode,
		region:  region,
		bucket:  bucket,
	}
}

func (service *FileStorageService) Upload(filename string, file io.Reader) error {
	if service.mode == ModeDir {
		return service.uploadToDir(filename, file)
	} else if service.mode == ModeBucket {
		return service.uploadToBucket(filename, file)
	}
	return errors.New("Unknow mode")
}

func (service *FileStorageService) uploadToDir(filename string, file io.Reader) error {
	f, err := os.OpenFile(service.dirPath+filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, file)
	return err
}

func (service *FileStorageService) uploadToBucket(filename string, file io.Reader) error {
	// Initialize a session in us-west-2 that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials.
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(service.region)},
	)

	// Setup the S3 Upload Manager. Also see the SDK doc for the Upload Manager
	// for more information on configuring part size, and concurrency.
	//
	// http://docs.aws.amazon.com/sdk-for-go/api/service/s3/s3manager/#NewUploader
	uploader := s3manager.NewUploader(sess)

	// Upload the file's body to S3 bucket as an object with the key being the
	// same as the filename.
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(service.bucket),

		// Can also use the `filepath` standard library package to modify the
		// filename as need for an S3 object key. Such as turning absolute path
		// to a relative path.
		Key: aws.String(filename),

		// The file to be uploaded. io.ReadSeeker is preferred as the Uploader
		// will be able to optimize memory when uploading large content. io.Reader
		// is supported, but will require buffering of the reader's bytes for
		// each part.
		Body: file,
	})
	if err != nil {
		return fmt.Errorf("Unable to upload %q to %q, err: %v", filename, service.bucket, err)
	}
	return nil
}
