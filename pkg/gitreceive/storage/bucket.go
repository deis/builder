package storage

func CreateBucket(s3Client *s3.Client, bucketName string) error {
	_, err := s3Client.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	return err
}
