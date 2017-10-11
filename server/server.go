// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"strings"
)

// организует канал связи с посетителем
type chatChannel struct {
	conn    net.Conn
	reader  *bufio.Reader
	writer  *bufio.Writer
	inChan  chan string
	outChan chan<- string
}

func (c *chatChannel) sendAnswer(str string) {

	n, err := c.writer.WriteString(str + "\n")
	if err != nil {
		fmt.Print(err)
	}

	c.writer.Flush()

	if *fDebug {
		fmt.Print(n)
	}
}

// cnstructor
func newChatChannel(conn net.Conn, outChan chan<- string) *chatChannel {

	return &chatChannel{
		conn:    conn,
		reader:  bufio.NewReader(conn),
		writer:  bufio.NewWriter(conn),
		inChan:  make(chan string),
		outChan: outChan,
	}
}

func (c *chatChannel) handle() {

	go func() {
		for str := range c.inChan {

			c.sendAnswer(str)
		}
	}()
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

			if *fDebug {
				fmt.Println(str)
			}
			c.outChan <- str
		}
	}
}

var (
	fPort     = flag.String("port", ":8080", "host address to listen on")
	fPortFile = flag.String("portFile", ":2121", "host port for file transfer")
	fDebug    = flag.Bool("debug", true, "debug mode")
)

func main() {
	flag.Parse()

	// созджаем систему для рассылки сообщений
	broadcast := make([]*chatChannel, 0)
	chanBroadcast := make(chan string)

	go func() {
		for {
			str := <-chanBroadcast
			for i, ch := range broadcast {
				if *fDebug {
					fmt.Print(i)
				}
				ch.inChan <- str
			}
		}
	}()

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

		broadcast = append(broadcast, c)

	}
}
