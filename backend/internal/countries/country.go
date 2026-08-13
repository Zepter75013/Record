package countries

import (
	"time"
)

type Country struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`        // Code ISO (ex: "FR", "US")
	Description string    `json:"description"` // Description optionnelle
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
