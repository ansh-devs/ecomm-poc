package repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	db "github.com/ansh-devs/commercelens/login-service/db/generated"
	"github.com/ansh-devs/commercelens/login-service/dto"
	"github.com/go-kit/log"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"golang.org/x/crypto/bcrypt"
)

type PGRepo struct {
	db     db.Querier
	logger log.Logger
	trace  opentracing.Tracer
}

func NewPostgresRepo(db db.Querier, logger log.Logger, tracer opentracing.Tracer) Repository {
	return &PGRepo{db: db, logger: log.With(logger, "layer", "repository"), trace: tracer}
}

func (r *PGRepo) CreateUser(ctx context.Context, usr dto.RegisterUserRequest) (dto.User, error) {
	resp, err := r.db.CheckUserIsRegistered(ctx, usr.Email)
	if err != nil {
		return dto.User{}, errors.New("can't create your account at this moment")
	}
	if resp == 0 {
		id := uuid.NewString()
		hashedPwd, err := bcrypt.GenerateFromPassword([]byte(usr.Password), bcrypt.MinCost)
		if err != nil {
			return dto.User{}, errors.New("can't create your account at this moment")
		}
		createUser, err := r.db.CreateUser(ctx, db.CreateUserParams{
			ID:        id,
			Fullname:  usr.FullName,
			Email:     usr.Email,
			Password:  string(hashedPwd),
			Address:   usr.Address,
			CreatedAt: time.Now(),
		})
		if err != nil {
			return dto.User{}, fmt.Errorf("error occured while creating your account : %v", err)
		}
		return dto.User{
			ID:       createUser.ID,
			FullName: createUser.Fullname,
			Email:    createUser.Email,
			Password: createUser.Password,
			Address:  createUser.Address,
		}, nil
	} else {
		return dto.User{}, errors.New("account with this email id already exists")
	}
}
func (r *PGRepo) GetUser(ctx context.Context, id string) (dto.User, error) {
	resp, err := r.db.GetUser(ctx, id)
	if err != nil {
		return dto.User{}, errors.New("can't fetch user details from database")
	}
	return dto.User{
		ID:       resp.ID,
		FullName: resp.Fullname,
		Email:    resp.Email,
		Password: resp.Password,
		Address:  resp.Address,
	}, nil
}

func (r *PGRepo) CheckUserByEmail(ctx context.Context, email string) (dto.User, error) {
	resp, err := r.db.CheckUserByEmail(ctx, email)
	if err != nil {
		return dto.User{}, err
	}
	return dto.User{
		ID:       resp.ID,
		FullName: resp.Fullname,
		Email:    resp.Email,
		Password: resp.Password,
		Address:  resp.Address,
	}, nil
}
