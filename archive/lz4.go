package archive

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pierrec/lz4"
)

func Compress(src string, archivePath string) error {
	// ensure the src actually exists before trying to tar it
	if _, err := os.Stat(src); err != nil {
		return err
	}

	// Create dir for output
	os.MkdirAll(filepath.Dir(archivePath), 0766)

	// Create file
	archive, err := os.OpenFile(archivePath+".tar.lz4", os.O_CREATE|os.O_RDWR, os.FileMode(0600))
	if err != nil {
		return err
	}
	defer archive.Close()

	// tar > lz4 > file
	lz4Writer := lz4.NewWriter(archive)
	defer lz4Writer.Close()
	tarWriter := tar.NewWriter(lz4Writer)
	defer tarWriter.Close()

	// Add dirs
	return filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !fi.Mode().IsRegular() {
			return nil
		}

		// generate tar header
		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		header.Name = filepath.Clean(strings.TrimPrefix(file, src))

		// write header
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// open file
		f, err := os.Open(file)
		if err != nil {
			return err
		}

		// copy file data
		if _, err := io.Copy(tarWriter, f); err != nil {
			return err
		}

		// manually close here after each file operation; defering would cause each file close
		// to wait until all operations have completed.
		f.Close()
		return nil
	})
}

func Uncompress(file io.Reader, dir string) error {
	lz4Reader := lz4.NewReader(file)
	tarReader := tar.NewReader(lz4Reader)

	for {
		header, err := tarReader.Next()

		switch {

		// if no more files are found return
		case err == io.EOF:
			return nil

		// return any other error
		case err != nil:
			return err

		// if the header is nil, just skip it
		case header == nil:
			continue
		}

		// the target location where the dir/file should be created
		target := filepath.Join(dir, header.Name)

		// check the file type
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}

		// if it's a file create it
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			// copy over contents
			if _, err := io.Copy(f, tarReader); err != nil {
				return err
			}

			// manually close here after each file operation; defering would cause each file close
			// to wait until all operations have completed.
			f.Close()
		}
	}
}
