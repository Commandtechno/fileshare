package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
)

func recv(ip string, output string) {
	if strings.Count(ip, ".") != 3 {
		ip = "192.168." + ip
	}

	os.MkdirAll(output, os.ModePerm)

	netConn, err := net.Dial("tcp", ip+":1234")
	if err != nil {
		panic(err)
	}

	fmt.Println("Connected")

	conn := &Conn{netConn}
	var dst *os.File

	for {
		op, length, err := conn.ParseHeader()
		if err != nil {
			panic(err)
		}

		switch op {
		case OP_FOLDER:
			folder := make([]byte, length)
			_, err = conn.Read(folder)
			if err != nil {
				panic(err)
			}

			os.Mkdir(filepath.Join(output, string(folder)), os.ModePerm)

		case OP_FILE:
			file := make([]byte, length)
			_, err = conn.Read(file)
			if err != nil {
				panic(err)
			}

			f, err := os.Create(filepath.Join(output, string(file)))
			if err != nil {
				panic(err)
			}

			dst = f

		case OP_DATA:
			io.CopyN(dst, conn, int64(length))

		case OP_FINISH:
			fmt.Println("Finished")
			return
		}
	}
}
