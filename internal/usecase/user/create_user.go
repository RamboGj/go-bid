package user

import (
	"context"
	"go-bid/internal/validator"
)

type CreateUserRequest struct {
	UserName     string `json:"user_name"`
	Email        string `json:"email"`
	PasswordHash []byte `json:"password_hash"`
	Bio          string `json:"bio"`
}

/*
* { "email": "must be a valid email"}
 */

func (req CreateUserRequest) Valid(ctx context.Context) validator.Evaluator {
	var eval validator.Evaluator

	return eval
}
