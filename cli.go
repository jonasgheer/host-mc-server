package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

func main() {
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
		ipAddress = start(sess, "minecraft")
		fmt.Printf("ip: %s\n", ipAddress)
	case "stop":
		stop(sess, "minecraft")
	case "test":
		test(sess)
	default:
		fmt.Printf("'%s' is not a command, type 'help'\n", command)
	}
}
