package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/golang-demos/chalk"
)

type WriteCounter struct {
	Total     uint64
	Completed uint64
	LastPrint time.Time
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Completed += uint64(n)
	wc.PrintProgress()
	return n, nil
}

func (wc *WriteCounter) PrintProgress() {
	if time.Since(wc.LastPrint) < 100*time.Millisecond {
		return
	}

	wc.LastPrint = time.Now()
	fmt.Printf("\033[2K\rDownloading... %s%s/%s%s", chalk.Bold(), humanize.Bytes(wc.Completed), humanize.Bytes(wc.Total), chalk.Reset())
}

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

			split := strings.Split(string(folder), "/")
			fmt.Println(strings.Repeat("    ", len(split)-1), "-", split[len(split)-1])

			os.Mkdir(filepath.Join(output, filepath.FromSlash(string(folder))), os.ModePerm)

		case OP_FILE:
			file := make([]byte, length)
			_, err = conn.Read(file)
			if err != nil {
				panic(err)
			}

			split := strings.Split(string(file), "/")
			fmt.Println(strings.Repeat("    ", len(split)-1), "-", split[len(split)-1])

			f, err := os.Create(filepath.Join(output, filepath.FromSlash(string(file))))
			if err != nil {
				panic(err)
			}

			dst = f

		case OP_DATA:
			counter := &WriteCounter{
				Total:     uint64(length),
				Completed: 0,
				LastPrint: time.Now(),
			}

			io.CopyN(dst, io.TeeReader(conn, counter), int64(length))
			fmt.Print("\033[2K\r")

		case OP_FINISH:
			fmt.Println("Finished")
			return
		}
	}
}
