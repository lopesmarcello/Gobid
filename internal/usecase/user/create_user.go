package user

import (
	"context"

	"github.com/lopesmarcello/gobid/internal/validator"
)

type CreateUserReq struct {
	UserName string `json:"user_name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Bio      string `json:"bio"`
}

/*
* { "email": "must be a valid email" }
 */

func (r CreateUserReq) Valid(ctx context.Context) validator.Evaluator {
	var eval validator.Evaluator
	eval.CheckField(validator.NotBlank(r.UserName), "user_name", "this field cannot be empty")

	eval.CheckField(validator.NotBlank(r.Email), "email", "this field cannot be empty")
	eval.CheckField(validator.Matches(r.Email, validator.EmailRX), "email", "must be a valid email")

	eval.CheckField(validator.NotBlank(r.Bio), "bio", "this field cannot be empty")
	eval.CheckField(
		validator.MinChars(r.Bio, 10) &&
			validator.MaxChars(r.Bio, 255),
		"bio", "this field mus have a length between 10 and 255")

	eval.CheckField(validator.MinChars(r.Password, 8), "password", "password length must be above 8")

	return eval
}
