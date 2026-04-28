package channels

import (
	"context"
	"errors"

	"kwadw0/WhatsCRM/internal/postgres/repo"
	"kwadw0/WhatsCRM/organizations"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type ChannelService interface {
	CreateChannel(ctx context.Context, dto CreateChannelDto) (ChannelResponseDto, error)
	UpdateChannel(ctx context.Context, channelID uuid.UUID, dto UpdateChannelDto) (ChannelResponseDto, error)
	GetChannelByID(ctx context.Context, channelID uuid.UUID) (ChannelResponseDto, error)
	ListChannelsByOrganization(ctx context.Context, organizationID uuid.UUID) ([]ChannelResponseDto, error)
	DeleteChannel(ctx context.Context, channelID uuid.UUID) (ChannelResponseDto, error)
	ConnectChannel(ctx context.Context, channelID uuid.UUID) (MetaConfigResponse, error)
}

type channelService struct {
	repo         *repo.Queries
	orgService   organizations.OrganizationService
	metaConfigID string
	metaAppID    string
}

func NewChannelService(channelRepo *repo.Queries, orgService organizations.OrganizationService, metaConfigID, metaAppID string) ChannelService {
	return &channelService{
		repo:         channelRepo,
		orgService:   orgService,
		metaConfigID: metaConfigID,
		metaAppID:    metaAppID,
	}
}

func (s *channelService) UpdateChannel(ctx context.Context, channelID uuid.UUID, dto UpdateChannelDto) (ChannelResponseDto, error) {
	// Implementation will go here
	return ChannelResponseDto{}, nil
}

func (s *channelService) CreateChannel(ctx context.Context, dto CreateChannelDto) (ChannelResponseDto, error) {
	orgIDStr := dto.OrganizationID
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		return ChannelResponseDto{}, err
	}

	// Check if organization exists
	_, err = s.orgService.GetOrganizationByID(ctx, orgID)
	if err != nil {
		return ChannelResponseDto{}, err
	}
	createChannel, err := s.repo.CreateChannel(ctx, repo.CreateChannelParams{
		OrganizationID:  orgID,
		Name:            dto.Name,
		Description:     pgtype.Text{String: dto.Description, Valid: dto.Description != ""},
		ChannelPlatform: repo.ChannelPlatform(dto.ChannelPlatform),
		AvatarUrl:       pgtype.Text{String: dto.AvatarUrl, Valid: dto.AvatarUrl != ""},
		AuthConfig:      []byte("{}"),
		PlatformConfig:  []byte("{}"),
		Capabilities:    []byte("{}"),
		WebhookUrl:      pgtype.Text{Valid: false},
	})
	if err != nil {
		return ChannelResponseDto{}, err
	}
	return mapChannelToResponse(createChannel), nil
}

func (s *channelService) ConnectChannel(ctx context.Context, channelID uuid.UUID) (MetaConfigResponse, error) {
	if s.metaConfigID == "" || s.metaAppID == "" {
		return MetaConfigResponse{}, errors.New("meta configuration is not set")
	}

	existingChannel, err := s.repo.GetChannelByID(ctx, channelID)
	if err != nil {
		return MetaConfigResponse{}, err
	}
	if existingChannel.Status != repo.ChannelStatusPending {
		return MetaConfigResponse{}, errors.New("channel status is not pending")
	}
	return MetaConfigResponse{
		ConfigID: s.metaConfigID,
		AppID:    s.metaAppID,
	}, nil
}

func (s *channelService) GetChannelByID(ctx context.Context, channelID uuid.UUID) (ChannelResponseDto, error) {
	// Implementation will go here
	channel, err := s.repo.GetChannelByID(ctx, channelID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
        return ChannelResponseDto{}, errors.New("channel not found")
    }
    return ChannelResponseDto{}, err
	}
	return mapChannelToResponse(channel), nil
}

func (s *channelService) ListChannelsByOrganization(ctx context.Context, organizationID uuid.UUID) ([]ChannelResponseDto, error) {
	// Implementation will go here
	return []ChannelResponseDto{}, nil
}

func (s *channelService) DeleteChannel(ctx context.Context, channelID uuid.UUID) (ChannelResponseDto, error) {
	// Implementation will go here
	return ChannelResponseDto{}, nil
}

func mapChannelToResponse(c repo.Channel) ChannelResponseDto {
	return ChannelResponseDto{
		ID:              c.ID.String(),
		OrganizationID:  c.OrganizationID.String(),
		Name:            c.Name,
		Description:     c.Description.String,
		ChannelPlatform: string(c.ChannelPlatform),
		AvatarUrl:       c.AvatarUrl.String,
		Status:          string(c.Status),
		StatusReason:    c.StatusReason.String,
		AuthConfig:      c.AuthConfig,
		PlatformConfig:  c.PlatformConfig,
		Capabilities:    c.Capabilities,
		WebhookVerified: c.WebhookVerified,
		WebhookUrl:      c.WebhookUrl.String,
		CreatedAt:       c.CreatedAt.Time.String(),
		UpdatedAt:       c.UpdatedAt.Time.String(),
	}
}