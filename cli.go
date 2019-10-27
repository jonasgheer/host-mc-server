package main

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"golang.org/x/crypto/ssh"
)

var sshConfig = &ssh.ClientConfig{
	User: "ec2-user",
	Auth: []ssh.AuthMethod{
		publicKey("myFirstKey"),
	},
	HostKeyCallback: ssh.InsecureIgnoreHostKey(),
}

func main() {
	const (
		tagKey   = "Name"
		tagValue = "minecraft"
	)
	command := os.Args[1]
	var ipAddress string
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-west-1"),
	})
	if err != nil {
		panic(err)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	switch command {
	case "help":
		fmt.Println("The commands are:\n",
			"\tstart-server			spawns server if needed and starts up\n",
			"\tsave-world			downloads the minecraft world and saves in current folder\n",
			"\tterminate-server		stops and destroys the server")
	case "start":
		ipAddress = start(sess, tagKey, tagValue)
		fmt.Printf("ip: %s\n", ipAddress)
	case "stop":
		instances := fetchRunningInstancesByTag(sess, tagKey, tagValue)
		if len(instances) == 0 {
			fmt.Println("No minecraft instances to shut down")
			os.Exit(1)
		} else if len(instances) > 1 {
			fmt.Println("More than one minecraft instance running. Check AWS")
			os.Exit(1)
		}
		fmt.Println(*instances[0].InstanceId)
		fmt.Println(len(instances))
		stopMcServer()
		time.Sleep(5 * time.Second) // give 5 sec for server to save world
		err := downloadWorld()
		if err != nil {
			fmt.Println("Could not download world, canceling instance termination")
			os.Exit(1)
		}
		//stop(sess, tagKey, tagValue)
	case "download":
		downloadWorld()
	case "fetch":
		instances := fetchRunningInstancesByTag(sess, tagKey, tagValue)
		fmt.Println(len(instances))
	default:
		fmt.Printf("'%s' is not a command, type 'help'\n", command)
	}
}
