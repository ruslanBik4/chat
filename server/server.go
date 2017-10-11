// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"net"
	"fmt"
	"bufio"
	"io"
	"strings"
	"flag"
)
type chatChannel struct {
	conn net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
	inChan chan string
}
func (c *chatChannel) sendAnswer( str string) {

	_, err := c.writer.WriteString(str + "\n")
	if err != nil {
		fmt.Print(err)
	}

	c.writer.Flush()

}
func newChatChannel(conn net.Conn, inChan chan string) *chatChannel {

	return &chatChannel{
		conn: conn,
		reader: bufio.NewReader(conn),
		writer: bufio.NewWriter(conn),
		inChan: inChan,
	}
}

func (c *chatChannel) handle(){
	for {
		str, err := c.reader.ReadString('\n')
		if err != nil {
			// завершение работы потока
			if err == io.EOF {
				fmt.Println(" ends of bytes from stream")
				break
			}
			fmt.Print(err)
			continue
		}

		str = strings.Trim(str, "\n\r ")
		switch str {
		case ":list":
			c.sendAnswer(fileList())
			c.sendAnswer("Вы можете скачать файл, введя команду :file")
		default:

			fmt.Println(str)
			c.inChan <- str
		}
	}
}
var (
	fPort    = flag.String("port", ":8080", "host address to listen on")
	fDebug = flag.Bool("debug", true, "debug mode")
)
func main()  {
	flag.Parse()
	chanBroadcast := make(chan string)

	go startFileServer()
	ln, err := net.Listen("tcp", *fPort)
	if err != nil {
		// handle error
		panic(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
			fmt.Println(err)
			continue
		}
		// запускаем отдельный поток для каждого соединения чата
		c := newChatChannel(conn, chanBroadcast)
		go c.handle()

	}
}