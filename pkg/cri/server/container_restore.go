package server

import (
	"fmt"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/pkg/cri/util"
	"golang.org/x/net/context"
	runtime "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
	"log"
)

// RestoreContainer restores a container from a previously created image.
// It's essentially the same as starting a container with the additon of loading a checkpoint.
func (c *criService) RestoreContainer(ctx context.Context, r *runtime.RestoreContainerRequest) (retRes *runtime.RestoreContainerResponse, retErr error) {
	checkPath := r.GetOptions().GetCheckpointPath()
	jId := util.GetId(checkPath)
	log.Println("Start download")
	err := util.DownloadDirAz(jId, "/mnt/migration")
	log.Println("Finish download")
	if err != nil {
		return nil, err
	}
	save := util.GetTmpPath(checkPath, "/mnt/migration")
	if err := c.startContainer(ctx, r.GetContainerId(), containerd.WithRestoreImagePath(save)); err != nil {
		return nil, fmt.Errorf("failed to restore container: %v", err)
	}
	defer util.DeleteAllFiles(save)
	return &runtime.RestoreContainerResponse{}, nil
}
