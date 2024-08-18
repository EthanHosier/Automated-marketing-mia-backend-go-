package storage

import "github.com/ethanhosier/mia-backend-go/types"

type Storage interface {
	// GetUserByID(id string) (*types.User, error)
	CreateUserFromEmailPassword(name, email, password string) (*types.User, error)
}
