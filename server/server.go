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
	"time"
)
// организует канал связи с посетителем
type chatChannel struct {
	conn    net.Conn
	reader  *bufio.Reader
	writer  *bufio.Writer
	inChan  chan string
	outChan chan<- string
	isClose bool
	nick string
}
func (c *chatChannel) Close() error {
	c.isClose = true
	err := usersStore.delUser(c.nick)
	if err != nil {
		fmt.Println(err)
	}
	return c.conn.Close()
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
func (c *chatChannel) readMessage() string {
	str, err := c.reader.ReadString('\n')
	if err != nil {
		// завершение работы потока
		if err == io.EOF {
			fmt.Println(c.nick + " покидает нас!")
			err = c.Close()
			c.outChan <- c.nick + " покидает нас!"
		}
		if err != nil {
			fmt.Print(err)
		}
		return ""
	}

	return strings.TrimSuffix(str, "\n" )
}

// constructor
func newChatChannel(conn net.Conn, outChan chan<- string, nick string) *chatChannel {

	return &chatChannel{
		conn:    conn,
		reader:  bufio.NewReader(conn),
		writer:  bufio.NewWriter(conn),
		inChan:  make(chan string),
		outChan: outChan,
		nick:	 nick,
	}
}
func (c *chatChannel) setNick(nick string) bool {
	if nick == ""   {
		return false
	}

	user := usersStore.newUserNick(nick)
	if user == nil {
		c.sendAnswer("такое имя уже используется")
		fmt.Println(user)
		return false
	}
	if !user.active {
		c.sendAnswer("введите пароль")
		pass := c.readMessage()
		if pass != user.Pass {
			c.sendAnswer("неверный пароль")
			return false
		}
		c.sendAnswer("А мы Вас ждем АЖ с " + user.LastLogin.String())
		user.LastLogin = time.Now()
		user.active = true
	}
	c.outChan <- "Пользователь '" + c.nick + "' сменил имя на - " + nick
	c.nick = nick
	return true
}

func (c *chatChannel) handle(userList [] string) {

	c.showGreeting()
	c.sendAnswer("Сейчас присутствуют : " + strings.Join( userList, ",") )
	c.outChan <- "К нам приходит " + c.nick
	go func() {
		for str := range c.inChan {
			c.sendAnswer(str)
			if c.isClose {
				break
			}
		}
	}()
	for !c.isClose {
		str := c.readMessage()

		switch str {
		case "":
			continue
		case ":exit":
			c.sayGoodBy()
			c.outChan <- "Нас покидает " + c.nick
		case ":list":
			c.sendAnswer(fileList())
		case ":nick":
			c.sendAnswer("введите имя:")
			nick := c.readMessage()
			if c.setNick(nick) {
				c.sendAnswer("успешно сменили имя" )
			} else {
				c.sendAnswer("смена ника не удалась" )
			}
		case ":register":
			pass := c.readMessage()
			oldNick := c.nick
			if err := usersStore.putUser(c.nick, pass); err == nil {
				c.sendAnswer("успешно зарегистрировались" )
				c.outChan <- "Пользователь '" + oldNick + "' сменил имя на - " + c.nick
			} else {
				c.sendAnswer("регистрация ника не удалась - " + err.Error() )
			}
		default:

			if *fDebug {
				fmt.Println(str)
			}
			c.outChan <- c.nick + ">" + str
		}
	}

}
func (c *chatChannel) showGreeting() {
	c.sendAnswer(`Добро пожаловать в наш тестовый чат,` + c.nick + `
	Вы можете отправлять сообщения нажатием клавиши Enter.
	Перечень доступных команд:
	"file:" - отправить файл,
	":list" - получить список файлов с сервера
	":file" - получить файл с сервера
	":nick" - сменить текущий нил
	":register" - зарегистрироть пароль для ника
	":exit"  - завершить работу
Приятной работы!`)
}
func (c *chatChannel) sayGoodBy() {
	c.sendAnswer("GoodBy")
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
				if !ch.isClose {
					ch.inChan <- str
				}
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
		newNick := fmt.Sprintf("Guest%d", len(broadcast))
		// запускаем отдельный поток для каждого соединения чата
		c := newChatChannel(conn, chanBroadcast, newNick)

		userList := make([]string, len(broadcast))
		for key, ch := range broadcast {
			userList[key] = ch.nick
		}
		go c.handle(userList)

		broadcast = append(broadcast, c)

	}
}
