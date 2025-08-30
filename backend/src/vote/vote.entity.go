package vote

import (
	"project-one/src/user"
	"time"
)

type Vote struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null"`
	VideoID   uint      `json:"video_id" gorm:"primaryKey"`
	VotedAt   time.Time `json:"voted_at"`
	CreatedAt time.Time `json:"created_at"`

	User user.User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}
