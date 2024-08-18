package types

import (
	"fmt"
	"regexp"
)

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UserCreateFromEmailPasswordRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func ValidateUser(u *User) bool {
	return true
}

func ValidateUserCreateFromEmailPasswordRequest(u *UserCreateFromEmailPasswordRequest) (bool, error) {
	if len(u.Name) < 3 {
		return false, fmt.Errorf("Name must be at least 3 characters long")
	}
	if len(u.Password) < 8 {
		return false, fmt.Errorf("Password must be at least 8 characters long")
	}

	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !re.MatchString(u.Email) {
		return false, fmt.Errorf("Invalid email")
	}

	return true, nil
}
