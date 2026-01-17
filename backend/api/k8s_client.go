package api

import (
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
)

// Cache structure
type cachedClient struct {
	Clientset     *kubernetes.Clientset
	MetricsClient *metricsclientset.Clientset
	Config        *rest.Config
	ClientConfig  clientcmd.ClientConfig
	LastMod       time.Time
	KubeConfig    string
}

var (
	clientCache = make(map[string]*cachedClient)
	cacheMutex  sync.RWMutex
)

// GetClientInfo returns a kubernetes clientset and rest config for a specific context and user storage namespace
func GetClientInfo(storageNamespace string, contextName string) (*kubernetes.Clientset, *rest.Config, error) {
	client, err := getCachedClient(storageNamespace, contextName)
	if err != nil {
		return nil, nil, err
	}
	return client.Clientset, client.Config, nil
}

// GetMetricsClient returns a metrics clientset for a specific context and user storage namespace
func GetMetricsClient(storageNamespace string, contextName string) (*metricsclientset.Clientset, error) {
	client, err := getCachedClient(storageNamespace, contextName)
	if err != nil {
		return nil, err
	}
	// Return nil if initialization failed silently (though we handle it below)
	return client.MetricsClient, nil
}

// GetClientConfig returns the raw ClientConfig interface for use with other tools (like Helm)
func GetClientConfig(storageNamespace string, contextName string) (clientcmd.ClientConfig, error) {
	client, err := getCachedClient(storageNamespace, contextName)
	if err != nil {
		return nil, err
	}
	return client.ClientConfig, nil
}

func getCachedClient(storageNamespace string, contextName string) (*cachedClient, error) {
	if storageNamespace == "" {
		return nil, os.ErrNotExist // Enforce isolation
	}

	kubeconfig := GetUserKubeConfigPath(storageNamespace)
	fileInfo, err := os.Stat(kubeconfig)
	if err != nil {
		return nil, err
	}
	modTime := fileInfo.ModTime()

	cacheKey := storageNamespace + "::" + contextName

	// Check Cache
	cacheMutex.RLock()
	if cached, ok := clientCache[cacheKey]; ok {
		if cached.LastMod.Equal(modTime) && cached.KubeConfig == kubeconfig {
			cacheMutex.RUnlock()
			return cached, nil
		}
	}
	cacheMutex.RUnlock()

	log.Printf("Cache Miss for %s (ModTime changed or new)", cacheKey)

	// Cache Miss or Stale - Recreate
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	// Double check inside lock
	if cached, ok := clientCache[cacheKey]; ok {
		if cached.LastMod.Equal(modTime) && cached.KubeConfig == kubeconfig {
			return cached, nil
		}
	}

	// Load raw config to process contexts
	config, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		return nil, err
	}

	// If no context provided, use current
	if contextName == "" {
		contextName = config.CurrentContext
	}

	// Sanitize config: Replace "force-keyring" with "no" in glab exec args
	for _, authInfo := range config.AuthInfos {
		if authInfo.Exec != nil {
			for i, arg := range authInfo.Exec.Args {
				if arg == "force-keyring" {
					log.Printf("Sanitizing kubeconfig: replacing 'force-keyring' with 'no' for auth info")
					authInfo.Exec.Args[i] = "no"
				}
			}
			if storageNamespace != "" {
				glabConfigDir := filepath.Join("/data", storageNamespace, ".config", "glab-cli")
				authInfo.Exec.Env = append(authInfo.Exec.Env, clientcmdapi.ExecEnvVar{
					Name:  "GLAB_CONFIG_DIR",
					Value: glabConfigDir,
				})
			}
		}
	}

	clientConfig := clientcmd.NewNonInteractiveClientConfig(*config, contextName, &clientcmd.ConfigOverrides{}, nil)
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	// Initialize metrics client
	// We don't error out if this fails, just log it? Or maybe we should?
	// Usually invalid config fails both.
	metricsClient, err := metricsclientset.NewForConfig(restConfig)
	if err != nil {
		log.Printf("Warning: Failed to create metrics client: %v", err)
		// We continue without metrics client
	}

	// Update Cache
	cached := &cachedClient{
		Clientset:     clientset,
		MetricsClient: metricsClient,
		Config:        restConfig,
		ClientConfig:  clientConfig,
		LastMod:       modTime,
		KubeConfig:    kubeconfig,
	}
	clientCache[cacheKey] = cached

	return cached, nil
}
