package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"kwadw0/WhatsCRM/utils"
	"net/http"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type Handler interface {
	RegisterUser(w http.ResponseWriter, r *http.Request)
	LoginUser(w http.ResponseWriter, r *http.Request)
}

type authHandler struct {
	service Service
}

func AuthHandler(service Service) Handler {
	return &authHandler{service: service}
}



// @Summary Register a new user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RegisterUserDto true "Registration details"
// @Success 201 {object} utils.JsonResponse{Data=users.UserResponseDto}
// @Router /auth/register [post]
func (h *authHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var dto RegisterUserDto
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		utils.WriteJson(w, http.StatusBadRequest, "Invalid request body", nil, err.Error())
		return
	}

	user, err := h.service.RegisterUser(r.Context(), dto)
	if err != nil {
		var pgErr *pgconn.PgError
		
		// If it's a database error, let's check what kind it is!
		if errors.As(err, &pgErr) {
			// 1. Check for Duplicate Email or Duplicate Phone
			if pgErr.Code == pgerrcode.UniqueViolation {
				var msg string
				
				// Postgres tells us exactly which constraint was blocked!
				if pgErr.ConstraintName == "users_email_key" {
					msg = fmt.Sprintf("A user with email %s already exists", dto.Email)
					utils.WriteJson(w, http.StatusConflict, "Email already in use", nil, msg)
					return
				}
				
				if pgErr.ConstraintName == "users_phone_key" {
					msg = fmt.Sprintf("A user with phone number %s already exists", dto.Phone)
					utils.WriteJson(w, http.StatusConflict, "Phone number already in use", nil, msg)
					return
				}
			}
			
			// 2. Add other checks here if needed in the future
			// if pgErr.Code == pgerrcode.ForeignKeyViolation { ... }
		}
		
		// If it's not a known database error, just return a generic 500
		utils.WriteJson(w, http.StatusInternalServerError, "Failed to register user", nil, err.Error())
		return
	}

	utils.WriteJson(w, http.StatusCreated, "User registered successfully", user, nil)
}

// @Summary Login an existing user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginUserDto true "Login credentials"
// @Success 200 {object} utils.JsonResponse{Data=users.LoginResponse}
// @Router /auth/login [post]
func (h *authHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var dto LoginUserDto
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		utils.WriteJson(w, http.StatusBadRequest, "Invalid request body", nil, err.Error())
		return
	}

	user, err := h.service.LoginUser(r.Context(), dto)
	if err != nil {
		if errors.Is(err, utils.ErrInvalidCredentials) {
			utils.WriteJson(w, http.StatusUnauthorized, "Invalid credentials", nil, "Incorrect email or password")
			return
		}
		
		// 500 Internal Server error for anything corrupted like "hashedSecret too short"
		utils.WriteJson(w, http.StatusInternalServerError, "Internal Server Error", nil, err.Error())
		return
	}

	utils.WriteJson(w, http.StatusOK, "User logged in successfully", user, nil)
}