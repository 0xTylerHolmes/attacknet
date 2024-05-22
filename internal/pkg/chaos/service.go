package chaos

import (
	chaos_mesh "attacknet/cmd/internal/pkg/chaos/chaos-mesh"
	"attacknet/cmd/internal/pkg/kubernetes"
	log "github.com/sirupsen/logrus"
)

// Service facilitates the chaos injection within a kubernetes environment
type Service struct {
	ChaosClient *chaos_mesh.ChaosClient
	KubeClient  *kubernetes.KubeClient
}

func NewService(kubernetesNamespace string) (*Service, error) {
	kubeClient, err := kubernetes.CreateKubeClient(kubernetesNamespace)
	if err != nil {
		return nil, err
	}

	// create chaos-mesh client
	log.Infof("Creating a chaos-mesh client")
	chaosClient, err := chaos_mesh.CreateClient(kubernetesNamespace, kubeClient)
	if err != nil {
		return nil, err
	}
	return &Service{
		ChaosClient: chaosClient,
		KubeClient:  kubeClient,
	}, nil
}
