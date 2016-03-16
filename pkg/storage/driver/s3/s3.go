package s3

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	storagedriver "github.com/deis/builder/pkg/storage/driver"
	"github.com/deis/builder/pkg/storage/driver/base"
	"github.com/deis/builder/pkg/storage/driver/factory"
)

const driverName = "s3"

// minChunkSize defines the minimum multipart upload chunk size
// S3 API requires multipart upload chunks to be at least 5MB
const minChunkSize = 5 << 20

const defaultChunkSize = 2 * minChunkSize

// listMax is the largest amount of objects you can request from S3 in a list call
const listMax = 1000

// validRegions maps known s3 region identifiers to region descriptors
var validRegions = map[string]struct{}{}

//DriverParameters A struct that encapsulates all of the driver parameters after all values have been set
type DriverParameters struct {
	AccessKey      string
	SecretKey      string
	Bucket         string
	RegionEndpoint string
	Region         string
	Encrypt        bool
	Secure         bool
	ChunkSize      int64
	RootDirectory  string
	StorageClass   string
}

func init() {
	for _, region := range []string{
		"us-east-1",
		"us-west-1",
		"us-west-2",
		"eu-west-1",
		"eu-central-1",
		"ap-southeast-1",
		"ap-southeast-2",
		"ap-northeast-1",
		"ap-northeast-2",
		"sa-east-1",
	} {
		validRegions[region] = struct{}{}
	}

	factory.Register(driverName, &s3DriverFactory{})
	factory.Register("minio", &s3DriverFactory{})
}

// s3DriverFactory implements the factory.StorageDriverFactory interface
type s3DriverFactory struct{}

func (factory *s3DriverFactory) Create(parameters map[string]string) (storagedriver.StorageDriver, error) {
	return FromParameters(parameters)
}

type driver struct {
	S3            *s3.S3
	Bucket        string
	ChunkSize     int64
	Encrypt       bool
	RootDirectory string
	StorageClass  string

	pool  sync.Pool // pool []byte buffers used for WriteStream
	zeros []byte    // shared, zero-valued buffer used for WriteStream
}

type baseEmbed struct {
	base.Base
}

// Driver is a storagedriver.StorageDriver implementation backed by Amazon S3
// Objects are stored at absolute keys in the provided bucket.
type Driver struct {
	baseEmbed
}

// FromParameters constructs a new Driver with a given parameters map
// Required parameters:
// - accesskey
// - secretkey
// - region
// - bucket
// - encrypt
func FromParameters(parameters map[string]string) (*Driver, error) {
	// Providing no values for these is valid in case the user is authenticating
	// with an IAM on an ec2 instance (in which case the instance credentials will
	// be summoned when GetAuth is called)
	var err error
	accessKey, ok := parameters["accesskey"]
	if !ok {
		accessKey = ""
	}
	secretKey, ok := parameters["secretkey"]
	if !ok {
		secretKey = ""
	}

	regionName, ok := parameters["region"]
	if !ok || fmt.Sprint(regionName) == "" {
		return nil, fmt.Errorf("No region parameter provided")
	}
	region := fmt.Sprint(regionName)
	_, ok = validRegions[region]
	if !ok {
		return nil, fmt.Errorf("Invalid region provided: %v", region)
	}

	bucket, ok := parameters["builder-bucket"]
	if !ok || fmt.Sprint(bucket) == "" {
		return nil, fmt.Errorf("No bucket parameter provided")
	}

	regionEndpoint, ok := parameters["regionendpoint"]
	if !ok || regionEndpoint == "" {
		regionEndpoint = ""
	}

	encryptBool := false
	encrypt, ok := parameters["encrypt"]
	if ok {
		encryptBool, err = strconv.ParseBool(encrypt)
		if err != nil {
			return nil, fmt.Errorf("The encrypt parameter should be a boolean")
		}
	}

	secureBool := true
	secure, ok := parameters["secure"]
	if ok {
		secureBool, err = strconv.ParseBool(secure)
		if err != nil {
			return nil, fmt.Errorf("The secure parameter should be a boolean")
		}
	}

	chunkSize := int64(defaultChunkSize)
	chunkSizeParam, ok := parameters["chunksize"]
	if ok {
		vv, err := strconv.ParseInt(chunkSizeParam, 0, 64)
		if err != nil {
			return nil, fmt.Errorf("chunksize parameter must be an integer, %v invalid", chunkSizeParam)
		}
		chunkSize = vv

		if chunkSize < minChunkSize {
			return nil, fmt.Errorf("The chunksize %#v parameter should be a number that is larger than or equal to %d", chunkSize, minChunkSize)
		}
	}

	rootDirectory, ok := parameters["rootdirectory"]
	if !ok {
		rootDirectory = ""
	}

	storageClass := s3.StorageClassStandard
	storageClassParam, ok := parameters["storageclass"]
	if ok {
		storageClassString := storageClassParam

		// All valid storage class parameters are UPPERCASE, so be a bit more flexible here
		storageClassString = strings.ToUpper(storageClassString)
		if storageClassString != s3.StorageClassStandard && storageClassString != s3.StorageClassReducedRedundancy {
			return nil, fmt.Errorf("The storageclass parameter must be one of %v, %v invalid", []string{s3.StorageClassStandard, s3.StorageClassReducedRedundancy}, storageClassParam)
		}
		storageClass = storageClassString
	}

	params := DriverParameters{
		accessKey,
		secretKey,
		bucket,
		regionEndpoint,
		region,
		encryptBool,
		secureBool,
		chunkSize,
		fmt.Sprint(rootDirectory),
		storageClass,
	}

	return New(params)
}

