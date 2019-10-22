package main

import (
	"encoding/base64"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ssm"
)

const startupScript = `#!/bin/bash
					   yum update -y
					   yum install java-1.8.0-openjdk-headless.x86_64 -y
					   wget https://launcher.mojang.com/v1/objects/3dc3d84a581f14691199cf6831b71ed1296a9fdf/server.jar -O minecraft.jar
					   java -Xmx1024M -Xms512M -jar minecraft.jar nogui
					   sed -i -e s/eula=false/eula=true/g eula.txt
					   screen -dmS minecraft java -Xmx1024M -Xms512M -jar minecraft.jar nogui`

func start(session *session.Session, tagKey, tagValue string) (ipAddress string) {
	svc := ec2.New(session)

	instances := fetchRunningInstancesByTag(session, tagKey, tagValue)

	if len(instances) > 0 {
		for _, instance := range instances {
			fmt.Println("Instance already exists")
			return *instance.PublicIpAddress
		}
	}

	runResult, err := svc.RunInstances(&ec2.RunInstancesInput{
		ImageId:        aws.String("ami-0ce71448843cb18a1"),
		InstanceType:   aws.String("t2.micro"),
		MinCount:       aws.Int64(1),
		MaxCount:       aws.Int64(1),
		SecurityGroups: []*string{aws.String(tagValue)},
		UserData:       aws.String(base64.StdEncoding.EncodeToString([]byte(startupScript))),
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
				Key:   aws.String(tagKey),
				Value: aws.String(tagValue),
			},
		},
	})
	if errtag != nil {
		log.Println("Could not create tags for instance:", runResult.Instances[0].InstanceId, errtag)
		return
	}

	fmt.Println("Successfully tagged instance")

	return *minecraftInstance.Reservations[0].Instances[0].PublicIpAddress
}

func fetchRunningInstancesByTag(session *session.Session, tagKey, tagValue string) []ec2.Instance {
	svc := ec2.New(session)
	var instances []ec2.Instance
	instancesInfo, err := svc.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String(fmt.Sprintf("tag:%s", tagKey)),
				Values: []*string{aws.String(tagValue)},
			},
		},
	})
	if err != nil {
		panic(err)
	} else {
		for _, reservation := range instancesInfo.Reservations {
			for _, instance := range reservation.Instances {
				if *instance.State.Name == "running" {
					instances = append(instances, *instance)
				}
			}
		}
	}
	return instances
}

func stop(session *session.Session, tagValue string) {
	svc := ec2.New(session)
	instancesToDelete := fetchRunningInstancesByTag(session, tagKey, tagValue)

	var instanceIds []string
	for _, instance := range instancesToDelete {
		instanceIds = append(instanceIds, *instance.InstanceId)
	}

	fmt.Println(instancesToDelete)
	terminateOutput, err := svc.TerminateInstances(&ec2.TerminateInstancesInput{
		InstanceIds: aws.StringSlice(instanceIds),
	})
	if err != nil {
		panic(err)
	}
	for _, instanceStateChange := range terminateOutput.TerminatingInstances {
		fmt.Printf("Instance %s %s\n", *instanceStateChange.InstanceId, *instanceStateChange.CurrentState)
	}
}

func test(session *session.Session) {
	svc := ssm.New(session)

	sendcommandOutput, err := svc.SendCommand(&ssm.SendCommandInput{
		DocumentName: aws.String("AWS-RunShellScript"),
		Parameters: map[string][]*string{"commands": []*string{
			aws.String("echo more-hello"),
			aws.String("echo hello-again"),
		}},
		InstanceIds: []*string{aws.String("i-0041d402a819f00b3")},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(*sendcommandOutput)
}
