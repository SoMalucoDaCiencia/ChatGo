package share

import (
	"fmt"
	mgu "github.com/artking28/myGoUtils"
	"strings"
)

// ConnMsg struct representa uma mensagem entre client e servidor
// =================================================================>
type ConnMsg struct {

	// Control tipo de comando(login, signup, etc...)
	Control string

	// Token Session uuid, se houver algum
	Token string

	// Content é conteúdo, parecido com um "body" do http
	Content string

	// Status de resposta do servidor
	Status int
}

func CreateEmptyMsg(ctl string, token string, status int) ConnMsg {
	content := fmt.Sprintf("%s [token: %s]", ctl, token)
	return ConnMsg{
		Control: ctl,
		Status:  status,
		Token:   token,
		Content: content,
	}
}

func CreateMsg(ctl string, token, content string, status int) ConnMsg {
	return ConnMsg{
		Control: ctl,
		Status:  status,
		Token:   token,
		Content: content,
	}
}

// Parse interpreta os bytes para uma struct
// =============================================>
func Parse(bytes []byte) ConnMsg {
	sBytes := strings.ReplaceAll(string(bytes), "\x00", "")
	ret := ConnMsg{}
	input := strings.Split(sBytes, "\n")
	for _, line := range input {
		lineSplit := strings.Split(line, " ")
		switch lineSplit[0] {
		case "Control":
			ret.Control = lineSplit[1]
		case "Token":
			ret.Token = lineSplit[1]
		case "Status":
			s, _ := mgu.Int[int](lineSplit[1])
			ret.Status = s
		}
	}
	ret.Content = strings.Split(sBytes, "@")[1]
	return ret
}

// String transforma a struct na string a ser escrita na comunicação
// ====================================================================>
func (this ConnMsg) String() string {
	return fmt.Sprintf("Control %s\nToken %s\nStatus %d\n@%s", this.Control, this.Token, this.Status, this.Content)
}
