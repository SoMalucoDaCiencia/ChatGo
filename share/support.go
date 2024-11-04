package share

import (
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"math"
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

func GetUser(input string) (User, error) {
	vec := strings.Split(input, " ")
	if len(vec) != 4 || vec[1] != "-u" || vec[3] != "-p" {
		return NullUser(), errors.New(InvalidLoginComendMsg)
	}
	return NewUser(vec[2], GetHiddenInput()), nil
}

func EmitError(err error, origin string) {
	WriteLog(LogErr, err.Error(), origin)
	println(OperationCancelMsg)
}

func HslToRgb(h, s, l float64) (int, int, int) {
	c := (1 - math.Abs(2*l-1)) * s
	x := c * (1 - math.Abs(math.Mod(h/60.0, 2)-1))
	m := l - c/2

	var r, g, b float64
	switch {
	case 0 <= h && h < 60:
		r, g, b = c, x, 0
	case 60 <= h && h < 120:
		r, g, b = x, c, 0
	case 120 <= h && h < 180:
		r, g, b = 0, c, x
	case 180 <= h && h < 240:
		r, g, b = 0, x, c
	case 240 <= h && h < 300:
		r, g, b = x, 0, c
	case 300 <= h && h < 360:
		r, g, b = c, 0, x
	}

	r = (r + m) * 255
	g = (g + m) * 255
	b = (b + m) * 255

	return int(r), int(g), int(b)
}

func WrapColor(name string, colorByte byte) string {
	// Converte o hue para RGB com saturação de 80% e luminosidade de 90%
	r, g, b := HslToRgb(float64(colorByte), 0.8, 0.9)

	// Formata a string com a cor RGB gerada
	return fmt.Sprintf("\033[38;2;%d;%d;%dm%s\033[0m", r, g, b, name)
}

func EnsureConn(conn *net.Conn) error {
	if conn != nil {
		return nil
	}
	c, err := net.Dial("tcp", ":1110")
	if err != nil {
		return err
	}
	conn = &c
	return nil
}

func PrintHelp(isLogged bool) {
	sb := strings.Builder{}
	sb.WriteString(" > Commands:\n")
	if isLogged {
		sb.WriteString("   - login -u <USER_NAME> -p: Login to server.\n")
		sb.WriteString("   - signup -u <USER_NAME> -p: SignUp to server.\n")
	} else {
		sb.WriteString("   - msg: Logout from the server.\n")
		sb.WriteString("   - hidden <TARGET_USER>: Logout from the server.\n")
		sb.WriteString("   - users: List all online users.\n")
	}
	sb.WriteString("   - help: Get help panel.\n")
	sb.WriteString("   - exit: Get out the chat.\n")
	sb.WriteString("   - clear: Clear the terminal.\n")
	println(sb.String())
}
