package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func startServer() {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-west-1"),
	})
	if err != nil {
		panic(err)
	}

	svc := ec2.New(sess)

	// check if minecraft instance already exists
	result, err := svc.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				//Name:   aws.String("tag-key"),
				//Values: []*string{aws.String("minecraft")},
				Name:   aws.String("tag:Name"),
				Values: []*string{aws.String("My first instance")},
			},
		},
	})
	if err != nil {
		fmt.Println("Error", err)
	} else {
		for _, reservation := range result.Reservations {
			for _, instance := range reservation.Instances {
				for _, tag := range instance.Tags {
					if *tag.Key == "Name" && *tag.Value == "My first instance" {
						// panics if instance does not have public ip
						fmt.Printf("Instance already exists: %s", *instance.PublicIpAddress)

					}
				}
			}
		}

		//fmt.Println("Success", result.Reservations[0].Instances[0])
		//fmt.Println(len(result.Reservations))
	}
	os.Exit(0)
	/*
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

		fmt.Println("Created instance:", *runResult.Instances[0].InstanceId)

		_, errtag := svc.CreateTags(&ec2.CreateTagsInput{
			Resources: []*string{runResult.Instances[0].InstanceId},
			Tags: []*ec2.Tag{
				{
					Key:   aws.String("Name"),
					Value: aws.String("minecraft"),
				},
			},
		})
		if errtag != nil {
			log.Println("Could not create tags for instance:", runResult.Instances[0].InstanceId, errtag)
			return
		}

		fmt.Println("Successfully tagged instance")
	*/
}
