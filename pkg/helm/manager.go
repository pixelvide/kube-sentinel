package helm

import (
	"log"
	"os"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/client-go/rest"
)

func GetRelease(config *rest.Config, namespace, name string) (*release.Release, error) {
	actionConfig := new(action.Configuration)
	clientGetter := &simpleRESTClientGetter{config: config}

	if err := actionConfig.Init(clientGetter, namespace, os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		return nil, err
	}

	client := action.NewGet(actionConfig)
	return client.Run(name)
}

func GetReleaseHistory(config *rest.Config, namespace, name string) ([]*release.Release, error) {
	actionConfig := new(action.Configuration)
	clientGetter := &simpleRESTClientGetter{config: config}

	if err := actionConfig.Init(clientGetter, namespace, os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		return nil, err
	}

	client := action.NewHistory(actionConfig)
	client.Max = 256
	return client.Run(name)
}

func RollbackRelease(config *rest.Config, namespace, name string, revision int) error {
	actionConfig := new(action.Configuration)
	clientGetter := &simpleRESTClientGetter{config: config}

	if err := actionConfig.Init(clientGetter, namespace, os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		return err
	}

	client := action.NewRollback(actionConfig)
	client.Version = revision
	return client.Run(name)
}
