package organizations

import (
	"context"
	"kwadw0/WhatsCRM/internal/postgres/repo"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type OrganizationService interface {
	AddOrganization(ctx context.Context, userID uuid.UUID, dto CreateOrganizationDto) (OrganizationResponseDto, error)
	GetOrganizationByID(ctx context.Context, orgID uuid.UUID) (OrganizationResponseDto, error)
	ListOrganizations(ctx context.Context) ([]OrganizationResponseDto, error)
	UpdateOrganizationById(ctx context.Context, orgID uuid.UUID, dto UpdateOrganizationDto) (OrganizationResponseDto, error)
	DeleteOrganization(ctx context.Context, orgID uuid.UUID) (OrganizationResponseDto, error)
}

type organizationService struct {
	repo *repo.Queries
}

func NewOrganizationService(organizationRepo *repo.Queries) OrganizationService {
	return &organizationService{repo: organizationRepo}
}

func (s *organizationService) AddOrganization(ctx context.Context, userID uuid.UUID, dto CreateOrganizationDto) (OrganizationResponseDto, error) {
	organization, err := s.repo.CreateOrganization(ctx, repo.CreateOrganizationParams{
		Name:                dto.Name,
		Description:         pgtype.Text{String: dto.Description, Valid: dto.Description != ""},
		WebsiteUrl:          pgtype.Text{String: dto.WebsiteUrl, Valid: dto.WebsiteUrl != ""},
		Industry:            pgtype.Text{String: dto.Industry, Valid: dto.Industry != ""},
		TeamSize:            pgtype.Text{String: dto.TeamSize, Valid: dto.TeamSize != ""},
		PrimaryCustomerType: pgtype.Text{String: dto.PrimaryCustomerType, Valid: dto.PrimaryCustomerType != ""},
		PrimaryUseCase:      dto.PrimaryUseCase,
		OwnerRole:           dto.OwnerRole,
		ReferralSource:      pgtype.Text{String: dto.ReferralSource, Valid: dto.ReferralSource != ""},
		IsActive:            pgtype.Bool{Bool: true, Valid: true},
	})

	if err != nil {
		return OrganizationResponseDto{}, err
	}

	err = s.repo.UpdateUserOrganization(ctx, repo.UpdateUserOrganizationParams{
		ID:             userID,
		OrganizationID: pgtype.UUID{Bytes: organization.ID, Valid: true},
	})
	if err != nil {
		return OrganizationResponseDto{}, err
	}

	return mapOrgToResponse(organization), nil
}

func (s *organizationService) GetOrganizationByID(ctx context.Context, orgID uuid.UUID) (OrganizationResponseDto, error) {
	organization, err := s.repo.FindOrganizationByID(ctx, orgID)
	if err != nil {
		return OrganizationResponseDto{}, err
	}

	return mapOrgToResponse(organization), nil
}

func (s *organizationService) ListOrganizations(ctx context.Context) ([]OrganizationResponseDto, error) {
	orgs, err := s.repo.FindAllOrganizations(ctx)
	if err != nil {
		return nil, err
	}

	var dtos []OrganizationResponseDto
	for _, o := range orgs {
		dtos = append(dtos, mapOrgToResponse(o))
	}

	return dtos, nil
}

func (s *organizationService) UpdateOrganizationById(ctx context.Context, orgID uuid.UUID, dto UpdateOrganizationDto) (OrganizationResponseDto, error) {
	_, err := s.repo.FindOrganizationByID(ctx, orgID)
	if err != nil {
		return OrganizationResponseDto{}, err
	}

	organization, err := s.repo.UpdateOrganization(ctx, repo.UpdateOrganizationParams{
		ID:                  orgID,
		Industry:            pgtype.Text{String: dto.Industry, Valid: dto.Industry != ""},
		IsActive:            pgtype.Bool{Bool: dto.IsActive, Valid: true},
		Name:                dto.Name,
		PrimaryCustomerType: pgtype.Text{String: dto.PrimaryCustomerType, Valid: dto.PrimaryCustomerType != ""},
		PrimaryUseCase:      dto.PrimaryUseCase,
		Description:         pgtype.Text{String: dto.Description, Valid: dto.Description != ""},
		WebsiteUrl:          pgtype.Text{String: dto.WebsiteUrl, Valid: dto.WebsiteUrl != ""},
	})
	if err != nil {
		return OrganizationResponseDto{}, err
	}
	return mapOrgToResponse(organization), nil
}

func (s *organizationService) DeleteOrganization(ctx context.Context, orgID uuid.UUID) (OrganizationResponseDto, error) {
	deleteOrganization, err := s.repo.DeleteOrganization(ctx, orgID)
	if err != nil {
		return OrganizationResponseDto{}, err
	}
	return mapOrgToResponse(deleteOrganization), nil
}

func mapOrgToResponse(org repo.Organization) OrganizationResponseDto {
	return OrganizationResponseDto{
		ID:                  org.ID.String(),
		Name:                org.Name,
		Description:         org.Description.String,
		WebsiteUrl:          org.WebsiteUrl.String,
		Industry:            org.Industry.String,
		TeamSize:            org.TeamSize.String,
		PrimaryCustomerType: org.PrimaryCustomerType.String,
		PrimaryUseCase:      org.PrimaryUseCase,
		OwnerRole:           org.OwnerRole,
		ReferralSource:      org.ReferralSource.String,
		IsActive:            org.IsActive.Bool,
		CreatedAt:           org.CreatedAt.Time.String(),
		UpdatedAt:           org.UpdatedAt.Time.String(),
	}
}
