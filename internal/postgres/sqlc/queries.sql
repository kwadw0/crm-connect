-- name: CreateUser :one
INSERT INTO users (
  first_name,
  last_name,
  email,
  password,
  phone,
  role,
  avatar_url
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  $6,
  $7
) RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: UpdateUser :one
UPDATE users SET
  first_name = $2,
  last_name = $3,
  email = $4,
  password = $5,
  phone = $6,
  role = $7,
  avatar_url = $8
WHERE id = $1 RETURNING *;

-- name: DeleteUser :one
DELETE FROM users WHERE id = $1 RETURNING *;

-- name: ListUsers :many
SELECT * FROM users;

-- name: UpdateUserOrganization :exec
UPDATE users SET organization_id = $2 WHERE id = $1;

-- name: CreateOrganization :one
INSERT INTO organizations (
  name,
  description,
  website_url,
  industry,
  team_size,
  primary_customer_type,
  primary_use_case,
  owner_role,
  referral_source,
  is_active
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  $6,
  $7,
  $8,
  $9,
  $10
) RETURNING *;

-- name: FindOrganizationByID :one
SELECT * FROM organizations WHERE id = $1;

-- name: FindAllOrganizations :many
SELECT * FROM organizations;

-- name: UpdateOrganization :one
UPDATE organizations SET
  name = $2,
  description = $3,
  website_url = $4,
  industry = $5,
  team_size = $6,
  primary_customer_type = $7,
  primary_use_case = $8,
  owner_role = $9,
  referral_source = $10,
  is_active = $11
WHERE id = $1 RETURNING *;

-- name: DeleteOrganization :one
DELETE FROM organizations WHERE id = $1 RETURNING *;

  
