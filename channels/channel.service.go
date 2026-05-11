package channels

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

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
	InitiateConnection(ctx context.Context, channelID uuid.UUID) (MetaConfigResponse, error)
	ConnectChannel(ctx context.Context, channelID uuid.UUID, dto ConnectChannelDto) (ChannelResponseDto, error)
	ChannelWebhook(ctx context.Context, channelID uuid.UUID, payload []byte) (map[string]interface{}, error)
}

type channelService struct {
	repo            *repo.Queries
	orgService      organizations.OrganizationService
	metaConfigID    string
	metaAppID       string
	metaAppSecret   string
	metaRedirectURI string
}

func NewChannelService(
	channelRepo *repo.Queries,
	orgService organizations.OrganizationService,
	metaConfigID,
	metaAppID,
	metaAppSecret,
	metaRedirectURI string,
) ChannelService {
	return &channelService{
		repo:            channelRepo,
		orgService:      orgService,
		metaConfigID:    metaConfigID,
		metaAppID:       metaAppID,
		metaAppSecret:   metaAppSecret,
		metaRedirectURI: metaRedirectURI,
	}
}

func (s *channelService) CreateChannel(ctx context.Context, dto CreateChannelDto) (ChannelResponseDto, error) {
	orgID, err := uuid.Parse(dto.OrganizationID)
	if err != nil {
		return ChannelResponseDto{}, fmt.Errorf("invalid organization ID: %w", err)
	}

	_, err = s.orgService.GetOrganizationByID(ctx, orgID)
	if err != nil {
		return ChannelResponseDto{}, fmt.Errorf("organization not found: %w", err)
	}

	channel, err := s.repo.CreateChannel(ctx, repo.CreateChannelParams{
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
		return ChannelResponseDto{}, fmt.Errorf("failed to create channel: %w", err)
	}

	return mapChannelToResponse(channel), nil
}

func (s *channelService) InitiateConnection(ctx context.Context, channelID uuid.UUID) (MetaConfigResponse, error) {
	if s.metaConfigID == "" || s.metaAppID == "" {
		return MetaConfigResponse{}, errors.New("meta configuration is not set")
	}

	existing, err := s.repo.GetChannelByID(ctx, channelID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return MetaConfigResponse{}, errors.New("channel not found")
		}
		return MetaConfigResponse{}, err
	}

	if existing.Status != repo.ChannelStatusPending {
		return MetaConfigResponse{}, errors.New("channel is already connected")
	}

	return MetaConfigResponse{
		ConfigID: s.metaConfigID,
		AppID:    s.metaAppID,
	}, nil
}

func (s *channelService) ConnectChannel(ctx context.Context, channelID uuid.UUID, dto ConnectChannelDto) (ChannelResponseDto, error) {
	// 1. fetch existing channel to preserve all fields
	existing, err := s.repo.GetChannelByID(ctx, channelID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ChannelResponseDto{}, errors.New("channel not found")
		}
		return ChannelResponseDto{}, err
	}

	// 2. exchange code for access token
	tokenData := url.Values{}
	tokenData.Set("client_id", s.metaAppID)
	tokenData.Set("client_secret", s.metaAppSecret)
	tokenData.Set("redirect_uri", s.metaRedirectURI)
	tokenData.Set("code", dto.Code)

	tokenResp, err := http.PostForm(
		"https://graph.facebook.com/v20.0/oauth/access_token",
		tokenData,
	)
	if err != nil {
		return ChannelResponseDto{}, fmt.Errorf("failed to exchange code with Meta: %w", err)
	}
	defer tokenResp.Body.Close()

	var tokenResult struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}
	if err := json.NewDecoder(tokenResp.Body).Decode(&tokenResult); err != nil {
		return ChannelResponseDto{}, fmt.Errorf("failed to decode token response: %w", err)
	}
	if tokenResult.AccessToken == "" {
		return ChannelResponseDto{}, errors.New("no access token returned from Meta")
	}

	// 3. fetch WABA ID
	wabaReq, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://graph.facebook.com/v20.0/me/whatsapp_business_accounts", nil)
	if err != nil {
		return ChannelResponseDto{}, fmt.Errorf("failed to build WABA request: %w", err)
	}
	wabaReq.Header.Set("Authorization", "Bearer "+tokenResult.AccessToken)

	wabaResp, err := http.DefaultClient.Do(wabaReq)
	if err != nil {
		return ChannelResponseDto{}, fmt.Errorf("failed to fetch WABA: %w", err)
	}
	defer wabaResp.Body.Close()

	var wabaResult struct {
		Data []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}
	if err := json.NewDecoder(wabaResp.Body).Decode(&wabaResult); err != nil {
		return ChannelResponseDto{}, fmt.Errorf("failed to decode WABA response: %w", err)
	}
	if len(wabaResult.Data) == 0 {
		return ChannelResponseDto{}, errors.New("no WhatsApp Business Account found")
	}
	wabaID := wabaResult.Data[0].ID

	// 4. fetch phone number details
	phoneReq, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("https://graph.facebook.com/v20.0/%s/phone_numbers", wabaID), nil)
	if err != nil {
		return ChannelResponseDto{}, fmt.Errorf("failed to build phone number request: %w", err)
	}
	phoneReq.Header.Set("Authorization", "Bearer "+tokenResult.AccessToken)

	phoneResp, err := http.DefaultClient.Do(phoneReq)
	if err != nil {
		return ChannelResponseDto{}, fmt.Errorf("failed to fetch phone numbers: %w", err)
	}
	defer phoneResp.Body.Close()

	var phoneResult struct {
		Data []struct {
			ID                 string `json:"id"`
			DisplayPhoneNumber string `json:"display_phone_number"`
			VerifiedName       string `json:"verified_name"`
		} `json:"data"`
	}
	if err := json.NewDecoder(phoneResp.Body).Decode(&phoneResult); err != nil {
		return ChannelResponseDto{}, fmt.Errorf("failed to decode phone response: %w", err)
	}
	if len(phoneResult.Data) == 0 {
		return ChannelResponseDto{}, errors.New("no phone number found for this WABA")
	}
	phone := phoneResult.Data[0]

	// 5. build auth_config
	authConfig, err := json.Marshal(map[string]interface{}{
		"access_token": tokenResult.AccessToken,
		"expires_at":   nil,
	})
	if err != nil {
		return ChannelResponseDto{}, fmt.Errorf("failed to build auth config: %w", err)
	}

	// 6. build platform_config
	platformConfig, err := json.Marshal(map[string]string{
		"waba_id":         wabaID,
		"phone_number_id": phone.ID,
		"phone_number":    phone.DisplayPhoneNumber,
		"verified_name":   phone.VerifiedName,
		"meta_page_id":    "",
	})
	if err != nil {
		return ChannelResponseDto{}, fmt.Errorf("failed to build platform config: %w", err)
	}

	// 7. update channel — preserve existing fields, only change what connect gives us
	updatedChannel, err := s.repo.UpdateChannel(ctx, repo.UpdateChannelParams{
		ID:              channelID,
		Name:            existing.Name,
		Description:     existing.Description,
		AvatarUrl:       existing.AvatarUrl,
		Status:          repo.ChannelStatusActive,
		StatusReason:    pgtype.Text{String: "Connected via Meta", Valid: true},
		AuthConfig:      authConfig,
		PlatformConfig:  platformConfig,
		Capabilities:    existing.Capabilities,
		WebhookVerified: existing.WebhookVerified,
		WebhookUrl:      existing.WebhookUrl,
	})
	if err != nil {
		return ChannelResponseDto{}, fmt.Errorf("failed to update channel: %w", err)
	}

	return mapChannelToResponse(updatedChannel), nil
}

