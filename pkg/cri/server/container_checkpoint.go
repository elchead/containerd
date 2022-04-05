package server

import (
	"fmt"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/pkg/cri/util"
	"golang.org/x/net/context"
	runtime "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
	// "syscall"
	"os"
	"path/filepath"
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
	save := "/mnt"
	opts := []containerd.CheckpointTaskOpts{containerd.WithCheckpointImagePath(save)}
	if !r.GetOptions().LeaveRunning {
		opts = append(opts, containerd.WithCheckpointExit())
	}
	fmt.Println(ctx)
	_, err = task.Checkpoint(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to checkpoint container: %v", err)
	}

	checkPath := r.GetOptions().GetCheckpointPath()
	zipPath := filepath.Join(filepath.Dir(checkPath), "check.zip")
	err = util.RecursiveZip(save, zipPath)
	if err != nil {
		return nil, fmt.Errorf("failed to zip checkpoint: %v, %s, %s", err, save, zipPath)
	}
	os.Remove(save)
	// if !r.GetOptions().GetLeaveRunning() {
	// 	task.Kill(ctx, syscall.SIGKILL)
	// }

	return &runtime.CheckpointContainerResponse{}, nil
}
