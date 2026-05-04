package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/mi-michi/backend/internal/cats"
	"github.com/mi-michi/backend/internal/db"
	"github.com/mi-michi/backend/internal/health"
	"github.com/mi-michi/backend/internal/middleware"
	"github.com/mi-michi/backend/internal/notifications"
	"github.com/mi-michi/backend/internal/profile"
	"github.com/mi-michi/backend/internal/vaccines"
)

func main() {
	// Cargar .env si existe
	if err := godotenv.Load(); err != nil {
		log.Println("No se encontró .env, usando variables de entorno del sistema")
	}

	// Conectar MongoDB
	db.Connect()
	defer db.Disconnect()

	// Iniciar job de notificaciones diarias
	notifications.StartDailyJob()

	// Configurar Gin
	r := gin.Default()

	// CORS — permite Flutter en desarrollo y producción
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
	}))

	// ── Rutas públicas ──────────────────────────────────────────
	r.POST("/auth/google", profile.HandleGoogleLogin)
	r.POST("/auth/google/desktop", profile.HandleGoogleDesktopLogin)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "mi-michi-backend"})
	})

	// ── Rutas protegidas (sin auth de momento — modo desarrollo) ──
	api := r.Group("/api", middleware.DevAuth())
	{
		// Perfil
		api.GET("/profile", profile.HandleGetProfile)
		api.PUT("/profile", profile.HandleUpdateProfile)

		// Gatos
		api.GET("/cats", cats.HandleList)
		api.POST("/cats", cats.HandleCreate)
		api.GET("/cats/:id", cats.HandleGet)
		api.PUT("/cats/:id", cats.HandleUpdate)
		api.DELETE("/cats/:id", cats.HandleDelete)
		api.POST("/cats/:id/photo", cats.HandleUploadPhoto)

		// Salud
		api.GET("/cats/:id/health", health.HandleList)
		api.POST("/cats/:id/health", health.HandleCreate)
		api.PUT("/cats/:id/health/:logId", health.HandleUpdate)
		api.DELETE("/cats/:id/health/:logId", health.HandleDelete)

		// Vacunas
		api.GET("/cats/:id/vaccines", vaccines.HandleList)
		api.POST("/cats/:id/vaccines", vaccines.HandleCreate)
		api.PUT("/cats/:id/vaccines/:vacId", vaccines.HandleUpdate)
		api.DELETE("/cats/:id/vaccines/:vacId", vaccines.HandleDelete)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("🚀 Mi Michi Backend corriendo en :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Error al iniciar servidor: %v", err)
	}
}
