package internal

import (
	"os"
	"path/filepath"

	"github.com/pixelvide/kube-sentinel/pkg/cluster"
	"github.com/pixelvide/kube-sentinel/pkg/model"
	"github.com/pixelvide/kube-sentinel/pkg/rbac"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"
)

var (
	kubeSentinelUsername = os.Getenv("KUBE_SENTINEL_USERNAME")
	kubeSentinelPassword = os.Getenv("KUBE_SENTINEL_PASSWORD")
)

func loadUser() error {
	if kubeSentinelUsername != "" && kubeSentinelPassword != "" {
		uc, err := model.CountUsers()
		if err == nil && uc == 0 {
			klog.Infof("Creating super user %s from environment variables", kubeSentinelUsername)
			u := &model.User{
				Username: kubeSentinelUsername,
				Password: kubeSentinelPassword,
			}
			err := model.AddSuperUser(u)
			if err == nil {
				rbac.SyncNow <- struct{}{}
			} else {
				return err
			}
		}
	}

	return nil
}

func loadClusters() error {
	cc, err := model.CountClusters()
	if err != nil || cc > 0 {
		return err
	}
	kubeconfigpath := ""
	if home := homedir.HomeDir(); home != "" {
		kubeconfigpath = filepath.Join(home, ".kube", "config")
	}

	if envKubeconfig := os.Getenv("KUBECONFIG"); envKubeconfig != "" {
		kubeconfigpath = envKubeconfig
	}

	config, _ := os.ReadFile(kubeconfigpath)

	if len(config) == 0 {
		return nil
	}
	kubeconfig, err := clientcmd.Load(config)
	if err != nil {
		return err
	}

	klog.Infof("Importing clusters from kubeconfig: %s", kubeconfigpath)
	cluster.ImportClustersFromKubeconfig(kubeconfig)
	return nil
}

// LoadConfigFromEnv loads configuration from environment variables.
func LoadConfigFromEnv() {
	if err := loadUser(); err != nil {
		klog.Warningf("Failed to migrate env to db user: %v", err)
	}

	if err := loadClusters(); err != nil {
		klog.Warningf("Failed to migrate env to db cluster: %v", err)
	}
}
