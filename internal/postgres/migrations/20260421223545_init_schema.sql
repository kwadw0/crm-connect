-- +goose Up
-- paste everything from schema.sql here

create extension if not exists "uuid-ossp";

-- +goose StatementBegin
create or replace function set_updated_at()
returns trigger as $$
begin
  new.updated_at = now();
  return new;
end;
$$ language plpgsql;
-- +goose StatementEnd

create type channel_platform as enum (
  'whatsapp',
  'telegram',
  'instagram',
  'facebook',
  'twitter',
  'linkedin',
  'email',
  'sms',
  'other'
);

create type channel_status as enum (
  'active',
  'inactive',
  'suspended',
  'pending',
  'disconnected'
);

create table "roles" (
  "id"          uuid         primary key default gen_random_uuid(),
  "name"        varchar(255) not null unique,
  "description" text         null,
  "created_at"  timestamptz  not null default now(),
  "updated_at"  timestamptz  not null default now()
);

insert into "roles" ("name", "description") values
  ('owner',  'Full access to the organization'),
  ('admin',  'Can manage members and settings'),
  ('member', 'Standard access');

create trigger trg_roles_updated_at
  before update on "roles"
  for each row execute function set_updated_at();


create table "organizations" (
  "id"                    uuid         primary key default gen_random_uuid(),
  "name"                  varchar(255) not null unique,
  "description"           text         null,
  "website_url"           varchar(255) null,
  "industry"              varchar(255) null check (industry in (
                            'fashion_and_apparel',
                            'food_and_beverage',
                            'beauty_and_cosmetics',
                            'health_and_wellness',
                            'retail_and_ecommerce',
                            'professional_services',
                            'education_and_tutoring',
                            'real_estate',
                            'logistics_and_delivery',
                            'events_and_entertainment',
                            'tech_and_software',
                            'finance_and_fintech',
                            'agriculture',
                            'travel_and_tourism',
                            'other'
                          )),
  "team_size"             varchar(10)  null check (team_size in (
                            '1',
                            '2-5',
                            '6-10',
                            '11-50',
                            '51+'
                          )),
  "primary_customer_type" varchar(10)  null check (primary_customer_type in (
                            'b2b',
                            'b2c',
                            'both'
                          )),
  "primary_use_case"      varchar(255) not null check (primary_use_case in (
                            'manage_customer_conversations',
                            'track_orders_and_sales',
                            'automate_follow_ups',
                            'broadcast_to_customers',
                            'organise_contacts',
                            'replace_spreadsheets',
                            'team_collaboration',
                            'other'
                          )),
  "owner_role"            varchar(255) not null check (owner_role in (
                            'business_owner',
                            'sales_manager',
                            'customer_support_manager',
                            'marketing_manager',
                            'operations_manager',
                            'freelancer_or_consultant',
                            'developer_or_technical',
                            'other'
                          )),
  "referral_source"       varchar(255) null check (referral_source in (
                            'facebook_group',
                            'instagram',
                            'whatsapp',
                            'google_search',
                            'friend_or_colleague',
                            'youtube',
                            'tiktok',
                            'twitter_x',
                            'product_hunt',
                            'other'
                          )),
  "is_active"             boolean      not null default true,
  "created_at"            timestamptz  not null default now(),
  "updated_at"            timestamptz  not null default now()
);

create index idx_organizations_name      on "organizations"("name");
create index idx_organizations_industry  on "organizations"("industry");
create index idx_organizations_is_active on "organizations"("is_active");

create trigger trg_organizations_updated_at
  before update on "organizations"
  for each row execute function set_updated_at();

create table "users" (
  "id"              uuid         primary key default gen_random_uuid(),
  "first_name"      varchar(255) not null,
  "last_name"       varchar(255) not null,
  "email"           varchar(255) not null unique,
  "password"        varchar(255) not null,
  "phone"           varchar(255) not null unique,
  "role_id"         uuid         not null references roles(id),
  "organization_id" uuid         null references organizations(id) on delete set null,
  "avatar_url"      varchar(255) null,  
  "is_active"       boolean      not null default true,
  "email_verified"  boolean      not null default false,
  "phone_verified"  boolean      not null default false,
  "refresh_token"   text        null,
  "refresh_token_expires_at" timestamptz null,  
  "created_at"      timestamptz  not null default now(),
  "updated_at"      timestamptz  not null default now()
);

create index idx_users_email           on "users"("email");
create index idx_users_phone           on "users"("phone");
create index idx_users_role_id         on "users"("role_id");
create index idx_users_organization_id on "users"("organization_id");

create trigger trg_users_updated_at
  before update on "users"
  for each row execute function set_updated_at();


create table "channels" (
  "id"               uuid             primary key default gen_random_uuid(),
  "organization_id"  uuid             not null references organizations(id) on delete cascade,
  "name"             varchar(255)     not null,
  "description"      text             null,
  "channel_platform" channel_platform not null,
  "avatar_url"       text             null,
  "status"           channel_status   not null default 'pending',
  "status_reason"    text             null,

  -- auth tokens — structure varies per platform
  -- whatsapp / instagram / facebook:
  --   { "access_token": "...", "expires_at": "ISO8601|null" }
  -- telegram:
  --   { "bot_token": "..." }
  "auth_config"      jsonb            not null default '{}',

  -- platform-specific ids — structure varies per platform
  -- whatsapp:
  --   { "waba_id": "...", "phone_number_id": "...", "phone_number": "+233...", "meta_page_id": "..." }
  -- instagram:
  --   { "ig_account_id": "...", "meta_page_id": "...", "username": "@..." }
  -- facebook:
  --   { "meta_page_id": "...", "page_name": "..." }
  -- telegram:
  --   { "bot_id": "...", "bot_username": "@...", "webhook_secret": "..." }
  "platform_config"  jsonb            not null default '{}',

  -- drives ui conditional rendering
  -- { "inbound_messages": true, "outbound_messages": true, "media": true, "templates": true }
  "capabilities"     jsonb            not null default '{}',

  "webhook_verified" boolean          not null default false,
  "webhook_url"      text             null,
  "created_at"       timestamptz      not null default now(),
  "updated_at"       timestamptz      not null default now()
);

create index idx_channels_org      on "channels"("organization_id");
create index idx_channels_platform on "channels"("channel_platform");
create index idx_channels_status   on "channels"("status");
create index idx_channels_config   on "channels" using gin("platform_config");
create index idx_channels_auth     on "channels" using gin("auth_config");

create trigger trg_channels_updated_at
  before update on "channels"
  for each row execute function set_updated_at();
-- +goose Down
drop trigger if exists trg_channels_updated_at on "channels";
drop trigger if exists trg_users_updated_at on "users";
drop trigger if exists trg_organizations_updated_at on "organizations";
drop trigger if exists trg_roles_updated_at on "roles";

drop index if exists idx_channels_auth;
drop index if exists idx_channels_config;
drop index if exists idx_channels_status;
drop index if exists idx_channels_platform;
drop index if exists idx_channels_org;

drop index if exists idx_users_organization_id;
drop index if exists idx_users_role_id;
drop index if exists idx_users_phone;
drop index if exists idx_users_email;

drop index if exists idx_organizations_is_active;
drop index if exists idx_organizations_industry;
drop index if exists idx_organizations_name;

drop table if exists "channels";
drop table if exists "users";
drop table if exists "organizations";
drop table if exists "roles";

drop type if exists channel_status;
drop type if exists channel_platform;

drop function if exists set_updated_at;