func (s *channelService) UpdateChannel(ctx context.Context, channelID uuid.UUID, dto UpdateChannelDto) (ChannelResponseDto, error) {
	existing, err := s.repo.GetChannelByID(ctx, channelID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ChannelResponseDto{}, errors.New("channel not found")
		}
		return ChannelResponseDto{}, err
	}

	// only update fields the user is allowed to change
	// never let UpdateChannel overwrite auth_config or platform_config
	name := existing.Name
	if dto.Name != "" {
		name = dto.Name
	}

	description := existing.Description
	if dto.Description != "" {
		description = pgtype.Text{String: dto.Description, Valid: true}
	}

	avatarUrl := existing.AvatarUrl
	if dto.AvatarUrl != "" {
		avatarUrl = pgtype.Text{String: dto.AvatarUrl, Valid: true}
	}

	channel, err := s.repo.UpdateChannel(ctx, repo.UpdateChannelParams{
		ID:              channelID,
		Name:            name,
		Description:     description,
		AvatarUrl:       avatarUrl,
		Status:          existing.Status,
		StatusReason:    existing.StatusReason,
		AuthConfig:      existing.AuthConfig,
		PlatformConfig:  existing.PlatformConfig,
		Capabilities:    existing.Capabilities,
		WebhookVerified: existing.WebhookVerified,
		WebhookUrl:      existing.WebhookUrl,
	})
	if err != nil {
		return ChannelResponseDto{}, fmt.Errorf("failed to update channel: %w", err)
	}

	return mapChannelToResponse(channel), nil
}

func (s *channelService) GetChannelByID(ctx context.Context, channelID uuid.UUID) (ChannelResponseDto, error) {
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
	channels, err := s.repo.ListChannelsByOrganization(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to list channels: %w", err)
	}

	var dtos []ChannelResponseDto
	for _, c := range channels {
		dtos = append(dtos, mapChannelToResponse(c))
	}
	return dtos, nil
}

func (s *channelService) DeleteChannel(ctx context.Context, channelID uuid.UUID) (ChannelResponseDto, error) {
	channel, err := s.repo.DeleteChannel(ctx, channelID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ChannelResponseDto{}, errors.New("channel not found")
		}
		return ChannelResponseDto{}, fmt.Errorf("failed to delete channel: %w", err)
	}
	return mapChannelToResponse(channel), nil
}

func (s *channelService) ChannelWebhook(ctx context.Context, channelID uuid.UUID, payload []byte) (map[string]interface{}, error) {
	channel, err := s.repo.GetChannelByID(ctx, channelID)
	if err != nil {
		return nil, err
	}
	if channel.Status != repo.ChannelStatusActive {
		return nil, errors.New("channel is not active")
	}

	var payloadMap map[string]interface{}
	if err := json.Unmarshal(payload, &payloadMap); err != nil {
		return nil, err
	}
	fmt.Printf("Received Webhook Payload for channel %s:\n%+v\n", channelID, payloadMap)
	return payloadMap, nil
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
