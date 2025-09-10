// En backend/src/database/database.go
package database

import (
	"anb-app/src/user"
	"anb-app/src/video"
	"anb-app/src/vote"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
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
	log.Println("Verificando estado de las tablas...")

	// Ejecutar migraciones automáticas
	err := db.AutoMigrate(&user.User{}, &video.Video{}, &vote.Vote{})
	if err != nil {
		log.Fatalf("Error al ejecutar las migraciones: %v", err)
	}
	log.Println("Migraciones completadas.")

	// Poblar con datos de prueba si está habilitado
	SeedDatabase(db)
}

func SeedDatabase(db *gorm.DB) {
	// Verificar si ya hay datos en la base
	var userCount int64
	db.Model(&user.User{}).Count(&userCount)

	if userCount > 0 {
		log.Println("La base de datos ya contiene datos, omitiendo seeding.")
		return
	}

	log.Println("Poblando base de datos con datos de prueba...")

	// Hash para password "password"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error al generar hash de contraseña: %v", err)
		return
	}

	// Crear usuarios de prueba
	users := []user.User{
		{FirstName: "Carlos", LastName: "González", Email: "carlos@anb.com", Password: string(hashedPassword), City: "Bogotá", Country: "Colombia"},
		{FirstName: "María", LastName: "Rodríguez", Email: "maria@anb.com", Password: string(hashedPassword), City: "Medellín", Country: "Colombia"},
		{FirstName: "Luis", LastName: "Martínez", Email: "luis@anb.com", Password: string(hashedPassword), City: "Cali", Country: "Colombia"},
		{FirstName: "Ana", LastName: "López", Email: "ana@anb.com", Password: string(hashedPassword), City: "Barranquilla", Country: "Colombia"},
		{FirstName: "Miguel", LastName: "Hernández", Email: "miguel@anb.com", Password: string(hashedPassword), City: "Cartagena", Country: "Colombia"},
		{FirstName: "Sofía", LastName: "García", Email: "sofia@anb.com", Password: string(hashedPassword), City: "Bucaramanga", Country: "Colombia"},
		{FirstName: "Diego", LastName: "Vargas", Email: "diego@anb.com", Password: string(hashedPassword), City: "Pereira", Country: "Colombia"},
		{FirstName: "Camila", LastName: "Torres", Email: "camila@anb.com", Password: string(hashedPassword), City: "Manizales", Country: "Colombia"},
	}

	for i := range users {
		if err := db.Create(&users[i]).Error; err != nil {
			log.Printf("Error al crear usuario %s %s: %v", users[i].FirstName, users[i].LastName, err)
			continue
		}
	}
	log.Printf("Creados %d usuarios de prueba", len(users))

	// Crear videos de prueba usando archivos reales
	now := time.Now()
	processedTime := now.Add(-24 * time.Hour)

	// Usar el archivo real que existe en el volumen
	realVideoFile := "dunk_example.mp4"

	videos := []video.Video{
		// Videos procesados (usan archivos reales)
		{UserID: 1, Title: "Jugada Espectacular de Carlos - Triple desde media cancha", Status: "processed", OriginalURL: "/uploads/originals/" + realVideoFile, ProcessedURL: "/uploads/processed/" + realVideoFile, VoteCount: 45, UploadedAt: now, ProcessedAt: &processedTime},
		{UserID: 2, Title: "Defensa Perfecta de María - Robo y contraataque", Status: "processed", OriginalURL: "/uploads/originals/" + realVideoFile, ProcessedURL: "/uploads/processed/" + realVideoFile, VoteCount: 52, UploadedAt: now, ProcessedAt: &processedTime},
		{UserID: 3, Title: "Dunk Espectacular de Luis - Mate con giro 360°", Status: "processed", OriginalURL: "/uploads/originals/" + realVideoFile, ProcessedURL: "/uploads/processed/" + realVideoFile, VoteCount: 67, UploadedAt: now, ProcessedAt: &processedTime},
		{UserID: 4, Title: "Tiro libre bajo presión - Secuencia de 10/10", Status: "processed", OriginalURL: "/uploads/originals/" + realVideoFile, ProcessedURL: "/uploads/processed/" + realVideoFile, VoteCount: 25, UploadedAt: now, ProcessedAt: &processedTime},
		{UserID: 5, Title: "Salto vertical impresionante - 95cm de elevación", Status: "processed", OriginalURL: "/uploads/originals/" + realVideoFile, ProcessedURL: "/uploads/processed/" + realVideoFile, VoteCount: 19, UploadedAt: now, ProcessedAt: &processedTime},
		{UserID: 6, Title: "Combinación perfecta - Dribleo y mate", Status: "processed", OriginalURL: "/uploads/originals/" + realVideoFile, ProcessedURL: "/uploads/processed/" + realVideoFile, VoteCount: 28, UploadedAt: now, ProcessedAt: &processedTime},
		{UserID: 7, Title: "Tiro de 3 puntos desde esquina - Técnica perfecta", Status: "processed", OriginalURL: "/uploads/originals/" + realVideoFile, ProcessedURL: "/uploads/processed/" + realVideoFile, VoteCount: 33, UploadedAt: now, ProcessedAt: &processedTime},
		{UserID: 8, Title: "Asistencia no-look espectacular", Status: "processed", OriginalURL: "/uploads/originals/" + realVideoFile, ProcessedURL: "/uploads/processed/" + realVideoFile, VoteCount: 26, UploadedAt: now, ProcessedAt: &processedTime},

		// Videos solo subidos (sin archivo procesado)
		{UserID: 1, Title: "Triple Decisivo en el último segundo", Status: "uploaded", OriginalURL: "/uploads/originals/" + realVideoFile, VoteCount: 38, UploadedAt: now},
		{UserID: 1, Title: "Secuencia de tiros libres perfectos", Status: "uploaded", OriginalURL: "/uploads/originals/" + realVideoFile, VoteCount: 22, UploadedAt: now},
		{UserID: 2, Title: "Asistencia Increíble sin mirar", Status: "uploaded", OriginalURL: "/uploads/originals/" + realVideoFile, VoteCount: 34, UploadedAt: now},
		{UserID: 2, Title: "Técnica de dribleo avanzado entre conos", Status: "uploaded", OriginalURL: "/uploads/originals/" + realVideoFile, VoteCount: 18, UploadedAt: now},
		{UserID: 3, Title: "Robo y Contraataque Lightning", Status: "uploaded", OriginalURL: "/uploads/originals/" + realVideoFile, VoteCount: 41, UploadedAt: now},
		{UserID: 3, Title: "Bloqueo defensivo épico - Rechazo total", Status: "uploaded", OriginalURL: "/uploads/originals/" + realVideoFile, VoteCount: 29, UploadedAt: now},
		{UserID: 4, Title: "Jugada colectiva perfecta - Asistencia de lujo", Status: "uploaded", OriginalURL: "/uploads/originals/" + realVideoFile, VoteCount: 31, UploadedAt: now},
		{UserID: 5, Title: "Entrenamiento de resistencia - Sprint continuo", Status: "uploaded", OriginalURL: "/uploads/originals/" + realVideoFile, VoteCount: 0, UploadedAt: now},
		{UserID: 6, Title: "Defensa 1 vs 1 - Técnica de marcaje", Status: "uploaded", OriginalURL: "/uploads/originals/" + realVideoFile, VoteCount: 0, UploadedAt: now},
		{UserID: 7, Title: "Rebote ofensivo y segunda jugada", Status: "uploaded", OriginalURL: "/uploads/originals/" + realVideoFile, VoteCount: 0, UploadedAt: now},
		{UserID: 8, Title: "Entrenamiento de coordinación", Status: "uploaded", OriginalURL: "/uploads/originals/" + realVideoFile, VoteCount: 0, UploadedAt: now},

		// Videos con fallo en procesamiento
		{UserID: 1, Title: "Entrenamiento matutino - Técnica de dribleo", Status: "failed", OriginalURL: "/uploads/originals/" + realVideoFile, VoteCount: 0, UploadedAt: now},
		{UserID: 3, Title: "Salto vertical - Entrenamiento de potencia", Status: "failed", OriginalURL: "/uploads/originals/" + realVideoFile, VoteCount: 0, UploadedAt: now},
		{UserID: 4, Title: "Técnica de pivoteo y finalizacion", Status: "failed", OriginalURL: "/uploads/originals/" + realVideoFile, VoteCount: 0, UploadedAt: now},
	}

	for i := range videos {
		if err := db.Create(&videos[i]).Error; err != nil {
			log.Printf("Error al crear video: %v", err)
			continue
		}
	}
	log.Printf("Creados %d videos de prueba", len(videos))

	// Crear votos de prueba simulando interacciones reales
	votePatterns := map[uint][]uint{
		// user_id: [video_ids que vota - solo videos procesados de otros usuarios]
		1: {5, 6, 9, 10, 13, 16, 18, 20, 22},   // Carlos vota por videos de otros
		2: {1, 2, 9, 10, 13, 16, 18, 20, 22},   // María vota por videos de otros
		3: {1, 2, 5, 6, 13, 16, 18, 20, 22},    // Luis vota por videos de otros
		4: {1, 2, 5, 6, 9, 10, 16, 18, 20, 22}, // Ana vota por videos de otros
		5: {1, 2, 5, 6, 9, 10, 13, 18, 20},     // Miguel vota por videos de otros
		6: {1, 2, 5, 9, 10, 13, 16, 20, 22},    // Sofía vota por videos de otros
		7: {1, 2, 5, 6, 9, 13, 16, 18, 22},     // Diego vota por videos de otros
		8: {1, 2, 5, 6, 9, 10, 13, 16, 18, 20}, // Camila vota por videos de otros
	}

	var totalVotes int
	for userID, videoIDs := range votePatterns {
		for _, videoID := range videoIDs {
			voteRecord := vote.Vote{
				UserID:    userID,
				VideoID:   videoID,
				VotedAt:   now,
				CreatedAt: now,
			}
			if err := db.Create(&voteRecord).Error; err != nil {
				log.Printf("Error al crear voto user %d -> video %d: %v", userID, videoID, err)
				continue
			}
			totalVotes++
		}
	}
	log.Printf("Creados %d votos de prueba", totalVotes)

	log.Println("Base de datos poblada exitosamente con datos de prueba")
	log.Println("Usuarios de prueba - email/password:")
	log.Println("  carlos@anb.com/password")
	log.Println("  maria@anb.com/password")
	log.Println("  luis@anb.com/password")
	log.Println("  ana@anb.com/password")
	log.Println("  miguel@anb.com/password")
	log.Println("  sofia@anb.com/password")
	log.Println("  diego@anb.com/password")
	log.Println("  camila@anb.com/password")
}
