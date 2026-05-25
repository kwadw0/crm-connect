package channels

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"log/slog"

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
	defaultPin      string
}

func NewChannelService(
	channelRepo *repo.Queries,
	orgService organizations.OrganizationService,
	metaConfigID,
	metaAppID,
	metaAppSecret,
	metaRedirectURI,
	defaultPin string,
) ChannelService {
	return &channelService{
		repo:            channelRepo,
		orgService:      orgService,
		metaConfigID:    metaConfigID,
		metaAppID:       metaAppID,
		metaAppSecret:   metaAppSecret,
		metaRedirectURI: metaRedirectURI,
		defaultPin:      defaultPin,
	}
}

const graphAPIVersion = "v25.0"
 
// httpClient is a package-level client with a timeout.
// Inject this into your service struct instead of using http.DefaultClient.
var httpClient = &http.Client{Timeout: 10 * time.Second}
 

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
 
	// -------------------------------------------------------------------------
	// 1. Fetch existing channel to preserve all fields.
	// -------------------------------------------------------------------------
	existing, err := s.repo.GetChannelByID(ctx, channelID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ChannelResponseDto{}, errors.New("channel not found")
		}
		return ChannelResponseDto{}, err
	}
 
	slog.Info("Starting ConnectChannel", "channel_id", channelID)
 
	// -------------------------------------------------------------------------
	// 2. Exchange the embedded signup code for an access token.
	//
	//    Meta returns a "business integration system user access token" here.
	//    This is already long-lived for the embedded signup flow, but we
	//    immediately exchange it for a proper long-lived token (60-day TTL)
	//    to be safe and consistent across all token types.
	//
	//    Rules:
	//      - Must be a GET request with query params (NOT a POST form).
	//      - Do NOT include redirect_uri for embedded signup code exchange.
	// -------------------------------------------------------------------------
	tokenURL := fmt.Sprintf(
		"https://graph.facebook.com/%s/oauth/access_token?client_id=%s&client_secret=%s&code=%s",
		graphAPIVersion,
		url.QueryEscape(s.metaAppID),
		url.QueryEscape(s.metaAppSecret),
		url.QueryEscape(dto.Code),
	)
 
	tokenReq, err := http.NewRequestWithContext(ctx, http.MethodGet, tokenURL, nil)
	if err != nil {
		return ChannelResponseDto{}, fmt.Errorf("failed to build token request: %w", err)
	}
 
	tokenResp, err := httpClient.Do(tokenReq)
	if err != nil {
		return ChannelResponseDto{}, fmt.Errorf("failed to exchange code with Meta: %w", err)
	}
	defer tokenResp.Body.Close()
 
	tokenBody, err := io.ReadAll(tokenResp.Body)
	if err != nil {
		return ChannelResponseDto{}, fmt.Errorf("failed to read token response: %w", err)
	}
 
	slog.Info("Meta token exchange response", "status", tokenResp.StatusCode, "body", string(tokenBody))
 
	if tokenResp.StatusCode != http.StatusOK {
		return ChannelResponseDto{}, fmt.Errorf("meta token exchange failed (%d): %s", tokenResp.StatusCode, string(tokenBody))
	}
 
	// Parse token — Meta returns JSON for the embedded signup code exchange.
	shortLivedToken := ""
	var tokenJSON struct {
		AccessToken string `json:"access_token"`
		Error       *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(tokenBody, &tokenJSON); err == nil {
		if tokenJSON.Error != nil {
			return ChannelResponseDto{}, fmt.Errorf("meta token exchange error: %s", tokenJSON.Error.Message)
		}
		shortLivedToken = tokenJSON.AccessToken
	}
	// Fallback: some responses are a raw token string (not JSON-wrapped).
	if shortLivedToken == "" {
		shortLivedToken = strings.TrimSpace(string(tokenBody))
	}
	// Guard against HTML error pages or garbage responses.
	if shortLivedToken == "" || strings.HasPrefix(shortLivedToken, "<") {
		return ChannelResponseDto{}, errors.New("invalid or empty token response from Meta")
	}
 
	// -------------------------------------------------------------------------
	// 2b. Exchange the short-lived token for a long-lived token (60-day TTL).
	//
	//     Even though embedded signup tokens are already long-lived in practice,
	//     this explicit exchange guarantees a 60-day token and gives us a known
	//     expiry time to store in auth_config.
	// -------------------------------------------------------------------------
	longLivedToken, expiresAt, err := s.exchangeForLongLivedToken(ctx, shortLivedToken)
	if err != nil {
		// Non-fatal: fall back to the original token if exchange fails.
		// Log the error so you can investigate, but don't block the connection.
		slog.Warn("Failed to exchange for long-lived token, using original", "error", err)
		longLivedToken = shortLivedToken
		expiresAt = time.Now().Add(1 * time.Hour) // conservative fallback TTL
	}
 
	// -------------------------------------------------------------------------
	// 3. Use the WABA ID and phone number ID from the FINISH event (in dto).
	//    These are captured by the frontend from the session logging message
	//    event — do NOT re-fetch from /me/whatsapp_business_accounts.
	// -------------------------------------------------------------------------
	wabaID := dto.WabaID
	phoneNumberID := dto.PhoneNumberID
 
	// -------------------------------------------------------------------------
	// 4. Fetch phone number display details.
	// -------------------------------------------------------------------------
	phoneURL := fmt.Sprintf(
		"https://graph.facebook.com/%s/%s?fields=display_phone_number,verified_name,account_mode",
		graphAPIVersion,
		phoneNumberID,
	)
	phoneReq, err := http.NewRequestWithContext(ctx, http.MethodGet, phoneURL, nil)
	if err != nil {
		return ChannelResponseDto{}, fmt.Errorf("failed to build phone number request: %w", err)
	}
	phoneReq.Header.Set("Authorization", "Bearer "+longLivedToken)
 
	phoneResp, err := httpClient.Do(phoneReq)
	if err != nil {
		return ChannelResponseDto{}, fmt.Errorf("failed to fetch phone number details: %w", err)
	}
	defer phoneResp.Body.Close()
 
	phoneBody, _ := io.ReadAll(phoneResp.Body)
	slog.Info("Meta phone response", "status", phoneResp.StatusCode, "body", string(phoneBody))
 
	if phoneResp.StatusCode != http.StatusOK {
		return ChannelResponseDto{}, fmt.Errorf("failed to fetch phone number details (%d): %s", phoneResp.StatusCode, string(phoneBody))
	}
 
	var phoneResult struct {
		ID                 string `json:"id"`
		DisplayPhoneNumber string `json:"display_phone_number"`
		VerifiedName       string `json:"verified_name"`
		AccountMode        string `json:"account_mode"` // "LIVE" or "SANDBOX"
	}
	if err := json.Unmarshal(phoneBody, &phoneResult); err != nil {
		return ChannelResponseDto{}, fmt.Errorf("failed to decode phone number response: %w", err)
	}
 
	// -------------------------------------------------------------------------
	// 5. Subscribe your app to webhooks on the customer's WABA.
	//
	//    IMPORTANT: This requires your app-level webhook to already be verified
	//    and subscribed in the Meta App Dashboard (App Dashboard → WhatsApp →
	//    Configuration). If it isn't, Meta returns error #100.
	//
	//    This call links the customer's WABA to your app so webhook events
	//    (messages, statuses) flow to your registered callback URL.
	// -------------------------------------------------------------------------
	subURL := fmt.Sprintf("https://graph.facebook.com/%s/%s/subscribed_apps", graphAPIVersion, wabaID)
	subReq, err := http.NewRequestWithContext(ctx, http.MethodPost, subURL, nil)
	if err != nil {
		return ChannelResponseDto{}, fmt.Errorf("failed to build webhook subscription request: %w", err)
	}
	subReq.Header.Set("Authorization", "Bearer "+longLivedToken)
 
	subResp, err := httpClient.Do(subReq)
	if err != nil {
		return ChannelResponseDto{}, fmt.Errorf("failed to subscribe to WABA webhooks: %w", err)
	}
	defer subResp.Body.Close()
 
	subBody, _ := io.ReadAll(subResp.Body)
	slog.Info("Meta subscribed_apps response", "status", subResp.StatusCode, "body", string(subBody))
 
	if subResp.StatusCode != http.StatusOK {
		return ChannelResponseDto{}, fmt.Errorf("webhook subscription failed (%d): %s", subResp.StatusCode, string(subBody))
	}
 
	// Verify Meta confirmed the subscription (200 with success:false is possible).
	var subResult struct {
		Success bool `json:"success"`
	}
	if err := json.Unmarshal(subBody, &subResult); err != nil || !subResult.Success {
		return ChannelResponseDto{}, fmt.Errorf("webhook subscription did not confirm success: %s", string(subBody))
	}
 
	// -------------------------------------------------------------------------
	// 6. Register the phone number for Cloud API (conditional).
	//
	//    Only needed when migrating a number from on-premise BSP to Cloud API.
	//    Numbers onboarded fresh via embedded signup are already on Cloud API
	//    and will return an error if you call /register again.
	//
	//    We check account_mode: if it's "SANDBOX" or the number isn't in LIVE
	//    mode yet, we attempt registration. Otherwise we skip it.
	//
	//    The error is treated as non-fatal — log it and continue, since the
	//    most common case (already on Cloud API) should not block the flow.
	// -------------------------------------------------------------------------
	if phoneResult.AccountMode != "LIVE" {
		if regErr := s.registerPhoneNumber(ctx, phoneNumberID, longLivedToken); regErr != nil {
			slog.Warn("Phone number registration skipped or failed (may already be registered)",
				"phone_number_id", phoneNumberID,
				"error", regErr,
			)
			// Non-fatal: continue. If the number truly can't send messages,
			// it will surface when the user tries to send their first message.
		}
	} else {
		slog.Info("Phone number already in LIVE/Cloud API mode, skipping registration",
			"phone_number_id", phoneNumberID,
		)
	}
 
	// -------------------------------------------------------------------------
	// 7. Build auth_config and platform_config for storage.
	// -------------------------------------------------------------------------
	authConfig, err := json.Marshal(map[string]interface{}{
		"access_token": longLivedToken,
		"token_type":   "long_lived",
		"issued_at":    time.Now().UTC().Format(time.RFC3339),
		"expires_at":   expiresAt.UTC().Format(time.RFC3339),
	})
	if err != nil {
		return ChannelResponseDto{}, fmt.Errorf("failed to build auth config: %w", err)
	}
 
	platformConfig, err := json.Marshal(map[string]string{
		"waba_id":         wabaID,
		"phone_number_id": phoneNumberID,
		"phone_number":    phoneResult.DisplayPhoneNumber,
		"verified_name":   phoneResult.VerifiedName,
		"account_mode":    phoneResult.AccountMode,
		"meta_page_id":    "",
	})
	if err != nil {
		return ChannelResponseDto{}, fmt.Errorf("failed to build platform config: %w", err)
	}
 
	// -------------------------------------------------------------------------
	// 8. Persist the channel update.
	// -------------------------------------------------------------------------
	updatedChannel, err := s.repo.UpdateChannel(ctx, repo.UpdateChannelParams{
		ID:              channelID,
		Name:            existing.Name,
		Description:     existing.Description,
		AvatarUrl:       existing.AvatarUrl,
		Status:          repo.ChannelStatusActive,
		StatusReason:    pgtype.Text{String: "Connected via Meta Embedded Signup", Valid: true},
		AuthConfig:      authConfig,
		PlatformConfig:  platformConfig,
		Capabilities:    existing.Capabilities,
		WebhookVerified: existing.WebhookVerified,
		WebhookUrl:      existing.WebhookUrl,
	})
	if err != nil {
		return ChannelResponseDto{}, fmt.Errorf("failed to update channel: %w", err)
	}
 
	slog.Info("Channel connected successfully",
		"channel_id", channelID,
		"waba_id", wabaID,
		"phone_number_id", phoneNumberID,
		"account_mode", phoneResult.AccountMode,
		"token_expires_at", expiresAt.UTC().Format(time.RFC3339),
	)
 
	return mapChannelToResponse(updatedChannel), nil
}
 
