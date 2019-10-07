package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	fmt.Println("MC server tool")
	reader := bufio.NewReader(os.Stdin)
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
			startServer()
			/*
				case "save-world":
					saveWorld()
				case "terminate-server":
					terminateServer()
			*/
		default:
			fmt.Printf("'%s' is not a command, type 'help'\n", cmdString)
		}
	}
}
