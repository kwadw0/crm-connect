-- ============================================================
-- USERS
-- ============================================================

-- name: CreateUser :one
INSERT INTO users (
  first_name,
  last_name,
  email,
  password,
  phone,
  role_id,
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
  role_id = $7,
  avatar_url = $8
WHERE id = $1 RETURNING *;

-- name: DeleteUser :one
DELETE FROM users WHERE id = $1 RETURNING *;

-- name: ListUsers :many
SELECT * FROM users;

-- name: UpdateUserOrganization :one
UPDATE users SET
  organization_id = $2
WHERE id = $1 RETURNING *;

-- name: UpdateUserRefreshToken :one
UPDATE users SET
  refresh_token = $2,
  refresh_token_expires_at = $3
WHERE id = $1 RETURNING *;

-- name: GetUserByRefreshToken :one
SELECT * FROM users
WHERE refresh_token = $1
  AND refresh_token_expires_at > now();

-- name: RevokeRefreshToken :exec
UPDATE users SET
  refresh_token = NULL,
  refresh_token_expires_at = NULL
WHERE id = $1;


-- ============================================================
-- ORGANIZATIONS
-- ============================================================

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

-- name: GetOrganizationByID :one
SELECT * FROM organizations WHERE id = $1;

-- name: ListOrganizations :many
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


-- ============================================================
-- ROLES
-- ============================================================

-- name: CreateRole :one
INSERT INTO roles (
  name,
  description
) VALUES (
  $1,
  $2
) RETURNING *;

-- name: GetRoleByID :one
SELECT * FROM roles WHERE id = $1;

-- name: GetRoleByName :one
SELECT * FROM roles WHERE name = $1;

-- name: ListRoles :many
SELECT * FROM roles;

-- name: UpdateRole :one
UPDATE roles SET
  name = $2,
  description = $3
WHERE id = $1 RETURNING *;

-- name: DeleteRole :one
DELETE FROM roles WHERE id = $1 RETURNING *;


-- ============================================================
-- CHANNELS
-- ============================================================

-- name: CreateChannel :one
INSERT INTO channels (
  organization_id,
  name,
  description,
  channel_platform,
  avatar_url,
  status,
  status_reason,
  auth_config,
  platform_config,
  capabilities,
  webhook_verified,
  webhook_url
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
  $10,
  $11,
  $12
) RETURNING *;

-- name: GetChannelByID :one
SELECT * FROM channels WHERE id = $1;

-- name: ListChannels :many
SELECT * FROM channels;

-- name: ListChannelsByOrganization :many
SELECT * FROM channels WHERE organization_id = $1;

-- name: ListChannelsByPlatform :many
SELECT * FROM channels
WHERE organization_id = $1
  AND channel_platform = $2;

-- name: UpdateChannel :one
UPDATE channels SET
  name = $2,
  description = $3,
  avatar_url = $4,
  status = $5,
  status_reason = $6,
  auth_config = $7,
  platform_config = $8,
  capabilities = $9,
  webhook_verified = $10,
  webhook_url = $11
WHERE id = $1 RETURNING *;

-- name: UpdateChannelStatus :one
UPDATE channels SET
  status = $2,
  status_reason = $3
WHERE id = $1 RETURNING *;

-- name: UpdateChannelAuthConfig :one
UPDATE channels SET
  auth_config = $2
WHERE id = $1 RETURNING *;

-- name: UpdateChannelWebhook :one
UPDATE channels SET
  webhook_verified = $2,
  webhook_url = $3
WHERE id = $1 RETURNING *;

-- name: DeleteChannel :one
DELETE FROM channels WHERE id = $1 RETURNING *;