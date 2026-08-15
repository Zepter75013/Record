-- Ajoute le support de la tracklist (liste des pistes) par disque,
-- récupérée automatiquement depuis Discogs à l'ajout du disque.
--
-- Pas de système de migration automatisé dans ce projet — ce fichier sert
-- de référence, à exécuter à la main contre la base MySQL cible.

ALTER TABLE vinyls ADD COLUMN discogs_release_id BIGINT NULL AFTER isrc;

CREATE TABLE tracks (
  id INT AUTO_INCREMENT PRIMARY KEY,
  vinyl_id INT NOT NULL,
  track_order INT NOT NULL,
  position VARCHAR(10),
  title VARCHAR(255) NOT NULL,
  duration VARCHAR(10),
  FOREIGN KEY (vinyl_id) REFERENCES vinyls(id) ON DELETE CASCADE
);
