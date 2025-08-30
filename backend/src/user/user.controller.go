package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type UserService interface {
	Register(ctx *gin.Context, req *CreateUserRequest) (*UserResponse, error)
	Login(ctx *gin.Context, req *LoginRequest) (*TokenResponse, error)
}

type UserController struct {
	userService UserService
	validate    *validator.Validate
}

func NewUserController(userService UserService) *UserController {
	return &UserController{
		userService: userService,
		validate:    validator.New(),
	}
}

func (uc *UserController) Register(c *gin.Context) {
	req := new(CreateUserRequest)

	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	if err := uc.validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userResponse, err := uc.userService.Register(c, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, userResponse)
}

func (uc *UserController) Login(c *gin.Context) {
	req := new(LoginRequest)

	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	if err := uc.validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokenResponse, err := uc.userService.Login(c, req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tokenResponse)
}
