package sshx

import (
	"chihqiang/deply/utilx"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func OpenSftp(client *ssh.Client) (*sftp.Client, error) {
	return sftp.NewClient(client)
}

// IsSymlink determines whether the remote path is a symbolic link
func IsSymlink(sftpClient *sftp.Client, remotePath string) (bool, error) {
	// 1. Get file information
	fi, err := sftpClient.Lstat(remotePath)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}
	// 2. Check if it's a symbolic link
	return fi.Mode()&os.ModeSymlink != 0, nil
}

// RemoteExists determines whether the remote path exists
func RemoteExists(sftpClient *sftp.Client, remotePath string) bool {
	_, err := sftpClient.Lstat(remotePath)
	return err == nil
}

// ReadLink reads the actual path that the remote symbolic link points to
func ReadLink(sftpClient *sftp.Client, remotePath string) (string, error) {
	return sftpClient.ReadLink(remotePath)
}

// FileInfo encapsulates remote file information
type FileInfo struct {
	FileInfo os.FileInfo // Detailed file information
	Path     string      // Complete file path
	Name     string      // File name
	IsLink   bool        // Whether it's a soft link (current version)
}

// List lists all file information under the remote directory and marks the current version link
func List(sftpClient *sftp.Client, remotePath, linkPath string) ([]FileInfo, error) {
	// 2. Read the actual version directory that currentLink points to
	linkRemotePath, err := sftpClient.ReadLink(linkPath)
	if err != nil {
		return nil, err
	}
	// 3. Read all files under the remote version directory
	dirs, err := sftpClient.ReadDir(remotePath)
	if err != nil {
		return nil, err
	}
	// 4. Iterate through file list and encapsulate FileInfo
	var fis []FileInfo
	for _, dir := range dirs {
		fi := FileInfo{
			FileInfo: dir,
			Path:     path.Join(remotePath, dir.Name()),
			Name:     dir.Name(),
			IsLink:   path.Join(remotePath, dir.Name()) == linkRemotePath, // Current active version marker
		}
		fis = append(fis, fi)
	}
	return fis, nil
}

// Mkdir creates a directory on the remote server, creating it if it doesn't exist
func Mkdir(sftpClient *sftp.Client, remotePath string) error {
	//  Path safety check
	if filepath.IsAbs(remotePath) && !strings.HasPrefix(remotePath, "/") {
		return fmt.Errorf("unsupported path formats: %s", remotePath)
	}
	if strings.Contains(remotePath, "..") {
		return fmt.Errorf("the path contains illegal characters: %s", remotePath)
	}
	// Check if remote path exists
	stat, err := sftpClient.Lstat(remotePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Path doesn't exist, create directory recursively
			if err := sftpClient.MkdirAll(remotePath); err != nil {
				return fmt.Errorf("mkdir all %s failed: %w", remotePath, err)
			}
			return nil
		} else {
			// Other errors
			return fmt.Errorf("lstat %s failed: %w", remotePath, err)
		}
	}

	// Path exists but is not a directory
	if !stat.IsDir() {
		return fmt.Errorf("%s already exists and is not a directory", remotePath)
	}

	// Directory already exists, return directly
	return nil
}

// UploadFile uploads a local file to remote
func UploadFile(sftpClient *sftp.Client, localPath, remotePath string) error {

	// 2. Open local file
	srcFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("open local file: %w", err)
	}
	defer srcFile.Close()

	// 3. Get local file size for progress bar
	stat, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("stat local file: %w", err)
	}

	// 4. Create remote file
	dstFile, err := sftpClient.Create(remotePath)
	if err != nil {
		return fmt.Errorf("create remote file: %w", err)
	}
	defer dstFile.Close()

	// 5. Initialize progress bar
	bar := utilx.NewProgress(stat.Size(), "Uploading")

	// 6. Select buffer size based on file size
	const (
		smallFileThreshold = 1024 * 1024       // 1MB
		largeFileThreshold = 100 * 1024 * 1024 // 100MB
	)
	var bufSize int
	switch {
	case stat.Size() < smallFileThreshold:
		bufSize = 16 * 1024 // Small files use 16KB
	case stat.Size() > largeFileThreshold:
		bufSize = 256 * 1024 // Large files use 256KB
	default:
		bufSize = 64 * 1024 // Medium files use 64KB
	}
	buf := make([]byte, bufSize)

	// 7. Loop to read local file and write to remote
	for {
		n, err := srcFile.Read(buf)
		if n > 0 {
			written, writeErr := dstFile.Write(buf[:n])
			if writeErr != nil {
				return fmt.Errorf("write failed: %w", writeErr)
			}
			_ = bar.Add(written)
		}
		if err != nil {
			if err == io.EOF {
				break // File reading completed
			}
			return fmt.Errorf("read failed: %w", err)
		}
	}

	// Upload completed
	return nil
}
