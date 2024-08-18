package storage

import "github.com/ethanhosier/mia-backend-go/types"

type Storage interface {
	StoreBusinessSummary(userId string, businessSummary types.BusinessSummary) error
}
