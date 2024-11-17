package share

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ssh/terminal"
	"math"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// Bcrypt encripta uma string usando o "bcript"
// ===============================================>
func Bcrypt(word string) (string, error) {
	ret, err := bcrypt.GenerateFromPassword([]byte(word), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(ret), nil
}

// TryMatch compara um hash e uma string pura
// =============================================>
func TryMatch(original, attempt string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(original), []byte(attempt)); err != nil {
		return err
	}
	return nil
}

// GetHiddenInput pega uma senha pelo terminal escondendo o conteúdo
// ===================================================================>
func GetHiddenInput() string {
	println(" > Por favor, digite sua senha.")
	bytes, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	println()
	for len(bytes) < 5 {
		println(" > Por favor, digite uma senha de no mínimo 5 dígitos.")
		bytes, err = terminal.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			panic(err)
		}
		println()
	}
	h := sha256.New()
	h.Write(bytes)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// ClearConsole limpa o console do terminal
// ===================================================================>
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
		WriteLog(LogErr, err.Error(), "")
	}
}

// GetUser simula um "form" pelo terminal e monta a struct de usuário
// ===================================================================>
func GetUser(input string) (User, error) {
	vec := strings.Split(input, " ")
	if len(vec) != 4 || vec[1] != "-u" || vec[3] != "-p" {
		return NullUser(), errors.New(InvalidLoginCommandMsg)
	}
	return NewUser(vec[2], GetHiddenInput()), nil
}

// HslToRgb converte HSL pra RGB
// =================================>
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

	r, g, b = (r+m)*255, (g+m)*255, (b+m)*255
	return int(r), int(g), int(b)
}

// WrapColor colore uma string usando a refêrencia de hude do HSL, fixando a saturação em 1 e a luminosidade em 0.9
// ===================================================================================================================>
func WrapColor(name string, colorByte byte) string {
	r, g, b := HslToRgb(float64(colorByte), 1, 0.9)
	return fmt.Sprintf("\033[38;2;%d;%d;%dm%s\033[0m", r, g, b, name)
}

// PrintHelp imprime referencias de como usar o client
// =====================================================>
func PrintHelp(isLogged bool) {
	sb := strings.Builder{}
	sb.WriteString(" > Commands:\n")
	if isLogged {
		sb.WriteString("   - login -u <USER_NAME> -p: Login to server.\n")
		sb.WriteString("   - signup -u <USER_NAME> -p: SignUp to server.\n")
	} else {
		sb.WriteString("   - logout: Logout from the server.\n")
		sb.WriteString("   - msg: Send a simple message.\n")
		sb.WriteString("   - hidden <TARGET_USER>: Send a hidden message to someone.\n")
		sb.WriteString("   - users: List all online users.\n")
	}
	sb.WriteString("   - help: Get help panel.\n")
	sb.WriteString("   - exit: Get out the chat.\n")
	sb.WriteString("   - clear: Clear the terminal.\n")
	println(sb.String())
}
