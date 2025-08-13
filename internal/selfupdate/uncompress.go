package selfupdate

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func mkdirAllRoot(root *os.Root, dir string, mode os.FileMode) error {
	err := root.Mkdir(dir, mode)
	if err == nil{
		return nil
	}
	
	if os.IsExist(err){
		return nil
	}
	
	if os.IsNotExist(err){
		parentDir := filepath.Dir(dir)
		if parentDir != dir{
			if err := mkdirAllRoot(root, parentDir, 0755); err != nil {
				return err
			}
			err = root.Mkdir(dir, mode)
			if err == nil || os.IsExist(err){
				return nil
			}
		}
	}
	return err
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("open zip reader: %w", err)
	}
	defer func() {
		if err := r.Close(); err != nil {
			log.Printf("close zip reader: %v", err)
		}
	}()

	err = os.MkdirAll(dest, 0755)
	if err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	root, err := os.OpenRoot(dest)
	if err != nil {
		return fmt.Errorf("open root directory: %w", err)
	}
	defer root.Close()

	extractAndWriteFile := func(f *zip.File, root *os.Root) error {
		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("open zip file: %w", err)
		}
		defer func() {
			if err := rc.Close(); err != nil {
				log.Printf("close zip file: %v", err)
			}
		}()

		if f.FileInfo().IsDir() {
			err := root.Mkdir(f.Name, f.Mode())
			if err != nil {
				return fmt.Errorf("create directory: %w", err)
			}
		} else {
			file, err := root.OpenFile(f.Name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return fmt.Errorf("open file: %w", err)
			}
			defer func() {
				if err := file.Close(); err != nil {
					log.Printf("close file: %v", err)
				}
			}()

			const maxDecompressedSize int64 = 100 << 20 // 100 MB

			// Limit the size of the file to prevent zip bombs. G110 gosec
			limitedReader := &io.LimitedReader{R: rc, N: maxDecompressedSize}
			_, err = io.Copy(file, limitedReader)
			if err != nil {
				return fmt.Errorf("copy file: %w", err)
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f, root)
		if err != nil {
			return fmt.Errorf("extract and write file: %w", err)
		}
	}

	return nil
}

func untarGz(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("create gzip reader: %w", err)
	}
	defer gzReader.Close()

	root, err := os.OpenRoot(dest)
	if err != nil {
		return fmt.Errorf("open root directory: %w", err)
	}
	defer root.Close()

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar header: %w", err)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			err := root.Mkdir(header.Name, 0755)
			if err != nil {
				return fmt.Errorf("create directory: %w", err)
			}
		case tar.TypeReg:
			err := writeTarFile(tarReader, root, header.Name, os.FileMode(header.Mode)) // #nosec G115
			if err != nil {
				return fmt.Errorf("write file: %w", err)
			}
		}
	}
	return nil
}

func writeTarFile(tarReader *tar.Reader, root *os.Root, name string, mode os.FileMode) error {
	if mode > 0o777 {
		return fmt.Errorf("invalid file mode: %d", mode)
	}

	dir := filepath.Clean(filepath.Dir(name))
	if dir != "." && dir != "" {
		if err := mkdirAllRoot(root, dir, 0755); err != nil {
			return fmt.Errorf("create directory: %w", err)
		}
	}

	outFile, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer func() {
		if err := outFile.Close(); err != nil {
			fmt.Printf("close file: %v\n", err)
		}
	}()

	_, err = io.Copy(outFile, tarReader)
	return err
}

func unarchive(src, dest string) error {
	switch {
	case strings.HasSuffix(src, ".zip"):
		return unzip(src, dest)
	case strings.HasSuffix(src, ".tar.gz"), strings.HasSuffix(src, ".tgz"):
		return untarGz(src, dest)
	default:
		return fmt.Errorf("unsupported archive format: %s", src)
	}
}
