package cluster

import (
	"testing"

	"github.com/pixelvide/kube-sentinel/pkg/common"
	"github.com/pixelvide/kube-sentinel/pkg/model"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func TestImportClustersFromKubeconfig_Sanitization(t *testing.T) {
	// Setup in-memory DB
	common.DBType = "sqlite"
	common.DBDSN = ":memory:"
	model.InitDB()

	config := clientcmdapi.NewConfig()
	config.Clusters["test-cluster"] = &clientcmdapi.Cluster{Server: "https://localhost:6443"}
	config.AuthInfos["test-user"] = &clientcmdapi.AuthInfo{
		Exec: &clientcmdapi.ExecConfig{
			Command: "glab",
			Args:    []string{"auth", "token"},
		},
	}
	// Context with slash
	contextName := "production/us-east-1"
	config.Contexts[contextName] = &clientcmdapi.Context{
		Cluster:  "test-cluster",
		AuthInfo: "test-user",
	}
	config.CurrentContext = contextName

	// Run import
	count := ImportClustersFromKubeconfig(config)
	assert.Equal(t, int64(1), count)

	// Verify sanitization
	expectedName := "production-us-east-1"
	cluster, err := model.GetClusterByName(expectedName)
	assert.NoError(t, err)
	assert.NotNil(t, cluster)
	assert.Equal(t, expectedName, cluster.Name)
	assert.True(t, cluster.IsDefault)
}

func TestImportClustersFromKubeconfig_AWS(t *testing.T) {
	// Setup in-memory DB
	common.DBType = "sqlite"
	common.DBDSN = ":memory:"
	model.InitDB()

	config := clientcmdapi.NewConfig()
	config.Clusters["aws-cluster"] = &clientcmdapi.Cluster{Server: "https://aws-eks.com"}
	config.AuthInfos["aws-user"] = &clientcmdapi.AuthInfo{
		Exec: &clientcmdapi.ExecConfig{
			Command: "aws",
			Args:    []string{"eks", "get-token", "--cluster-name", "my-eks-cluster", "--region", "us-west-2"},
		},
	}
	config.Contexts["aws-ctx"] = &clientcmdapi.Context{
		Cluster:  "aws-cluster",
		AuthInfo: "aws-user",
	}

	// Run import
	ImportClustersFromKubeconfig(config)

	// Verify conversion
	cluster, err := model.GetClusterByName("aws-ctx")
	assert.NoError(t, err)
	assert.NotNil(t, cluster)
	assert.True(t, cluster.SkipSystemSync)

	// Check updated kubeconfig
	importedConfig, err := clientcmd.Load([]byte(cluster.Config))
	assert.NoError(t, err)
	authInfo := importedConfig.AuthInfos["aws-user"]
	assert.NotNil(t, authInfo)
	assert.NotNil(t, authInfo.Exec)
	assert.Equal(t, "aws-iam-authenticator", authInfo.Exec.Command)
	assert.Contains(t, authInfo.Exec.Args, "token")
	assert.Contains(t, authInfo.Exec.Args, "-i")
	assert.Contains(t, authInfo.Exec.Args, "my-eks-cluster")
	assert.NotContains(t, authInfo.Exec.Args, "--region")

	// Verify Env variables
	var regionEnv, stsEnv string
	for _, env := range authInfo.Exec.Env {
		if env.Name == "AWS_REGION" {
			regionEnv = env.Value
		}
		if env.Name == "AWS_STS_REGIONAL_ENDPOINTS" {
			stsEnv = env.Value
		}
	}
	assert.Equal(t, "us-west-2", regionEnv)
	assert.Equal(t, "regional", stsEnv)
}
