-- Préfixe les tables par domaine : records_ (disques), games_ (jeux vidéo), common_ (partagé).
-- RENAME TABLE conserve les données, index, clés étrangères et AUTO_INCREMENT.
RENAME TABLE
  vinyls TO records_vinyls,
  tracks TO records_tracks,
  artists TO records_artists,
  countries TO records_countries,
  formats TO records_formats,
  genres TO records_genres,
  labels TO records_labels,
  games TO games_games,
  platforms TO games_platforms,
  game_genres TO games_genres,
  publishers TO games_publishers,
  users TO common_users,
  password_reset_tokens TO common_password_reset_tokens;
