package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go/aws/awserr"
)

func main() {
	var (
		instanceId string
		err        error
	)
	ctx := context.Background()
	instanceId, err = createEc2(ctx, "us-east-1", "sagar")
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	fmt.Printf("Instance ID is: %s", instanceId)
}

func createEc2(ctx context.Context, region string, profile string) (string, error) {

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region), config.WithSharedConfigProfile(profile))
	if err != nil {
		return "", fmt.Errorf("unable to load SDK config, %v", err)
	}

	ec2Client := ec2.NewFromConfig(cfg)

	key, err := ec2Client.CreateKeyPair(ctx, &ec2.CreateKeyPairInput{
		KeyName: aws.String("sagar"),
	})
	if err != nil {
		return "", fmt.Errorf("Create KeyPair error:  %s", err)
	}

	err = os.WriteFile("sagar.pem", []byte(*key.KeyMaterial), 0600)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "InvalidKeyPair.NotFound" {
				return "", fmt.Errorf("error in creating key file, %s", awsErr)
			}
		}
		return "", fmt.Errorf("Write KeyPair error:  %s", err)
	}

	imageOutput, err := ec2Client.DescribeImages(ctx, &ec2.DescribeImagesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("name"),
				Values: []string{"ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"},
			},
			{
				Name:   aws.String("virtualization-type"),
				Values: []string{"hvm"},
			},
		},
		Owners: []string{"099720109477"},
	})
	if err != nil {
		return "", fmt.Errorf("Describe Image error:  %s", err)
	}

	if len(imageOutput.Images) == 0 {
		return "", fmt.Errorf("Image output is empty ")
	}

	instance, err := ec2Client.RunInstances(ctx, &ec2.RunInstancesInput{
		ImageId:      imageOutput.Images[0].ImageId,
		KeyName:      aws.String("sagar"),
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
		InstanceType: types.InstanceTypeT2Micro,
	})
	if err != nil {
		return "", fmt.Errorf("Run instance error error:  %s", err)
	}
	if len(instance.Instances) == 0 {
		return "", fmt.Errorf("Instance output is empty ")
	}

	return *instance.Instances[0].InstanceId, nil
}
