package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func main() {
	var (
		instanceId string
		err        error
	)
	if instanceId, err = createEC2(context.Background(), "us-east-1"); err != nil {
		fmt.Printf("Error: %s", err)
		os.Exit(1)
	}

	fmt.Printf("instance id: %s\n", instanceId)
}

func createEC2(ctx context.Context, region string) (string, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return "", fmt.Errorf("LoadDefaultConfig error: %s", err)
	}

	ec2Client := ec2.NewFromConfig(cfg)

	existingKeyPairs, err := ec2Client.DescribeKeyPairs(ctx, &ec2.DescribeKeyPairsInput{
		KeyNames: []string{"go-aws-ec2"},
		// or:
		//Filters: []types.Filter{
		//	{
		//		Name:   aws.String("key-name"),
		//		Values: []string{"go-aws-ec2"},
		//	},
		//},
	})
	if err != nil && !strings.Contains(err.Error(), "InvalidKeyPair.NotFound") {
		return "", fmt.Errorf("DescribeKeyPairs error: %s", err)
	}

	if existingKeyPairs == nil || len(existingKeyPairs.KeyPairs) == 0 {
		keyPair, err := ec2Client.CreateKeyPair(ctx, &ec2.CreateKeyPairInput{
			KeyName: aws.String("go-aws-ec2"),
		})
		if err != nil {
			return "", fmt.Errorf("CreateKeyPair error: %s", err)
		}

		err = os.WriteFile("go-aws-ec2.pem", []byte(*keyPair.KeyMaterial), 0600)
		if err != nil {
			return "", fmt.Errorf("WriteFile (keypair) error: %s", err)
		}
	}

	describeImages, err := ec2Client.DescribeImages(ctx, &ec2.DescribeImagesInput{
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
		Owners: []string{"099720109477"}, // see https://ubuntu.com/server/docs/cloud-images/amazon-ec2
	})
	if err != nil {
		return "", fmt.Errorf("DescribeImages error: %s", err)
	}
	if len(describeImages.Images) == 0 {
		return "", fmt.Errorf("describeImages has empty length (%d)", len(describeImages.Images))
	}
	runInstance, err := ec2Client.RunInstances(ctx, &ec2.RunInstancesInput{
		ImageId:      describeImages.Images[0].ImageId,
		InstanceType: types.InstanceTypeT3Micro,
		KeyName:      aws.String("go-aws-ec2"),
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
	})
	if err != nil {
		return "", fmt.Errorf("RunInstance error: %s", err)
	}

	if len(runInstance.Instances) == 0 {
		return "", fmt.Errorf("RunInstance has empty length (%d)", len(runInstance.Instances))
	}

	return *runInstance.Instances[0].InstanceId, nil
}
