package roles

import (
	"context"
	"kwadw0/WhatsCRM/internal/postgres/repo"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type RoleService interface {
	CreateRole(ctx context.Context, req CreateRoleDTO) (RoleResponseDTO, error)
	UpdateRole(ctx context.Context, roleID uuid.UUID, dto UpdateRoleDTO) (RoleResponseDTO, error)
	DeleteRole(ctx context.Context, roleID uuid.UUID) (RoleResponseDTO, error)
	GetRoleByID(ctx context.Context, roleID uuid.UUID) (RoleResponseDTO, error)
	GetAllRoles(ctx context.Context) ([]RoleResponseDTO, error)
}

type roleService struct {
	queries *repo.Queries
}

func NewService(queries *repo.Queries) RoleService {
	return &roleService{queries: queries}
}

func (s *roleService) CreateRole(ctx context.Context, req CreateRoleDTO) (RoleResponseDTO, error) {
	role, err := s.queries.CreateRole(ctx, repo.CreateRoleParams{
		Name:        req.Name,
		Description: pgtype.Text{String: req.Description, Valid: req.Description != ""},
	})
	if err != nil {
		return RoleResponseDTO{}, err
	}
	return RoleResponseDTO{
		ID:          role.ID.String(),
		Name:        role.Name,
		Description: role.Description.String,
		CreatedAt:   role.CreatedAt.Time.String(),
		UpdatedAt:   role.UpdatedAt.Time.String(),
	}, nil
}

func (s *roleService) UpdateRole(ctx context.Context, roleID uuid.UUID, dto UpdateRoleDTO) (RoleResponseDTO, error) {
	_, err := s.queries.GetRoleByID(ctx, roleID)
	if err != nil {
		return RoleResponseDTO{}, err
	}
	updatedRole, err := s.queries.UpdateRole(ctx, repo.UpdateRoleParams{
		ID:          roleID,
		Name:        dto.Name,
		Description: pgtype.Text{String: dto.Description, Valid: dto.Description != ""},
	})
	if err != nil {
		return RoleResponseDTO{}, err
	}
	return RoleResponseDTO{
		ID:          updatedRole.ID.String(),
		Name:        updatedRole.Name,
		Description: updatedRole.Description.String,
		CreatedAt:   updatedRole.CreatedAt.Time.String(),
		UpdatedAt:   updatedRole.UpdatedAt.Time.String(),
	}, nil
}

func (s *roleService) DeleteRole(ctx context.Context, roleID uuid.UUID) (RoleResponseDTO, error) {
	deletedRole, err := s.queries.DeleteRole(ctx, roleID)
	if err != nil {
		return RoleResponseDTO{}, err
	}

	return RoleResponseDTO{
		ID:          deletedRole.ID.String(),
		Name:        deletedRole.Name,
		Description: deletedRole.Description.String,
		CreatedAt:   deletedRole.CreatedAt.Time.String(),
		UpdatedAt:   deletedRole.UpdatedAt.Time.String(),
	}, nil
}

func (s *roleService) GetRoleByID(ctx context.Context, roleID uuid.UUID) (RoleResponseDTO, error) {
	role, err := s.queries.GetRoleByID(ctx, roleID)
	if err != nil {
		return RoleResponseDTO{}, err
	}
	return RoleResponseDTO{
		ID:          role.ID.String(),
		Name:        role.Name,
		Description: role.Description.String,
		CreatedAt:   role.CreatedAt.Time.String(),
		UpdatedAt:   role.UpdatedAt.Time.String(),
	}, nil
}

func (s *roleService) GetAllRoles(ctx context.Context) ([]RoleResponseDTO, error) {
	roles, err := s.queries.ListRoles(ctx)
	if err != nil {
		return nil, err
	}
	var dtos []RoleResponseDTO
	for _, role := range roles {
		dtos = append(dtos, RoleResponseDTO{
			ID:          role.ID.String(),
			Name:        role.Name,
			Description: role.Description.String,
			CreatedAt:   role.CreatedAt.Time.String(),
			UpdatedAt:   role.UpdatedAt.Time.String(),
		})
	}
	return dtos, nil
}