// exchangeForLongLivedToken exchanges a short-lived user access token for a
// long-lived token with a ~60-day TTL.
//
// Endpoint: GET /oauth/access_token
//   - grant_type=fb_exchange_token
//   - client_id, client_secret, fb_exchange_token
//
// Returns the long-lived token string and its expiry time.
func (s *channelService) exchangeForLongLivedToken(ctx context.Context, shortLivedToken string) (string, time.Time, error) {
	llURL := fmt.Sprintf(
		"https://graph.facebook.com/%s/oauth/access_token?grant_type=fb_exchange_token&client_id=%s&client_secret=%s&fb_exchange_token=%s",
		graphAPIVersion,
		url.QueryEscape(s.metaAppID),
		url.QueryEscape(s.metaAppSecret),
		url.QueryEscape(shortLivedToken),
	)
 
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, llURL, nil)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to build long-lived token request: %w", err)
	}
 
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to request long-lived token: %w", err)
	}
	defer resp.Body.Close()
 
	body, _ := io.ReadAll(resp.Body)
	slog.Info("Meta long-lived token exchange response", "status", resp.StatusCode, "body", string(body))
 
	if resp.StatusCode != http.StatusOK {
		return "", time.Time{}, fmt.Errorf("long-lived token exchange failed (%d): %s", resp.StatusCode, string(body))
	}
 
	var result struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int64  `json:"expires_in"` // seconds
		Error       *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", time.Time{}, fmt.Errorf("failed to parse long-lived token response: %w", err)
	}
	if result.Error != nil {
		return "", time.Time{}, fmt.Errorf("long-lived token error: %s", result.Error.Message)
	}
	if result.AccessToken == "" {
		return "", time.Time{}, errors.New("empty access token in long-lived token response")
	}
 
	expiresAt := time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)
	return result.AccessToken, expiresAt, nil
}
 
