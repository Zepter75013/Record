package domain

// User est le modèle représentant un utilisateur dans la base de données
type User struct {
	ID           uint
	Email        string
	PasswordHash string
	AvatarPath   *string
}
