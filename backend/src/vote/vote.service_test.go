package vote

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type MockVoteRepository struct {
	mock.Mock
}

func (m *MockVoteRepository) FindByUserAndVideo(userID, videoID uint) (*Vote, error) {
	args := m.Called(userID, videoID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Vote), args.Error(1)
}

func (m *MockVoteRepository) Create(vote *Vote) (*Vote, error) {
	args := m.Called(vote)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Vote), args.Error(1)
}

func (m *MockVoteRepository) DeleteByUserAndVideo(userID, videoID uint) error {
	args := m.Called(userID, videoID)
	return args.Error(0)
}

// Test simplificado sin transacciones
func TestVoteService_Simple(t *testing.T) {
	t.Run("CreateVote_AlreadyVoted", func(t *testing.T) {
		mockRepo := new(MockVoteRepository)
		mockDB := &gorm.DB{}
		voteSvc := NewVoteService(mockRepo, mockDB)

		userID := uint(1)
		videoID := uint(1)
		existingVote := &Vote{
			ID:      1,
			UserID:  userID,
			VideoID: videoID,
		}

		// Ya existe un voto
		mockRepo.On("FindByUserAndVideo", userID, videoID).Return(existingVote, nil)

		err := voteSvc.CreateVote(userID, videoID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already voted")
		mockRepo.AssertExpectations(t)
	})

	t.Run("CreateVote_DatabaseError", func(t *testing.T) {
		mockRepo := new(MockVoteRepository)
		mockDB := &gorm.DB{}
		voteSvc := NewVoteService(mockRepo, mockDB)

		userID := uint(1)
		videoID := uint(1)

		// Error al buscar voto
		mockRepo.On("FindByUserAndVideo", userID, videoID).Return(nil, errors.New("database error"))

		err := voteSvc.CreateVote(userID, videoID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
		mockRepo.AssertExpectations(t)
	})

	t.Run("DeleteVote_NotExists", func(t *testing.T) {
		mockRepo := new(MockVoteRepository)
		mockDB := &gorm.DB{}
		voteSvc := NewVoteService(mockRepo, mockDB)

		userID := uint(1)
		videoID := uint(1)

		// No existe voto para eliminar
		mockRepo.On("FindByUserAndVideo", userID, videoID).Return(nil, nil)

		err := voteSvc.DeleteVote(userID, videoID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not exist")
		mockRepo.AssertExpectations(t)
	})

	t.Run("DeleteVote_DatabaseError", func(t *testing.T) {
		mockRepo := new(MockVoteRepository)
		mockDB := &gorm.DB{}
		voteSvc := NewVoteService(mockRepo, mockDB)

		userID := uint(1)
		videoID := uint(1)

		// Error al buscar voto
		mockRepo.On("FindByUserAndVideo", userID, videoID).Return(nil, errors.New("database error"))

		err := voteSvc.DeleteVote(userID, videoID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
		mockRepo.AssertExpectations(t)
	})
}
