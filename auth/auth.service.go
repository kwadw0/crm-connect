package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"kwadw0/WhatsCRM/auth/jwt"
	"kwadw0/WhatsCRM/internal/postgres/repo"
	"kwadw0/WhatsCRM/users"
	"kwadw0/WhatsCRM/utils"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	RegisterUser(ctx context.Context, dto RegisterUserDto) (users.UserResponseDto, error)
	LoginUser(ctx context.Context, dto LoginUserDto) (users.LoginResponse, error)
}

type authService struct {
	userRepo *repo.Queries
	jwt_secret []byte
	jwt_ttl time.Duration
}

func NewService(userRepo *repo.Queries, secret []byte, exp time.Duration) Service {
	return &authService{
		userRepo: userRepo,
		jwt_secret: secret,
		jwt_ttl: exp,

	}
}

func (s *authService) RegisterUser(ctx context.Context, dto RegisterUserDto) (users.UserResponseDto, error) {
	hashedPassword, err := utils.HashPassword(dto.Password)
	if err != nil {
		return users.UserResponseDto{}, err
	}

	// Find the default "user" role
	role, err := s.userRepo.GetRoleByName(ctx, "user")
	if err != nil {
		return users.UserResponseDto{}, fmt.Errorf("failed to assign default role: %w", err)
	}
	roleID := role.ID

	user, err := s.userRepo.CreateUser(ctx, repo.CreateUserParams{
		FirstName: dto.FirstName,		
		LastName:  dto.LastName,
		Email:     dto.Email,
		Password:  hashedPassword,
		Phone:     dto.Phone,
		RoleID:    roleID,
		AvatarUrl: pgtype.Text{Valid: false}, // Default empty avatar
	})

	if err != nil {
		// If it's a different DB error, return it normally
		return users.UserResponseDto{}, err
	}

	// Map to response struct just like you did in users!
	return users.UserResponseDto{
		ID:        user.ID.String(),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Phone:     user.Phone,
		RoleID:    user.RoleID.String(),
		AvatarURL: user.AvatarUrl.String,
		CreatedAt: user.CreatedAt.Time.String(),
		UpdatedAt: user.UpdatedAt.Time.String(),
	}, nil
}

func (s *authService) LoginUser(ctx context.Context, dto LoginUserDto) (users.LoginResponse, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, dto.Email)
	if err != nil {
		// If the user does not exist in the DB, it's an invalid credential!
		if errors.Is(err, pgx.ErrNoRows) {
			return users.LoginResponse{}, utils.ErrInvalidCredentials
		}
		return users.LoginResponse{}, err
	}

	if err := utils.VerifyPassword(user.Password, dto.Password); err != nil {
		// If the password was just wrong, it's an invalid credential!
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return users.LoginResponse{}, utils.ErrInvalidCredentials
		}
		// If the hash was completely corrupted (like too short), pass through the REAL error (500)
		return users.LoginResponse{}, fmt.Errorf("password verification failed: %w", err)
	}

	token, err := authjwt.GenerateToken(s.jwt_secret, user.ID.String(), s.jwt_ttl)
	if err != nil {
		return users.LoginResponse{}, err
	}

	return users.LoginResponse{
		FirstName: user.FirstName,
		Email:     user.Email,
		AccessToken: token,
	}, nil
}