// registerPhoneNumber registers a phone number for the WhatsApp Cloud API.
// This is only needed when migrating from on-premise BSP to Cloud API.
// Numbers already on Cloud API will fail this call — that's expected and handled
// by the caller as non-fatal.
func (s *channelService) registerPhoneNumber(ctx context.Context, phoneNumberID, accessToken string) error {
	regURL := fmt.Sprintf("https://graph.facebook.com/%s/%s/register", graphAPIVersion, phoneNumberID)
	regPayload, _ := json.Marshal(map[string]string{
		"messaging_product": "whatsapp",
		"pin":               s.defaultPin, // 6-digit PIN from config
	})
 
	regReq, err := http.NewRequestWithContext(ctx, http.MethodPost, regURL, bytes.NewReader(regPayload))
	if err != nil {
		return fmt.Errorf("failed to build register request: %w", err)
	}
	regReq.Header.Set("Authorization", "Bearer "+accessToken)
	regReq.Header.Set("Content-Type", "application/json")
 
	regResp, err := httpClient.Do(regReq)
	if err != nil {
		return fmt.Errorf("failed to call register endpoint: %w", err)
	}
	defer regResp.Body.Close()
 
	regBody, _ := io.ReadAll(regResp.Body)
	slog.Info("Meta register phone response", "status", regResp.StatusCode, "body", string(regBody))
 
	if regResp.StatusCode != http.StatusOK {
		return fmt.Errorf("phone number registration failed (%d): %s", regResp.StatusCode, string(regBody))
	}
 
	return nil
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
