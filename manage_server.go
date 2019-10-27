package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/pkg/sftp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ssm"
	"golang.org/x/crypto/ssh"
)

const startupScript = `#!/bin/bash
				       cd /home/ec2-user/
					   yum update -y
					   yum install java-1.8.0-openjdk-headless.x86_64 -y
					   runuser -l ec2-user -c 'wget https://launcher.mojang.com/v1/objects/3dc3d84a581f14691199cf6831b71ed1296a9fdf/server.jar -O minecraft.jar'
					   runuser -l ec2-user -c 'java -Xms512M -Xmx1024M -jar minecraft.jar nogui'
					   sed -i -e s/eula=false/eula=true/g eula.txt
					   runuser -l ec2-user -c 'screen -dmS minecraft java -Xmx1024M -Xms512M -jar minecraft.jar nogui'`

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
		SecurityGroups: []*string{aws.String(tagValue), aws.String("allow-ssh")},
		KeyName:        aws.String("myFirstKey"),
		UserData:       aws.String(base64.StdEncoding.EncodeToString([]byte(startupScript))),
	})
	if err != nil {
		fmt.Println("Could not create instance:", err)
		return
	}
	mcInstanceId := *runResult.Instances[0].InstanceId

	fmt.Println("Waiting for instance to boot...")
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

	fmt.Printf("Created instance: %s\n", *minecraftInstance.Reservations[0].Instances[0].InstanceId)

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

func stop(session *session.Session, tagKey, tagValue string) {
	svc := ec2.New(session)
	instancesToDelete := fetchRunningInstancesByTag(session, tagKey, tagValue)
	if len(instancesToDelete) == 0 {
		fmt.Println("No instances to delete")
		return
	}

	var instanceIds []string
	for _, instance := range instancesToDelete {
		instanceIds = append(instanceIds, *instance.InstanceId)
	}

	terminateOutput, err := svc.TerminateInstances(&ec2.TerminateInstancesInput{
		InstanceIds: aws.StringSlice(instanceIds),
	})
	if err != nil {
		panic(err)
	}
	for _, instanceStateChange := range terminateOutput.TerminatingInstances {
		fmt.Printf("Instance %s %s\n", *instanceStateChange.InstanceId, *instanceStateChange.CurrentState.Name)
	}
}

func downloadWorld() error {
	/*
		config := &ssh.ClientConfig{
			User: "ec2-user",
			Auth: []ssh.AuthMethod{
				publicKey("myFirstKey"),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
	*/
	conn, err := ssh.Dial("tcp", "52.209.252.105:22", sshConfig)
	if err != nil {
		panic(err)
	}

	sftpClient, err := sftp.NewClient(conn)
	if err != nil {
		panic(err)
	}
	defer sftpClient.Close()

	err = copyRemoteDirToLocal("/home/ec2-user/world", "world", sftpClient)
	if err != nil {
		return fmt.Errorf("Could not download world %v", err)
	}
	/*err = copyLocalDirToRemote("cake", "/home/ec2-user", sftpClient)
	if err != nil {
		return fmt.Errorf("Could not upload dir %v", err)
	}*/
	return nil
}

func publicKey(path string) ssh.AuthMethod {
	key, err := ioutil.ReadFile("/home/jonas/.ssh/myFirstKey.pem")
	if err != nil {
		panic(err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		panic(err)
	}
	return ssh.PublicKeys(signer)
}

func stopMcServer() {
	conn, err := ssh.Dial("tcp", "52.214.47.40:22", sshConfig)
	if err != nil {
		panic(err)
	}

	sess, err := conn.NewSession()
	if err != nil {
		panic(err)
	}

	err = sess.Run(`screen -S minecraft -p 0 -X stuff "stop^M"`)
	if err != nil {
		panic(err)
	}
}

func ssmStopMcServer(session *session.Session, instanceId string) {
	svc := ssm.New(session)

	sendcommandOutput, err := svc.SendCommand(&ssm.SendCommandInput{
		DocumentName: aws.String("AWS-RunShellScript"),
		Parameters: map[string][]*string{"commands": []*string{
			aws.String(`screen -S minecraft -p 0 -X stuff "stop^M"`),
		}},
		InstanceIds: []*string{aws.String(instanceId)},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(*sendcommandOutput)
}
