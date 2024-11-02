package share

import (
	"github.com/artking28/myGoUtils"
	"github.com/google/uuid"
)

type User struct {
	Uuid     *string
	Name     string
	Password string
	Color    int32
}

func NullUser() User {
	return User{
		Uuid:     nil,
		Name:     "",
		Password: "",
	}
}

func NewUser(name, password string) User {
	return User{
		Uuid:     myGoUtils.Ptr(uuid.New().String()),
		Name:     name,
		Password: password,
	}
}

func (this User) IsLogged() bool {
	return this.Uuid != nil
}
