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
	localUser := ChatGo.NewUser()
	var conn net.Conn

	scanner := bufio.NewScanner(os.Stdin)
	println(ChatGo.WelcomeMsg)
	for i := 1; ; i++ {
		fmt.Printf("%sChatGo(%d) > %s", ChatGo.Bold, i, ChatGo.Reset)
		if !scanner.Scan() {
			ChatGo.EmitError(scanner.Err(), "")
		}
		original := scanner.Text()

		// Divide a entrada por espaços
		input := strings.Split(original, " ")
		switch input[0] {

		// Começa o sign up caso n haja um usuário logado
		case ChatGo.SignUp:
			if localUser.Name != "" {
				println(ChatGo.AlreadyLoggedInMsg)
				continue
			}
			localUser = ChatGo.GetUser(original)
			if err := SendServer(conn, original, localUser); err != nil {
				ChatGo.EmitError(err, "server")
				localUser = ChatGo.NewUser()
			}
			break

		// Começa o login caso n haja um usuário logado
		case ChatGo.Login:
			if localUser.Name != "" {
				println(ChatGo.AlreadyLoggedInMsg)
				continue
			}
			localUser = ChatGo.GetUser(original)
			if localUser.Name != "" {
				//if err := SendServer(conn, original, localUser); err != nil {
				//	ChatGo.EmitError(err, "server")
				//	localUser = ChatGo.NewUser()
				//}
			}
			println(localUser.Name, localUser.Password, localUser.Uuid)
			break

		// Send messages or get users list
		case ChatGo.Message:
		case ChatGo.Hidden:
		case ChatGo.Users:
			if localUser.Name == "" {
				println(ChatGo.NotLoggedInMsg)
				continue
			}
			if err := SendServer(conn, original, localUser); err != nil {
				ChatGo.EmitError(err, "server")
			}
			break

		// Desloga o usuário.
		case ChatGo.Logout:
			if localUser.Name == "" {
				println(ChatGo.NotLoggedInMsg)
				continue
			}
			if err := SendServer(conn, original, localUser); err != nil {
				ChatGo.EmitError(err, "server")
				localUser = ChatGo.NewUser()
				continue
			}
			localUser = ChatGo.NewUser()
			if err := conn.Close(); err != nil {
				ChatGo.EmitError(err, "")
			}
			conn = nil
			ChatGo.ClearConsole()

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
			os.Exit(0)

		// Comando n reconhecido.
		default:
			println(ChatGo.UnexpectMsg)
		}
	}
}

func SendServer(conn net.Conn, input string, _ ChatGo.User) error {

	err := ChatGo.EnsureConn(conn)
	if err != nil {
		ChatGo.EmitError(err, "server")
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
		return errors.New(r)
	}
	return nil
}
