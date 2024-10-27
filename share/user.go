package share

import (
	"github.com/google/uuid"
)

type User struct {
	Uuid     string
	Name     string
	Password string
}

func NewUser() User {
	return User{
		Uuid:     uuid.New().String(),
		Name:     "",
		Password: "",
	}
}

func (this *User) CheckExists() (bool, error) {
	return false, nil
}
