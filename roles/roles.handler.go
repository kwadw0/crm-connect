package roles

import (
	"encoding/json"
	"kwadw0/WhatsCRM/utils"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type RoleHandler interface {
	CreateRole(w http.ResponseWriter, r *http.Request)
	UpdateRole(w http.ResponseWriter, r *http.Request)
	DeleteRole(w http.ResponseWriter, r *http.Request)
	GetRoleByID(w http.ResponseWriter, r *http.Request)
	GetAllRoles(w http.ResponseWriter, r *http.Request)
}

type roleHandler struct {
	service RoleService
	validator *validator.Validate
}

func NewHandler(service RoleService, v *validator.Validate) RoleHandler {
	return &roleHandler{service: service, validator: v}
}

// @Summary Create a role
// @Tags Roles
// @Accept json
// @Produce json
// @Param request body CreateRoleDTO true "Role details"
// @Success 201 {object} utils.JsonResponse{Data=RoleResponseDTO}
// @Router /roles [post]
func (h *roleHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	var dto CreateRoleDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		utils.WriteJson(w, http.StatusBadRequest, "Invalid request body", nil, err.Error())
		return
	}

	if err := h.validator.Struct(dto); err != nil {
		utils.WriteJson(w, http.StatusBadRequest, "Validation failed", nil, err.Error())
		return
	}

	role, err := h.service.CreateRole(r.Context(), dto)
	if err != nil {
		utils.WriteJson(w, http.StatusInternalServerError, "Failed to create role", nil, err.Error())
		return
	}

	utils.WriteJson(w, http.StatusCreated, "Role created successfully", role, nil)
}

// @Summary Update an existing role
// @Tags Roles
// @Accept json
// @Produce json
// @Param id path string true "Role UUID"
// @Param request body UpdateRoleDTO true "Updated details"
// @Success 200 {object} utils.JsonResponse{Data=RoleResponseDTO}
// @Router /roles/{id} [put]
func (h *roleHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	roleIDStr := chi.URLParam(r, "id")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		utils.WriteJson(w, http.StatusBadRequest, "Invalid role ID format", nil, err.Error())
		return
	}	
	var dto UpdateRoleDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		utils.WriteJson(w, http.StatusBadRequest, "Invalid request body", nil, err.Error())
		return
	}

	if err := h.validator.Struct(dto); err != nil {
		utils.WriteJson(w, http.StatusBadRequest, "Validation failed", nil, err.Error())
		return
	}

	role, err := h.service.UpdateRole(r.Context(), roleID, dto)
	if err != nil {
		utils.WriteJson(w, http.StatusInternalServerError, "Failed to update role", nil, err.Error())
		return
	}

	utils.WriteJson(w, http.StatusOK, "Role updated successfully", role, nil)
}

// @Summary Delete a role
// @Tags Roles
// @Produce json
// @Param id path string true "Role UUID"
// @Success 200 {object} utils.JsonResponse{Data=RoleResponseDTO}
// @Router /roles/{id} [delete]
func (h *roleHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	roleIDStr := chi.URLParam(r, "id")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		utils.WriteJson(w, http.StatusBadRequest, "Invalid role ID format", nil, err.Error())
		return
	}	

	role, err := h.service.DeleteRole(r.Context(), roleID)
	if err != nil {
		utils.WriteJson(w, http.StatusInternalServerError, "Failed to delete role", nil, err.Error())
		return
	}

	utils.WriteJson(w, http.StatusOK, "Role deleted successfully", role, nil)
}

// @Summary Get a role by ID
// @Tags Roles
// @Produce json
// @Param id path string true "Role UUID"
// @Success 200 {object} utils.JsonResponse{Data=RoleResponseDTO}
// @Router /roles/{id} [get]
func (h *roleHandler) GetRoleByID(w http.ResponseWriter, r *http.Request) {
	roleIDStr := chi.URLParam(r, "id")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		utils.WriteJson(w, http.StatusBadRequest, "Invalid role ID format", nil, err.Error())
		return
	} 	

	role, err := h.service.GetRoleByID(r.Context(), roleID)
	if err != nil {
		utils.WriteJson(w, http.StatusInternalServerError, "Failed to get role", nil, err.Error())
		return
	}

	utils.WriteJson(w, http.StatusOK, "Role fetched successfully", role, nil)
}

// @Summary List all roles
// @Tags Roles
// @Produce json
// @Success 200 {object} utils.JsonResponse{Data=[]RoleResponseDTO}
// @Router /roles [get]
func (h *roleHandler) GetAllRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.service.GetAllRoles(r.Context())
	if err != nil {
		utils.WriteJson(w, http.StatusInternalServerError, "Failed to get roles", nil, err.Error())
		return
	}

	utils.WriteJson(w, http.StatusOK, "Roles fetched successfully", roles, nil)
}
