package vote

import (
	"anb-app/src/video"
	"errors"
	"time"

	"gorm.io/gorm"
)

type voteService struct {
	voteRepo VoteRepository
	db       *gorm.DB
}

func NewVoteService(voteRepo VoteRepository, db *gorm.DB) VoteService {
	return &voteService{
		voteRepo: voteRepo,
		db:       db,
	}
}

func (s *voteService) CreateVote(userID uint, videoID uint) error {
	existingVote, err := s.voteRepo.FindByUserAndVideo(userID, videoID)
	if err != nil {
		return errors.New("database error when checking for existing vote")
	}
	if existingVote != nil {
		return errors.New("user has already voted for this video")
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	newVote := &Vote{
		UserID:  userID,
		VideoID: videoID,
		VotedAt: time.Now(),
	}
	if err := tx.Create(newVote).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(&video.Video{}).Where("id = ?", videoID).UpdateColumn("vote_count", gorm.Expr("vote_count + 1")).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *voteService) DeleteVote(userID uint, videoID uint) error {
	existingVote, err := s.voteRepo.FindByUserAndVideo(userID, videoID)
	if err != nil {
		return errors.New("database error when checking for vote to delete")
	}
	if existingVote == nil {
		return errors.New("vote does not exist")
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if err := s.voteRepo.DeleteByUserAndVideo(userID, videoID); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(&video.Video{}).Where("id = ?", videoID).UpdateColumn("vote_count", gorm.Expr("vote_count - 1")).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
