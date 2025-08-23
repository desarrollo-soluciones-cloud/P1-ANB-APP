package vote

import (
	"project-one/src/user"
	"time"
	// NO importes video para evitar dependencia circular
)

type Vote struct {
	ID      uint      `json:"id" gorm:"primaryKey"`
	UserID  uint      `json:"user_id" gorm:"not null"`
	VideoID uint      `json:"video_id" gorm:"not null"` // Solo el ID, no el objeto
	VotedAt time.Time `json:"voted_at"`

	// Solo relación con User
	User user.User `json:"user,omitempty" gorm:"foreignKey:UserID"`
	// NO incluyas video.Video aquí
}
