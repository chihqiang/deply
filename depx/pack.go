package depx

import (
	"archive/tar"
	"chihqiang/deply/utilx"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// PackDir compresses directory dir to tar.gz with a beautiful progress bar
func PackDir(config *Config) (string, error) {
	if err := config.Validate(); err != nil {
		return "", err
	}
	tarName := fmt.Sprintf("%s.tar.gz", config.Version)
	tarPath := filepath.Join(os.TempDir(), tarName)
	file, err := os.Create(tarPath)
	if err != nil {
		return "", fmt.Errorf("create tar.gz file failed: %w", err)
	}
	defer file.Close()
	gw := gzip.NewWriter(file)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()
	// 1. Collect files to be packed and calculate total size
	var totalSize int64
	var files []string
	err = filepath.Walk(config.Dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, _ := filepath.Rel(config.Dir, path)
		if relPath == "." {
			return nil
		}
		// Include / Exclude
		if len(config.Include) > 0 {
			found := false
			for _, p := range config.Include {
				if strings.HasPrefix(relPath, p) {
					found = true
					break
				}
			}
			if !found {
				return nil
			}
		}
		if len(config.Exclude) > 0 {
			for _, p := range config.Exclude {
				if strings.HasPrefix(relPath, p) {
					return nil
				}
			}
		}

		if info.Mode().IsRegular() {
			totalSize += info.Size()
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("walk directory failed: %w", err)
	}

	// 2. Create progress bar
	bar := utilx.NewProgress(totalSize, "Packing")
	// 3. Write files and update progress bar
	var written int64
	for _, filename := range files {
		info, err := os.Stat(filename)
		if err != nil {
			return "", err
		}
		relPath, _ := filepath.Rel(config.Dir, filename)
		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return "", err
		}
		header.Name = relPath
		if err := tw.WriteHeader(header); err != nil {
			return "", err
		}
		if info.Mode().IsRegular() {
			f, err := os.Open(filename)
			if err != nil {
				return "", err
			}
			buf := make([]byte, 32*1024)
			for {
				n, err := f.Read(buf)
				if n > 0 {
					if _, errw := tw.Write(buf[:n]); errw != nil {
						return "", errw
					}
					written += int64(n)
					_ = bar.Set64(written)
				}
				if err != nil {
					if err == io.EOF {
						break
					}
					return "", err
				}
			}
			f.Close()
		}
	}
	return tarPath, nil
}
