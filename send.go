package main

import (
	"fmt"
	"io"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-demos/chalk"
)

func send(files []string) {
	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		panic(err)
	}

	localIp := getLocalIp()
	if localIp == "" {
		panic("Could not find local ip")
	}

	fmt.Printf(
		"On the receiver, run %sfileshare receive %s [output]%s\n",
		chalk.Bold(), strings.TrimPrefix(localIp, "192.168."), chalk.Reset(),
	)

	for {
		// Wait for a connection.
		netConn, err := listener.Accept()
		if err != nil {
			panic(err)
		}

		conn := &Conn{netConn}

		fmt.Println("Received connection")
		go func(conn *Conn) {
			for _, path := range files {
				info, err := os.Stat(path)
				if err != nil {
					panic(err)
				}

				if info.IsDir() {
					fs.WalkDir(os.DirFS(cwd), filepath.Clean(path), func(path string, entry fs.DirEntry, err error) error {
						if err != nil {
							panic(err)
						}

						if entry.IsDir() {
							conn.WriteHeader(OP_FOLDER, uint64(len(path)))
							conn.WriteString(filepath.ToSlash(path))
						} else {
							conn.WriteHeader(OP_FILE, uint64(len(path)))
							conn.WriteString(filepath.ToSlash(path))

							info, err := entry.Info()
							if err != nil {
								panic(err)
							}

							conn.WriteHeader(OP_DATA, uint64(info.Size()))
							file, err := os.Open(path)
							if err != nil {
								panic(err)
							}

							if _, err := io.Copy(conn, file); err != nil {
								panic(err)
							}
						}
						return nil
					})
				} else {
					conn.WriteHeader(OP_FILE, uint64(len(path)))
					conn.WriteString(path)

					conn.WriteHeader(OP_DATA, uint64(info.Size()))
					file, err := os.Open(path)
					if err != nil {
						panic(err)
					}

					if _, err := io.Copy(conn, file); err != nil {
						panic(err)
					}
				}
			}

			conn.WriteHeader(OP_FINISH, 0)
			conn.Close()
			fmt.Println("Finished sending files")
		}(conn)
	}
}
