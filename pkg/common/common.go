package common

import (
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
	JwtSecret   = "kube-sentinel-default-jwt-secret-key-change-in-production"
	Host        = ""
	Base        = ""
	GitlabHosts = ""

	NodeTerminalImage = "busybox:latest"
	DBType            = "sqlite"
	DBDSN             = "dev.db"

	KubeSentinelEncryptKey = "kube-sentinel-default-encryption-key-change-in-production"

	CookieExpirationSeconds = 2 * JWTExpirationSeconds // double jwt

	DisableGZIP         = true
	DisableVersionCheck = false
	InsecureSkipVerify  = false

	APIKeyProvider = "api_key"
)

func LoadEnvs() {
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		JwtSecret = secret
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

	if key := os.Getenv("CLOUD_SENTINEL_K8S_ENCRYPT_KEY"); key != "" {
		KubeSentinelEncryptKey = key
	} else {
		klog.Warningf("CLOUD_SENTINEL_K8S_ENCRYPT_KEY is not set, using default key, this is not secure for production!")
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

	if v := os.Getenv("CLOUD_SENTINEL_K8S_BASE"); v != "" {
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
}
