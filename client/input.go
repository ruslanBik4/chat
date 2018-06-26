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
	"os"
	"runtime"
	"runtime/pprof"
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
		case ":exit":
			err = sendMessage(writer, str)
			isExit = true
		case ":file":
			getFile()
		case "file:":
			sendFile()
		default:
			err = sendMessage(writer, str)
		}
		if err != nil {
			fmt.Println(err)
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

// переменные конфигурации
var (
	fPort     = flag.String("port", ":8080", "host address to connected")
	fPortFile = flag.String("portFile", ":2121", "host port for file transfer")
	fDebug    = flag.Bool("debug", true, "debug mode")
	memprofile = flag.String("memprofile", "", "write memory profile to `file`")
)
func RunProfiler() {
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			fmt.Println("could not create memory profile: ", err)
			return
		}
		defer f.Close()

		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			fmt.Println("could not write memory profile: ", err)
		}
	}
}
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

	wg.Wait()

}
