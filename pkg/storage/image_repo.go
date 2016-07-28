package storage

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
)

// CreateImageRepo create a repository for the image on amazon's ECR(EC2 Container Repository)
// if it doesn't exist as repository needs to be present before pushing and image into it.
func CreateImageRepo(reponame string, params map[string]string) error {

	accessKey, ok := params["accesskey"]
	if !ok {
		accessKey = ""
	}
	secretKey, ok := params["secretkey"]
	if !ok {
		secretKey = ""
	}
	regionName, ok := params["region"]
	if !ok || fmt.Sprint(regionName) == "" {
		return fmt.Errorf("No region parameter provided")
	}
	region := fmt.Sprint(regionName)
	creds := credentials.NewChainCredentials([]credentials.Provider{
		&credentials.StaticProvider{
			Value: credentials.Value{
				AccessKeyID:     fmt.Sprint(accessKey),
				SecretAccessKey: fmt.Sprint(secretKey),
			},
		},
		&credentials.EnvProvider{},
		&credentials.SharedCredentialsProvider{},
		&ec2rolecreds.EC2RoleProvider{Client: ec2metadata.New(session.New())},
	})
	awsConfig := aws.NewConfig()
	awsConfig.WithCredentials(creds)
	awsConfig.WithRegion(region)
	svc := ecr.New(session.New(awsConfig))

	repoInput := &ecr.CreateRepositoryInput{
		RepositoryName: aws.String(reponame),
	}

	_, err := svc.CreateRepository(repoInput)
	if err != nil {
		if s3Err, ok := err.(awserr.Error); ok && s3Err.Code() == "RepositoryAlreadyExistsException" {
			return nil
		}
		return err
	}
	return nil
}
