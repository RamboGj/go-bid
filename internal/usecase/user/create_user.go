package user

import (
	"context"
	"go-bid/internal/validator"
)

type CreateUserRequest struct {
	UserName string `json:"user_name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Bio      string `json:"bio"`
}

/*
* { "email": "must be a valid email"}
 */

func (req CreateUserRequest) Valid(ctx context.Context) validator.Evaluator {
	var eval validator.Evaluator

	eval.CheckField(validator.NotBlank(req.UserName), "user_name", "must not be blank")
	eval.CheckField(validator.NotBlank(req.Email), "email", "must not be blank")
	eval.CheckField(validator.NotBlank(req.Bio), "bio", "must not be blank")
	eval.CheckField(
		validator.MinChars(req.Bio, 10) && validator.MaxChars(req.Bio, 255),
		"bio",
		"must have a length between 10 and 255 characters",
	)
	eval.CheckField(validator.MinChars(req.Password, 8), "password", "must be at least 8 characters long")
	eval.CheckField(validator.Matches(req.Email, validator.EmailRX), "email", "must be a valid email")

	return eval
}
