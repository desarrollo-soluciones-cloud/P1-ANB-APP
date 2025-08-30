package video

import "time"

type UploadVideoRequest struct {
	Title string `json:"title" form:"title" validate:"required"`
}

type VideoResponse struct {
	ID           uint       `json:"id"`
	UserID       uint       `json:"user_id"`
	Title        string     `json:"title"`
	Status       string     `json:"status"`
	OriginalURL  string     `json:"original_url"`
	ProcessedURL string     `json:"processed_url,omitempty"`
	VoteCount    int        `json:"votes"`
	UploadedAt   time.Time  `json:"uploaded_at"`
	ProcessedAt  *time.Time `json:"processed_at,omitempty"`
}
