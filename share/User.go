package share

import (
	"math/rand/v2"
)

type User struct {
	SessionId *string `json:"-"`
	Name      string  `json:"name"`
	Password  string  `json:"password"`
	Color     byte    `json:"color"`
}

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

func (this User) IsLogged() bool {
	return this.SessionId != nil
}
