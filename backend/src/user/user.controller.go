package user

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// UserService define la interfaz para la lógica de negocio de usuarios.
// Actualizamos el contexto para que sea el de Gin (*gin.Context).
type UserService interface {
	Register(ctx *gin.Context, req *CreateUserRequest) (*UserResponse, error)
}

// UserController maneja las peticiones HTTP para la entidad User.
type UserController struct {
	userService UserService
	validate    *validator.Validate
}

// NewUserController crea una nueva instancia del controlador de usuarios.
func NewUserController(userService UserService) *UserController {
	return &UserController{
		userService: userService,
		validate:    validator.New(),
	}
}

// Register es el handler para el endpoint de registro con Gin.
func (uc *UserController) Register(c *gin.Context) {
	// 1. Instanciar el DTO de la solicitud
	req := new(CreateUserRequest)

	// 2. Vincular el JSON de la petición al DTO
	// Gin usa c.ShouldBindJSON() en lugar de c.Bind().
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// 3. Validar la estructura usando las etiquetas 'validate' del DTO
	if err := uc.validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 4. Llamada al servicio (COMENTADA)
	/*
		userResponse, err := uc.userService.Register(c, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	*/

	// 5. Devolver una respuesta exitosa y HARCODEADA
	mockResponse := &UserResponse{
		ID:        1,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		City:      req.City,
		Country:   req.Country,
		CreatedAt: time.Now(),
	}

	c.JSON(http.StatusCreated, mockResponse)
}
