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
	// GORM se encarga de generar la sentencia SQL "INSERT INTO users..."
	result := r.db.Create(user)
	if result.Error != nil {
		// Si hay un error (ej. email duplicado por una restricción UNIQUE), lo devolvemos.
		return nil, result.Error
	}
	return user, nil
}

// FindByEmail busca un usuario por su dirección de email.
func (r *userRepository) FindByEmail(email string) (*User, error) {
	var user User
	// GORM genera el "SELECT * FROM users WHERE email = ? LIMIT 1"
	result := r.db.Where("email = ?", email).First(&user)

	if result.Error != nil {
		// GORM tiene un error específico para cuando no encuentra un registro.
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Esto no es un error del sistema, es un resultado esperado.
			// Devolvemos 'nil' para indicar que no se encontró el usuario.
			return nil, nil
		}
		// Para cualquier otro tipo de error, lo devolvemos.
		return nil, result.Error
	}

	return &user, nil
}
