package main

import (
	ChatGo "chatGo/share"
	"encoding/json"
	"errors"
	"fmt"
	mgu "github.com/artking28/myGoUtils"
	"net"
	"os"
	"strings"
)

// map[username]User
var userDB map[string]*ChatGo.User

// map[address]username
var online map[string]string

var msgStack map[string][]string

var tc = mgu.NewThreadControl(200)

func main() {

	filePath := "users.json"
	c, err := os.ReadFile(filePath)
	if err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
		err := os.WriteFile(filePath, []byte("[]"), 777)
		if err != nil {
			panic(err)
		}
	}
	var list []ChatGo.User
	if err = json.Unmarshal(c, &list); err != nil {
		panic(err)
	}
	for _, u := range list {
		userDB[u.Name] = &u
	}

	ln, err := net.Listen("tcp", ":1110")
	if err != nil {
		panic(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			panic(err)
		}

		buf := make([]byte, ChatGo.ClientBuffer)
		_, err = conn.Read(buf)
		if err != nil {
			ChatGo.EmitError(err, "")
			continue
		}

		tc.Begin()
		go func() {
			handleConnection(conn, string(buf))
			tc.Done()
		}()
	}
}

func handleConnection(conn net.Conn, buf string) {

	connAddress := conn.RemoteAddr().String()
	input := strings.Split(buf, " ")
	if len(buf)*len(input) == 0 {
		_, err := conn.Write([]byte(ChatGo.EmptyMessageMsg))
		if err != nil {
			ChatGo.WriteLog(ChatGo.LogErr, err.Error(), "internal")
			err = nil
		}
		err = conn.Close()
		if err != nil {
			ChatGo.WriteLog(ChatGo.LogErr, err.Error(), "internal")
		}
		return
	}

	LogCommand(buf, connAddress)
	switch input[0] {
	case ChatGo.Login:
		if u := userDB[input[2]]; u == nil || u.Password != input[4] {
			CloseErr(&conn, errors.New("incorrect username or password"))
			return
		}
		tc.Lock()
		online[connAddress] = input[2]
		tc.Unlock()
		CloseOk(&conn, "sign-in complete!")
		fallthrough

	case ChatGo.SignUp:
		if u := userDB[input[2]]; u != nil {
			CloseErr(&conn, errors.New("this username is already taken"))
			return
		}
		tc.Lock()
		userDB[input[2]] = mgu.Ptr(ChatGo.NewUser(input[2], input[4]))
		online[connAddress] = input[2]
		tc.Unlock()
		CloseOk(&conn, "sign-up complete! You are now automatically logged in")
		fallthrough

	case ChatGo.Message:
		u := userDB[online[connAddress]]
		msg := fmt.Sprintf("%s%s: %s", ChatGo.Bold, ChatGo.WrapColor(u.Name, u.Color), input[1])
		tc.Lock()
		for k := range msgStack {
			msgStack[k] = append(msgStack[k], msg)
		}
		tc.Unlock()
		fallthrough

	case ChatGo.Hidden:
		uOrigin := userDB[online[connAddress]]
		uTarget := userDB[input[1]].Name
		msg := fmt.Sprintf("%s(private)%s %s%s:",
			ChatGo.DGray, ChatGo.Reset, ChatGo.Bold,
			ChatGo.WrapColor(uOrigin.Name, uOrigin.Color),
		)
		msg += input[2]
		tc.Lock()
		msgStack[uOrigin.Name] = append(msgStack[uOrigin.Name], msg)
		msgStack[uTarget] = append(msgStack[uTarget], msg)
		tc.Unlock()
		fallthrough

	case ChatGo.Users:
		users := mgu.MapKeys(userDB)
		tc.Lock()
		for k := range msgStack {
			msgStack[k] = append(msgStack[k], users...)
		}
		tc.Unlock()
		fallthrough

	case ChatGo.Logout:
		if u := userDB[input[2]]; u != nil {
			CloseErr(&conn, errors.New("incorrect username"))
			return
		}
		if online[connAddress] == "" {
			CloseErr(&conn, errors.New("you are not logged in"))
			return
		}
		delete(online, connAddress)
		fallthrough

	default:
		ChatGo.WriteLog(ChatGo.LogWarn, "unrecognized command: "+buf, "")
		CloseErr(&conn, errors.New("this command is invalid"))
		return
	}
}

func LogCommand(msg, connAddress string) {
	s := fmt.Sprintf("from: [%s%s%s] command: [%s%s%s]",
		ChatGo.DGray, connAddress, ChatGo.Reset,
		ChatGo.DGray, msg, ChatGo.Reset,
	)
	ChatGo.WriteLog(ChatGo.LogInfo, s, "")
}

func CloseOk(conn *net.Conn, msg string) {
	_, err := (*conn).Write([]byte("ok -m" + msg))
	if err != nil {
		ChatGo.WriteLog(ChatGo.LogErr, err.Error(), "internal")
		err = nil
	}
}

func CloseErr(conn *net.Conn, errInput error) {
	_, err := (*conn).Write([]byte("error -m " + errInput.Error()))
	if err != nil {
		ChatGo.WriteLog(ChatGo.LogErr, err.Error(), "internal")
		err = nil
	}
}
