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

// Usuário logado(se existir)
var localUser = ChatGo.NullUser()

func main() {

	// Conexão aberta
	// ================>
	var conn net.Conn

	// Canal para capturar os sinais
	// ===============================>
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Código de emergência par lidar com erros críticos repentinos
	// ================================================================>
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recuperado de um panic:", r)
		}
		sigs <- syscall.SIGABRT
	}()

	// Goroutine para capturar o sinal e executar o código de limpeza
	// ================================================================>
	go Cleanup(sigs)

	// Goroutine para buscar as mensagens se houver usuário logado
	// =============================================================>
	go func() {
		for {
			if localUser.IsLogged() {
				msg := ChatGo.CreateMsg(ChatGo.Fetch, *localUser.SessionId, "", ChatGo.StatusNeutral)
				_, err := SendServer(msg)
				if err != nil && err.Error() != "you are not logged in" {
					ChatGo.EmitError(err, "")
					os.Exit(1)
				}
			}
			time.Sleep(500 * time.Millisecond)
		}
	}()

	// For para capturar entradas do usuário
	// ========================================>
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
		// ===============================>
		input := strings.Split(original, " ")
		switch input[0] {

		// Começa o sign up caso não haja um usuário logado
		// ====================================================>
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
			msg := ChatGo.CreateMsg(ChatGo.SignUp, "", strings.Join(input[1:], " "), ChatGo.StatusNeutral)
			if conn, err = SendServer(msg); err != nil {
				ChatGo.EmitError(err, "server")
				localUser = ChatGo.NullUser()
			}
			break

		// Começa o login caso n haja um usuário logado
		// ================================================>
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
			msg := ChatGo.CreateMsg(ChatGo.Login, "", strings.Join(input[1:], " "), ChatGo.StatusNeutral)
			if conn, err = SendServer(msg); err != nil {
				if strings.Contains(err.Error(), "dial tcp") {
					ChatGo.WriteLog(ChatGo.LogInfo, "bad connection or offline server", "")
					localUser = ChatGo.NullUser()
					break
				}
				ChatGo.EmitError(err, "server")
				localUser = ChatGo.NullUser()
			}
			break

		// Desloga o usuário.
		// ====================>
		case ChatGo.Logout:
			if !localUser.IsLogged() {
				println(ChatGo.NotLoggedInMsg)
				continue
			}
			msg := ChatGo.CreateMsg(ChatGo.Logout, *localUser.SessionId, "", ChatGo.StatusNeutral)
			conn, err = SendServer(msg)
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
		// ==================================>
		case ChatGo.Message:
			fallthrough
		case ChatGo.Hidden:
			fallthrough
		case ChatGo.Users:
			if !localUser.IsLogged() {
				println(ChatGo.NotLoggedInMsg)
				continue
			}
			msg := ChatGo.CreateMsg(input[0], "", strings.Join(input[1:], " "), ChatGo.StatusNeutral)
			if conn, err = SendServer(msg); err != nil {
				ChatGo.EmitError(err, "server")
			}
			break

		// Mostra um guia simples no console.
		// ====================================>
		case ChatGo.Help:
			ChatGo.PrintHelp(localUser.Name == "")
			break

		// Limpa o console.
		// ==================>
		case ChatGo.Clear:
			ChatGo.ClearConsole()
			break

		// Termina o programa.
		// =====================>
		case ChatGo.Exit:
			if localUser.IsLogged() {
				msg := ChatGo.CreateMsg(ChatGo.Logout, *localUser.SessionId, "", ChatGo.StatusNeutral)
				if conn, err = SendServer(msg); err != nil {
					ChatGo.EmitError(err, "server")
					ChatGo.WriteLog(ChatGo.LogErr, "falha ao realizar logout para sair", "internal")
					continue
				}
				localUser = ChatGo.NullUser()
			}
			Cleanup(sigs)
			os.Exit(0)

		// Lida com comandos não reconhecidos.
		// ======================================>
		default:
			println(ChatGo.UnexpectMsg)
		}
	}
}

func SendServer(content ChatGo.ConnMsg) (net.Conn, error) {

	conn, err := net.Dial("tcp", ":1110")
	if err != nil {
		return nil, err
	}

	_, err = conn.Write([]byte(content.String()))
	if err != nil {
		ChatGo.EmitError(err, "server")
	}

	reply := make([]byte, ChatGo.ServerBuffer)
	_, err = conn.Read(reply)
	if err != nil {
		ChatGo.EmitError(err, "server")
	}

	resp := ChatGo.Parse(reply)
	switch resp.Status {
	case ChatGo.StatusSuccess:
		println(resp.Content)
		return conn, nil
	case ChatGo.StatusError:
		return conn, errors.New(resp.Content)
	}
	return conn, nil
}

func Cleanup(sigs chan os.Signal) {
	<-sigs
	if localUser.IsLogged() {
		ChatGo.CreateMsg("ChatGo.Logout", *localUser.SessionId, "", ChatGo.StatusNeutral)
		_, err := SendServer("logout "+*localUser.SessionId, localUser)
		if err != nil && err.Error() != "you are not logged in" {
			ChatGo.EmitError(err, "")
			os.Exit(1)
		}
	}
	os.Exit(0)
}
