package services

import (
	"context"
	"errors"
	"go-bid/internal/store/pgstore"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var ErrDuplicatedEmailOrEmail = errors.New("invalid username or email")

type UserService struct {
	pool    *pgxpool.Pool
	queries *pgstore.Queries
}

func NewUserService(pool *pgxpool.Pool) UserService {
	return UserService{
		pool:    pool,
		queries: pgstore.New(pool),
	}
}

func (us *UserService) CreateUser(ctx context.Context, username, password, email, bio string) (uuid.UUID, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)

	if err != nil {
		return uuid.UUID{}, err
	}

	args := pgstore.CreateUserParams{
		UserName:     username,
		PasswordHash: hash,
		Email:        email,
		Bio:          bio,
	}

	id, err := us.queries.CreateUser(ctx, args)

	if err != nil {
		/* check if its constraint SQL error */
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return uuid.UUID{}, ErrDuplicatedEmailOrEmail
		}

		return uuid.UUID{}, err
	}

	return id, nil
}
