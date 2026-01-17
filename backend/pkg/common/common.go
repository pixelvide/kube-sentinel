package common

import (
	"os"

	"k8s.io/klog/v2"
)

const (
	// AppName is the name of the application
	AppName = "cloud-sentinel-k8s"

	// CookieExpirationSeconds is the expiration time for auth cookies (1 day)
	CookieExpirationSeconds = 3600 * 24
)

var (
	// CloudSentinalEncryptKey is the key used for encryption
	CloudSentinalEncryptKey = "cloud-sentinel-encrypt-key"

	DBType = "sqlite"
	DBDSN  = "dev.db"

	JWTSecret = []byte("cloud-sentinel-default-secret-change-me")
)

func LoadEnvs() {
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		JWTSecret = []byte(secret)
	}

	if key := os.Getenv("CLOUD_SENTINAL_ENCRYPT_KEY"); key != "" {
		CloudSentinalEncryptKey = key
	}

	if dbDSN := os.Getenv("DB_DSN"); dbDSN != "" {
		DBDSN = dbDSN
	}

	if dbType := os.Getenv("DB_TYPE"); dbType != "" {
		if dbType != "sqlite" && dbType != "mysql" && dbType != "postgres" {
			klog.Fatalf("Invalid DB_TYPE: %s, must be one of sqlite, mysql, postgres", dbType)
		}
		DBType = dbType
	}
}
