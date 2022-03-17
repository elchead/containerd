package server

import (
	"fmt"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/pkg/cri/util"
	"golang.org/x/net/context"
	runtime "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
	"log"
	// "syscall"
)

func (c *criService) CheckpointContainer(ctx context.Context, r *runtime.CheckpointContainerRequest) (retRes *runtime.CheckpointContainerResponse, retErr error) {
	cntr, err := c.containerStore.Get(r.GetContainerId())
	if err != nil {
		return nil, fmt.Errorf("failed to find container: %v", err)
	}
	task, err := cntr.Container.Task(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to checkpoint container task: %v", err)
	}
	checkPath := r.GetOptions().GetCheckpointPath()
	save := util.GetTmpPath(checkPath, "/mnt/migration")
	opts := []containerd.CheckpointTaskOpts{containerd.WithCheckpointImagePath(save)}
	if !r.GetOptions().LeaveRunning {
		opts = append(opts, containerd.WithCheckpointExit())
	}
	_, err = task.Checkpoint(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to checkpoint container: %v", err)
	}

	log.Println("Start upload")
	err = util.UploadDirAz(save)
	if err != nil {
		return nil, err
	}
	log.Println("Finish upload")
	defer util.DeleteAllFiles(save)
	return &runtime.CheckpointContainerResponse{}, nil
}
