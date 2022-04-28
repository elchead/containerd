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
	"fmt"
	gzip "github.com/klauspost/pgzip"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

var (
	suffixes [5]string
)

func round(val float64, roundOn float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	newVal = round / pow
	return
}

func humanFileSize(size float64) string {
	suffixes[0] = "B"
	suffixes[1] = "KB"
	suffixes[2] = "MB"
	suffixes[3] = "GB"
	suffixes[4] = "TB"

	base := math.Log(size) / math.Log(1024)
	getSize := round(math.Pow(1024, base-math.Floor(base)), .5, 2)
	getSuffix := suffixes[int(math.Floor(base))]
	return strconv.FormatFloat(getSize, 'f', -1, 64) + string(getSuffix)
}

func createArchiveSizeFile(zipPath string) {
	i, _ := os.Stat(zipPath)
	sz := float64(i.Size())
	sizeFile := filepath.Join(filepath.Dir(filepath.Dir(zipPath)), filepath.Base(filepath.Dir(zipPath))+"_"+humanFileSize(sz))
	os.Create(sizeFile)
}

func RecursiveZip(pathToZip, zipPath string) error {
	fmt.Println("Creating zip..")
	os.MkdirAll(filepath.Base(zipPath), os.ModePerm)
	fs, err := ioutil.ReadDir(pathToZip)
	if err != nil {
		return fmt.Errorf("error reading directory: %v", err)
	}
	var files []string
	for _, f := range fs {
		files = append(files, filepath.Join(pathToZip, f.Name()))
	}
	out, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("error writing archive: %v", err)
	}
	defer out.Close()

	// Create the archive and write the output to the "out" Writer
	err = createArchive(files, out)
	if err != nil {
		return fmt.Errorf("failed to create archive: %v", err)
	}
	createArchiveSizeFile(zipPath)
	fmt.Println("Created archive")
	return nil
}

func ExtractTarGz(src, dest string) {
	r, err := os.Open(src)
	if err != nil {
		log.Fatalf("could not open zip file: %v", err)
	}
	copy, err := os.Create("/mnt/migration/check.tar.gz")
	if err != nil {
		log.Fatalf("could not copy zip file: %v", err)
	}
	fmt.Println("Start copy gz")
	io.Copy(copy, r)
	fmt.Println("Finish copy gz")
	uncompressedStream, err := gzip.NewReader(copy)
	if err != nil {
		log.Fatal("ExtractTarGz: NewReader failed")
	}

	tarReader := tar.NewReader(uncompressedStream)
	fmt.Println("Start untar gz")
	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("ExtractTarGz: Next() failed: %s", err.Error())
		}

		switch header.Typeflag {
		case tar.TypeDir:
			path := filepath.Join(dest, header.Name)
			if err := os.Mkdir(path, 0755); err != nil {
				log.Fatalf("ExtractTarGz: Mkdir() failed: %s", err.Error())
			}
		case tar.TypeReg:
			path := filepath.Join(dest, header.Name)
			outFile, err := os.Create(path)
			if err != nil {
				log.Fatalf("ExtractTarGz: Create() failed: %s", err.Error())
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				log.Fatalf("ExtractTarGz: Copy() failed: %s", err.Error())
			}
			outFile.Close()

		default:
			log.Fatalf(
				"ExtractTarGz: uknown type: %s in %s",
				header.Typeflag,
				header.Name)
		}

	}
	fmt.Println("Finish untar gz")
}

func Unzip(src, dest string) error {
	// cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("mkdir -p %s", dest))
	os.MkdirAll(dest, os.ModePerm)
	ExtractTarGz(src, dest)
	// cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("mkdir -p %s && tar -xf %s -C %s", dest, src, dest))
	// fmt.Printf("mkdir -p %s && tar -xf %s -C %s\n", dest, src, dest)
	return nil //cmd.Run()
}

func DeleteAllFiles(dir string) error {
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("rm -r %s", filepath.Join(dir, "*")))
	fmt.Printf("rm -r %s", filepath.Join(dir, "*"))
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
