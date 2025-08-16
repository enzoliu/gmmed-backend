CREATE TABLE IF NOT EXISTS public.serials (
	id uuid DEFAULT uuid_generate_v4() NOT NULL,
	serial_number varchar(20) NOT NULL,
	full_serial_number varchar(100) NOT NULL,
	product_id uuid NULL,
	created_at timestamp DEFAULT now() NULL,
	updated_at timestamp DEFAULT now() NULL,
	CONSTRAINT serials_pkey PRIMARY KEY (id)
);
CREATE INDEX idx_serials_serial_number ON public.serials USING btree (serial_number);
CREATE INDEX idx_serials_full_serial_number ON public.serials USING btree (full_serial_number);
CREATE INDEX idx_serials_product_id ON public.serials USING btree (product_id);