package user

import (
	"errors"

	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) Create(user *User) (*User, error) {
	result := r.db.Create(user)
	if result.Error != nil {
		return nil, result.Error
	}
	return user, nil
}

func (r *userRepository) FindByEmail(email string) (*User, error) {
	var user User
	result := r.db.Where("email = ?", email).First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}

	return &user, nil
}

func (r *userRepository) GetRankings() ([]RankingResponse, error) {
	var results []struct {
		Username   string
		City       string
		TotalVotes int
	}

	queryResult := r.db.Model(&User{}).
		Select("users.first_name || ' ' || users.last_name as username, users.city, SUM(videos.vote_count) as total_votes").
		Joins("LEFT JOIN videos ON videos.user_id = users.id").
		Group("users.id").
		Order("total_votes DESC").
		Scan(&results)

	if queryResult.Error != nil {
		return nil, queryResult.Error
	}

	rankings := make([]RankingResponse, len(results))
	for i, result := range results {
		rankings[i] = RankingResponse{
			Position: i + 1,
			Username: result.Username,
			City:     result.City,
			Votes:    result.TotalVotes,
		}
	}

	return rankings, nil
}
