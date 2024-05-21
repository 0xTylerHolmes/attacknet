package ethereum

import (
	kubernetes2 "attacknet/cmd/internal/kubernetes"
	"attacknet/cmd/internal/pkg/chaos/chaos-mesh"
	"context"
)

func getPodsToHealthCheck(
	ctx context.Context,
	kubeClient *kubernetes2.KubeClient,
	podsUnderTest []*chaos_mesh.PodUnderTest,
	podsUnderTestLookup map[string]*chaos_mesh.PodUnderTest,
	labelKey, labelValue string,
) ([]kubernetes2.KubePod, error) {

	var podsToHealthCheck []kubernetes2.KubePod
	// add pods under test that match the label criteria _and_ aren't expected to die
	// todo: depending on whether we're testing network recovery or node recovery, we may want to health check nodes we're expecting to die
	for _, pod := range podsUnderTest {
		if pod.MatchesLabel(labelKey, labelValue) && !pod.ExpectDeath {
			podsToHealthCheck = append(podsToHealthCheck, pod)
		}
	}

	// add pods that were not targeted by a fault
	bystanders, err := kubeClient.PodsMatchingLabel(ctx, labelKey, labelValue)
	if err != nil {
		return nil, err
	}
	for _, pod := range bystanders {
		_, match := podsUnderTestLookup[pod.GetName()]
		// don't add pods we've already added
		if !match {
			podsToHealthCheck = append(podsToHealthCheck, pod)
		}
	}
	return podsToHealthCheck, nil
}
