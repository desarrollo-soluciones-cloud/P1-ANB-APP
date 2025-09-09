// En backend/src/database/database.go
package database

import (
	"anb-app/src/user"
	"anb-app/src/video"
	"anb-app/src/vote"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectDB() *gorm.DB {
	err := godotenv.Load()
	if err != nil {
		log.Println("Advertencia: No se pudo encontrar el archivo .env, se usarán las variables de entorno del sistema.")
	}

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbname, port)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error fatal al conectar a la base de datos: %v", err)
	}

	log.Println("Conexión a la base de datos establecida exitosamente.")
	return db
}

func MigrateTables(db *gorm.DB) {
	// Verificar si las tablas ya existen
	if db.Migrator().HasTable(&user.User{}) {
		log.Println("Las tablas ya existen, omitiendo migraciones automáticas.")
		return
	}

	log.Println("Ejecutando migraciones...")
	err := db.AutoMigrate(&user.User{}, &video.Video{}, &vote.Vote{})
	if err != nil {
		log.Fatalf("Error al ejecutar las migraciones: %v", err)
	}
	log.Println("Migraciones completadas.")
}
