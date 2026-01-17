package db

import (
	"fmt"
	"log"
	"os"
	"time"

	"cloud-sentinel-k8s/pkg/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbSSLMode := os.Getenv("DB_SSLMODE")
	if dbSSLMode == "" {
		dbSSLMode = "require"
	}
	dbTimeZone := os.Getenv("DB_TIMEZONE")
	if dbTimeZone == "" {
		dbTimeZone = "UTC"
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		dbPort,
		dbSSLMode,
		dbTimeZone,
	)

	var err error
	for i := 0; i < 10; i++ {
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("Failed to connect to database: %v. Retrying in 2 seconds...", err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatal("Failed to connect to database after retries: ", err)
	}

	log.Println("Database connection established")

	// Migration
	err = DB.AutoMigrate(&models.App{}, &models.OAuthProvider{}, &models.User{}, &models.UserIdentity{}, &models.AuditLog{}, &models.GitlabConfig{}, &models.GitlabK8sAgentConfig{}, &models.K8sClusterContextMapping{}, &models.KubeConfig{}, &models.AWSConfig{}, &models.EKSCluster{})
	if err != nil {
		log.Fatal("Failed to migrate database: ", err)
	}
	log.Println("Database migration completed")

	// Ensure cloud-sentinel-k8s app exists
	var app models.App
	result := DB.Where("name = ?", "cloud-sentinel-k8s").First(&app)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			app = models.App{
				Name:    "cloud-sentinel-k8s",
				Enabled: true,
			}
			if err := DB.Create(&app).Error; err != nil {
				log.Printf("Failed to create cloud-sentinel-k8s app: %v", err)
			} else {
				log.Println("Created cloud-sentinel-k8s app")
			}
		} else {
			log.Printf("Error checking for cloud-sentinel-k8s app: %v", result.Error)
		}
	}
}
