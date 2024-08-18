package storage

import (
	"context"
	"fmt"

	"github.com/ethanhosier/mia-backend-go/types"
	supa "github.com/nedpals/supabase-go"
)

type SupabaseStorage struct {
	client *supa.Client
}

func NewSupabaseStorage(client *supa.Client) *SupabaseStorage {
	return &SupabaseStorage{
		client: client,
	}
}

func (s *SupabaseStorage) GetUserByID(id string) (*types.User, error) {
	return &types.User{
		ID:   id,
		Name: "Alice",
	}, nil
}

func (s *SupabaseStorage) CreateUserFromEmailPassword(name, email, password string) (*types.User, error) {
	ctx := context.Background()

	user, err := s.client.Auth.SignUp(ctx, supa.UserCredentials{
		Email:    email,
		Password: password,
	})

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	fmt.Println(user)
	return &types.User{
		ID:    user.ID,
		Name:  name,
		Email: user.Email,
	}, nil
}
