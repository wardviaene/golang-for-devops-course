package users

import "fmt"

type User struct {
	Login  string
	Groups []string
	Email  string
}

func Auth(login, password, mfa string) (bool, User, error) {
	if login == "edward" && password == "password" {
		return true, User{
			Login:  "edward",
			Groups: []string{"admin"},
			Email:  "edward@domain.inv",
		}, nil
	}
	return false, User{}, fmt.Errorf("Invalid login or password")
}
