package user

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type UserService interface {
	SignUp(ctx *gin.Context, req *CreateUserRequest) (*UserResponse, error)
	Login(ctx *gin.Context, req *LoginRequest) (*TokenResponse, error)
	GetRankings() ([]RankingResponse, error) // <-- NUEVO
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

func (uc *UserController) SignUp(c *gin.Context) {
	req := new(CreateUserRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or malformed input data."})
		return
	}

	if err := uc.validate.Struct(req); err != nil {
		if validationErrs, ok := err.(validator.ValidationErrors); ok {
			for _, fieldErr := range validationErrs {
				if fieldErr.Field() == "Password2" && fieldErr.Tag() == "eqfield" {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Passwords do not match."})
					return
				}
			}
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation error. Please check the provided fields."})
		return
	}

	userResponse, err := uc.userService.SignUp(c, req)
	if err != nil {
		if strings.Contains(err.Error(), "email already exists") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "The provided email is already registered."})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "An error occurred while registering the user."})
		return
	}

	c.JSON(http.StatusCreated, userResponse)
}

func (uc *UserController) Login(c *gin.Context) {
	req := new(LoginRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data."})
		return
	}
	if err := uc.validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokenResponse, err := uc.userService.Login(c, req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials."})
		return
	}

	c.JSON(http.StatusOK, tokenResponse)
}

func (uc *UserController) GetRankings(c *gin.Context) {
	rankings, err := uc.userService.GetRankings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve rankings."})
		return
	}

	c.JSON(http.StatusOK, rankings)
}
