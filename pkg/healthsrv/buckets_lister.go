package healthsrv

import (
	s3 "github.com/aws/aws-sdk-go/service/s3"
)

// BucketLister is a *(github.com/aws/aws-sdk-go/service/s3).Client compatible interface that provides just the ListBuckets cross-section of functionality. It can also be implemented for unit tests
type BucketLister interface {
	// ListBuckets lists all the buckets in the object storage system
	ListBuckets(*s3.ListBucketsInput) (*s3.ListBucketsOutput, error)
}
