// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"fmt"
	"strings"
)

func sendMessage(writer *bufio.Writer, str string) error {
	n, err := writer.WriteString(str + "\n")
	if err != nil {
		fmt.Printf("Error by sending: %#v\n", err)
		return err
	}

	writer.Flush()

	if *fDebug {
		fmt.Println(n)
	}

	return nil
}

func readMessage(reader *bufio.Reader) (string, error) {
	str, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error by received: %#v\n", err)
		return "", err
	}

	return strings.TrimSuffix(str, "\n"), nil
}
