package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-demos/chalk"
)

type Conn struct {
	net.Conn
}

const (
	OP_FOLDER = byte(iota)
	OP_FILE   = byte(iota)
	OP_DATA   = byte(iota)
	OP_FINISH = byte(iota)
)

var cwd, _ = os.Getwd()

func (c *Conn) WriteString(s string) (int, error) {
	return c.Write([]byte(s))
}

func (c *Conn) WriteHeader(op byte, l uint32) error {
	if _, err := c.Write([]byte{op}); err != nil {
		return err
	}

	lbuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lbuf, l)
	if _, err := c.Write(lbuf); err != nil {
		return err
	}

	return nil
}

func (c *Conn) ParseHeader() (byte, uint32, error) {
	op := make([]byte, 1)
	if _, err := c.Read(op); err != nil {
		return 0, 0, err
	}

	header := make([]byte, 4)
	if _, err := c.Read(header); err != nil {
		return 0, 0, err
	}

	return op[0], binary.BigEndian.Uint32(header), nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println(" fileshare send [...files]")
		fmt.Println(" fileshare recv [code] [output]")
		return
	}

	cmd := os.Args[1]
	if cmd == "send" || cmd == "s" {
		l, err := net.Listen("tcp", ":1234")
		if err != nil {
			panic(err)
		}

		// print local ip address
		addrs, err := net.InterfaceAddrs()
		if err != nil {
			panic(err)
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					fmt.Printf(
						"on the receiver, run %sfileshare receive %s [output]%s\n",
						chalk.Bold(), strings.TrimPrefix(ipnet.IP.String(), "192.168."), chalk.Reset(),
					)
				}
			}
		}

		for {
			// Wait for a connection.
			netConn, err := l.Accept()
			if err != nil {
				panic(err)
			}

			conn := &Conn{netConn}

			fmt.Println("received connection")
			go func(conn *Conn) {
				for _, path := range os.Args[2:] {
					info, err := os.Stat(path)
					if err != nil {
						panic(err)
					}

					if info.IsDir() {
						fs.WalkDir(os.DirFS(cwd), path, func(path string, entry fs.DirEntry, err error) error {
							if err != nil {
								panic(err)
							}

							if entry.IsDir() {
								conn.WriteHeader(OP_FOLDER, uint32(len(path)))
								conn.WriteString(path)
							} else {
								conn.WriteHeader(OP_FILE, uint32(len(path)))
								conn.WriteString(path)

								info, err := entry.Info()
								if err != nil {
									panic(err)
								}

								conn.WriteHeader(OP_DATA, uint32(info.Size()))
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
						conn.WriteHeader(OP_FILE, uint32(len(path)))
						conn.WriteString(path)

						conn.WriteHeader(OP_DATA, uint32(info.Size()))
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
			}(conn)
		}
	} else if cmd == "receive" || cmd == "recv" || cmd == "r" {
		ip := os.Args[2]
		if strings.Count(ip, ".") != 3 {
			ip = "192.168." + ip
		}

		output := os.Args[3]
		os.MkdirAll(output, os.ModePerm)

		netConn, err := net.Dial("tcp", ip+":1234")
		if err != nil {
			panic(err)
		}

		fmt.Println("connected")

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

				fmt.Println("folder", string(folder))
				os.Mkdir(filepath.Join(output, string(folder)), os.ModePerm)

			case OP_FILE:
				file := make([]byte, length)
				_, err = conn.Read(file)
				if err != nil {
					panic(err)
				}

				fmt.Println("file", string(file))
				f, err := os.Create(filepath.Join(output, string(file)))
				if err != nil {
					panic(err)
				}

				dst = f

			case OP_DATA:
				fmt.Println("data", length, "bytes")
				io.CopyN(dst, conn, int64(length))
				fmt.Println("done")

			case OP_FINISH:
				fmt.Println("finished")
				return
			}
		}
	}
}
