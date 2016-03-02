package storage

import (
	s3 "github.com/minio/minio-go"
)

type BucketCreator interface {
	MakeBucket(bucketName string, acl s3.BucketACL, location string) error
}

type ObjectStatter interface {
	StatObject(bucketName, objectKey string) (s3.ObjectInfo, error)
}

type ObjectPutter interface {
	PutObject(bucketName, objectKey string, reader io.Reader, contentType string) (int64, error)
}
