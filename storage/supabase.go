package storage

import "github.com/ethanhosier/mia-backend-go/types"

type SupabaseStorage struct{}

func (s *SupabaseStorage) GetUserByID(id string) (*types.User, error) {
	return &types.User{
		ID:   id,
		Name: "Alice",
	}, nil
}
