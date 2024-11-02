package main

//import "C"
import (
	"bufio"
	ChatGo "chatGo/share"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {

	// Usuário logado(se estiver) e conexão aberta
	localUser := ChatGo.NullUser()
	var conn net.Conn

	println(ChatGo.WelcomeMsg)
	scanner := bufio.NewScanner(os.Stdin)
	for i := 1; ; i++ {
		var err error
		fmt.Printf("%sChatGo(%d) > %s", ChatGo.Bold, i, ChatGo.Reset)
		if !scanner.Scan() {
			ChatGo.EmitError(scanner.Err(), "")
		}
		original := scanner.Text()

		// Divide a entrada por espaços
		input := strings.Split(original, " ")
		switch input[0] {

		// Começa o sign up caso não haja um usuário logado
		case ChatGo.SignUp:
			if localUser.IsLogged() {
				println(ChatGo.AlreadyLoggedInMsg)
				continue
			}
			localUser, err = ChatGo.GetUser(original)
			if err != nil {
				ChatGo.EmitError(err, "")
				continue
			}
			if conn, err = SendServer(original, localUser); err != nil {
				ChatGo.EmitError(err, "server")
				localUser = ChatGo.NullUser()
			}
			break

		// Começa o login caso n haja um usuário logado
		case ChatGo.Login:
			if localUser.IsLogged() {
				println(ChatGo.AlreadyLoggedInMsg)
				continue
			}
			var err error
			localUser, err = ChatGo.GetUser(original)
			if err != nil {
				ChatGo.EmitError(err, "")
				continue
			}
			if conn, err = SendServer(original, localUser); err != nil {
				ChatGo.EmitError(err, "server")
				localUser = ChatGo.NullUser()
			}
			break

		// Desloga o usuário.
		case ChatGo.Logout:
			if !localUser.IsLogged() {
				println(ChatGo.NotLoggedInMsg)
				continue
			}
			if conn, err = SendServer(original, localUser); err != nil {
				ChatGo.EmitError(err, "")
				continue
			}
			localUser = ChatGo.NullUser()
			if err := conn.Close(); err != nil {
				ChatGo.EmitError(err, "")
			}
			conn = nil
			break

		// Send messages or get users list
		case ChatGo.Message:
		case ChatGo.Hidden:
		case ChatGo.Users:
			if !localUser.IsLogged() {
				println(ChatGo.NotLoggedInMsg)
				continue
			}
			if conn, err = SendServer(original, localUser); err != nil {
				ChatGo.EmitError(err, "server")
			}
			break

		// Mostra um guia simples no console.
		case ChatGo.Help:
			ChatGo.PrintHelp(localUser.Name == "")
			break

		// Limpa o console.
		case ChatGo.Clear:
			ChatGo.ClearConsole()
			break

		// Termina o programa.
		case ChatGo.Exit:
			if !localUser.IsLogged() {
				if conn, err = SendServer(original, localUser); err != nil {
					ChatGo.EmitError(err, "server")
					ChatGo.WriteLog(ChatGo.LogErr, "falha ao realizar logout para sair", "internal")
					continue
				}
				localUser = ChatGo.NullUser()
			}
			os.Exit(0)

		// Comando n reconhecido.
		default:
			println(ChatGo.UnexpectMsg)
		}
	}
}

func SendServer(input string, _ ChatGo.User) (net.Conn, error) {

	conn, err := net.Dial("tcp", ":1110")
	if err != nil {
		return nil, err
	}

	_, err = conn.Write([]byte(input))
	if err != nil {
		ChatGo.EmitError(err, "server")
	}

	reply := make([]byte, ChatGo.ServerBuffer)
	_, err = conn.Read(reply)
	if err != nil {
		ChatGo.EmitError(err, "server")
	}

	if r := string(reply); r == "error" {
		return nil, errors.New(r)
	}
	return conn, nil
}
