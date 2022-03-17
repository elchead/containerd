package util

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"github.com/joho/godotenv"
)

var url string
var ssa string
const envFile = "./token.env"

func init() {
	err := godotenv.Load(envFile)
	url = os.Getenv("URL")
	ssa = os.Getenv("SSA")
}

func DeleteAllFiles(dir string) error {
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("rm -r %s", filepath.Join(dir, "*")))
	fmt.Printf("rm -r %s", filepath.Join(dir, "*"))
	return cmd.Run()
}

func GetId(containerPath string) string {
	return filepath.Base(filepath.Dir(containerPath))
}

func GetTmpPath(containerPath, tmpPath string) string {
	path := filepath.Join(tmpPath, GetId(containerPath))
	os.MkdirAll(path, 0755)
	return path
}

func createArchiveSizeFile(zipPath string) {
	i, _ := os.Stat(zipPath)
	sz := float64(i.Size())
	sizeFile := filepath.Join(filepath.Dir(filepath.Dir(zipPath)), filepath.Base(filepath.Dir(zipPath))+"_"+humanFileSize(sz))
	os.Create(sizeFile)
}

func UploadDirAz(copyPath string) error {
	completeUrl := url+"/"+os.Getenv("SSA")
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("azcopy copy %s \"%s\" --recursive", copyPath, completeUrl))
	outfile, err := os.Create("/mnt/upload.txt")
	if err != nil {
		panic(err)
	}
	defer outfile.Close()
	cmd.Stdout = outfile
	return cmd.Run()
}

func DownloadDirAz(remotePath, localPath string) error {
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("azcopy copy \"%s/%s/%s\" %s --recursive", url,remotePath, ssa, localPath))
	// open the out file for writing
	outfile, err := os.Create("/mnt/download.txt")
	if err != nil {
		panic(err)
	}
	defer outfile.Close()
	cmd.Stdout = outfile
	return cmd.Run()
}

func CopyFile(copyPath, originalPath string) error {
	r, err := os.Open(originalPath)
	if err != nil {
		log.Fatalf("could not open zip file: %v", err)
	}
	copy, err := os.Create(copyPath)
	if err != nil {
		log.Fatalf("could not copy zip file: %v", err)
	}
	io.Copy(copy, r)
	copy.Close()
	return nil
}
