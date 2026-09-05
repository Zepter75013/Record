package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"records-manager/backend/config"
	artisthandler "records-manager/backend/internal/artists/handler"
	artistrepo "records-manager/backend/internal/artists/repository"
	artistservice "records-manager/backend/internal/artists/service"
	authhandler "records-manager/backend/internal/auth/handler"
	authservice "records-manager/backend/internal/auth/service"
	"records-manager/backend/internal/backup"
	countryhandler "records-manager/backend/internal/countries/handler"
	countryrepo "records-manager/backend/internal/countries/repository"
	countryservice "records-manager/backend/internal/countries/service"
	dischandler "records-manager/backend/internal/discs/handler"
	discrepo "records-manager/backend/internal/discs/repository"
	discservice "records-manager/backend/internal/discs/service"
	"records-manager/backend/internal/email" // 🆕 AJOUTÉ
	formathandler "records-manager/backend/internal/formats/handler"
	formatrepo "records-manager/backend/internal/formats/repository"
	formatservice "records-manager/backend/internal/formats/service"
	gamegenrehandler "records-manager/backend/internal/gamegenres/handler"
	gamegenrerepo "records-manager/backend/internal/gamegenres/repository"
	gamegenreservice "records-manager/backend/internal/gamegenres/service"
	gamehandler "records-manager/backend/internal/games/handler"
	gamerepo "records-manager/backend/internal/games/repository"
	gameservice "records-manager/backend/internal/games/service"
	genrehandler "records-manager/backend/internal/genres/handler"
	genrerepo "records-manager/backend/internal/genres/repository"
	genreservice "records-manager/backend/internal/genres/service"
	labelhandler "records-manager/backend/internal/labels/handler"
	labelrepo "records-manager/backend/internal/labels/repository"
	labelservice "records-manager/backend/internal/labels/service"
	platformhandler "records-manager/backend/internal/platforms/handler"
	platformrepo "records-manager/backend/internal/platforms/repository"
	platformservice "records-manager/backend/internal/platforms/service"
	publisherhandler "records-manager/backend/internal/publishers/handler"
	publisherrepo "records-manager/backend/internal/publishers/repository"
	publisherservice "records-manager/backend/internal/publishers/service"
	"records-manager/backend/internal/report"
	statshandler "records-manager/backend/internal/stats/handler"
	statsrepo "records-manager/backend/internal/stats/repository"
	statsservice "records-manager/backend/internal/stats/service"
	tracksrepo "records-manager/backend/internal/tracks/repository"
	userhandler "records-manager/backend/internal/users/handler"
	userrepo "records-manager/backend/internal/users/repository"
	userservice "records-manager/backend/internal/users/service"
)

// loadEnvFromFile charge manuellement les variables depuis le fichier .env
func loadEnvFromFile() error {
	content, err := os.ReadFile(".env")
	if err != nil {
		return fmt.Errorf("fichier .env non trouvé: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, `"'`)

		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}

	return nil
}

// loadEnvDefaults définit les variables d'environnement par défaut
func loadEnvDefaults() {
	if os.Getenv("HTTP_PORT") == "" {
		os.Setenv("HTTP_PORT", "8080")
	}

	if os.Getenv("JWT_SECRET") == "" {
		os.Setenv("JWT_SECRET", "default_jwt_secret_change_this_in_production")
		log.Println("⚠️ JWT_SECRET non défini, utilisation d'une valeur par défaut")
	}

	if os.Getenv("DB_HOST") == "" {
		os.Setenv("DB_HOST", "localhost")
	}

	if os.Getenv("DB_PORT") == "" {
		os.Setenv("DB_PORT", "3306")
	}

	if os.Getenv("DB_USER") == "" {
		os.Setenv("DB_USER", "root")
	}

	if os.Getenv("DB_PASSWORD") == "" {
		os.Setenv("DB_PASSWORD", "")
	}

	if os.Getenv("DB_NAME") == "" {
		os.Setenv("DB_NAME", "discdb")
	}

	if os.Getenv("DISCOGS_TOKEN") == "" {
		log.Println("⚠️ DISCOGS_TOKEN non défini - téléchargement des pochettes désactivé")
	}

	if os.Getenv("UPLOADS_DIR") == "" {
		os.Setenv("UPLOADS_DIR", "./uploads")
	}

	// 🆕 Défaults SMTP
	if os.Getenv("SMTP_HOST") == "" {
		os.Setenv("SMTP_HOST", "smtp.gmail.com")
	}
	if os.Getenv("SMTP_PORT") == "" {
		os.Setenv("SMTP_PORT", "587")
	}
	if os.Getenv("SMTP_FROM_NAME") == "" {
		os.Setenv("SMTP_FROM_NAME", "Vinyl Manager")
	}
	if os.Getenv("FRONTEND_URL") == "" {
		os.Setenv("FRONTEND_URL", "http://localhost:5173")
	}
}

// ResponseWriter personnalisé pour capturer le status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// loggingMiddleware logue toutes les requêtes HTTP
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("➡️  %s %s", r.Method, r.URL.Path)

		if r.Method == "POST" || r.Method == "PUT" {
			bodyBytes, _ := io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			if len(bodyBytes) > 0 && len(bodyBytes) < 1000 {
				log.Printf("  📦 Body: %s", string(bodyBytes))
			}
		}

		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		start := time.Now()
		next.ServeHTTP(rw, r)
		duration := time.Since(start)

		log.Printf("⬅️  %s %s - Status: %d - Durée: %v",
			r.Method, r.URL.Path, rw.statusCode, duration)
	})
}

