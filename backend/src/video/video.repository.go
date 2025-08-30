package video

import "gorm.io/gorm"

type videoRepository struct {
	db *gorm.DB
}

func NewVideoRepository(db *gorm.DB) VideoRepository {
	return &videoRepository{
		db: db,
	}
}

func (r *videoRepository) Create(video *Video) (*Video, error) {
	result := r.db.Create(video)
	if result.Error != nil {
		return nil, result.Error
	}

	return video, nil
}

func (r *videoRepository) FindByUserID(userID uint) ([]Video, error) {
	var videos []Video

	result := r.db.Where("user_id = ?", userID).Order("uploaded_at DESC").Find(&videos)
	if result.Error != nil {
		return nil, result.Error
	}

	return videos, nil
}

func (r *videoRepository) FindByID(videoID uint) (*Video, error) {
	var video Video
	result := r.db.First(&video, videoID)

	if result.Error != nil {

		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &video, nil
}

func (r *videoRepository) Delete(videoID uint) error {
	result := r.db.Delete(&Video{}, videoID)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *videoRepository) FindPublic() ([]Video, error) {
	var videos []Video

	result := r.db.Where("status = ?", "processed").Order("vote_count DESC").Find(&videos)
	if result.Error != nil {
		return nil, result.Error
	}
	return videos, nil
}

func (r *videoRepository) Update(video *Video) error {
	result := r.db.Save(video)
	return result.Error
}
