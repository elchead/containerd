package server

import (
	"fmt"
	"github.com/containerd/containerd"
	"golang.org/x/net/context"
	runtime "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
	"path/filepath"
)

// RestoreContainer restores a container from a previously created image.
// It's essentially the same as starting a container with the additon of loading a checkpoint.
func (c *criService) RestoreContainer(ctx context.Context, r *runtime.RestoreContainerRequest) (retRes *runtime.RestoreContainerResponse, retErr error) {
	// fmt.Println("Waiting 60s before restore")
	// time.Sleep(60 * time.Second)
	// fmt.Println("Finished waiting restore")
	checkPath := r.GetOptions().GetCheckpointPath()
	zipPath := filepath.Join(filepath.Dir(checkPath), "check.zip")
	fmt.Println("Restore here", checkPath, zipPath)
	err := util.Unzip(zipPath, checkPath)
	if err != nil {
		return nil, err
	}
	if err := c.startContainer(ctx, r.GetContainerId(), containerd.WithRestoreImagePath(r.GetOptions().GetCheckpointPath())); err != nil {
		return nil, fmt.Errorf("failed to restore container: %v", err)
	}
	return &runtime.RestoreContainerResponse{}, nil
}
