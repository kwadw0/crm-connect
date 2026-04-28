package channels

import (
	"encoding/json"

)

type CreateChannelDto struct {
    OrganizationID  string `json:"organization_id"  validate:"required,uuid"`
    Name            string `json:"name"             validate:"required"`
    Description     string `json:"description"`
    ChannelPlatform string `json:"channel_platform" validate:"required,oneof=whatsapp telegram instagram facebook twitter linkedin email sms other"`
    AvatarUrl       string `json:"avatar_url"`
}

// type CreateChannelDto struct {
// 	Name            string          `json:"name" validate:"required"`
// 	OrganizationID  string          `json:"organization_id" validate:"required"`
// 	Description     string          `json:"description"`
// 	ChannelPlatform string          `json:"channel_platform" validate:"required"`
// 	AvatarUrl       string          `json:"avatar_url"`
// 	AuthConfig      json.RawMessage `json:"auth_config" validate:"required"`
// 	PlatformConfig  json.RawMessage `json:"platform_config" validate:"required"`
// 	Capabilities    json.RawMessage `json:"capabilities"`
// 	WebhookUrl      string          `json:"webhook_url"`
// }

type UpdateChannelDto struct {
	Name            string          `json:"name"`
	Description     string          `json:"description"`
	AvatarUrl       string          `json:"avatar_url"`
	Status          string          `json:"status"`
	StatusReason    string          `json:"status_reason"`
	AuthConfig      json.RawMessage `json:"auth_config"`
	PlatformConfig  json.RawMessage `json:"platform_config"`
	Capabilities    json.RawMessage `json:"capabilities"`
	WebhookUrl      string          `json:"webhook_url"`
}

type ChannelResponseDto struct {
	ID              string          `json:"id"`
	OrganizationID  string          `json:"organization_id"`
	Name            string          `json:"name"`
	Description     string          `json:"description"`
	ChannelPlatform string          `json:"channel_platform"`
	AvatarUrl       string          `json:"avatar_url"`
	Status          string          `json:"status"`
	StatusReason    string          `json:"status_reason"`
	AuthConfig      json.RawMessage `json:"auth_config"`
	PlatformConfig  json.RawMessage `json:"platform_config"`
	Capabilities    json.RawMessage `json:"capabilities"`
	WebhookVerified bool            `json:"webhook_verified"`
	WebhookUrl      string          `json:"webhook_url"`
	CreatedAt       string          `json:"created_at"`
	UpdatedAt       string          `json:"updated_at"`
}
type MetaConfigResponse struct {
	ConfigID string `json:"config_id"`
	AppID    string `json:"app_id"`
}
