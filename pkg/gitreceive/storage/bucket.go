package storage

import (
	s3 "github.com/minio/minio-go"
)

const (
	bucketAlreadyExistsCode = "BucketAlreadyExists"
	nonExistentBucketCode   = "NoSuchBucket"
)

var (
	// ACLPublicRead default ACL for objects in the S3 API compatible storage
	ACLPublicRead = s3.BucketACL("public-read")
)

// CreateBucket creates a new bucket in the S3 API compatible storage or
// return an error in case the bucket already exists
func CreateBucket(creator BucketCreator, bucketName string) error {
	_, err := creator.MakeBucket(bucketName, ACLPublicRead, "")
	if err != nil {
		minioErr := s3.ToErrorResponse(err)
		if minioErr.Code == bucketAlreadyExistsCode {
			return nil
		}
		return err
	}
	return nil
}
