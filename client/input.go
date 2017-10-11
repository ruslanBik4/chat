// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	_ "net/http/pprof"
	"sync"
)

func getUserInput(conn net.Conn, wg *sync.WaitGroup) {

	writer := bufio.NewWriter(conn)
	for isExit := false; !isExit; {
		var str string
		_, err := fmt.Scanln(&str)
		if err != nil {
			fmt.Printf("%#v", err)
			break
		}
		switch str {
		case "exit":
			isExit = true
		case ":file":
			getFile()
		case "file:":
			sendFile()
		default:
			sendMessage(writer, str)
		}
	}

	wg.Done()

}
func readAnswer(conn net.Conn, wg *sync.WaitGroup) {
	reader := bufio.NewReader(conn)

	for {
		str, err := reader.ReadString('\n')

		if err != nil {
			fmt.Printf("Error from answer:  %#v\n", err)
			break
		}

		fmt.Println(str)
	}
	wg.Done()

}
func showGreeting() {
	fmt.Println(`Добро пожаловать в наш тестовый чат,
	Вы можете отправлять сообщения нажатием клавиши Enter.
	Перечень доступных команд:
	"file:" - отправить файл,
	":list" - получить список файлов с сервера
	":file" - получить файл с сервера
	"exit"  - завершить работу
Приятной работы!`)
}

// переменные конфигурации
var (
	fPort     = flag.String("port", ":8080", "host address to connected")
	fPortFile = flag.String("portFile", ":2121", "host port for file transfer")
	fDebug    = flag.Bool("debug", true, "debug mode")
)

func main() {

	flag.Parse()

	conn, err := net.Dial("tcp", *fPort)
	if err != nil {
		panic(err)
	}
	wg := &sync.WaitGroup{}

	wg.Add(1)
	go getUserInput(conn, wg)
	go readAnswer(conn, wg)

	showGreeting()
	wg.Wait()

}
