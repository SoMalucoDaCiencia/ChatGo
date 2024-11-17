package share

import (
	"fmt"
	"math/rand/v2"
)

// User é a struct para representar um usuário
// ===============================================>
type User struct {

	// SessionId é o uuid da sessão e pode ser nulo
	SessionId *string `json:"-"`

	// Name é o nome do usuário
	Name string `json:"name"`

	// Password é a senha do usuário criptografada
	Password string `json:"password"`

	// Color é a cor do chat de 0 a 255 representando o hue de um HSL
	Color byte `json:"color"`
}

// NullUser cria um usuário nulo
func NullUser() User {
	return User{
		Name:     "",
		Password: "",
		Color:    byte(rand.IntN(256)),
	}
}

func NewUser(name, password string) User {
	return User{
		Name:     name,
		Password: password,
		Color:    byte(rand.IntN(256)),
	}
}

// GetMessage Formata uma string pra uma mensagem do usuário
// =============================================================>
func (this User) GetMessage(message string, hidden bool) string {
	ret := fmt.Sprintf("%s%s: %v", WrapColor(this.Name, this.Color), Reset, message)
	if hidden {
		return fmt.Sprintf("%s(private)%s %s", DGray, Reset, ret)
	}
	return ret
}

// IsLogged vê se está logado
// ==============================>
func (this User) IsLogged() bool {
	return this.SessionId != nil
}
