package users

type CreateUserDto struct {
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=6"`
	Phone     string `json:"phone" validate:"required"`
	RoleID    string `json:"roleID" validate:"required"`
	AvatarURL string `json:"avatar_url"`
}

type UpdateUserDto struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	RoleID    string `json:"roleID"`
	AvatarURL string `json:"avatar_url"`
}

type UserResponseDto struct {
	ID             string `json:"id"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	RoleID         string `json:"role_id"`
	OrganizationID string `json:"organization_id"`
	AvatarURL      string `json:"avatar_url"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
	AccessToken    string `json:"access_token"`
}

type LoginResponse struct {
	FirstName   string `json:"first_name"`
	Email       string `json:"email"`
	AccessToken string `json:"access_token"`
}
