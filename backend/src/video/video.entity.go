package video

import (
	"project-one/src/user"
	"time"
)

type Video struct {
	ID           uint       `json:"id" gorm:"primaryKey"`
	UserID       uint       `json:"user_id" gorm:"not null"`
	Title        string     `json:"title" gorm:"not null"`
	Status       string     `json:"status" gorm:"default:'uploaded'"`
	OriginalURL  string     `json:"original_url"`
	ProcessedURL string     `json:"processed_url,omitempty"`
	VoteCount    int        `json:"votes" gorm:"default:0"`
	UploadedAt   time.Time  `json:"uploaded_at"`
	ProcessedAt  *time.Time `json:"processed_at,omitempty"`
	IsPublic     bool       `json:"is_public" gorm:"default:false"`

	// Solo relación con User
	User user.User `json:"user" gorm:"foreignKey:UserID"`
	// NO incluyas []vote.Vote aquí para evitar dependencia circular
}
