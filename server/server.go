package main

import (
	ChatGo "chatGo/share"
	"encoding/json"
	"fmt"
	mgu "github.com/artking28/myGoUtils"
	"github.com/google/uuid"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var (
	cachePath = ".cache"
	userCache = fmt.Sprintf("%s%cusers.json", cachePath, os.PathSeparator)

	// map[username]User
	// Todos os usuários colocados em memória pelo nome
	userDB = map[string]*ChatGo.User{}

	// map[address]username
	// Todas as sessions guardadas por um UUID
	online = map[string]string{}

	// Pilha de mensagens pendentes de cada usuário referenciada pelo nome
	msgStack = map[string][]string{}

	// Meu controle de threads customizado usado em vários dos meus projetos, você pode checar em "github.com/artking28/myGoUtils"
	tc = mgu.NewThreadControl(200)
)

func main() {

	// Carrega o "cache" de usuários em memória
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
			ChatGo.WriteLog(ChatGo.LogInfo, "Recovering panic...", "")
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

	ChatGo.WriteLog(ChatGo.LogOk, "Server started with success", "")

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
			ChatGo.WriteLog(ChatGo.LogErr, err.Error(), "")
			continue
		}

		// Lida com as conexões por concorrência
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
			ChatGo.WriteLog(ChatGo.LogErr, err.Error(), "")
			err = nil
		}
		err = conn.Close()
		if err != nil {
			ChatGo.WriteLog(ChatGo.LogErr, err.Error(), "")
		}
		return
	}
	command := ChatGo.Parse(buf)
	input := strings.Split(command.Content, " ")

	// Registra um log de comandos de usuários.
	// ==============================================================>
	s := fmt.Sprintf("from: [%s%s%s] command: [%s%s%s]",
		ChatGo.DGray, connAddress, ChatGo.Reset,
		ChatGo.DGray, command.Content, ChatGo.Reset,
	)
	ChatGo.WriteLog(ChatGo.LogInfo, s, "")

	switch command.Control {

	// Começa o login de um usuário
	// ================================================>
	case ChatGo.Login:
		if u := userDB[input[2]]; u == nil || ChatGo.TryMatch(u.Password, input[4]) != nil {
			Close(&conn, "incorrect username or password", ChatGo.StatusError)
			return
		}
		u, err := uuid.NewUUID()
		if err != nil {
			ChatGo.WriteLog(ChatGo.LogErr, err.Error(), "")
			Close(&conn, err.Error(), ChatGo.StatusError)
		}
		tk := u.String()
		tc.Lock()
		online[tk] = input[2]
		msgStack[input[2]] = []string{}
		tc.Unlock()
		Close(&conn, tk, ChatGo.StatusSuccess)
		break

	// Começa o processo de sign-in e ja loga o usuário se bem-sucedido
	// ===================================================================>
	case ChatGo.SignUp:
		if u := userDB[input[2]]; u != nil {
			Close(&conn, "this username is already taken", ChatGo.StatusError)
			return
		}
		u, err := uuid.NewUUID()
		if err != nil {
			ChatGo.WriteLog(ChatGo.LogErr, err.Error(), "")
			Close(&conn, err.Error(), ChatGo.StatusError)
		}
		hash, err := ChatGo.Bcrypt(input[4])
		if err != nil {
			ChatGo.WriteLog(ChatGo.LogErr, err.Error(), "")
			Close(&conn, err.Error(), ChatGo.StatusError)
		}
		input[4] = hash
		tk := u.String()
		tc.Lock()
		userDB[input[2]] = mgu.Ptr(ChatGo.NewUser(input[2], input[4]))
		online[tk] = input[2]
		msgStack[input[2]] = []string{}
		tc.Unlock()
		Close(&conn, tk, ChatGo.StatusSuccess)
		break

	case ChatGo.Fetch:
		if online[command.Token] == "" {
			Close(&conn, "unauthorized, missing token", ChatGo.StatusError)
		}
		tc.Lock()
		uOrigin := userDB[online[command.Token]]
		if uOrigin != nil {
			messages := msgStack[uOrigin.Name]
			msgStack[uOrigin.Name] = []string{}
			Close(&conn, strings.Join(messages, "\n"), ChatGo.StatusSuccess)
			tc.Unlock()
			break
		}
		tc.Unlock()
		Close(&conn, "", ChatGo.StatusSuccess)
		break

	//case ChatGo.Refresh:

	// Envia uma mensagem pata o chat global.
	// ======================================================>
	case ChatGo.Message:
		if online[command.Token] == "" {
			Close(&conn, "unauthorized, missing token", ChatGo.StatusError)
		}
		u := userDB[online[command.Token]]
		msg := u.GetMessage(strings.Join(input[1:], " "), false)
		tc.Lock()
		for k := range msgStack {
			if k != u.Name {
				msgStack[k] = append(msgStack[k], msg)
			}
		}
		tc.Unlock()
		Close(&conn, "message sent", ChatGo.StatusSuccess)
		break

	// Envia uma mensagem oculta de um usuário para outro
	// ======================================================>
	case ChatGo.Hidden:
		if online[command.Token] == "" {
			Close(&conn, "unauthorized, missing token", ChatGo.StatusError)
		}
		u := userDB[online[command.Token]]
		uTarget := userDB[input[1]].Name
		msg := u.GetMessage(strings.Join(input[2:], " "), true)
		tc.Lock()
		msgStack[u.Name] = append(msgStack[u.Name], msg)
		msgStack[uTarget] = append(msgStack[uTarget], msg)
		tc.Unlock()
		Close(&conn, "hidden message sent", ChatGo.StatusSuccess)
		break

	// Lista todos os usuários logados para os usuários
	// ====================================================>
	case ChatGo.Users:
		if online[command.Token] == "" {
			Close(&conn, "unauthorized, missing token", ChatGo.StatusError)
		}
		users := mgu.MapValues(online)
		Close(&conn, strings.Join(users, "\n"), ChatGo.StatusSuccess)
		break

	// Desloga o usuário.
	// ====================>
	case ChatGo.Logout:
		name := online[command.Token]
		if name == "" {
			Close(&conn, "you are not logged in", ChatGo.StatusError)
			return
		}
		if u := userDB[name]; u == nil {
			Close(&conn, "incorrect username", ChatGo.StatusError)
			return
		}
		tc.Lock()
		delete(msgStack, command.Token)
		for k := range msgStack {
			msgStack[k] = append(msgStack[k], online[command.Token]+" logged out!")
		}
		tc.Unlock()
		delete(online, command.Token)
		Close(&conn, "logout complete", ChatGo.StatusSuccess)
		break

	// Lida com comandos não reconhecidos.
	// ======================================>
	default:
		ChatGo.WriteLog(ChatGo.LogWarn, "unrecognized command: "+command.Content, "")
		Close(&conn, "this command is invalid", ChatGo.StatusError)
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
		ChatGo.WriteLog(ChatGo.LogErr, err.Error(), "")
		os.Exit(1)
	}
	os.Exit(0)
}

// Close responde para o client uma mensagem simples
// ========================================================>
func Close(conn *net.Conn, content string, status int) {
	res := ChatGo.CreateMsg(ChatGo.Response, "", content, status)
	_, err := (*conn).Write([]byte(res.String()))
	if err != nil {
		ChatGo.WriteLog(ChatGo.LogErr, err.Error(), "")
	}
}
