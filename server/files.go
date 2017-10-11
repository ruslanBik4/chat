// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

type fileChannel struct {
	conn   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
}

func (fc *fileChannel) sendAnswer(str string) {

	_, err := fc.writer.WriteString(str + "\n")
	if err != nil {
		fmt.Print(err)
	}

	fc.writer.Flush()

}
func (fc *fileChannel) sendMessage(str string) {
	fc.sendAnswer(str)
}
func (fc *fileChannel) readMessage() (string, error) {
	str, err := fc.reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error by received: %#v\n", err)
		return "", err
	}

	return strings.TrimSuffix(str, "\n"), nil
}
func (fc *fileChannel) readFrom(file *os.File) (int64, error) {
	return fc.writer.ReadFrom(file)
}

const dirUploadFiles = "files"

func (fc *fileChannel) saveFile(fileName string) (n int64, err error) {
	writeFile, err := os.Create(path.Join(dirUploadFiles, fileName))

	if err != nil {
		return 0, err
	}
	return fc.reader.WriteTo(writeFile)
}
func (fc *fileChannel) Close() {
	fc.conn.Close()
}
func newFileChannel(conn net.Conn) *fileChannel {

	return &fileChannel{
		conn:   conn,
		reader: bufio.NewReader(conn),
		writer: bufio.NewWriter(conn),
	}
}

func fileList() string {
	files, err := filepath.Glob(dirUploadFiles + "/*")
	if err != nil {
		return err.Error()
	}

	list := ""
	for i, name := range files {

		shortName := strings.TrimPrefix(name, dirUploadFiles+"/")
		if strings.HasPrefix(shortName, ".") {
			continue
		}
		list += fmt.Sprintf("%d. %s\n", i, shortName)
	}
	return list
}

// прием и запись файла
func (fc *fileChannel) getFile() error {

	var n int64

	fc.sendAnswer("ready")
	// должно быть передано имя файла
	fileName, err := fc.readMessage()
	if err == nil {

		n, err = fc.saveFile(fileName)
	}
	if err != nil {
		return err
	}

	fmt.Printf("received bytes - %d", n)
	return nil
}

// отправка файла
func (fc *fileChannel) sendFile() error {
	fc.sendAnswer("ready")

	str, err := fc.readMessage()
	if err != nil {
		return err
	}
	fileNumber, err := strconv.Atoi(str)
	if err == nil {

		files, err := filepath.Glob(dirUploadFiles + "/*")
		if err != nil {
			return err
		}
		if fileNumber < len(files) {
			reader, err := os.Open(files[fileNumber])
			if err != nil {
				return err
			}

			defer reader.Close()
			fc.sendMessage(path.Base(reader.Name()))
			_, err = fc.readFrom(reader)
		} else {
			return fmt.Errorf("Number %d not in file list!", fileNumber)
		}
	}

	return err
}
func (fc *fileChannel) handle() {
	str, err := fc.readMessage()
	if err == nil {

		switch str {
		case "file:":
			err = fc.getFile()
		case ":file":
			err = fc.sendFile()
		}
		if err != nil {
			fmt.Println(err)
			fc.sendAnswer(err.Error())
		}
	}
}
func startFileServer() {
	ln, err := net.Listen("tcp", *fPortFile)
	if err != nil {
		// handle error
		panic(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
			panic(err)

		}
		// запускаем отдельный поток для каждого принимаемого файла
		fc := newFileChannel(conn)
		go fc.handle()

	}

}
