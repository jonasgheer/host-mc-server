package main

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func startServer(tagValue string) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-west-1"),
	})
	if err != nil {
		panic(err)
	}

	svc := ec2.New(sess)

	// check if minecraft instance already exists
	instancesInfo, err := svc.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("tag:Name"),
				Values: []*string{aws.String(tagValue)},
			},
		},
	})
	if err != nil {
		fmt.Println("Error", err)
	} else {
		for _, reservation := range instancesInfo.Reservations {
			for _, instance := range reservation.Instances {
				if *instance.State.Name == "terminated" {
					continue
				}
				for _, tag := range instance.Tags {
					if *tag.Key == "Name" && *tag.Value == tagValue {
						// panics if instance does not have public ip
						fmt.Printf("Instance already exists: %s\n", *instance.PublicIpAddress)
						return
					}
				}
			}
		}
	}

	runResult, err := svc.RunInstances(&ec2.RunInstancesInput{
		ImageId:      aws.String("ami-0ce71448843cb18a1"),
		InstanceType: aws.String("t2.micro"),
		MinCount:     aws.Int64(1),
		MaxCount:     aws.Int64(1),
	})
	if err != nil {
		fmt.Println("Could not create instance:", err)
		return
	}
	mcInstanceId := *runResult.Instances[0].InstanceId

	err = svc.WaitUntilInstanceRunning(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(mcInstanceId)},
	})
	if err != nil {
		panic(err)
	}

	minecraftInstance, err := svc.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(mcInstanceId)},
	})
	if err != nil {
		panic(err)
	}

	// does not exist yet, need to wait
	fmt.Println("Created instance:", *minecraftInstance.Reservations[0].Instances[0].PublicIpAddress)

	_, errtag := svc.CreateTags(&ec2.CreateTagsInput{
		Resources: []*string{runResult.Instances[0].InstanceId},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String(tagValue),
			},
		},
	})
	if errtag != nil {
		log.Println("Could not create tags for instance:", runResult.Instances[0].InstanceId, errtag)
		return
	}

	fmt.Println("Successfully tagged instance")
}
