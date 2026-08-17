CREATE TABLE platforms (
  id INT AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(100) NOT NULL UNIQUE,
  description TEXT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- description explicite (chaîne vide, pas NULL) : le code Go scanne cette
-- colonne dans un `string` non-nullable (voir platforms.go), une valeur NULL
-- ferait échouer le SELECT.
INSERT INTO platforms (name, description) VALUES
  ('PS5', ''), ('Xbox Series X|S', ''), ('Nintendo Switch', ''), ('Nintendo Switch 2', '');

CREATE TABLE game_genres (
  id INT AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(100) NOT NULL UNIQUE,
  description TEXT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE publishers (
  id INT AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(255) NOT NULL UNIQUE,
  description TEXT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE games (
  id INT AUTO_INCREMENT PRIMARY KEY,
  title VARCHAR(255) NOT NULL,
  platform_id INT NOT NULL,
  genre_id INT NULL,
  publisher_id INT NULL,
  release_year INT NULL,
  barcode VARCHAR(50) NULL,
  cover_url VARCHAR(500) NULL,
  notes TEXT NULL,
  price DECIMAL(10,2) NULL,
  quantity INT DEFAULT 1,
  rawg_id BIGINT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (platform_id) REFERENCES platforms(id),
  FOREIGN KEY (genre_id) REFERENCES game_genres(id),
  FOREIGN KEY (publisher_id) REFERENCES publishers(id)
);
