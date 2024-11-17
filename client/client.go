package main

//import "C"
import (
	"bufio"
	ChatGo "chatGo/share"
	"errors"
	"fmt"
	"math"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// Lista de sinais do OS para detectar qualquer tipo de saida e lidar com isso.
// ===============================================================================>
var sigsVec = []os.Signal{
	syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP,
	syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGCONT,
	syscall.SIGQUIT, syscall.SIGTSTP,
}

// Usuário logado(se existir)
// ===============================>
var localUser = ChatGo.NullUser()

func main() {

	// Conexão aberta
	// ================>
	var conn net.Conn
	var line int

	// Canal para capturar os sinais
	// ===============================>
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, sigsVec...)

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

	// Goroutine para buscar as mensagens, se houver usuário logado
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
					ChatGo.WriteLog(ChatGo.LogErr, err.Error(), "server")
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

	// Da boas vindas ao usuário.
	// =============================>
	println(ChatGo.WelcomeMsg)

	// For para capturar entradas do usuário
	// ========================================>
	scanner := bufio.NewScanner(os.Stdin)
	for ; ; line++ {
		var err error

		// Imprime uma referência de linha para o usuário
		var u string
		if localUser.IsLogged() {
			u = fmt.Sprintf("[%s] ", localUser.Name)
		}
		fmt.Printf("%sChatGo(%d) %s> %s", ChatGo.Bold, line, u, ChatGo.Reset)

		// Lê o input do usuário
		// ===============================>
		if !scanner.Scan() {
			ChatGo.WriteLog(ChatGo.LogErr, scanner.Err().Error(), "")
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
				ChatGo.WriteLog(ChatGo.LogErr, err.Error(), "")
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
				ChatGo.WriteLog(ChatGo.LogErr, err.Error(), "server")
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
				ChatGo.WriteLog(ChatGo.LogErr, err.Error(), "")
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
				ChatGo.WriteLog(ChatGo.LogErr, err.Error(), "server")
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
				ChatGo.WriteLog(ChatGo.LogErr, err.Error(), "server")
				localUser = ChatGo.NullUser()
				continue
			}
			localUser = ChatGo.NullUser()
			if conn != nil {
				if err = conn.Close(); err != nil {
					ChatGo.WriteLog(ChatGo.LogErr, err.Error(), "")
				}
			}
			conn = nil
			break

		// Send messages or get users list
		// ==================================>
		case ChatGo.Message:
			msg := localUser.GetMessage(strings.Join(input[1:], " "), false)
			size := int(math.Ceil(math.Log10(float64(line))+16)) + len(localUser.Name)
			println("\033[A", msg, strings.Repeat(" ", size))
			fallthrough
		case ChatGo.Hidden:
			if input[0] == ChatGo.Hidden {
				msg := localUser.GetMessage(strings.Join(input[2:], " "), false)
				size := int(math.Ceil(math.Log10(float64(line))+28)) + len(localUser.Name)
				println("\033[A", msg, strings.Repeat(" ", size))
			}
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
				ChatGo.WriteLog(ChatGo.LogErr, err.Error(), "server")
				continue
			}
			if input[0] == ChatGo.Users {
				println(res)
			}
			break

		// Altera o nome do usuário
		// ============================>
		case ChatGo.Changenick:
			if !localUser.IsLogged() {
				println(ChatGo.NotLoggedInMsg)
				continue
			}
			msg := ChatGo.CreateMsg(input[0], *localUser.SessionId, original, ChatGo.StatusNeutral)
			conn, _, err = SendServer(msg)
			if err != nil && err.Error() != "you are not logged in" {
				if strings.Contains(err.Error(), "dial tcp") {
					ChatGo.WriteLog(ChatGo.LogInfo, "bad connection or offline server", "")
					localUser = ChatGo.NullUser()
					continue
				}
				ChatGo.WriteLog(ChatGo.LogErr, err.Error(), "server")
				localUser = ChatGo.NullUser()
				continue
			}
			localUser.Name = input[1]
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
					ChatGo.WriteLog(ChatGo.LogErr, err.Error(), "server")
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

	// Conecta com o servidor
	// ==========================>
	conn, err := net.Dial("tcp", ":1110")
	if err != nil {
		return nil, "", err
	}

	// Escreve no fd da conexão os bytes
	// =====================================>
	_, err = conn.Write([]byte(content.String()))
	if err != nil {
		ChatGo.WriteLog(ChatGo.LogErr, err.Error(), "server")
		return nil, "", err
	}

	// Lê a resposta do servido através fd
	//========================================>
	reply := make([]byte, ChatGo.ServerBuffer)
	_, err = conn.Read(reply)
	if err != nil {
		ChatGo.WriteLog(ChatGo.LogErr, err.Error(), "server")
		return nil, "", err
	}

	// Faz o "parse" da resposta e lida com ela.
	// ================================================>
	resp := ChatGo.Parse(reply)
	switch resp.Status {
	case ChatGo.StatusSuccess:
		return conn, resp.Content, nil
	case ChatGo.StatusError:
		return conn, "", errors.New(resp.Content)
	}
	return conn, resp.Content, nil
}

// Cleanup Finaliza o programa por qualquer sinal de saida repentina e envia um sinal de logout pro servidor
// =============================================================================================================>
func Cleanup(sigs chan os.Signal) {
	<-sigs
	if localUser.IsLogged() {
		msg := ChatGo.CreateEmptyMsg(ChatGo.Logout, *localUser.SessionId, ChatGo.StatusNeutral)
		_, _, err := SendServer(msg)
		if err != nil && err.Error() != "you are not logged in" {
			ChatGo.WriteLog(ChatGo.LogErr, err.Error(), "")
			os.Exit(1)
		}
	}
	os.Exit(0)
}
