ALTER TABLE records_labels
  ADD COLUMN country_id INT NULL,
  ADD CONSTRAINT fk_records_labels_country FOREIGN KEY (country_id) REFERENCES records_countries(id);
