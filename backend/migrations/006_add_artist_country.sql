ALTER TABLE records_artists
  ADD COLUMN country_id INT NULL,
  ADD CONSTRAINT fk_records_artists_country FOREIGN KEY (country_id) REFERENCES records_countries(id);
