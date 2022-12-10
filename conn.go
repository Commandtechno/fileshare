package main

import (
	"encoding/binary"
	"net"
)

type Conn struct {
	net.Conn
}

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
