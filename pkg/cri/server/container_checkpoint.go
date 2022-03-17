package server

import (
	"fmt"
	"github.com/containerd/containerd"
	"golang.org/x/net/context"
	runtime "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
	"syscall"
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
	opts := []containerd.CheckpointTaskOpts{containerd.WithCheckpointImagePath(r.GetOptions().GetCheckpointPath())}
	if !r.GetOptions().LeaveRunning {
		opts = append(opts, containerd.WithCheckpointExit())
	}
	_, err = task.Checkpoint(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to checkpoint container: %v", err)
	}

	if !r.GetOptions().GetLeaveRunning() {
		task.Kill(ctx, syscall.SIGKILL)
	}

	return &runtime.CheckpointContainerResponse{}, nil
}
