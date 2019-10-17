package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

func main() {
	fmt.Println("MC server tool")
	var ipAddress string
	reader := bufio.NewReader(os.Stdin)
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-west-1"),
	})
	if err != nil {
		panic(err)
	}
	for {
		fmt.Print("> ")
		cmdString, err := reader.ReadString('\n')
		cmdString = strings.TrimSpace(cmdString)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		switch cmdString {
		case "help":
			fmt.Println("The commands are:\n",
				"\tstart-server			spawns server if needed and starts up\n",
				"\tsave-world			downloads the minecraft world and saves in current folder\n",
				"\tterminate-server		stops and destroys the server")
		case "start-server":
			ipAddress = startServer(sess, "minecraft")
			fmt.Printf("ip: %s\n", ipAddress)
		case "terminate-server":
			terminateServers(sess, "minecraft")
		default:
			fmt.Printf("'%s' is not a command, type 'help'\n", cmdString)
		}
	}
}
