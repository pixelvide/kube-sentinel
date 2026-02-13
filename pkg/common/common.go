package common

import (
	"crypto/rand"
	"encoding/base64"
	"os"
	"strings"
	"time"

	"k8s.io/klog/v2"
)

const (
	AppName = "kube-sentinel"

	JWTExpirationSeconds = 24 * 60 * 60 // 24 hours

	NodeTerminalPodName = "kube-sentinel-node-terminal-agent"

	KubectlAnnotation = "kubectl.kubernetes.io/last-applied-configuration"

	// db connection max idle time
	DBMaxIdleTime  = 10 * time.Minute
	DBMaxOpenConns = 100
	DBMaxIdleConns = 10
)

var (
	Port        = "8080"
	JwtSecret   = ""
	Host        = ""
	Base        = ""
	GitlabHosts = ""

	NodeTerminalImage = "busybox:latest"
	DBType            = "sqlite"
	DBDSN             = "dev.db"
	DBSchemaCore      = "public"
	DBSchemaApp       = "public"

	KubeSentinelEncryptKey = "kube-sentinel-default-encryption-key-change-in-production"

	CookieExpirationSeconds = 2 * JWTExpirationSeconds // double jwt

	DisableGZIP         = true
	DisableVersionCheck = false
	InsecureSkipVerify  = false

	APIKeyProvider = "api_key"

	AllowedOrigins []string
)

func GetTableName(schema, baseName string) string {
	if DBType == "postgres" && schema != "public" {
		return schema + "." + baseName
	}
	return baseName
}

func GetAppTableName(baseName string) string {
	return GetTableName(DBSchemaApp, baseName)
}

func GetCoreTableName(baseName string) string {
	return GetTableName(DBSchemaCore, baseName)
}

func LoadEnvs() {
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		JwtSecret = secret
	} else {
		// Generate random secret if not provided to avoid hardcoded secrets
		b := make([]byte, 32)
		_, err := rand.Read(b)
		if err != nil {
			klog.Fatalf("Failed to generate random JWT secret: %v", err)
		}
		JwtSecret = base64.StdEncoding.EncodeToString(b)
		klog.Warning("JWT_SECRET is not set, using a randomly generated secret. Sessions will be invalidated on restart.")
	}

	if port := os.Getenv("PORT"); port != "" {
		Port = port
	}

	if nodeTerminalImage := os.Getenv("NODE_TERMINAL_IMAGE"); nodeTerminalImage != "" {
		NodeTerminalImage = nodeTerminalImage
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

	if schema := os.Getenv("DB_SCHEMA_CORE"); schema != "" {
		DBSchemaCore = schema
	}

	if schema := os.Getenv("DB_SCHEMA_APP"); schema != "" {
		DBSchemaApp = schema
	}

	if key := os.Getenv("KUBE_SENTINEL_ENCRYPT_KEY"); key != "" {
		KubeSentinelEncryptKey = key
	} else {
		klog.Warningf("KUBE_SENTINEL_ENCRYPT_KEY is not set, using default key, this is not secure for production!")
	}

	if v := os.Getenv("HOST"); v != "" {
		Host = v
	}
	if v := os.Getenv("DISABLE_GZIP"); v != "" {
		DisableGZIP = v == "true"
	}

	if v := os.Getenv("DISABLE_VERSION_CHECK"); v == "true" {
		DisableVersionCheck = true
	}

	if v := os.Getenv("KUBE_SENTINEL_BASE"); v != "" {
		if v[0] != '/' {
			v = "/" + v
		}
		Base = strings.TrimRight(v, "/")
		klog.Infof("Using base path: %s", Base)
	}

	if v := os.Getenv("GITLAB_HOSTS"); v != "" {
		GitlabHosts = v
	}

	if v := os.Getenv("INSECURE_SKIP_VERIFY"); v == "true" {
		InsecureSkipVerify = true
		klog.Warning("INSECURE_SKIP_VERIFY is set to true, SSL certificate verification will be skipped")
	}

	if v := os.Getenv("ALLOWED_ORIGINS"); v != "" {
		AllowedOrigins = strings.Split(v, ",")
		for i := range AllowedOrigins {
			AllowedOrigins[i] = strings.TrimSpace(AllowedOrigins[i])
		}
		klog.Infof("CORS Allowed Origins: %v", AllowedOrigins)
	}
}
