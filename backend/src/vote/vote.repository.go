package vote

import (
	"errors"

	"gorm.io/gorm"
)

type VoteRepository interface {
	FindByUserAndVideo(userID uint, videoID uint) (*Vote, error)
	Create(vote *Vote) (*Vote, error)
	DeleteByUserAndVideo(userID uint, videoID uint) error
}

type voteRepository struct {
	db *gorm.DB
}

func NewVoteRepository(db *gorm.DB) VoteRepository {
	return &voteRepository{
		db: db,
	}
}

func (r *voteRepository) FindByUserAndVideo(userID uint, videoID uint) (*Vote, error) {
	var vote Vote
	result := r.db.Where("user_id = ? AND video_id = ?", userID, videoID).First(&vote)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}

	return &vote, nil
}

func (r *voteRepository) Create(vote *Vote) (*Vote, error) {
	result := r.db.Create(vote)
	if result.Error != nil {
		return nil, result.Error
	}
	return vote, nil
}

func (r *voteRepository) DeleteByUserAndVideo(userID uint, videoID uint) error {
	result := r.db.Where("user_id = ? AND video_id = ?", userID, videoID).Delete(&Vote{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
