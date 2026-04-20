package organizations

type CreateOrganizationDto struct {
	Name                string `json:"name" validate:"required"`
	Description         string `json:"description" validate:"required"`
	WebsiteUrl          string `json:"website_url"`
	Industry            string `json:"industry" validate:"required"`
	TeamSize            string `json:"team_size" validate:"required"`
	PrimaryCustomerType string `json:"primary_customer_type" validate:"required"`
	PrimaryUseCase      string `json:"primary_use_case" validate:"required"`
	OwnerRole           string `json:"owner_role" validate:"required"`
	ReferralSource      string `json:"referral_source"`
}

type UpdateOrganizationDto struct {
	Name                string `json:"name"`
	Description         string `json:"description"`
	WebsiteUrl          string `json:"website_url"`
	Industry            string `json:"industry"`
	TeamSize            string `json:"team_size"`
	PrimaryCustomerType string `json:"primary_customer_type"`
	PrimaryUseCase      string `json:"primary_use_case"`
	OwnerRole           string `json:"owner_role"`
	ReferralSource      string `json:"referral_source"`
	IsActive            bool   `json:"is_active"`
}

type OrganizationResponseDto struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	Description         string `json:"description"`
	WebsiteUrl          string `json:"website_url"`
	Industry            string `json:"industry"`
	TeamSize            string `json:"team_size"`
	PrimaryCustomerType string `json:"primary_customer_type"`
	PrimaryUseCase      string `json:"primary_use_case"`
	OwnerRole           string `json:"owner_role"`
	ReferralSource      string `json:"referral_source"`
	IsActive            bool   `json:"is_active"`
	CreatedAt           string `json:"created_at"`
	UpdatedAt           string `json:"updated_at"`
}