package storage

import (
	"github.com/mitchellh/goamz/s3"
)

const (
	ACLPublicRead = s3.ACL("public-read")
)

func BucketExists(svc *s3.S3, bucketName string) (bool, error) {
	buckets, err := svc.ListBuckets()
	if err != nil {
		return false, err
	}
	for _, bucket := range buckets.Buckets {
		if bucketName == bucket.Name {
			return true, nil
		}
	}
	return false, nil
}

func CreateBucket(svc *s3.S3, bucketName string) error {
	return svc.Bucket(bucketName).PutBucket(s3.ACL("public-read"))
}
