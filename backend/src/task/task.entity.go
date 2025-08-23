package task

import "time"

type Task struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	VideoID   uint      `json:"video_id"`
	Status    string    `json:"status"` // pending, processing, completed, failed
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	ErrorMsg  string    `json:"error_message,omitempty"`
}
