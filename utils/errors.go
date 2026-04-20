package utils

import "errors"

var ErrInvalidCredentials = errors.New("invalid email or password")
var ErrOrgNotFound = errors.New("organization not found")
var ErrOrgAlreadyExists = errors.New("organization with this name already exists")
var ErrInvalidInput = errors.New("the provided input is invalid or missing required fields")