// authMiddlewareWithService vérifie le token JWT
func authMiddlewareWithService(authService authservice.AuthService) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// ⭐ Routes publiques (sans auth JWT)
			if r.URL.Path == "/api/login" ||
				r.URL.Path == "/api/health" ||
				r.URL.Path == "/api/password-reset/request" || // 🆕 AJOUTÉ
				r.URL.Path == "/api/password-reset/reset" || // 🆕 AJOUTÉ
				r.URL.Path == "/api/search-discogs" ||
				r.URL.Path == "/api/select-discogs-result" ||
				r.URL.Path == "/api/upload-cover" ||
				r.URL.Path == "/api/search-rawg" ||
				r.URL.Path == "/api/select-rawg-result" ||
				strings.HasPrefix(r.URL.Path, "/api/check-barcode/") ||
				strings.HasPrefix(r.URL.Path, "/uploads/") {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "Token manquant"})
				return
			}

			if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "Format de token invalide"})
				return
			}

			tokenString := authHeader[7:]

			// 🆕 Validation ET extraction des claims
			claims, err := authService.ValidateTokenAndGetClaims(tokenString)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "Token invalide"})
				return
			}

			// 🆕 Injecter les claims dans le contexte
			ctx := context.WithValue(r.Context(), "user", claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func main() {
	log.Println("🚀 Démarrage de Records Manager API...")

	// 1. Chargement de la configuration
	log.Println("🔍 Chargement du fichier .env...")
	if err := loadEnvFromFile(); err != nil {
		log.Printf("⚠️  %v", err)
		log.Println("   Utilisation des variables d'environnement système")
	} else {
		log.Println("✅ Fichier .env chargé")
	}

	loadEnvDefaults()

	// 2. Affichage de la configuration
	log.Println("🔧 Configuration:")
	log.Printf("  📌 DB_HOST: %s", os.Getenv("DB_HOST"))
	log.Printf("  📌 DB_PORT: %s", os.Getenv("DB_PORT"))
	log.Printf("  📌 DB_USER: %s", os.Getenv("DB_USER"))
	log.Printf("  📌 DB_NAME: %s", os.Getenv("DB_NAME"))
	log.Printf("  📌 HTTP_PORT: %s", os.Getenv("HTTP_PORT"))

	discogsToken := os.Getenv("DISCOGS_TOKEN")
	if discogsToken != "" {
		log.Printf("  📌 DISCOGS_TOKEN: %s...%s (✅ configuré)",
			discogsToken[:min(5, len(discogsToken))],
			discogsToken[max(0, len(discogsToken)-5):])
	} else {
		log.Println("  ⚠️  DISCOGS_TOKEN: non configuré")
	}

	rawgAPIKey := os.Getenv("RAWG_API_KEY")
	if rawgAPIKey != "" {
		log.Printf("  📌 RAWG_API_KEY: %s...%s (✅ configuré)",
			rawgAPIKey[:min(5, len(rawgAPIKey))],
			rawgAPIKey[max(0, len(rawgAPIKey)-5):])
	} else {
		log.Println("  ⚠️  RAWG_API_KEY: non configuré")
	}

	deeplAPIKey := os.Getenv("DEEPL_API_KEY")
	if deeplAPIKey != "" {
		log.Printf("  📌 DEEPL_API_KEY: %s...%s (✅ configuré)",
			deeplAPIKey[:min(5, len(deeplAPIKey))],
			deeplAPIKey[max(0, len(deeplAPIKey)-5):])
	} else {
		log.Println("  ⚠️  DEEPL_API_KEY: non configuré — descriptions/biographies suggérées en anglais")
	}

	uploadsDir := os.Getenv("UPLOADS_DIR")
	log.Printf("  📌 UPLOADS_DIR: %s", uploadsDir)

	// 🆕 Affichage config SMTP
	smtpUser := os.Getenv("SMTP_USER")
	if smtpUser != "" {
		log.Printf("  📧 SMTP_HOST: %s", os.Getenv("SMTP_HOST"))
		log.Printf("  📧 SMTP_USER: %s", smtpUser)
		log.Printf("  📧 SMTP_FROM: %s", os.Getenv("SMTP_FROM"))
		log.Println("  ✅ Email configuré")
	} else {
		log.Println("  ⚠️  SMTP non configuré - réinitialisation mot de passe désactivée")
	}

	// 3. Création du dossier uploads
	if err := os.MkdirAll(filepath.Join(uploadsDir, "covers"), 0755); err != nil {
		log.Fatalf("❌ Erreur création dossier uploads: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(uploadsDir, "avatars"), 0755); err != nil {
		log.Fatalf("❌ Erreur création dossier uploads: %v", err)
	}
	log.Printf("✅ Dossier uploads créé: %s", uploadsDir)

	// 4. Configuration base de données
	dbConfig := config.LoadDBConfig()
	if dbConfig.Host == "" || dbConfig.User == "" || dbConfig.Name == "" {
		log.Fatal("❌ Configuration DB incomplète")
	}

	// 5. Connexion à la base de données
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Name)

	log.Println("🔌 Connexion à MySQL...")
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("❌ Erreur ouverture MySQL: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("❌ MySQL inaccessible: %v", err)
	}

	log.Println("✅ Connecté à MySQL")
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// 6. Initialisation des repositories
	userRepository := userrepo.NewMySQLUserRepository(db)
	resetTokenRepo := userrepo.NewPasswordResetRepository(db) // 🆕 AJOUTÉ
	formatRepository := formatrepo.NewMySQLFormatRepository(db)
	genreRepository := genrerepo.NewMySQLGenreRepository(db)
	artistRepository := artistrepo.NewMySQLArtistRepository(db)
	countryRepository := countryrepo.NewMySQLCountryRepository(db)
	labelRepository := labelrepo.NewMySQLLabelRepository(db)
	statsRepository := statsrepo.NewMySQLStatsRepository(db)
	discRepository := discrepo.NewMySQLDiscRepository(db)
	trackRepository := tracksrepo.NewMySQLTrackRepository(db)
	platformRepository := platformrepo.NewMySQLPlatformRepository(db)
	gameGenreRepository := gamegenrerepo.NewMySQLGameGenreRepository(db)
	publisherRepository := publisherrepo.NewMySQLPublisherRepository(db)
	gameRepository := gamerepo.NewMySQLGameRepository(db)

	// 7. Initialisation des services
	authService := authservice.NewAuthService(userRepository)
	userService := userservice.NewUserService(userRepository, uploadsDir)
	emailService := email.NewEmailService()                                                                   // 🆕 AJOUTÉ
	passwordResetService := userservice.NewPasswordResetService(userRepository, resetTokenRepo, emailService) // 🆕 AJOUTÉ
	formatService := formatservice.NewFormatService(formatRepository)
	genreService := genreservice.NewGenreService(genreRepository)
	artistService := artistservice.NewArtistService(artistRepository, countryRepository, discogsToken, deeplAPIKey)
	countryService := countryservice.NewCountryService(countryRepository)
	labelService := labelservice.NewLabelService(labelRepository, discogsToken, deeplAPIKey)
	statsService := statsservice.NewStatsService(statsRepository)
	discService := discservice.NewDiscService(discRepository, discogsToken, uploadsDir, trackRepository)
	platformService := platformservice.NewPlatformService(platformRepository)
	gameGenreService := gamegenreservice.NewGameGenreService(gameGenreRepository)
	publisherService := publisherservice.NewPublisherService(publisherRepository)
	gameService := gameservice.NewGameService(gameRepository, rawgAPIKey, uploadsDir)

	backupService, err := backup.NewService(dbConfig, "./backups")
	if err != nil {
		log.Fatalf("❌ Erreur initialisation service de sauvegarde: %v", err)
	}

	reportService, err := report.NewService("./reports")
	if err != nil {
		log.Fatalf("❌ Erreur initialisation service de rapports: %v", err)
	}

	// 8. Création utilisateur test
	testEmail := "test@vinylmanager.com"
	testPassword := "password123"

	if user, _ := userRepository.FindByEmail(testEmail); user == nil {
		if err := authService.CreateTestUser(context.Background(), testEmail, testPassword); err != nil {
			log.Fatalf("❌ Erreur création user test: %v", err)
		}
		log.Printf("👤 User test créé: %s / %s", testEmail, testPassword)
	} else {
		log.Printf("👤 User test existe: %s", testEmail)
	}

	// 9. Initialisation des handlers
	authHandler := authhandler.NewAuthHandler(authService)
	userHandler := userhandler.NewUserHandler(userService, uploadsDir)
	passwordResetHandler := userhandler.NewPasswordResetHandler(passwordResetService) // 🆕 AJOUTÉ
	formatHandler := formathandler.NewFormatHandler(formatService)
	genreHandler := genrehandler.NewGenreHandler(genreService)
	artistHandler := artisthandler.NewArtistHandler(artistService)
	countryHandler := countryhandler.NewCountryHandler(countryService)
	labelHandler := labelhandler.NewLabelHandler(labelService)
	statsHandler := statshandler.NewStatsHandler(statsService)
	discHandler := dischandler.NewDiscHandler(discService)
	platformHandler := platformhandler.NewPlatformHandler(platformService)
	gameGenreHandler := gamegenrehandler.NewGameGenreHandler(gameGenreService)
	publisherHandler := publisherhandler.NewPublisherHandler(publisherService)
	gameHandler := gamehandler.NewGameHandler(gameService)
	backupHandler := backup.NewHandler(backupService)
	reportHandler := report.NewHandler(reportService)

	// 10. Configuration du routeur
	r := mux.NewRouter()

	// ✅ Serveur de fichiers statiques (avant middlewares)
	r.PathPrefix("/uploads/").Handler(
		http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadsDir))),
	).Methods("GET")

	// Middlewares
	authMiddleware := authMiddlewareWithService(authService)
	r.Use(loggingMiddleware)
	r.Use(authMiddleware)

	// Routes publiques
	r.HandleFunc("/api/health", authhandler.HealthCheckHandler).Methods("GET")
	r.HandleFunc("/api/login", authHandler.Login).Methods("POST")

	// 🆕 Routes réinitialisation mot de passe (publiques)
	r.HandleFunc("/api/password-reset/request", passwordResetHandler.RequestPasswordReset).Methods("POST")
	r.HandleFunc("/api/password-reset/reset", passwordResetHandler.ResetPassword).Methods("POST")

	// Route changement de mot de passe (protégée)
	r.HandleFunc("/api/user/change-password", userHandler.ChangePassword).Methods("POST")

	// Photo de profil (protégées)
	r.HandleFunc("/api/user/avatar", userHandler.UploadAvatar).Methods("POST")
	r.HandleFunc("/api/user/avatar", userHandler.RemoveAvatar).Methods("DELETE")

	// Sauvegardes (protégées — pas dans la liste blanche du middleware auth)
	r.HandleFunc("/api/backups", backupHandler.List).Methods("GET")
	r.HandleFunc("/api/backups", backupHandler.Create).Methods("POST")
	r.HandleFunc("/api/backups/restore-upload", backupHandler.RestoreUpload).Methods("POST")
	r.HandleFunc("/api/backups/settings", backupHandler.GetSettings).Methods("GET")
	r.HandleFunc("/api/backups/settings", backupHandler.UpdateSettings).Methods("PUT")
	r.HandleFunc("/api/backups/settings/pick-directory", backupHandler.PickDirectory).Methods("POST")
	r.HandleFunc("/api/backups/{name}/download", backupHandler.Download).Methods("GET")
	r.HandleFunc("/api/backups/{name}/restore", backupHandler.Restore).Methods("POST")
	r.HandleFunc("/api/backups/{name}", backupHandler.Delete).Methods("DELETE")

	// Rapports (export PDF/XLSX/DOCX/CSV de la collection, un seul "dernier rapport" conservé)
	r.HandleFunc("/api/reports", reportHandler.Upload).Methods("POST")
	r.HandleFunc("/api/reports/latest", reportHandler.GetLatestMetadata).Methods("GET")
	r.HandleFunc("/api/reports/latest/file", reportHandler.DownloadLatestFile).Methods("GET")

	// Streaming (résolution d'album Deezer, proxy pour contourner le CORS de l'API Deezer)
	r.HandleFunc("/api/streaming/deezer-search", handleDeezerSearch).Methods("GET")

	// Stats
	r.HandleFunc("/api/stats", statsHandler.GetStats).Methods("GET")

	// Formats
	r.HandleFunc("/api/formats", formatHandler.GetAllFormats).Methods("GET")
	r.HandleFunc("/api/formats", formatHandler.CreateFormat).Methods("POST")
	r.HandleFunc("/api/formats/create-if-not-exists", formatHandler.CreateFormatIfNotExists).Methods("POST")
	r.HandleFunc("/api/formats/{id}", formatHandler.GetFormatByID).Methods("GET")
	r.HandleFunc("/api/formats/{id}", formatHandler.UpdateFormat).Methods("PUT")
	r.HandleFunc("/api/formats/{id}", formatHandler.DeleteFormat).Methods("DELETE")

	// Genres
	r.HandleFunc("/api/genres", genreHandler.GetAllGenres).Methods("GET")
	r.HandleFunc("/api/genres", genreHandler.CreateGenre).Methods("POST")
	r.HandleFunc("/api/genres/create-if-not-exists", genreHandler.CreateGenreIfNotExists).Methods("POST")
	r.HandleFunc("/api/genres/{id}", genreHandler.GetGenreByID).Methods("GET")
	r.HandleFunc("/api/genres/{id}", genreHandler.UpdateGenre).Methods("PUT")
	r.HandleFunc("/api/genres/{id}", genreHandler.DeleteGenre).Methods("DELETE")

	// Artists
	r.HandleFunc("/api/artists", artistHandler.GetAllArtists).Methods("GET")
	r.HandleFunc("/api/artists", artistHandler.CreateArtist).Methods("POST")
	r.HandleFunc("/api/artists/{id}", artistHandler.GetArtistByID).Methods("GET")
	r.HandleFunc("/api/artists/{id}", artistHandler.UpdateArtist).Methods("PUT")
	r.HandleFunc("/api/artists/{id}", artistHandler.DeleteArtist).Methods("DELETE")
	r.HandleFunc("/api/artists/{id}/country/suggest", artistHandler.SuggestArtistCountry).Methods("POST")
	r.HandleFunc("/api/artists/{id}/biography/suggest", artistHandler.SuggestArtistBiography).Methods("POST")

	// Countries
	r.HandleFunc("/api/countries", countryHandler.GetAllCountries).Methods("GET")
	r.HandleFunc("/api/countries", countryHandler.CreateCountry).Methods("POST")
	r.HandleFunc("/api/countries/{id}", countryHandler.GetCountryByID).Methods("GET")
	r.HandleFunc("/api/countries/{id}", countryHandler.UpdateCountry).Methods("PUT")
	r.HandleFunc("/api/countries/{id}", countryHandler.DeleteCountry).Methods("DELETE")

	// Labels
	r.HandleFunc("/api/labels", labelHandler.GetAllLabels).Methods("GET")
	r.HandleFunc("/api/labels", labelHandler.CreateLabel).Methods("POST")
	r.HandleFunc("/api/labels/{id}", labelHandler.GetLabelByID).Methods("GET")
	r.HandleFunc("/api/labels/{id}", labelHandler.UpdateLabel).Methods("PUT")
	r.HandleFunc("/api/labels/{id}", labelHandler.DeleteLabel).Methods("DELETE")
	r.HandleFunc("/api/labels/{id}/description/suggest", labelHandler.SuggestLabelDescription).Methods("POST")

	// Platforms (jeux vidéo)
	r.HandleFunc("/api/platforms", platformHandler.GetAllPlatforms).Methods("GET")
	r.HandleFunc("/api/platforms", platformHandler.CreatePlatform).Methods("POST")
	r.HandleFunc("/api/platforms/create-if-not-exists", platformHandler.CreatePlatformIfNotExists).Methods("POST")
	r.HandleFunc("/api/platforms/{id}", platformHandler.GetPlatformByID).Methods("GET")
	r.HandleFunc("/api/platforms/{id}", platformHandler.UpdatePlatform).Methods("PUT")
	r.HandleFunc("/api/platforms/{id}", platformHandler.DeletePlatform).Methods("DELETE")

	// Game genres (jeux vidéo — vocabulaire distinct des genres musicaux)
	r.HandleFunc("/api/game-genres", gameGenreHandler.GetAllGameGenres).Methods("GET")
	r.HandleFunc("/api/game-genres", gameGenreHandler.CreateGameGenre).Methods("POST")
	r.HandleFunc("/api/game-genres/create-if-not-exists", gameGenreHandler.CreateGameGenreIfNotExists).Methods("POST")
	r.HandleFunc("/api/game-genres/{id}", gameGenreHandler.GetGameGenreByID).Methods("GET")
	r.HandleFunc("/api/game-genres/{id}", gameGenreHandler.UpdateGameGenre).Methods("PUT")
	r.HandleFunc("/api/game-genres/{id}", gameGenreHandler.DeleteGameGenre).Methods("DELETE")

	// Publishers (jeux vidéo)
	r.HandleFunc("/api/publishers", publisherHandler.GetAllPublishers).Methods("GET")
	r.HandleFunc("/api/publishers", publisherHandler.CreatePublisher).Methods("POST")
	r.HandleFunc("/api/publishers/create-if-not-exists", publisherHandler.CreatePublisherIfNotExists).Methods("POST")
	r.HandleFunc("/api/publishers/{id}", publisherHandler.GetPublisherByID).Methods("GET")
	r.HandleFunc("/api/publishers/{id}", publisherHandler.UpdatePublisher).Methods("PUT")
	r.HandleFunc("/api/publishers/{id}", publisherHandler.DeletePublisher).Methods("DELETE")

	// ⭐ Disques - Routes principales
	r.HandleFunc("/api/discs", discHandler.GetAllDiscs).Methods("GET")
	r.HandleFunc("/api/discs", discHandler.CreateDisc).Methods("POST")

	// Routes spécifiques (sans auth → déjà dans middleware whitelist)
	r.HandleFunc("/api/discs/preview-cover", discHandler.PreviewCover).Methods("POST")
	r.HandleFunc("/api/discs/check-barcode", discHandler.CheckBarcodeExists).Methods("POST")

	// ✅ Routes compatibles frontend (sans /discs/ préfixe, déjà publiques via middleware)
	r.HandleFunc("/api/check-barcode/{barcode}", discHandler.CheckBarcodeExistsGET).Methods("GET")
	r.HandleFunc("/api/search-discogs", discHandler.SearchDiscogs).Methods("POST")
	r.HandleFunc("/api/select-discogs-result", discHandler.SelectDiscogsResult).Methods("POST")
	r.HandleFunc("/api/upload-cover", discHandler.UploadCover).Methods("POST")

	// Routes authentifiées (CRUD disques)
	r.HandleFunc("/api/discs/{id}", discHandler.GetDiscByID).Methods("GET")
	r.HandleFunc("/api/discs/{id}", discHandler.UpdateDisc).Methods("PUT")
	r.HandleFunc("/api/discs/{id}", discHandler.DeleteDisc).Methods("DELETE")
	r.HandleFunc("/api/discs/{id}/download-cover", discHandler.DownloadCover).Methods("POST")
	r.HandleFunc("/api/discs/{id}/tracks", discHandler.GetTracks).Methods("GET")
	r.HandleFunc("/api/discs/{id}/tracks", discHandler.UpdateTracks).Methods("PUT")
	r.HandleFunc("/api/discs/{id}/tracks/fetch", discHandler.FetchTracks).Methods("POST")
	r.HandleFunc("/api/discs/{id}/tracks/preview", discHandler.PreviewTracks).Methods("GET")

	// ⭐ Jeux vidéo - Routes principales
	r.HandleFunc("/api/games", gameHandler.GetAllGames).Methods("GET")
	r.HandleFunc("/api/games", gameHandler.CreateGame).Methods("POST")
	r.HandleFunc("/api/games/check-barcode", gameHandler.CheckBarcodeExists).Methods("POST")

	// Recherche RAWG (publiques, même logique que search-discogs/select-discogs-result)
	r.HandleFunc("/api/search-rawg", gameHandler.SearchRAWG).Methods("POST")
	r.HandleFunc("/api/select-rawg-result", gameHandler.SelectRAWGResult).Methods("POST")

	// Routes authentifiées (CRUD jeux)
	r.HandleFunc("/api/games/{id}", gameHandler.GetGameByID).Methods("GET")
	r.HandleFunc("/api/games/{id}", gameHandler.UpdateGame).Methods("PUT")
	r.HandleFunc("/api/games/{id}", gameHandler.DeleteGame).Methods("DELETE")

	// CORS preflight
	r.Methods("OPTIONS").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.WriteHeader(http.StatusOK)
	})

	// 11. Configuration CORS
	allowedOrigins := handlers.AllowedOrigins([]string{
		"http://localhost:5173",
		"http://localhost:5174",
		"http://localhost:3000",
	})
	allowedHeaders := handlers.AllowedHeaders([]string{
		"X-Requested-With", "Content-Type", "Authorization", "Accept",
	})
	allowedMethods := handlers.AllowedMethods([]string{
		"GET", "POST", "PUT", "DELETE", "OPTIONS",
	})

	// 12. Démarrage du serveur
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("🌈 Serveur démarré!")
	log.Printf("🌐 URL: http://localhost:%s", port)
	log.Println("📊 Endpoints (extraits):")
	log.Printf("   GET  /api/health")
	log.Printf("   POST /api/login")
	log.Printf("   POST /api/password-reset/request ✅ PUBLIC 🆕") // 🆕 AJOUTÉ
	log.Printf("   POST /api/password-reset/reset ✅ PUBLIC 🆕")   // 🆕 AJOUTÉ
	log.Printf("   POST /api/user/change-password 🔐 Auth")
	log.Printf("   POST /api/user/avatar 🔐 Auth")
	log.Printf("   DELETE /api/user/avatar 🔐 Auth")
	log.Printf("   GET/POST/PUT/DELETE /api/backups/* 🔐 Auth")
	log.Printf("   GET  /api/stats")
	log.Printf("   POST /api/search-discogs ✅ PUBLIC")
	log.Printf("   POST /api/select-discogs-result ✅ PUBLIC")
	log.Printf("   POST /api/upload-cover ✅ PUBLIC")
	log.Printf("   GET  /api/check-barcode/{bc} ✅ PUBLIC")
	log.Printf("   POST /api/discs/check-barcode 🔐 Auth")
	log.Printf("   GET/POST/PUT/DELETE /api/discs/* 🔐 Auth")
	log.Printf("   POST /api/search-rawg ✅ PUBLIC")
	log.Printf("   POST /api/select-rawg-result ✅ PUBLIC")
	log.Printf("   GET/POST/PUT/DELETE /api/games/* 🔐 Auth")
	log.Println("🔒 CORS: http://localhost:5173")
	log.Println("🔐 Auth JWT activée")

	handler := handlers.CORS(allowedHeaders, allowedOrigins, allowedMethods)(r)

	log.Printf("✅ Écoute sur le port %s...", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("❌ Erreur serveur: %v", err)
	}
}

// handleDeezerSearch proxifie une recherche d'album Deezer (l'API Deezer ne renvoie pas
// d'en-têtes CORS, donc l'appel doit passer par le backend plutôt que par le navigateur).
func handleDeezerSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		http.Error(w, "paramètre q manquant", http.StatusBadRequest)
		return
	}

	deezerURL := "https://api.deezer.com/search/album?limit=1&q=" + url.QueryEscape(q)
	resp, err := http.Get(deezerURL)
	if err != nil {
		http.Error(w, "recherche Deezer indisponible", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			ID   int    `json:"id"`
			Link string `json:"link"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		http.Error(w, "réponse Deezer invalide", http.StatusBadGateway)
		return
	}
	if len(result.Data) == 0 {
		http.Error(w, "aucun album trouvé", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":   result.Data[0].ID,
		"link": result.Data[0].Link,
	})
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
