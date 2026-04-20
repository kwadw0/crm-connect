package organizations

import (
	"encoding/json"
	"errors"
	"kwadw0/WhatsCRM/auth"
	"kwadw0/WhatsCRM/users"
	"kwadw0/WhatsCRM/utils"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type OrganizationHandler interface {
	AddOrganization(w http.ResponseWriter, r *http.Request)
	GetOrganizationByID(w http.ResponseWriter, r *http.Request)
	GetCurrentUserOrganization(w http.ResponseWriter, r *http.Request)
	ListOrganizations(w http.ResponseWriter, r *http.Request)
	UpdateOrganization(w http.ResponseWriter, r *http.Request)
	DeleteOrganization(w http.ResponseWriter, r *http.Request)
}

type organizationHandler struct {
	service     OrganizationService
	userService users.Service
	validator   *validator.Validate
}

func NewOrganizationHandler(service OrganizationService, userService users.Service, v *validator.Validate) OrganizationHandler {
	return &organizationHandler{
		service:     service,
		userService: userService,
		validator:   v,
	}
}

// @Summary Create a new organization
// @Tags Organizations
// @Accept json
// @Produce json
// @Param request body CreateOrganizationDto true "Organization creation details"
// @Success 201 {object} utils.JsonResponse{Data=OrganizationResponseDto}
// @Router /organizations [post]
func (h *organizationHandler) AddOrganization(w http.ResponseWriter, r *http.Request) {
	var dto CreateOrganizationDto
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		utils.WriteJson(w, http.StatusBadRequest, "Invalid request body", nil, err.Error())
		return
	}

	if err := h.validator.Struct(dto); err != nil {
		utils.WriteJson(w, http.StatusBadRequest, "Validation failed", nil, err.Error())
		return
	}

	userIDStr := auth.GetUserIDFromContext(r.Context())
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.WriteJson(w, http.StatusUnauthorized, "Invalid user ID format", nil, err.Error())
		return
	}

	org, err := h.service.AddOrganization(r.Context(), userID, dto)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			utils.WriteJson(w, http.StatusConflict, "Organization name already exists", nil, "Try a unique name")
			return
		}
		utils.WriteJson(w, http.StatusInternalServerError, "Failed to create organization", nil, err.Error())
		return
	}

	utils.WriteJson(w, http.StatusCreated, "Organization created successfully", org, nil)
}

// @Summary Get organization by ID
// @Tags Organizations
// @Produce json
// @Param id path string true "Organization ID"
// @Success 200 {object} utils.JsonResponse{Data=OrganizationResponseDto}
// @Router /organizations/{id} [get]
func (h *organizationHandler) GetOrganizationByID(w http.ResponseWriter, r *http.Request) {
	orgIDStr := chi.URLParam(r, "id")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		utils.WriteJson(w, http.StatusBadRequest, "Invalid Organization ID format", nil, err.Error())
		return
	}

	org, err := h.service.GetOrganizationByID(r.Context(), orgID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			utils.WriteJson(w, http.StatusNotFound, "Organization not found", nil, "no rows in result set")
			return
		}
		utils.WriteJson(w, http.StatusInternalServerError, "Failed to get organization", nil, err.Error())
		return
	}
	utils.WriteJson(w, http.StatusOK, "Organization fetched successfully", org, nil)
}

// @Summary List all organizations
// @Tags Organizations
// @Produce json
// @Success 200 {object} utils.JsonResponse{Data=[]OrganizationResponseDto}
// @Router /organizations [get]
func (h *organizationHandler) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	orgs, err := h.service.ListOrganizations(r.Context())
	if err != nil {
		utils.WriteJson(w, http.StatusInternalServerError, "Failed to fetch organizations", nil, err.Error())
		return
	}
	utils.WriteJson(w, http.StatusOK, "Organizations fetched successfully", orgs, nil)
}

// @Summary Get current user's organization
// @Tags Organizations
// @Produce json
// @Success 200 {object} utils.JsonResponse{Data=OrganizationResponseDto}
// @Router /organizations/current [get]
func (h *organizationHandler) GetCurrentUserOrganization(w http.ResponseWriter, r *http.Request) {
	userIDStr := auth.GetUserIDFromContext(r.Context())
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.WriteJson(w, http.StatusUnauthorized, "Invalid User ID format", nil, err.Error())
		return
	}

	// 1. Fetch the User record using the UserService
	user, err := h.userService.GetUserByID(r.Context(), userID)
	if err != nil {
		utils.WriteJson(w, http.StatusNotFound, "User not found", nil, err.Error())
		return
	}

	// 2. Check if the user even belongs to an organization
	if user.OrganizationID == "" {
		utils.WriteJson(w, http.StatusNotFound, "User does not belong to any organization", nil, "No organization assigned")
		return
	}

	orgID, _ := uuid.Parse(user.OrganizationID)

	// 3. Fetch the actual Organization details
	org, err := h.service.GetOrganizationByID(r.Context(), orgID)
	if err != nil {
		utils.WriteJson(w, http.StatusInternalServerError, "Failed to get organization details", nil, err.Error())
		return
	}

	utils.WriteJson(w, http.StatusOK, "Organization fetched successfully", org, nil)
}

// @Summary Update organization by ID
// @Tags Organizations
// @Accept json
// @Produce json
// @Param id path string true "Organization ID"
// @Param request body UpdateOrganizationDto true "Update details"
// @Success 200 {object} utils.JsonResponse{Data=OrganizationResponseDto}
// @Router /organizations/{id} [put]
func (h *organizationHandler) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	orgIDStr := chi.URLParam(r, "id")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		utils.WriteJson(w, http.StatusBadRequest, "Invalid organization ID format", nil, err.Error())
		return
	}

	var dto UpdateOrganizationDto
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		utils.WriteJson(w, http.StatusBadRequest, "Invalid request body", nil, err.Error())
		return
	}

	org, err := h.service.UpdateOrganizationById(r.Context(), orgID, dto)
	if err != nil {
		utils.WriteJson(w, http.StatusInternalServerError, "Failed updating organization", nil, err.Error())
		return
	}
	utils.WriteJson(w, http.StatusOK, "Organization updated successfully", org, nil)
}

// @Summary Delete organization by ID
// @Tags Organizations
// @Produce json
// @Param id path string true "Organization ID"
// @Success 200 {object} utils.JsonResponse{Data=OrganizationResponseDto}
// @Router /organizations/{id} [delete]
func (h *organizationHandler) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	orgIDStr := chi.URLParam(r, "id")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		utils.WriteJson(w, http.StatusBadRequest, "Invalid organization ID format", nil, err.Error())
		return
	}
	
	org, err := h.service.DeleteOrganization(r.Context(), orgID)
	if err != nil {
		utils.WriteJson(w, http.StatusInternalServerError, "Failed to delete organization", nil, err.Error())
		return
	}
	utils.WriteJson(w, http.StatusOK, "Organization deleted successfully", org, nil)
}