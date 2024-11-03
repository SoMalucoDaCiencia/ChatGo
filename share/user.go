package share

import (
	"github.com/artking28/myGoUtils"
	"github.com/google/uuid"
	"math/rand/v2"
)

type User struct {
	Uuid     *string
	Name     string
	Password string
	Color    byte
}

func NullUser() User {
	return User{
		Uuid:     nil,
		Name:     "",
		Password: "",
		Color:    byte(rand.IntN(256)),
	}
}

func NewUser(name, password string) User {
	return User{
		Uuid:     myGoUtils.Ptr(uuid.New().String()),
		Name:     name,
		Password: password,
		Color:    byte(rand.IntN(256)),
	}
}

func (this User) IsLogged() bool {
	return this.Uuid != nil
}
