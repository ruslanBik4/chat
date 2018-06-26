// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"time"
	"sync"
	"os"
	"encoding/gob"
	"fmt"
	"errors"
)

type userNick struct {
	Pass      string
	LastLogin time.Time
	active bool
}
type userData struct {
	users map[string]*userNick
	lock *sync.Mutex
}

func (u *userData) newUserNick(nick string) *userNick {
	user, ok := u.users[nick]
	if ok {
		if user.active {
			return nil
		}

		return user
	}
	u.lock.Lock()
	defer u.lock.Unlock()

	u.users[nick] = &userNick{active: true}

	return u.users[nick]
}
func (u *userData) putUser(nick, pass string) error {
	user, ok := u.users[nick]
	if !ok {
		return errors.New("нет пользователя с таким ником")
	}
	u.lock.Lock()
	defer u.lock.Unlock()

	user.Pass = pass
	user.LastLogin = time.Now()
	user.active = true

	err := saveStore()
	if err != nil {
		fmt.Println(err)
	}

	return err
}
func (u *userData) delUser(nick string) error {
	user, ok := u.users[nick]
	if !ok {
		return errors.New("нет пользователя с таким ником")
	}
	u.lock.Lock()
	defer u.lock.Unlock()

	if user.Pass > "" {
		user.active = false
	} else {
		delete(u.users, nick)
	}

	return nil
}
func NewUserStore() *userData {
	return &userData{
		users: make(map[string]*userNick, 0),
		lock: &sync.Mutex{},
	}
}
var usersStore = NewUserStore()

const dataFileName = "users.dat"
func saveStore() error {
	ioUsers, err := os.Create(dataFileName)
	if err == nil {
		defer ioUsers.Close()
		dec := gob.NewEncoder(ioUsers)
		regUsers := make(map[string]*userNick, 0)
		for key, user := range usersStore.users {
			if user.Pass > "" {
				regUsers[key] = user
			}
		}
		err = dec.Encode(&regUsers)
	}

	return err
}

func readStore() error {
	ioUsers, err := os.Open(dataFileName)
	if err == nil {
		defer ioUsers.Close()
		dec := gob.NewDecoder(ioUsers)
		err = dec.Decode(&usersStore.users)
	}

	return err
}
func init()  {

	err := readStore()
	if err != nil {
		fmt.Println(err)
	}
}



