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

	// Defined Tags
	tags := []types.Tag{
		{
			Key:   aws.String("Name"),
			Value: aws.String("Sagar"),
		},
		{
			Key:   aws.String("ENV"),
			Value: aws.String("Learning"),
		},
	}
	tagSpecifications := []types.TagSpecification{
		{
			ResourceType: types.ResourceTypeInstance,
			Tags:         tags,
		},
	}

	// Volume variables
	volumeSize := int32(11) // Adjust the size as needed
	volumeType := types.VolumeTypeGp2
	availabilityZone := "us-east-1a"

	// Describe KeyPair
	sshKey, err := ec2Client.DescribeKeyPairs(ctx, &ec2.DescribeKeyPairsInput{
		KeyNames: []string{"sagar"},
	})

	// Creating Key Pair.
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "InvalidKeyPair.NotFound" {
				// Key pair doesn't exist, create it
				key, err := ec2Client.CreateKeyPair(ctx, &ec2.CreateKeyPairInput{
					KeyName: aws.String("sagar"),
				})
				if err != nil {
					return "", fmt.Errorf("Errors while creating :  %s", err)
				}
				err = os.WriteFile("sagar.pem", []byte(*key.KeyMaterial), 0600)
				if err != nil {
					return "", fmt.Errorf("Write KeyPair error: %s", err)
				}
			} else if awsErr.Code() == "InvalidKeyPair.Duplicate" {
				// Handle other error cases
				log.Println("Key pair already exist")
			} else {
				return "", fmt.Errorf("Some error has occured while creating key: %s", err)
			}
		}
	}
	if sshKey == nil {
		sshKey, err := ec2Client.CreateKeyPair(ctx, &ec2.CreateKeyPairInput{
			KeyName: aws.String("sagar"),
		})
		if err != nil {
			return "", fmt.Errorf("Write sshKey error: %s", err)
		}

		if err = os.WriteFile("sagar.pem", []byte(*sshKey.KeyMaterial), 0600); err != nil {
			return "", fmt.Errorf("Write KeyPair error:  %s", err)
		}
	}

	// if len(keyPair.KeyPairs) == 0 {
	// 	key, err := ec2Client.CreateKeyPair(ctx, &ec2.CreateKeyPairInput{
	// 		KeyName: aws.String("sagar"),
	// 	})
	// 	if err != nil {
	// 		return "", fmt.Errorf("Create KeyPair error:  %s", err)
	// 	}
	// 	file := os.WriteFile("sagar.pem", []byte(*key.KeyMaterial), 0600)
	// 	if file != nil {
	// 		return "", fmt.Errorf("Write KeyPair error:  %s", file)
	// 	}
	// }

	// Selecting Image for the Instance
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

	//  EBS block size configuration and mapping.
	blockDeviceMappings := []types.BlockDeviceMapping{
		{
			DeviceName: aws.String("/dev/sda1"),
			Ebs: &types.EbsBlockDevice{
				DeleteOnTermination: aws.Bool(true),
				VolumeSize:          aws.Int32(volumeSize),
				VolumeType:          volumeType,
			},
		},
	}

	// Selecting Instance type
	instance, err := ec2Client.RunInstances(ctx, &ec2.RunInstancesInput{
		ImageId:      imageOutput.Images[0].ImageId,
		KeyName:      aws.String("sagar"),
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
		InstanceType: types.InstanceTypeT2Micro,
		Placement: &types.Placement{
			AvailabilityZone: aws.String(availabilityZone),
		},
		BlockDeviceMappings: blockDeviceMappings,
		TagSpecifications:   tagSpecifications,
	})

	if err != nil {
		return "", fmt.Errorf("Run instance error error:  %s", err)
	}
	if len(instance.Instances) == 0 {
		return "", fmt.Errorf("Instance output is empty ")
	}

	return *instance.Instances[0].InstanceId, nil
}
