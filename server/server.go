package main

import (
	ChatGo "chatGo/share"
	"encoding/json"
	"errors"
	"fmt"
	mgu "github.com/artking28/myGoUtils"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var (
	cachePath = ".cache"
	userCache = fmt.Sprintf("%s%cusers.json", cachePath, os.PathSeparator)
	msgCache  = fmt.Sprintf("%s%cmessages.txt", cachePath, os.PathSeparator)

	// map[username]User
	userDB = map[string]*ChatGo.User{}

	// map[address]username
	online = map[string]string{}

	//all = []string

	msgStack = map[string][]string{}

	tc = mgu.NewThreadControl(200)
)

func main() {

	// Salva o "cache" de usuários em memória
	// ===========================================>
	err := os.MkdirAll(".cache", 0777)
	if err != nil {
		panic(err)
	}
	c, err := os.ReadFile(userCache)
	if err != nil || c == nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
		c = []byte("[]")
		err := os.WriteFile(userCache, c, 0777)
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

	// Inicia o listening da porta
	// ================================>
	ln, err := net.Listen("tcp", ":1110")
	if err != nil {
		panic(err)
	}

	// For para capturar chamadas do client
	// ========================================>
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
			handleConnection(conn, buf)
			tc.Done()
		}()
	}
}

func handleConnection(conn net.Conn, buf []byte) {

	// Pega a origem do usuário e valida o input.
	// ==============================================================>
	connAddress := conn.RemoteAddr().String()
	if len(buf) == 0 {
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
	command := ChatGo.Parse(buf)

	// Registra um log de comandos de usuários.
	// ==============================================================>
	s := fmt.Sprintf("from: [%s%s%s] command: [%s%s%s]",
		ChatGo.DGray, connAddress, ChatGo.Reset,
		ChatGo.DGray, command.GetUserInput(), ChatGo.Reset,
	)
	ChatGo.WriteLog(ChatGo.LogInfo, s, "")

	switch command.Control {

	// Começa o login de um usuário
	// ================================================>
	case ChatGo.Login:
		if u := userDB[input[2]]; u == nil || u.Password != input[4] {
			CloseErr(&conn, errors.New("incorrect username or password"))
			return
		}
		tc.Lock()
		online[command.Token] = input[2]
		tc.Unlock()
		CloseOk(&conn, "sign-in complete!", &command.Token)
		break

	// Começa o processo de sign-in e ja loga o usuário se bem-sucedido
	// ===================================================================>
	case ChatGo.SignUp:
		if u := userDB[input[2]]; u != nil {
			CloseErr(&conn, errors.New("this username is already taken"))
			return
		}
		tc.Lock()
		userDB[input[2]] = mgu.Ptr(ChatGo.NewUser(input[2], input[4]))
		online[sessionID] = input[2]
		tc.Unlock()
		CloseOk(&conn, "sign-up complete! You are now automatically logged in", &sessionID)
		break

	case ChatGo.Fetch:
		uOrigin := userDB[online[sessionID]]
		if uOrigin != nil {
			messages := msgStack[uOrigin.Name]
			CloseOk(&conn, strings.Join(messages, "\n"), nil)
			break
		}
		CloseOk(&conn, "", nil)
		break

	case ChatGo.Refresh:

	// Envia uma mensagem pata o chat global.
	// ======================================================>
	case ChatGo.Message:
		u := userDB[online[sessionID]]
		msg := fmt.Sprintf("%s%s: %s", ChatGo.Bold, ChatGo.WrapColor(u.Name, u.Color), input[1])
		tc.Lock()
		for k := range msgStack {
			msgStack[k] = append(msgStack[k], msg)
		}
		tc.Unlock()
		CloseOk(&conn, "hidden message sent", nil)
		break

	// Envia uma mensagem oculta de um usuário para outro
	// ======================================================>
	case ChatGo.Hidden:
		uOrigin := userDB[online[sessionID]]
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
		CloseOk(&conn, "message sent", nil)
		break

	// Lista todos os usuários logados para os usuários
	// ====================================================>
	case ChatGo.Users:
		users := mgu.MapValues(online)
		CloseOk(&conn, strings.Join(users, "\n"), nil)
		break

	// Desloga o usuário.
	// ====================>
	case ChatGo.Logout:
		if u := userDB[online[input[1]]]; u == nil {
			CloseErr(&conn, errors.New("incorrect username"))
			return
		}
		if online[input[1]] == "" {
			CloseErr(&conn, errors.New("you are not logged in"))
			return
		}
		delete(online, input[1])
		CloseOk(&conn, "logout complete", nil)
		break

	// Lida com comandos não reconhecidos.
	// ======================================>
	default:
		ChatGo.WriteLog(ChatGo.LogWarn, "unrecognized command: "+command.GetUserInput(), "")
		CloseErr(&conn, errors.New("this command is invalid"))
		return
	}
}

// Cleanup finaliza o servidor e realiza processos necessários
// ===============================================================>
func Cleanup(sigs chan os.Signal) {
	<-sigs
	list := mgu.VecMap(mgu.MapValues(userDB), func(t *ChatGo.User) ChatGo.User {
		return *t
	})
	j, _ := json.Marshal(list)
	err := os.WriteFile(userCache, j, 0777)
	if err != nil {
		ChatGo.EmitError(err, "")
		os.Exit(1)
	}
	os.Exit(0)
}

// CloseOk responde para o client uma mensagem simples
// ========================================================>
func CloseOk(conn *net.Conn, msg string, session *string) {
	var prefix, sessionId = "ok", ""
	if session != nil {
		prefix = "session"
		sessionId = " " + *session
	}
	res := fmt.Sprintf("%s%s -m %s", prefix, sessionId, msg)
	_, err := (*conn).Write([]byte(res))
	if err != nil {
		ChatGo.WriteLog(ChatGo.LogErr, err.Error(), "internal")
		err = nil
	}
}

// CloseErr responde para o client uma mensagem de erro
// ========================================================>
func CloseErr(conn *net.Conn, errMsg error) {
	_, err := (*conn).Write([]byte("error -m " + errMsg.Error()))
	if err != nil {
		ChatGo.WriteLog(ChatGo.LogErr, err.Error(), "internal")
		err = nil
	}
}
