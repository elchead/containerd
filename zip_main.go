package main

// import (
// 	"github.com/containerd/containerd/pkg/cri/util"
// )

// func main() {
// 	util.RecursiveZip("./bin/zip.gzip", "./bin/test")

// }

import (
	"archive/tar"
	// "filepath"
	"fmt"
	gzip "github.com/klauspost/pgzip"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

func RecursiveZip(pathToZip, zipPath string) error {

	fmt.Println("Creating zip..")
	fs, err := ioutil.ReadDir(pathToZip)
	var files []string
	for _, f := range fs {
		files = append(files, pathToZip+f.Name())
	}
	out, err := os.Create("./bin/output.tar.gz")
	if err != nil {
		return fmt.Errorf("Error writing archive: %v", err)
	}
	defer out.Close()

	// Create the archive and write the output to the "out" Writer
	err = createArchive(files, out)
	if err != nil {
		return fmt.Errorf("failed to create archive: %v", err)
	}
	fmt.Println("Created zip!")
	return nil
}

func Unzip(src, dest string) error {
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("mkdir -p %s && tar -xf %s -C %s", dest, src, dest))
	return cmd.Run()
}

func main() {
	// Files which to include in the tar.gz archive
	destPath := "./bin/test/"
	fs, err := ioutil.ReadDir(destPath)
	var files []string
	for _, f := range fs {
		files = append(files, destPath+f.Name())
	}
	// files := []string{"/Users/I545428/go/src/github.com/containerd/containerd/bin/test/test.txt", "./bin/test/tem"}

	// Create output file
	fname := "./bin/output.tar.gz"
	out, err := os.Create(fname)
	if err != nil {
		log.Fatalln("Error writing archive:", err)
	}
	defer out.Close()

	// Create the archive and write the output to the "out" Writer
	err = createArchive(files, out)
	if err != nil {
		log.Fatalln("Error creating archive:", err)
	}

	fmt.Println("Archive created successfully")
	err = Unzip(fname, "./bin/output")
	if err != nil {
		log.Fatalln("Error unzipping archive:", err)
	}
}

func createArchive(files []string, buf io.Writer) error {
	// Create new Writers for gzip and tar
	// These writers are chained. Writing to the tar writer will
	// write to the gzip writer which in turn will write to
	// the "buf" writer
	gw := gzip.NewWriter(buf)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Iterate over files and add them to the tar archive
	for _, file := range files {
		err := addToArchive(tw, file)
		if err != nil {
			return err
		}
	}

	return nil
}

func addToArchive(tw *tar.Writer, filename string) error {
	// Open the file which will be written into the archive
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Get FileInfo about our file providing file size, mode, etc.
	info, err := file.Stat()
	if err != nil {
		return err
	}

	// Create a tar Header from the FileInfo data
	header, err := tar.FileInfoHeader(info, info.Name())
	if err != nil {
		return err
	}

	// Use full path as name (FileInfoHeader only takes the basename)
	// If we don't do this the directory strucuture would
	// not be preserved
	// https://golang.org/src/archive/tar/common.go?#L626
	// header.Name = filename

	// Write file header to the tar archive
	err = tw.WriteHeader(header)
	if err != nil {
		return err
	}

	// Copy file content to tar archive
	_, err = io.Copy(tw, file)
	if err != nil {
		return err
	}

	return nil
}
