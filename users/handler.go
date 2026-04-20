package users

import (
	"encoding/json"
	"kwadw0/WhatsCRM/utils"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler interface {
	CreateUser(w http.ResponseWriter, r *http.Request)
	GetAllUsers(w http.ResponseWriter, r *http.Request)
	UpdateUser(w http.ResponseWriter, r *http.Request)
	DeleteUser(w http.ResponseWriter, r *http.Request)
	GetUserByID(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	service Service
}

func NewHandler(service Service) Handler {
	return &handler{service: service}
}

// @Summary Create a user
// @Tags Users
// @Accept json
// @Produce json
// @Param request body CreateUserDto true "User details"
// @Success 201 {object} utils.JsonResponse{Data=UserResponseDto}
// @Router /users [post]
func (h *handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var dto CreateUserDto
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		utils.WriteJson(w, http.StatusBadRequest, "Invalid request body", nil, err.Error())
		return
	}

	user, err := h.service.CreateUser(r.Context(), dto)
	if err != nil {
		utils.WriteJson(w, http.StatusInternalServerError, "Failed to create user", nil, err.Error())
		return
	}

	utils.WriteJson(w, http.StatusCreated, "User created successfully", user, nil)
}

// @Summary List all users
// @Tags Users
// @Produce json
// @Success 200 {object} utils.JsonResponse{Data=[]UserResponseDto}
// @Router /users [get]
func (h *handler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.GetAllUsers(r.Context())
	if err != nil {
		utils.WriteJson(w, http.StatusInternalServerError, "Failed to fetch users", nil, err.Error())
		return
	}

	utils.WriteJson(w, http.StatusOK, "Users fetched successfully", users, nil)
}

// @Summary Update an existing user
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User UUID"
// @Param request body UpdateUserDto true "Updated details"
// @Success 200 {object} utils.JsonResponse{Data=UserResponseDto}
// @Router /users/{id} [put]
func (h *handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	// 1. Grab the "id" param from chi and parse it to UUID
	userIDStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.WriteJson(w, http.StatusBadRequest, "Invalid User ID format", nil, err.Error())
		return
	}

	var dto UpdateUserDto
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		utils.WriteJson(w, http.StatusBadRequest, "Invalid request body", nil, err.Error())
		return
	}

	// 2. Pass the parsed UUID to your refactored service
	user, err := h.service.UpdateUser(r.Context(), userID, dto)
	if err != nil {
		utils.WriteJson(w, http.StatusInternalServerError, "Failed to update user", nil, err.Error())
		return
	}

	utils.WriteJson(w, http.StatusOK, "User updated successfully", user, nil)
}

// @Summary Delete a user
// @Tags Users
// @Produce json
// @Param id path string true "User UUID"
// @Success 200 {object} utils.JsonResponse{Data=UserResponseDto}
// @Router /users/{id} [delete]
func (h *handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.WriteJson(w, http.StatusBadRequest, "Invalid User ID format", nil, err.Error())
		return
	}

	user, err := h.service.DeleteUser(r.Context(), userID)
	if err != nil {
		utils.WriteJson(w, http.StatusInternalServerError, "Failed to delete user", nil, err.Error())
		return
	}

	utils.WriteJson(w, http.StatusOK, "User deleted successfully", user, nil)
}

// @Summary Get a user by ID
// @Tags Users
// @Produce json
// @Param id path string true "User UUID"
// @Success 200 {object} utils.JsonResponse{Data=UserResponseDto}
// @Router /users/{id} [get]
func (h *handler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.WriteJson(w, http.StatusBadRequest, "Invalid User ID format", nil, err.Error())
		return
	}

	user, err := h.service.GetUserByID(r.Context(), userID)
	if err != nil {
		utils.WriteJson(w, http.StatusInternalServerError, "Failed to fetch user", nil, err.Error())
		return
	}

	utils.WriteJson(w, http.StatusOK, "User fetched successfully", user, nil)
}