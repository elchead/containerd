package server

import (
	"fmt"
	"github.com/containerd/containerd"
	"golang.org/x/net/context"
	runtime "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
	"time"
)

// RestoreContainer restores a container from a previously created image.
// It's essentially the same as starting a container with the additon of loading a checkpoint.
func (c *criService) RestoreContainer(ctx context.Context, r *runtime.RestoreContainerRequest) (retRes *runtime.RestoreContainerResponse, retErr error) {
	// fmt.Println("Waiting 60s before restore")
	// time.Sleep(60 * time.Second)
	// fmt.Println("Finished waiting restore")
	fmt.Println("Restore here:", r.GetOptions().GetCheckpointPath())
	if err := c.startContainer(ctx, r.GetContainerId(), containerd.WithRestoreImagePath(r.GetOptions().GetCheckpointPath())); err != nil {
		return nil, fmt.Errorf("failed to restore container: %v", err)
	}
	return &runtime.RestoreContainerResponse{}, nil
}
