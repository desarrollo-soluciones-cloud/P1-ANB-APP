package vote

import "time"

type CreateVoteRequest struct {
	VideoID uint `json:"video_id" validate:"required"`
}
type VoteResponse struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	VideoID   uint      `json:"video_id"`
	CreatedAt time.Time `json:"created_at"`
}
