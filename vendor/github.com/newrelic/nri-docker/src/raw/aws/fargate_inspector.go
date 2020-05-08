package aws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/infra-integrations-sdk/persist"
)

const fargateTaskMetadataCacheKey = "task-metadata-response"

// FargateInspector is responsible for listing containers and inspecting containers in Fargate.
// Both operations use the same data source and thus access it through a caching layer to avoid extra
// computations.
type FargateInspector struct {
	baseURL        *url.URL
	http           *http.Client
	containerStore persist.Storer
}

// NewFargateInspector creates a new FargateInspector
func NewFargateInspector(baseURL *url.URL) (*FargateInspector, error) {
	containerStore := persist.NewInMemoryStore()

	return &FargateInspector{
		baseURL:        baseURL,
		http:           fargateHTTPClient,
		containerStore: containerStore,
	}, nil
}

// ContainerList lists containers that the current Fargate container can see (only the container in the same
// task). It completely ignores any listing option for the moment.
func (i *FargateInspector) ContainerList(_ context.Context, _ types.ContainerListOptions) ([]types.Container, error) {
	var taskResponse TaskResponse
	err := i.taskResponseFromCacheOrNew(&taskResponse)
	if err != nil {
		return nil, err
	}

	containers := make([]types.Container, len(taskResponse.Containers))
	for index, container := range taskResponse.Containers {
		converted := containerResponseToDocker(container)
		containers[index] = converted
	}
	return containers, nil
}

// taskResponseFromCacheOrNew wraps the access to Fargate task metadata with a caching layer.
func (i *FargateInspector) taskResponseFromCacheOrNew(response *TaskResponse) error {
	defer func() {
		if err := i.containerStore.Save(); err != nil {
			log.Warn("error persisting Fargate task metadata: %s", err)
		}
	}()

	var err error
	_, err = i.containerStore.Get(fargateTaskMetadataCacheKey, response)
	if err == persist.ErrNotFound {
		err = i.fetchTaskResponse(response)
	}
	if err != nil {
		return fmt.Errorf("cannot fetch Fargate task metadata response: %s", err)
	}
	i.containerStore.Set(fargateTaskMetadataCacheKey, response)
	return nil
}

func (i *FargateInspector) fetchTaskResponse(taskResponse *TaskResponse) error {
	endpoint := TaskMetadataEndpoint(i.baseURL.String())

	response, err := metadataResponse(i.http, endpoint)
	if err != nil {
		return fmt.Errorf(
			"error when sending request to ECS task metadata endpoint (%s): %v",
			endpoint,
			err,
		)
	}
	log.Debug("task metadata response from endpoint %s: %s", endpoint, string(response))

	err = json.Unmarshal(response, taskResponse)
	if err != nil {
		return fmt.Errorf("error unmarshalling ECS task: %v", err)
	}
	return nil
}

// ContainerInspect returns metadata about a container given its container ID.
func (i *FargateInspector) ContainerInspect(_ context.Context, containerID string) (types.ContainerJSON, error) {
	var taskResponse TaskResponse
	err := i.taskResponseFromCacheOrNew(&taskResponse)
	if err != nil {
		return types.ContainerJSON{}, err
	}

	for _, container := range taskResponse.Containers {
		if container.ID == containerID {
			containerJSON := types.ContainerJSON{
				ContainerJSONBase: &types.ContainerJSONBase{ID: containerID},
			}
			return containerJSON, nil
		}
	}
	return types.ContainerJSON{}, errors.New("container not found")
}

func containerResponseToDocker(container ContainerResponse) types.Container {
	c := types.Container{
		ID:      container.ID,
		Names:   []string{container.Name},
		Image:   container.Image,
		ImageID: container.ImageID,
		Labels:  processFargateLabels(container.Labels),
		Status:  container.KnownStatus,
	}
	if created := container.CreatedAt; created != nil {
		c.Created = created.Unix()
	}
	return c
}

func processFargateLabels(labels map[string]string) map[string]string {
	for label, value := range labels {
		// The cluster name label has to be processed because in ECS with EC2 launch type it's only the cluster name.
		// Meanwhile, in ECS with Fargate launch type the same label has the full cluster ARN as value.
		if label == "com.amazonaws.ecs.cluster" && isECSARN(value) {
			labels[label] = ecsClusterNameFromARN(value)
		}
	}
	return labels
}

// ecsClusterNameFromARN extracts the cluster name from an ECS cluster ARN.
func ecsClusterNameFromARN(ecsClusterARN string) string {
	return strings.Split(ecsClusterARN, "/")[1]
}

// isARN returns whether the given string is an ECS ARN by looking for
// whether the string starts with "arn:aws:ecs" and contains the correct number
// of sections delimited by colons(:).
// Copied from: https://github.com/aws/aws-sdk-go/blob/81abf80dec07700b14a91ece14b8eca6c5e6b4f8/aws/arn/arn.go#L81
func isECSARN(arn string) bool {
	const arnPrefix = "arn:aws:ecs"
	const arnSections = 6

	return strings.HasPrefix(arn, arnPrefix) && strings.Count(arn, ":") >= arnSections-1
}
