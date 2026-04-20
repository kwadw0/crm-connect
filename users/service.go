package users

import (
	"context"
	"fmt"
	"kwadw0/WhatsCRM/internal/postgres/repo"
	"kwadw0/WhatsCRM/utils"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service interface {
	CreateUser(ctx context.Context, dto CreateUserDto) (UserResponseDto, error)
	GetAllUsers(ctx context.Context) ([]UserResponseDto, error)
	UpdateUser(ctx context.Context, userID uuid.UUID, dto UpdateUserDto) (UserResponseDto, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) (UserResponseDto, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (UserResponseDto, error)
}

type userService struct {
	repo *repo.Queries
}

func NewService(userRepo *repo.Queries) Service {
	return &userService{repo: userRepo}
}

func (s *userService) CreateUser(ctx context.Context, dto CreateUserDto) (UserResponseDto, error) {
	hashedPassword, err := utils.HashPassword(dto.Password)
	if err != nil {
		return UserResponseDto{}, err
	}
	user, err := s.repo.CreateUser(ctx, repo.CreateUserParams{
		FirstName: dto.FirstName,
		LastName:  dto.LastName,
		Email:     dto.Email,
		Password:  hashedPassword,
		Phone:     dto.Phone,
		Role:      dto.Role,
		AvatarUrl: pgtype.Text{String: dto.AvatarURL, Valid: dto.AvatarURL != ""},
	})
	if err != nil {
		return UserResponseDto{}, err
	}
	return mapUserToResponse(user), nil
}

func (s *userService) GetUserByEmail(ctx context.Context, email string) (repo.User, error) {
	existingUser, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return repo.User{}, err
	}

	return existingUser, nil
}

func (s *userService) GetAllUsers(ctx context.Context) ([]UserResponseDto, error) {
	users, err := s.repo.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	response := make([]UserResponseDto, len(users))
	for i, u := range users {
		response[i] = mapUserToResponse(u)
	}

	return response, nil
}

func (s *userService) UpdateUser(ctx context.Context, userID uuid.UUID, dto UpdateUserDto) (UserResponseDto, error) {
	// First, fetch the existing user safely
	existingUser, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return UserResponseDto{}, fmt.Errorf("user not found: %w", err)
	}

	// Now run the update with the standard UUID type and the existing password
	user, err := s.repo.UpdateUser(ctx, repo.UpdateUserParams{
		ID:        userID,
		FirstName: dto.FirstName,
		LastName:  dto.LastName,
		Email:     dto.Email,
		Phone:     dto.Phone,
		Role:      dto.Role,
		Password:  existingUser.Password, // Carry over existing password
		AvatarUrl: pgtype.Text{String: dto.AvatarURL, Valid: dto.AvatarURL != ""},
	})
	if err != nil {
		return UserResponseDto{}, err
	}
	return mapUserToResponse(user), nil
}

func (s *userService) DeleteUser(ctx context.Context, userID uuid.UUID) (UserResponseDto, error) {
	user, err := s.repo.DeleteUser(ctx, userID)
	if err != nil {
		return UserResponseDto{}, err
	}
	return mapUserToResponse(user), nil
}

func (s *userService) GetUserByID(ctx context.Context, userID uuid.UUID) (UserResponseDto, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return UserResponseDto{}, err
	}
	return mapUserToResponse(user), nil
}

func mapUserToResponse(u repo.User) UserResponseDto {
	// Thanks to google/uuid, this is basically instant now:
	idStr := u.ID.String()

	return UserResponseDto{
		ID:        idStr,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
		Phone:     u.Phone,
		Role:      u.Role,
		OrganizationID: u.OrganizationID.String(),
		AvatarURL: u.AvatarUrl.String,
		CreatedAt: u.CreatedAt.Time.String(),
		UpdatedAt: u.UpdatedAt.Time.String(),
	}
}
