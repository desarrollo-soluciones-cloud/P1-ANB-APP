package main

import (
	"log"
	"project-one/src/auth"
	"project-one/src/user"
	"project-one/src/video"
	"project-one/src/vote"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func main() {
	// 1. Conexión a la Base de Datos (SQLite)
	db, err := gorm.Open(sqlite.Open("anb.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// 2. Migración Automática de las Entidades
	// GORM creará las tablas si no existen cuando se inicie la app.
	err = db.AutoMigrate(&user.User{}, &video.Video{}, &vote.Vote{})
	if err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	// 3. Configuración Crítica (Clave Secreta para JWT)
	// !! IMPORTANTE: En una aplicación real, esta clave NUNCA debe estar en el código.
	// !! Se debe cargar de forma segura desde una variable de entorno.
	jwtSecret := "MI_CLAVE_SECRETA_SUPREMAMENTE_SEGURA"

	// 4. Inyección de Dependencias (El "Cableado" de la aplicación)
	// Se crean las instancias en orden, pasando una como dependencia a la siguiente.
	authSvc := auth.NewAuthService(jwtSecret)
	userRepo := user.NewUserRepository(db)
	userSvc := user.NewUserService(userRepo, authSvc)
	userController := user.NewUserController(userSvc)

	// 5. Configuración del Router de Gin
	router := gin.Default()

	// Creamos un grupo de rutas para la API, por ejemplo /api/v1
	apiV1 := router.Group("/api/v1")
	{
		// Registramos todas las rutas de usuario (/users/register, /users/login)
		user.RegisterUserRoutes(apiV1, userController)

		// En el futuro, aquí registraremos las rutas para 'video' y 'vote'.
	}

	// 6. Iniciar el Servidor
	log.Println("Server is running on port 8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
