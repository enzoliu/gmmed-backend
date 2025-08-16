

ALTER TABLE serials
ADD CONSTRAINT serials_serial_number_unique UNIQUE (serial_number);

ALTER TABLE serials
ADD CONSTRAINT serials_full_serial_number_unique UNIQUE (full_serial_number);