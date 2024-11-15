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
	var line int

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
			time.Sleep(500 * time.Millisecond)
			if !localUser.IsLogged() {
				continue
			}
			msg := ChatGo.CreateEmptyMsg(ChatGo.Fetch, *localUser.SessionId, ChatGo.StatusNeutral)
			_, res, err := SendServer(msg)
			if err != nil {
				if strings.Contains(err.Error(), "dial tcp") {
					ChatGo.WriteLog(ChatGo.LogInfo, "server connection has been lost for an unknown reason", "")
					localUser = ChatGo.NullUser()
				} else {
					ChatGo.EmitError(err, "server")
				}
				var u string
				if localUser.IsLogged() {
					u = fmt.Sprintf("[%s] ", localUser.Name)
				}
				line += 1
				fmt.Printf("%sChatGo(%d) %s> %s", ChatGo.Bold, line, u, ChatGo.Reset)
				continue
			}
			if len(res) > 0 {
				println("\r", res, strings.Repeat(" ", 10))
				var u string
				if localUser.IsLogged() {
					u = fmt.Sprintf("[%s] ", localUser.Name)
				}
				line += 1
				fmt.Printf("%sChatGo(%d) %s> %s", ChatGo.Bold, line, u, ChatGo.Reset)
			}
		}
	}()

	// For para capturar entradas do usuário
	// ========================================>
	println(ChatGo.WelcomeMsg)
	scanner := bufio.NewScanner(os.Stdin)
	for ; ; line++ {
		var err error
		var u string
		if localUser.IsLogged() {
			u = fmt.Sprintf("[%s] ", localUser.Name)
		}
		fmt.Printf("%sChatGo(%d) %s> %s", ChatGo.Bold, line, u, ChatGo.Reset)
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
			msg := ChatGo.CreateMsg(ChatGo.SignUp, "", original, ChatGo.StatusNeutral)
			var tk string
			if conn, tk, err = SendServer(msg); err != nil {
				if strings.Contains(err.Error(), "dial tcp") {
					ChatGo.WriteLog(ChatGo.LogInfo, "bad connection or offline server", "")
					localUser = ChatGo.NullUser()
					continue
				}
				ChatGo.EmitError(err, "server")
				localUser = ChatGo.NullUser()
				continue
			}
			localUser.SessionId = &tk
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
			msg := ChatGo.CreateMsg(ChatGo.Login, "", original, ChatGo.StatusNeutral)
			var tk string
			if conn, tk, err = SendServer(msg); err != nil {
				if strings.Contains(err.Error(), "dial tcp") {
					ChatGo.WriteLog(ChatGo.LogInfo, "bad connection or offline server", "")
					localUser = ChatGo.NullUser()
					continue
				}
				ChatGo.EmitError(err, "server")
				localUser = ChatGo.NullUser()
				continue
			}
			localUser.SessionId = &tk
			break

		// Desloga o usuário.
		// ====================>
		case ChatGo.Logout:
			if !localUser.IsLogged() {
				println(ChatGo.NotLoggedInMsg)
				continue
			}
			msg := ChatGo.CreateEmptyMsg(ChatGo.Logout, *localUser.SessionId, ChatGo.StatusNeutral)
			conn, _, err = SendServer(msg)
			if err != nil && err.Error() != "you are not logged in" {
				if strings.Contains(err.Error(), "dial tcp") {
					ChatGo.WriteLog(ChatGo.LogInfo, "bad connection or offline server", "")
					localUser = ChatGo.NullUser()
					continue
				}
				ChatGo.EmitError(err, "server")
				localUser = ChatGo.NullUser()
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
			msg := ChatGo.CreateMsg(input[0], *localUser.SessionId, original, ChatGo.StatusNeutral)
			var res string
			if conn, res, err = SendServer(msg); err != nil {
				if strings.Contains(err.Error(), "dial tcp") {
					ChatGo.WriteLog(ChatGo.LogInfo, "bad connection or offline server", "")
					localUser = ChatGo.NullUser()
					continue
				}
				ChatGo.EmitError(err, "server")
				continue
			}
			if input[0] == ChatGo.Users {
				println(res)
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
				msg := ChatGo.CreateEmptyMsg(ChatGo.Logout, *localUser.SessionId, ChatGo.StatusNeutral)
				if conn, _, err = SendServer(msg); err != nil {
					if strings.Contains(err.Error(), "dial tcp") {
						ChatGo.WriteLog(ChatGo.LogInfo, "bad connection or offline server", "")
						localUser = ChatGo.NullUser()
						continue
					}
					ChatGo.EmitError(err, "server")
					localUser = ChatGo.NullUser()
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

func SendServer(content ChatGo.ConnMsg) (net.Conn, string, error) {

	conn, err := net.Dial("tcp", ":1110")
	if err != nil {
		return nil, "", err
	}

	_, err = conn.Write([]byte(content.String()))
	if err != nil {
		ChatGo.EmitError(err, "server")
		return nil, "", err
	}

	reply := make([]byte, ChatGo.ServerBuffer)
	_, err = conn.Read(reply)
	if err != nil {
		ChatGo.EmitError(err, "server")
		return nil, "", err
	}

	resp := ChatGo.Parse(reply)
	switch resp.Status {
	case ChatGo.StatusSuccess:
		return conn, resp.Content, nil
	case ChatGo.StatusError:
		return conn, "", errors.New(resp.Content)
	}
	return conn, resp.Content, nil
}

func Cleanup(sigs chan os.Signal) {
	<-sigs
	if localUser.IsLogged() {
		msg := ChatGo.CreateEmptyMsg(ChatGo.Logout, *localUser.SessionId, ChatGo.StatusNeutral)
		_, _, err := SendServer(msg)
		if err != nil && err.Error() != "you are not logged in" {
			ChatGo.EmitError(err, "")
			os.Exit(1)
		}
	}
	os.Exit(0)
}
