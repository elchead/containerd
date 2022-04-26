package util

import (
	// "archive/zip"
	"archive/tar"
	"fmt"
	gzip "github.com/klauspost/pgzip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func compress(src string, buf io.Writer) error {
	// tar > gzip > buf
	zr := gzip.NewWriter(buf)
	tw := tar.NewWriter(zr)

	// walk through every file in the folder
	filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {
		// generate tar header
		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return err
		}

		// must provide real name
		// (see https://golang.org/src/archive/tar/common.go?#L626)
		header.Name = filepath.ToSlash(file)

		// write header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		// if not a dir, write file content
		if !fi.IsDir() {
			data, err := os.Open(file)
			if err != nil {
				return err
			}
			if _, err := io.Copy(tw, data); err != nil {
				return err
			}
		}
		return nil
	})

	// produce tar
	if err := tw.Close(); err != nil {
		return err
	}
	// produce gzip
	if err := zr.Close(); err != nil {
		return err
	}
	//
	return nil
}

// check for path traversal and correct forward slashes
func validRelPath(p string) bool {
	if p == "" || strings.Contains(p, `\`) || strings.HasPrefix(p, "/") || strings.Contains(p, "../") {
		return false
	}
	return true
}

func uncompress(src io.Reader, dst string) error {
	// ungzip
	zr, err := gzip.NewReader(src)
	if err != nil {
		return err
	}
	// untar
	tr := tar.NewReader(zr)

	// uncompress each element
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return err
		}
		target := 

		// validate name against path traversal
		if !validRelPath(header.Name) {
			return fmt.Errorf("tar contained invalid name error %q\n", target)
		}

		// add dst + re-format slashes according to system
		target = filepath.Join(dst, header.Name)
		// if no join is needed, replace with ToSlash:
		// target = filepath.ToSlash(header.Name)

		// check the type
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it (with 0755 permission)
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}
		// if it's a file create it (with same permission)
		case tar.TypeReg:
			fileToWrite, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			// copy over contents
			if _, err := io.Copy(fileToWrite, tr); err != nil {
				return err
			}
			// manually close here after each file operation; defering would cause each file close
			// to wait until all operations have completed.
			fileToWrite.Close()
		}
	}
}

func RecursiveZip(pathToZip, destinationPath string) error {
	destinationFile, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	_ := compress("./folderToCompress", &buf)
	if _, err := io.Copy(fileToWrite, &buf); err != nil {
		panic(err)
	}
	// myZip := gzip.NewWriter(destinationFile)
	// err = filepath.Walk(pathToZip, func(filePath string, info os.FileInfo, err error) error {
	// 	if info.IsDir() {
	// 		return nil
	// 	}
	// 	if err != nil {
	// 		return err
	// 	}
	// 	relPath := strings.TrimPrefix(filePath, filepath.Dir(pathToZip))
	// 	zipFile, err := myZip.Create(relPath)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	fsFile, err := os.Open(filePath)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	_, err = io.Copy(zipFile, fsFile)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	return nil
	// })
	// if err != nil {
	// 	return err
	// }
	// err = myZip.Close()
	// if err != nil {
	// 	return err
	// }
	// // defer os.Remove(pathToZip)
	// return nil
}

func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	uncompress(&r, dest)
	// if err != nil {
	// 	return err
	// }
	// defer func() {
	// 	if err := r.Close(); err != nil {
	// 		panic(err)
	// 	}
	// }()

	// os.MkdirAll(dest, 0755)

	// // Closure to address file descriptors issue with all the deferred .Close() methods
	// extractAndWriteFile := func(f *zip.File) error {
	// 	rc, err := f.Open()
	// 	if err != nil {
	// 		return err
	// 	}
	// 	defer func() {
	// 		if err := rc.Close(); err != nil {
	// 			panic(err)
	// 		}
	// 	}()

	// 	path := filepath.Join(dest, f.Name)

	// 	// Check for ZipSlip (Directory traversal)
	// 	if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
	// 		return fmt.Errorf("illegal file path: %s", path)
	// 	}

	// 	if f.FileInfo().IsDir() {
	// 		os.MkdirAll(path, f.Mode())
	// 	} else {
	// 		os.MkdirAll(filepath.Dir(path), f.Mode())
	// 		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	// 		if err != nil {
	// 			return err
	// 		}
	// 		defer func() {
	// 			if err := f.Close(); err != nil {
	// 				panic(err)
	// 			}
	// 		}()

	// 		_, err = io.Copy(f, rc)
	// 		if err != nil {
	// 			return err
	// 		}
	// 	}
	// 	return nil
	// }

	// for _, f := range r.File {
	// 	err := extractAndWriteFile(f)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	// return nil
}
