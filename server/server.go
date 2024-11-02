package main

import (
	ChatGo "chatGo/share"
	"encoding/json"
	"github.com/artking28/myGoUtils"
	"net"
	"os"
	"strings"
)

var allUsers map[string]ChatGo.User

var tc = myGoUtils.NewThreadControl(200)

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
		allUsers[*u.Uuid] = u
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

	switch input[0] {
	//case ChatGo.Login:
	//case ChatGo.SignUp:
	//case ChatGo.Message:
	//case ChatGo.Hidden:
	//case ChatGo.Users:
	//case ChatGo.Logout:
	default:
		s, _ := ChatGo.WrapInColor("client:", nil)
		tc.Lock()
		println(s, strings.Join(input, " "))
		tc.Unlock()
	}

	_, err := conn.Write([]byte("ok"))
	if err != nil {
		ChatGo.WriteLog(ChatGo.LogErr, err.Error(), "internal")
		err = nil
	}
}

//func handleConnection(conn net.Conn) {
//
//	for {
//		buf := make([]byte, ChatGo.ClientBuffer)
//		_, err := conn.Read(buf)
//		if err != nil {
//			return
//		}
//
//		fmt.Printf("Received: %s\n", buf)
//
//		_, err = conn.Write([]byte("Message received"))
//		if err != nil {
//			fmt.Println("Error sending response:", err)
//			return
//		}
//	}
//	//conn.Close()
//}
