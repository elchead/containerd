package server

import (
	"fmt"
	// "github.com/avast/retry-go/v4"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/pkg/cri/util"
	"golang.org/x/net/context"
	runtime "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
	"os"
	"path/filepath"
)

// RestoreContainer restores a container from a previously created image.
// It's essentially the same as starting a container with the additon of loading a checkpoint.
func (c *criService) RestoreContainer(ctx context.Context, r *runtime.RestoreContainerRequest) (retRes *runtime.RestoreContainerResponse, retErr error) {
	// fmt.Println("Waiting 60s before restore")
	// time.Sleep(60 * time.Second)
	// fmt.Println("Finished waiting restore")
	checkPath := r.GetOptions().GetCheckpointPath()
	save := util.GetTmpPath(checkPath, "/mnt/migration")
	defer util.DeleteAllFiles(save)
	copyPath := fmt.Sprintf("/mnt/%s_check.tar.gz",util.GetId(checkPath))
	defer os.Delete(copyPath)
	zipPath := filepath.Join(filepath.Dir(checkPath), "check.tar.gz")
	fmt.Println("Start copy gz")
	CopyFile(copyPath, zipPath))
	fmt.Println("Finish copy gz")
	// retry.Do(func() error {
		// 	if fileExists(zipPath) {
			// 		return nil
			// 	} else {
				// 		return fmt.Errorf("file not existent")
				// 	}
				// })
	fmt.Println("Starting unzip")
	err := util.Unzip(copyPath, save)
	if err != nil {
		return nil, err
	}
	fmt.Println("Unzip complete")
	if err := c.startContainer(ctx, r.GetContainerId(), containerd.WithRestoreImagePath(save)); err != nil {
		return nil, fmt.Errorf("failed to restore container: %v", err)
	}
	return &runtime.RestoreContainerResponse{}, nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
