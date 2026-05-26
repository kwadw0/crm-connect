package channels

import (
	"encoding/json"
	"kwadw0/WhatsCRM/utils"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type Handler interface {
	CreateChannel(w http.ResponseWriter, r *http.Request)
	InitiateConnection(w http.ResponseWriter, r *http.Request)
	ConnectChannel(w http.ResponseWriter, r *http.Request)
	//HandleWebhook(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	service   ChannelService
	validator *validator.Validate
}

func NewHandler(service ChannelService, v *validator.Validate) Handler {
	return &handler{
		service:   service,
		validator: v,
	}
}

// @Summary Create a new channel
// @Tags Channels
// @Accept json
// @Produce json
// @Param request body CreateChannelDto true "Channel details"
// @Success 201 {object} utils.JsonResponse{Data=ChannelResponseDto}
// @Router /channels [post]
func (h *handler) CreateChannel(w http.ResponseWriter, r *http.Request) {
	var dto CreateChannelDto
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		utils.WriteJson(w, http.StatusBadRequest, "Invalid request body", nil, err.Error())
		return
	}

	if err := h.validator.Struct(dto); err != nil {
		utils.WriteJson(w, http.StatusBadRequest, "Validation failed", nil, err.Error())
		return
	}

	channel, err := h.service.CreateChannel(r.Context(), dto)
	if err != nil {
		utils.WriteJson(w, http.StatusInternalServerError, "Failed to create channel", nil, err.Error())
		return
	}

	utils.WriteJson(w, http.StatusCreated, "Channel created successfully", channel, nil)
}

// @Summary Connect a channel (get Meta configuration)
// @Tags Channels
// @Accept json
// @Produce json
// @Param id path string true "Channel UUID"
// @Success 200 {object} utils.JsonResponse{Data=MetaConfigResponse}
// @Router /channels/{id}/initiate [post]
func (h *handler) InitiateConnection(w http.ResponseWriter, r *http.Request) {
	channelIDStr := chi.URLParam(r, "id")
	channelID, err := uuid.Parse(channelIDStr)
	if err != nil {
		utils.WriteJson(w, http.StatusBadRequest, "Invalid Channel ID format", nil, err.Error())
		return
	}

	config, err := h.service.InitiateConnection(r.Context(), channelID)
	if err != nil {
		utils.WriteJson(w, http.StatusInternalServerError, "Failed to initiate channel", nil, err.Error())
		return
	}

	utils.WriteJson(w, http.StatusOK, "Connection configuration fetched", config, nil)
}

// @Summary Connect a channel with Meta OAuth code
// @Tags Channels
// @Accept json
// @Produce json
// @Param id path string true "Channel UUID"
// @Param request body ConnectChannelDto true "OAuth Code, Waba ID, and Phone Number ID"
// @Success 200 {object} utils.JsonResponse{Data=ChannelResponseDto}
// @Router /channels/{id}/connect [post]
func (h *handler) ConnectChannel(w http.ResponseWriter, r *http.Request) {
	channelIDStr := chi.URLParam(r, "id")
	channelID, err := uuid.Parse(channelIDStr)
	if err != nil {
		utils.WriteJson(w, http.StatusBadRequest, "Invalid Channel ID format", nil, err.Error())
		return
	}

	var dto ConnectChannelDto
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		utils.WriteJson(w, http.StatusBadRequest, "Invalid request body", nil, err.Error())
		return
	}

	if err := h.validator.Struct(dto); err != nil {
		utils.WriteJson(w, http.StatusBadRequest, "Validation failed", nil, err.Error())
		return
	}

	channel, err := h.service.ConnectChannel(r.Context(), channelID, dto)
	if err != nil {
		utils.WriteJson(w, http.StatusInternalServerError, "Failed to connect channel", nil, err.Error())
		return
	}

	utils.WriteJson(w, http.StatusOK, "Channel connected successfully", channel, nil)
}

// @Summary Handle Webhook payload for a channel
// @Tags Webhooks
// @Accept json
// @Produce json
// @Param id path string true "Channel UUID"
// @Param payload body interface{} true "Webhook payload"
// @Success 200 {string} string "OK"
// @Router /channels/{id}/webhook [post]
func (h *handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	channelIDStr := chi.URLParam(r, "id")
	channelID, err := uuid.Parse(channelIDStr)
	if err != nil {
		utils.WriteJson(w, http.StatusBadRequest, "Invalid Channel ID format", nil, err.Error())
		return
	}

	var payload json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.WriteJson(w, http.StatusBadRequest, "Invalid webhook payload", nil, err.Error())
		return
	}

	_, err = h.service.ChannelWebhook(r.Context(), channelID, payload)
	if err != nil {
		utils.WriteJson(w, http.StatusInternalServerError, "Failed to process webhook", nil, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