// New constructs a new Driver with the given AWS credentials, region, encryption flag, and
// bucketName
func New(params DriverParameters) (*Driver, error) {
	awsConfig := aws.NewConfig()
	var creds *credentials.Credentials
	if params.RegionEndpoint == "" {
		creds = credentials.NewChainCredentials([]credentials.Provider{
			&credentials.StaticProvider{
				Value: credentials.Value{
					AccessKeyID:     params.AccessKey,
					SecretAccessKey: params.SecretKey,
				},
			},
			&credentials.EnvProvider{},
			&credentials.SharedCredentialsProvider{},
			&ec2rolecreds.EC2RoleProvider{Client: ec2metadata.New(session.New())},
		})

	} else {
		creds = credentials.NewChainCredentials([]credentials.Provider{
			&credentials.StaticProvider{
				Value: credentials.Value{
					AccessKeyID:     params.AccessKey,
					SecretAccessKey: params.SecretKey,
				},
			},
			&credentials.EnvProvider{},
		})
		awsConfig.WithS3ForcePathStyle(true)
		awsConfig.WithEndpoint(params.RegionEndpoint)
	}

	awsConfig.WithCredentials(creds)
	awsConfig.WithRegion(params.Region)
	awsConfig.WithDisableSSL(!params.Secure)

	s3obj := s3.New(session.New(awsConfig))
	_, err := s3obj.CreateBucket(&s3.CreateBucketInput{
		Bucket: &params.Bucket,
	})
	if err != nil {
		if s3err, ok := err.(awserr.Error); ok {
			if s3err.Code() != "BucketAlreadyOwnedByYou" && s3err.Code() != "BucketAlreadyExists" {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	if err := s3obj.WaitUntilBucketExists(&s3.HeadBucketInput{Bucket: &params.Bucket}); err != nil {
		return nil, err
	}

	d := &driver{
		S3:            s3obj,
		Bucket:        params.Bucket,
		ChunkSize:     params.ChunkSize,
		Encrypt:       params.Encrypt,
		RootDirectory: params.RootDirectory,
		StorageClass:  params.StorageClass,
		zeros:         make([]byte, params.ChunkSize),
	}

	d.pool.New = func() interface{} {
		return make([]byte, d.ChunkSize)
	}

	return &Driver{
		baseEmbed: baseEmbed{
			Base: base.Base{
				StorageDriver: d,
			},
		},
	}, nil
}

// Implement the storagedriver.StorageDriver interface

func (d *driver) Name() string {
	return driverName
}

func (d *driver) CheckConnectionStatus() (bool, error) {
	_, err := d.S3.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		return false, err
	}
	return true, err
}

// GetContent retrieves the content stored at "path" as a []byte.
func (d *driver) GetContent(path string) ([]byte, error) {
	reader, err := d.ReadStream(path, 0)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(reader)
}

// PutContent stores the []byte content at a location designated by "path".
func (d *driver) PutContent(path string, contents []byte) error {
	_, err := d.S3.PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(d.Bucket),
		Key:                  aws.String(d.s3Path(path)),
		ContentType:          d.getContentType(),
		ACL:                  d.getACL(),
		ServerSideEncryption: d.getEncryptionMode(),
		StorageClass:         d.getStorageClass(),
		Body:                 bytes.NewReader(contents),
	})
	return parseError(path, err)
}

// ReadStream retrieves an io.ReadCloser for the content stored at "path" with a
// given byte offset.
func (d *driver) ReadStream(path string, offset int64) (io.ReadCloser, error) {
	resp, err := d.S3.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(d.Bucket),
		Key:    aws.String(d.s3Path(path)),
		Range:  aws.String("bytes=" + strconv.FormatInt(offset, 10) + "-"),
	})

	if err != nil {
		if s3Err, ok := err.(awserr.Error); ok && s3Err.Code() == "InvalidRange" {
			return ioutil.NopCloser(bytes.NewReader(nil)), nil
		}

		return nil, parseError(path, err)
	}
	return resp.Body, nil
}

// Stat retrieves the FileInfo for the given path, including the current size
// in bytes and the creation time.
func (d *driver) Stat(path string) (storagedriver.FileInfo, error) {
	resp, err := d.S3.ListObjects(&s3.ListObjectsInput{
		Bucket:  aws.String(d.Bucket),
		Prefix:  aws.String(d.s3Path(path)),
		MaxKeys: aws.Int64(1),
	})
	if err != nil {
		return nil, err
	}

	fi := storagedriver.FileInfoFields{
		Path: path,
	}

	if len(resp.Contents) == 1 {
		if *resp.Contents[0].Key != d.s3Path(path) {
			fi.IsDir = true
		} else {
			fi.IsDir = false
			fi.Size = *resp.Contents[0].Size
			fi.ModTime = *resp.Contents[0].LastModified
		}
	} else if len(resp.CommonPrefixes) == 1 {
		fi.IsDir = true
	} else {
		return nil, storagedriver.PathNotFoundError{Path: path}
	}

	return storagedriver.FileInfoInternal{FileInfoFields: fi}, nil
}

func (d *driver) s3Path(path string) string {
	return strings.TrimLeft(strings.TrimRight(d.RootDirectory, "/")+path, "/")
}

func parseError(path string, err error) error {
	if s3Err, ok := err.(awserr.Error); ok && s3Err.Code() == "NoSuchKey" {
		return storagedriver.PathNotFoundError{Path: path}
	}

	return err
}

func (d *driver) getEncryptionMode() *string {
	if d.Encrypt {
		return aws.String("AES256")
	}
	return nil
}

func (d *driver) getContentType() *string {
	return aws.String("application/octet-stream")
}

func (d *driver) getACL() *string {
	return aws.String("private")
}

func (d *driver) getStorageClass() *string {
	return aws.String(d.StorageClass)
}
