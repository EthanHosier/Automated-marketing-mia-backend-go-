package storage

import "github.com/ethanhosier/mia-backend-go/types"

type MemoryStorage struct{}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{}
}

func (s *MemoryStorage) GetUserByID(id string) (*types.User, error) {
	return &types.User{
		ID:   id,
		Name: "Alice",
	}, nil
}
