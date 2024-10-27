package share

import (
	"golang.org/x/crypto/ssh/terminal"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func GetHiddenInput() string {
	println(" > Por favor, digite sua senha.")
	bytes, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	println()
	return string(bytes)
}

func ClearConsole() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		EmitError(err, "internal")
	}
}

func GetUser(input string) User {
	vec := strings.Split(input, " ")
	if len(vec) != 4 || vec[1] != "-u" || vec[3] != "-p" {
		println(InvalidLoginComendMsg)
		return NewUser()
	}
	user := NewUser()
	user.Name = vec[2]
	user.Password = GetHiddenInput()
	return user
}

func EmitError(err error, origin string) {
	WriteLog(LogErr, err.Error(), origin)
	println(OperationCancelMsg)
}

func EnsureConn(conn net.Conn) error {
	if conn != nil {
		return nil
	}
	c, err := net.Dial("tcp", ":1110")
	if err != nil {
		return err
	}
	conn = c
	return nil
}

func PrintHelp(isLogged bool) {
	sb := strings.Builder{}
	sb.WriteString(" > Commands:\n")
	if isLogged {
		sb.WriteString("   - login -u <USER_NAME> -p: Login to server.\n")
		sb.WriteString("   - signUp -u <USER_NAME> -p: SignUp to server.\n")
	} else {
		sb.WriteString("   - msg: Logout from the server.\n")
		sb.WriteString("   - hidden: Logout from the server.\n")
		sb.WriteString("   - users: List all online users.\n")
	}
	sb.WriteString("   - help: Get help panel.\n")
	sb.WriteString("   - exit: Get out the chat.\n")
	sb.WriteString("   - clear: Clear the terminal.\n")
	println(sb.String())
}
