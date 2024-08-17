package types

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func ValidateUser(u *User) bool {
	return true
}
