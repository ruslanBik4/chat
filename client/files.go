// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"os"
	"bufio"
	"path"
	"net"
	"fmt"
)
// отвечает за создание соединения с сервером для передачи файлов
type fileChannel struct {
	conn net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
}
func (fc *fileChannel) sendMessage(str string) {
	sendMessage(fc.writer, str)
}
func (fc *fileChannel) readMessage() (string, error) {
	return readMessage(fc.reader)
}
func (fc *fileChannel) readFrom(file *os.File) (int64, error) {
	return fc.writer.ReadFrom(file)
}
const dirDownloadFiles = "files"
func (fc *fileChannel) saveFile(fileName string) (n int64, err error) {
	writeFile, err := os.Create(path.Join(dirDownloadFiles, fileName))
	return fc.reader.WriteTo(writeFile)
}
func (fc *fileChannel) Close() {
	fc.conn.Close()
}
func newFileChannel() *fileChannel {

	conn, err := net.Dial("tcp", *fPortFile )
	if err != nil {
		panic(err)
	}

	return &fileChannel{
		conn: conn,
		reader: bufio.NewReader(conn),
		writer: bufio.NewWriter(conn),
	}
}
const maxFileSize = 1000000000
func sendFile() {

	defer func() {
		err := recover()
		switch err {
		case os.ErrNotExist:
			fmt.Print("not found file")
		default:
			fmt.Print(err)
		}
	}()

	fc := newFileChannel()
	defer fc.Close()

	fileName, err := inputStr("Введите имя файла (максимальный размер файла - 1ГБ):")

	reader, err := os.Open(fileName)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer reader.Close()
	stat, _ := reader.Stat()
	if stat.Size() > maxFileSize {
		fmt.Print(" file too big!")
		return
	}
	fc.sendMessage("file:")

	if str, err := fc.readMessage(); err != nil || (str != "ready") {
		fmt.Println("Отправка файла невозможна!")
		return
	}

	fc.sendMessage(path.Base(reader.Name()))

	n, err := fc.readFrom(reader)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("send %d bytes", n)
}
func inputStr(title string) (str string, err error) {
	fmt.Println(title)
	_, err = fmt.Scanln(&str)
	if err != nil {
		fmt.Printf("%#v", err)
		panic(err)
	}

	return
}
func getFile()  {
	fc := newFileChannel()
	defer fc.Close()

	fc.sendMessage(":file")
	fileNumber, _ := inputStr("Введите номер файла:")
	fc.sendMessage(fileNumber)

	if str, err := fc.readMessage(); err != nil || (str != "ready") {
		fmt.Printf("Ошибка при приеме имени файла - %v", err)
		return
	} else {
		fileName, err := fc.readMessage()

		if err != nil {
			fmt.Printf("Ошибка при приеме имени файла - %v", err)
			return

		}
		fc.saveFile(fileName)
	}

}