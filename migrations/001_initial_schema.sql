-- public.products definition

-- Drop table

-- DROP TABLE public.products;

CREATE TABLE IF NOT EXISTS public.products (
	id uuid DEFAULT uuid_generate_v4() NOT NULL,
	model_number varchar(100) NOT NULL,
	brand varchar(100) NOT NULL,
	"type" varchar(50) NOT NULL,
	"size" varchar(50) NULL,
	warranty_years int4 DEFAULT 10 NULL,
	description text NULL,
	is_active bool DEFAULT true NULL,
	created_at timestamp DEFAULT now() NULL,
	updated_at timestamp DEFAULT now() NULL,
	CONSTRAINT chk_products_warranty_years_non_negative CHECK ((warranty_years >= 0)),
	CONSTRAINT products_pkey PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_products_active_model ON public.products USING btree (model_number) WHERE (is_active = true);
CREATE INDEX IF NOT EXISTS idx_products_brand ON public.products USING btree (brand);
CREATE INDEX IF NOT EXISTS idx_products_is_active ON public.products USING btree (is_active);
CREATE INDEX IF NOT EXISTS idx_products_model_number ON public.products USING btree (model_number);
CREATE INDEX IF NOT EXISTS idx_products_type ON public.products USING btree (type);

-- Table Triggers

CREATE OR REPLACE TRIGGER update_products_updated_at BEFORE
UPDATE
    ON
    public.products FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();


-- public.schema_migrations definition

-- Drop table

-- DROP TABLE public.schema_migrations;

CREATE TABLE IF NOT EXISTS public.schema_migrations (
	id serial4 NOT NULL,
	filename varchar(255) NOT NULL,
	executed_at timestamp DEFAULT now() NULL,
	CONSTRAINT schema_migrations_filename_key UNIQUE (filename),
	CONSTRAINT schema_migrations_pkey PRIMARY KEY (id)
);


-- public.users definition

-- Drop table

-- DROP TABLE public.users;

CREATE TABLE IF NOT EXISTS public.users (
	id uuid DEFAULT uuid_generate_v4() NOT NULL,
	username varchar(50) NOT NULL,
	email varchar(255) NOT NULL,
	password_hash varchar(255) NOT NULL,
	"role" varchar(20) NOT NULL,
	is_active bool DEFAULT true NULL,
	last_login_at timestamp NULL,
	created_at timestamp DEFAULT now() NULL,
	updated_at timestamp DEFAULT now() NULL,
	CONSTRAINT chk_users_email_format CHECK (((email)::text ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'::text)),
	CONSTRAINT users_email_key UNIQUE (email),
	CONSTRAINT users_pkey PRIMARY KEY (id),
	CONSTRAINT users_role_check CHECK (((role)::text = ANY ((ARRAY['admin'::character varying, 'editor'::character varying, 'readonly'::character varying])::text[]))),
	CONSTRAINT users_username_key UNIQUE (username)
);
CREATE INDEX IF NOT EXISTS idx_users_email ON public.users USING btree (email);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON public.users USING btree (is_active);
CREATE INDEX IF NOT EXISTS idx_users_role ON public.users USING btree (role);
CREATE INDEX IF NOT EXISTS idx_users_username ON public.users USING btree (username);

-- Table Triggers

CREATE OR REPLACE TRIGGER update_users_updated_at BEFORE
UPDATE
    ON
    public.users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();


-- public.audit_logs definition

-- Drop table

-- DROP TABLE public.audit_logs;

CREATE TABLE IF NOT EXISTS public.audit_logs (
	id uuid DEFAULT uuid_generate_v4() NOT NULL,
	user_id uuid NULL,
	"action" varchar(50) NOT NULL,
	table_name varchar(50) NOT NULL,
	record_id uuid NULL,
	old_values jsonb NULL,
	new_values jsonb NULL,
	ip_address inet NULL,
	user_agent text NULL,
	created_at timestamp DEFAULT now() NULL,
	CONSTRAINT audit_logs_pkey PRIMARY KEY (id),
	CONSTRAINT audit_logs_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE SET NULL
);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON public.audit_logs USING btree (action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON public.audit_logs USING btree (created_at);
CREATE INDEX IF NOT EXISTS idx_audit_logs_record_id ON public.audit_logs USING btree (record_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_table_name ON public.audit_logs USING btree (table_name);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON public.audit_logs USING btree (user_id);


-- public.warranty_registrations definition

-- Drop table

-- DROP TABLE public.warranty_registrations;

CREATE TABLE IF NOT EXISTS public.warranty_registrations (
	id uuid DEFAULT uuid_generate_v4() NOT NULL,
	patient_name varchar(100) NULL,
	patient_id_encrypted text NULL,
	patient_birth_date date NULL,
	patient_phone_encrypted text NULL,
	patient_email varchar(255) NULL,
	hospital_name varchar(255) NULL,
	doctor_name varchar(100) NULL,
	surgery_date date NULL,
	product_id uuid NULL,
	warranty_start_date date NULL,
	warranty_end_date date NULL,
	confirmation_email_sent bool DEFAULT false NULL,
	email_sent_at timestamp NULL,
	status varchar(20) DEFAULT 'active'::character varying NULL,
	created_at timestamp DEFAULT now() NULL,
	updated_at timestamp DEFAULT now() NULL,
	product_serial_number varchar(100) DEFAULT ''::character varying NULL,
	serial_number_2 varchar(100) NULL,
	CONSTRAINT warranty_registrations_pkey PRIMARY KEY (id),
	CONSTRAINT warranty_registrations_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE RESTRICT
);
CREATE INDEX IF NOT EXISTS idx_warranty_registrations_active ON public.warranty_registrations USING btree (patient_name, surgery_date) WHERE ((status)::text = 'active'::text);
CREATE INDEX IF NOT EXISTS idx_warranty_registrations_common_hospitals ON public.warranty_registrations USING btree (hospital_name, created_at) WHERE ((hospital_name)::text = ANY ((ARRAY['台北美麗診所'::character varying, '高雄時尚醫美中心'::character varying, '台中雅緻整形外科'::character varying])::text[]));
CREATE INDEX IF NOT EXISTS idx_warranty_registrations_created_at ON public.warranty_registrations USING btree (created_at);
CREATE INDEX IF NOT EXISTS idx_warranty_registrations_doctor_name ON public.warranty_registrations USING btree (doctor_name);
CREATE INDEX IF NOT EXISTS idx_warranty_registrations_doctor_name_gin ON public.warranty_registrations USING gin (to_tsvector('simple'::regconfig, (doctor_name)::text));
CREATE INDEX IF NOT EXISTS idx_warranty_registrations_filled ON public.warranty_registrations USING btree (status, updated_at) WHERE (((status)::text = ANY ((ARRAY['active'::character varying, 'expired'::character varying, 'cancelled'::character varying])::text[])) AND (created_at <> updated_at));
CREATE INDEX IF NOT EXISTS idx_warranty_registrations_hospital_doctor ON public.warranty_registrations USING btree (hospital_name, doctor_name);
CREATE INDEX IF NOT EXISTS idx_warranty_registrations_hospital_name ON public.warranty_registrations USING btree (hospital_name);
CREATE INDEX IF NOT EXISTS idx_warranty_registrations_hospital_name_gin ON public.warranty_registrations USING gin (to_tsvector('simple'::regconfig, (hospital_name)::text));
CREATE INDEX IF NOT EXISTS idx_warranty_registrations_name_hospital ON public.warranty_registrations USING btree (patient_name, hospital_name);
CREATE INDEX IF NOT EXISTS idx_warranty_registrations_patient_email ON public.warranty_registrations USING btree (patient_email);
CREATE INDEX IF NOT EXISTS idx_warranty_registrations_patient_name ON public.warranty_registrations USING btree (patient_name);
CREATE INDEX IF NOT EXISTS idx_warranty_registrations_patient_name_gin ON public.warranty_registrations USING gin (to_tsvector('simple'::regconfig, (patient_name)::text));
CREATE INDEX IF NOT EXISTS idx_warranty_registrations_pending ON public.warranty_registrations USING btree (status, created_at) WHERE ((status)::text = 'pending'::text);
CREATE INDEX IF NOT EXISTS idx_warranty_registrations_product_serial ON public.warranty_registrations USING btree (product_serial_number);
CREATE INDEX IF NOT EXISTS idx_warranty_registrations_serial_number_2 ON public.warranty_registrations USING btree (serial_number_2) WHERE (serial_number_2 IS NOT NULL);
CREATE INDEX IF NOT EXISTS idx_warranty_registrations_serial_number_not_null ON public.warranty_registrations USING btree (product_serial_number) WHERE (product_serial_number IS NOT NULL);
CREATE INDEX IF NOT EXISTS idx_warranty_registrations_status ON public.warranty_registrations USING btree (status);
CREATE INDEX IF NOT EXISTS idx_warranty_registrations_status_warranty_end ON public.warranty_registrations USING btree (status, warranty_end_date);
CREATE INDEX IF NOT EXISTS idx_warranty_registrations_surgery_date ON public.warranty_registrations USING btree (surgery_date);
CREATE INDEX IF NOT EXISTS idx_warranty_registrations_surgery_date_desc ON public.warranty_registrations USING btree (surgery_date DESC);
CREATE INDEX IF NOT EXISTS idx_warranty_registrations_warranty_end_date ON public.warranty_registrations USING btree (warranty_end_date);

-- Table Triggers

CREATE OR REPLACE TRIGGER update_warranty_registrations_updated_at BEFORE
UPDATE
    ON
    public.warranty_registrations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();


-- public.email_logs definition

-- Drop table

-- DROP TABLE public.email_logs;

CREATE TABLE IF NOT EXISTS public.email_logs (
	id uuid DEFAULT uuid_generate_v4() NOT NULL,
	warranty_registration_id uuid NULL,
	recipient_email varchar(255) NOT NULL,
	subject varchar(255) NOT NULL,
	body text NOT NULL,
	status varchar(20) NOT NULL,
	mailgun_id varchar(255) NULL,
	error_message text NULL,
	sent_at timestamp NULL,
	created_at timestamp DEFAULT now() NULL,
	CONSTRAINT email_logs_pkey PRIMARY KEY (id),
	CONSTRAINT email_logs_status_check CHECK (((status)::text = ANY ((ARRAY['sent'::character varying, 'failed'::character varying, 'pending'::character varying])::text[]))),
	CONSTRAINT email_logs_warranty_registration_id_fkey FOREIGN KEY (warranty_registration_id) REFERENCES public.warranty_registrations(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_email_logs_created_at ON public.email_logs USING btree (created_at);
CREATE INDEX IF NOT EXISTS idx_email_logs_recipient_email ON public.email_logs USING btree (recipient_email);
CREATE INDEX IF NOT EXISTS idx_email_logs_status ON public.email_logs USING btree (status);
CREATE INDEX IF NOT EXISTS idx_email_logs_warranty_registration_id ON public.email_logs USING btree (warranty_registration_id);