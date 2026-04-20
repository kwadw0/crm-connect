-- +goose Up
-- +goose StatementBegin
CREATE TABLE "organizations" (
  "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  "name" varchar(255) NOT NULL,
  "description" text NULL,
  "website_url" varchar(255) NULL,
  "industry" varchar(255) NULL CHECK (industry IN (
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
  "team_size" varchar(255) NULL CHECK (team_size IN (
                                '1',
                                '2-5',
                                '6-10',
                                '11-50',
                                '51+'
                            )),
  "primary_customer_type" varchar(255) NULL CHECK (primary_customer_type IN (
                                'b2b',
                                'b2c',
                                'both'
                            )),
  "primary_use_case" varchar(255) NOT NULL CHECK (primary_use_case IN (
                                'manage_customer_conversations',
                                'track_orders_and_sales',
                                'automate_follow_ups',
                                'broadcast_to_customers',
                                'organise_contacts',
                                'replace_spreadsheets',
                                'team_collaboration',
                                'other'
                            )),     
  "owner_role" varchar(255) NOT NULL CHECK (owner_role IN (
                                'business_owner',
                                'sales_manager',
                                'customer_support_manager',
                                'marketing_manager',
                                'operations_manager',
                                'freelancer_or_consultant',
                                'developer_or_technical',
                                'other'
                            )),
  "referral_source" varchar(255) NULL CHECK (referral_source IN (
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
  "is_active" boolean DEFAULT true,                          
  "created_at" timestamp DEFAULT now(),
  "updated_at" timestamp DEFAULT now()
);

ALTER TABLE users ADD COLUMN organization_id uuid NULL;
ALTER TABLE users ADD CONSTRAINT users_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE SET NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_organization_id_fkey;
ALTER TABLE users DROP COLUMN IF EXISTS organization_id;
DROP TABLE "organizations";
-- +goose StatementEnd
