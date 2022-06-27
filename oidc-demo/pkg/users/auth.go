package users

import "fmt"

func Auth(login, password, mfa string) (bool, error) {
	if login == "placeholder" && password == "password" {
		return true, nil
	}
	return false, fmt.Errorf("Invalid login or password")
}
