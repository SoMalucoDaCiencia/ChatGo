package share

import (
	"fmt"
	mgu "github.com/artking28/myGoUtils"
	"strings"
)

type ConnMsg struct {
	Control string
	Token   string
	Content string
	Status  int
}

func CreateMsg(ctl string, token, content string, status int) ConnMsg {
	return ConnMsg{
		Control: ctl,
		Status:  status,
		Token:   token,
		Content: content,
	}
}

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

func (this ConnMsg) GetUserInput() string {
	return fmt.Sprintf("%s %s", this.Control, this.Content)
}

func (this ConnMsg) String() string {
	return fmt.Sprintf("Control %s\nToken %s\nStatus %d\n@%s", this.Control, this.Token, this.Status, this.Content)
}
