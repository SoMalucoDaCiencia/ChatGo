package main

//import "C"
import (
	"bufio"
	ChatGo "chatGo/share"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var localUser = ChatGo.NullUser()

func main() {

	// Usuário logado(se estiver) e conexão aberta
	var conn net.Conn

	// Canal para capturar os sinais
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recuperado de um panic:", r)
		}
		sigs <- syscall.SIGABRT
	}()

	// Goroutine para capturar o sinal e executar o código de limpeza
	go Cleanup(sigs)

	go func() {
		for {
			if localUser.IsLogged() {

			}
			time.Sleep(500 * time.Millisecond)
		}
	}()

	println(ChatGo.WelcomeMsg)
	scanner := bufio.NewScanner(os.Stdin)
	for i := 1; ; i++ {
		var err error
		var u string
		if localUser.IsLogged() {
			u = fmt.Sprintf("[%s] ", localUser.Name)
		}
		fmt.Printf("%sChatGo(%d) %s> %s", ChatGo.Bold, i, u, ChatGo.Reset)
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
			original = original + " " + localUser.Password
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
			original = original + " " + localUser.Password
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
			conn, err = SendServer(original, localUser)
			if err != nil && err.Error() != "you are not logged in" {
				ChatGo.EmitError(err, "")
				continue
			}
			localUser = ChatGo.NullUser()
			if conn != nil {
				if err = conn.Close(); err != nil {
					ChatGo.EmitError(err, "")
				}
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
			Cleanup(true, sigs)
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

	replySplit := strings.Split(string(reply), " -m ")
	for i := range replySplit {
		replySplit[i] = strings.ReplaceAll(replySplit[i], "\x00", "")
	}
	switch replySplit[0] {
	case "ok":
		println(replySplit[1])
		break
	case "error":
		return nil, errors.New(replySplit[1])
	}
	return conn, nil
}

func Cleanup(sigs chan os.Signal) {
	<-sigs
	println("passou aqui")
	if localUser.IsLogged() {
		_, err := SendServer("logout", localUser)
		if err != nil && err.Error() != "you are not logged in" {
			ChatGo.EmitError(err, "")
			os.Exit(1)
		}
	}
	os.Exit(0)
}
