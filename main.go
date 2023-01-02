package main

import (
	"fmt"
	"os"
)

const (
	OP_FOLDER = byte(iota)
	OP_FILE   = byte(iota)
	OP_DATA   = byte(iota)
	OP_FINISH = byte(iota)
)

var cwd, _ = os.Getwd()

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println(" fileshare send [...files]")
		fmt.Println(" fileshare recv [ip] [?output]")
		return
	}

	i, e := os.Stat(os.Args[2])
	fmt.Println(uint64(i.Size()), e)

	switch os.Args[1] {
	case "send", "s":
		if len(os.Args) < 3 {
			fmt.Println("No files specified")
			return
		}

		files := os.Args[2:]
		for _, file := range files {
			if _, err := os.Stat(file); err != nil {
				fmt.Println("File not found:", file)
				return
			}
		}

		send(files)
	case "receive", "recv", "r":
		if len(os.Args) < 3 {
			fmt.Println("No ip specified")
			return
		}

		ip := os.Args[2]
		output := "."
		if len(os.Args) > 3 {
			output = os.Args[3]
		}

		recv(ip, output)
	}
}
