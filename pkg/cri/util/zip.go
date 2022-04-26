package util

// import (
// 	// "archive/zip"
// 	"archive/tar"
// 	"bytes"
// 	"fmt"
// 	gzip "github.com/klauspost/pgzip"
// 	"io"
// 	"os"
// 	"path/filepath"
// 	"strings"
// )

import (
	"archive/tar"
	"filepath"
	"fmt"
	gzip "github.com/klauspost/pgzip"
	"io"
	"io/ioutil"
	// "log"
	"os"
	"os/exec"
)

func RecursiveZip(pathToZip, zipPath string) error {

	fmt.Println("Creating zip..")
	fs, err := ioutil.ReadDir(pathToZip)
	var files []string
	for _, f := range fs {
		files = append(files, filepath.Join(pathToZip, f.Name()))
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

// func Compress(src string, buf io.Writer) error {
// 	// tar > gzip > buf
// 	zr := gzip.NewWriter(buf)
// 	tw := tar.NewWriter(zr)

// 	// walk through every file in the folder
// 	filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {
// 		// generate tar header
// 		header, err := tar.FileInfoHeader(fi, file)
// 		if err != nil {
// 			return err
// 		}

// 		// must provide real name
// 		// (see https://golang.org/src/archive/tar/common.go?#L626)
// 		header.Name = filepath.ToSlash(file)

// 		// write header
// 		if err := tw.WriteHeader(header); err != nil {
// 			return err
// 		}
// 		// if not a dir, write file content
// 		if !fi.IsDir() {
// 			data, err := os.Open(file)
// 			if err != nil {
// 				return err
// 			}
// 			if _, err := io.Copy(tw, data); err != nil {
// 				return err
// 			}
// 		}
// 		return nil
// 	})

// 	// produce tar
// 	if err := tw.Close(); err != nil {
// 		return err
// 	}
// 	// produce gzip
// 	if err := zr.Close(); err != nil {
// 		return err
// 	}
// 	//
// 	return nil
// }

// // check for path traversal and correct forward slashes
// func validRelPath(p string) bool {
// 	if p == "" || strings.Contains(p, `\`) || strings.HasPrefix(p, "/") || strings.Contains(p, "../") {
// 		return false
// 	}
// 	return true
// }

// func Uncompress(src, dst string) error {
// 	// ungzip
// 	zr, err := os.Open(src)
// 	if err != nil {
// 		return err
// 	}
// 	// untar
// 	tr := tar.NewReader(zr)

// 	// uncompress each element
// 	for {
// 		header, err := tr.Next()
// 		if err == io.EOF {
// 			break // End of archive
// 		}
// 		if err != nil {
// 			return err
// 		}
// 		target := filepath.Join(dst, header.Name)

// 		// validate name against path traversal
// 		if !validRelPath(header.Name) {
// 			return fmt.Errorf("tar contained invalid name error %q\n", target)
// 		}

// 		// add dst + re-format slashes according to system
// 		target = filepath.Join(dst, header.Name)
// 		// if no join is needed, replace with ToSlash:
// 		// target = filepath.ToSlash(header.Name)

// 		// check the type
// 		switch header.Typeflag {

// 		// if its a dir and it doesn't exist create it (with 0755 permission)
// 		case tar.TypeDir:
// 			if _, err := os.Stat(target); err != nil {
// 				if err := os.MkdirAll(target, 0755); err != nil {
// 					return err
// 				}
// 			}
// 		// if it's a file create it (with same permission)
// 		case tar.TypeReg:
// 			fileToWrite, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
// 			if err != nil {
// 				return err
// 			}
// 			// copy over contents
// 			if _, err := io.Copy(fileToWrite, tr); err != nil {
// 				return err
// 			}
// 			// manually close here after each file operation; defering would cause each file close
// 			// to wait until all operations have completed.
// 			fileToWrite.Close()
// 		}
// 	}
// 	return nil
// }

// func RecursiveZip(pathToZip, destinationPath string) error {
// 	destinationFile, err := os.Create(destinationPath)
// 	if err != nil {
// 		return err
// 	}
// 	var buf bytes.Buffer
// 	Compress("./folderToCompress", &buf)
// 	if _, err := io.Copy(destinationFile, &buf); err != nil {
// 		panic(err)
// 	}
// 	return nil
// 	// myZip := gzip.NewWriter(destinationFile)
// 	// err = filepath.Walk(pathToZip, func(filePath string, info os.FileInfo, err error) error {
// 	// 	if info.IsDir() {
// 	// 		return nil
// 	// 	}
// 	// 	if err != nil {
// 	// 		return err
// 	// 	}
// 	// 	relPath := strings.TrimPrefix(filePath, filepath.Dir(pathToZip))
// 	// 	zipFile, err := myZip.Create(relPath)
// 	// 	if err != nil {
// 	// 		return err
// 	// 	}
// 	// 	fsFile, err := os.Open(filePath)
// 	// 	if err != nil {
// 	// 		return err
// 	// 	}
// 	// 	_, err = io.Copy(zipFile, fsFile)
// 	// 	if err != nil {
// 	// 		return err
// 	// 	}
// 	// 	return nil
// 	// })
// 	// if err != nil {
// 	// 	return err
// 	// }
// 	// err = myZip.Close()
// 	// if err != nil {
// 	// 	return err
// 	// }
// 	// // defer os.Remove(pathToZip)
// 	// return nil
// }

// func Unzip(src, dest string) error {
// 	// r, err := os.Open(src)
// 	Uncompress(src, dest)
// 	// if err != nil {
// 	// 	return err
// 	// }
// 	// defer func() {
// 	// 	if err := r.Close(); err != nil {
// 	// 		panic(err)
// 	// 	}
// 	// }()

// 	// os.MkdirAll(dest, 0755)

// 	// // Closure to address file descriptors issue with all the deferred .Close() methods
// 	// extractAndWriteFile := func(f *zip.File) error {
// 	// 	rc, err := f.Open()
// 	// 	if err != nil {
// 	// 		return err
// 	// 	}
// 	// 	defer func() {
// 	// 		if err := rc.Close(); err != nil {
// 	// 			panic(err)
// 	// 		}
// 	// 	}()

// 	// 	path := filepath.Join(dest, f.Name)

// 	// 	// Check for ZipSlip (Directory traversal)
// 	// 	if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
// 	// 		return fmt.Errorf("illegal file path: %s", path)
// 	// 	}

// 	// 	if f.FileInfo().IsDir() {
// 	// 		os.MkdirAll(path, f.Mode())
// 	// 	} else {
// 	// 		os.MkdirAll(filepath.Dir(path), f.Mode())
// 	// 		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
// 	// 		if err != nil {
// 	// 			return err
// 	// 		}
// 	// 		defer func() {
// 	// 			if err := f.Close(); err != nil {
// 	// 				panic(err)
// 	// 			}
// 	// 		}()

// 	// 		_, err = io.Copy(f, rc)
// 	// 		if err != nil {
// 	// 			return err
// 	// 		}
// 	// 	}
// 	// 	return nil
// 	// }

// 	// for _, f := range r.File {
// 	// 	err := extractAndWriteFile(f)
// 	// 	if err != nil {
// 	// 		return err
// 	// 	}
// 	// }

// 	return nil
// }
