// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"net"
	"bufio"
	"os"
	"path"
	"strings"
	"fmt"
	"path/filepath"
	"strconv"
)
type fileChannel struct {
	conn net.Conn
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
	fc.sendMessage(str)
}
func (fc *fileChannel) readMessage() (string, error) {
	str, err := fc.reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error by recieved: %#v\n", err)
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
	return fc.reader.WriteTo(writeFile)
}
func (fc *fileChannel) Close() {
	fc.conn.Close()
}
func newFileChannel(conn net.Conn) *fileChannel {

	return &fileChannel{
		conn: conn,
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

		shortName := strings.TrimPrefix(name, dirUploadFiles + "/")
		if strings.HasPrefix(shortName, ".") {
			continue
		}
		list += fmt.Sprintf("%d. %s\n", i, shortName )
	}
	return list
}
// прием и запись файла
func (fc *fileChannel) getFile()  error {

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

	fmt.Printf("recieved bytes - %d", n)
	return nil
}
// отправка файла
func (fc *fileChannel) sendFile() error{
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
		if fileNumber > len(files) {
			return err
		}
		reader, err := os.Open(files[fileNumber])
		if err != nil {
			return err
		}

		defer reader.Close()
		_, err = fc.readFrom(reader)
	}

	return err
}
func (fc *fileChannel) handle()  {
	str, err := fc.readMessage()
	if err == nil {

		switch str  {
		case "file:":
			err = fc.getFile()
		case ":file":
			err = fc.sendFile()
		}
		if err != nil {
			fmt.Println(err)
		}
	}
}
func startFileServer() {
	ln, err := net.Listen("tcp", ":2121")
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
