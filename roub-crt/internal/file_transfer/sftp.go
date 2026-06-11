package file_transfer

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

type FileInfo struct {
	Name      string
	Size      int64
	Mode      os.FileMode
	ModTime   time.Time
	IsDir     bool
	IsLink    bool
	LinkTarget string
	Owner     string
	Group     string
	Perms     string
}

type FileList struct {
	Path     string
	Files    []*FileInfo
	TotalSize int64
}

type TransferProgress struct {
	TotalBytes      int64
	TransferredBytes int64
	FileName        string
	Direction       string
}

type SFTPClient struct {
	client *ssh.Client
}

func NewSFTPClient(client *ssh.Client) (*SFTPClient, error) {
	return &SFTPClient{client: client}, nil
}

func (s *SFTPClient) ListDir(path string) (*FileList, error) {
	session, err := s.client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	var stdout bytes.Buffer
	session.Stdout = &stdout

	err = session.Run("ls -la " + path)
	if err != nil {
		return nil, fmt.Errorf("failed to list directory: %w", err)
	}

	fileList := &FileList{
		Path:  path,
		Files: make([]*FileInfo, 0),
	}

	lines := strings.Split(stdout.String(), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 8 {
			continue
		}

		if fields[0] == "total" {
			continue
		}

		info := &FileInfo{}

		perms := fields[0]
		info.Perms = perms
		info.IsDir = strings.HasPrefix(perms, "d")
		info.IsLink = strings.HasPrefix(perms, "l")

		info.Name = fields[8]

		if info.IsDir || info.Name == "." || info.Name == ".." {
			continue
		}

		info.Size = 0
		for _, c := range fields[4] {
			if c >= '0' && c <= '9' {
				info.Size = info.Size*10 + int64(c-'0')
			}
		}

		if len(fields) >= 9 {
			dateStr := strings.Join(fields[5:8], " ")
			info.ModTime, _ = time.Parse("Jan 02 2006", dateStr)
		}

		fileList.Files = append(fileList.Files, info)
		fileList.TotalSize += info.Size
	}

	return fileList, nil
}

func (s *SFTPClient) DownloadFile(remotePath, localPath string, progress chan<- TransferProgress) error {
	session, err := s.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	session.Stdout, _ = os.Create(localPath)
	session.Stderr = os.Stderr

	go func() {
		progress <- TransferProgress{
			TotalBytes:      0,
			TransferredBytes: 0,
			FileName:        remotePath,
			Direction:       "downloading",
		}
	}()

	err = session.Run("cat " + remotePath)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	stat, _ := os.Stat(localPath)
	progress <- TransferProgress{
		TotalBytes:      stat.Size(),
		TransferredBytes: stat.Size(),
		FileName:        remotePath,
		Direction:       "downloading",
	}

	return nil
}

func (s *SFTPClient) UploadFile(localPath, remotePath string, progress chan<- TransferProgress) error {
	session, err := s.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	session.Stdin, _ = os.Open(localPath)
	session.Stderr = os.Stderr

	stat, _ := os.Stat(localPath)
	go func() {
		progress <- TransferProgress{
			TotalBytes:      stat.Size(),
			TransferredBytes: 0,
			FileName:        localPath,
			Direction:       "uploading",
		}
	}()

	err = session.Run("cat > " + remotePath)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	progress <- TransferProgress{
		TotalBytes:      stat.Size(),
		TransferredBytes: stat.Size(),
		FileName:        localPath,
		Direction:       "uploading",
	}

	return nil
}

func (s *SFTPClient) DeleteFile(path string) error {
	session, err := s.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	return session.Run("rm " + path)
}

func (s *SFTPClient) DeleteDir(path string) error {
	session, err := s.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	return session.Run("rm -r " + path)
}

func (s *SFTPClient) Rename(oldPath, newPath string) error {
	session, err := s.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	return session.Run("mv " + oldPath + " " + newPath)
}

func (s *SFTPClient) MakeDir(path string) error {
	session, err := s.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	return session.Run("mkdir -p " + path)
}

func (s *SFTPClient) Stat(path string) (*FileInfo, error) {
	session, err := s.client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	var stdout bytes.Buffer
	session.Stdout = &stdout

	err = session.Run("stat " + path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat: %w", err)
	}

	return &FileInfo{
		Name: filepath.Base(path),
	}, nil
}

func (s *SFTPClient) Close() error {
	return nil
}

func (s *SFTPClient) WalkDir(path string, showHidden bool) ([]string, error) {
	var dirs []string

	fileList, err := s.ListDir(path)
	if err != nil {
		return nil, err
	}

	for _, entry := range fileList.Files {
		if !showHidden && strings.HasPrefix(entry.Name, ".") {
			continue
		}

		fullPath := filepath.Join(path, entry.Name)
		dirs = append(dirs, fullPath)

		if entry.IsDir {
			subDirs, err := s.WalkDir(fullPath, showHidden)
			if err == nil {
				dirs = append(dirs, subDirs...)
			}
		}
	}

	return dirs, nil
}

type LocalFileSystem struct{}

func NewLocalFileSystem() *LocalFileSystem {
	return &LocalFileSystem{}
}

func (l *LocalFileSystem) ListDir(path string) (*FileList, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	fileList := &FileList{
		Path:  path,
		Files: make([]*FileInfo, 0),
	}

	for _, entry := range entries {
		stat, err := entry.Info()
		if err != nil {
			continue
		}

		info := &FileInfo{
			Name:    entry.Name(),
			Size:    stat.Size(),
			Mode:    stat.Mode(),
			ModTime: stat.ModTime(),
			IsDir:   entry.IsDir(),
			IsLink:  entry.Type()&os.ModeSymlink != 0,
			Perms:   formatPerms(stat.Mode()),
		}

		if info.IsLink {
			target, err := os.Readlink(filepath.Join(path, entry.Name()))
			if err == nil {
				info.LinkTarget = target
			}
		}

		fileList.Files = append(fileList.Files, info)
		fileList.TotalSize += info.Size
	}

	return fileList, nil
}

func (l *LocalFileSystem) DeleteFile(path string) error {
	return os.Remove(path)
}

func (l *LocalFileSystem) DeleteDir(path string) error {
	return os.RemoveAll(path)
}

func (l *LocalFileSystem) Rename(oldPath, newPath string) error {
	return os.Rename(oldPath, newPath)
}

func (l *LocalFileSystem) MakeDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func (l *LocalFileSystem) Stat(path string) (*FileInfo, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	return &FileInfo{
		Name:    filepath.Base(path),
		Size:    stat.Size(),
		Mode:    stat.Mode(),
		ModTime: stat.ModTime(),
		IsDir:   stat.IsDir(),
		Perms:   formatPerms(stat.Mode()),
	}, nil
}

func (l *LocalFileSystem) WalkDir(path string, showHidden bool) ([]string, error) {
	var dirs []string

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !showHidden && strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		fullPath := filepath.Join(path, entry.Name())
		dirs = append(dirs, fullPath)

		if entry.IsDir() {
			subDirs, err := l.WalkDir(fullPath, showHidden)
			if err == nil {
				dirs = append(dirs, subDirs...)
			}
		}
	}

	return dirs, nil
}

type SortType int

const (
	SortByName SortType = iota
	SortBySize
	SortByDate
	SortByType
)

func SortFileList(fileList *FileList, sortType SortType, descending bool) {
	switch sortType {
	case SortByName:
		sort.Slice(fileList.Files, func(i, j int) bool {
			if descending {
				return fileList.Files[i].Name > fileList.Files[j].Name
			}
			return fileList.Files[i].Name < fileList.Files[j].Name
		})
	case SortBySize:
		sort.Slice(fileList.Files, func(i, j int) bool {
			if descending {
				return fileList.Files[i].Size > fileList.Files[j].Size
			}
			return fileList.Files[i].Size < fileList.Files[j].Size
		})
	case SortByDate:
		sort.Slice(fileList.Files, func(i, j int) bool {
			if descending {
				return fileList.Files[i].ModTime.After(fileList.Files[j].ModTime)
			}
			return fileList.Files[i].ModTime.Before(fileList.Files[j].ModTime)
		})
	case SortByType:
		sort.Slice(fileList.Files, func(i, j int) bool {
			if descending {
				if fileList.Files[i].IsDir != fileList.Files[j].IsDir {
					return fileList.Files[i].IsDir
				}
				return fileList.Files[i].Name > fileList.Files[j].Name
			}
			if fileList.Files[i].IsDir != fileList.Files[j].IsDir {
				return fileList.Files[j].IsDir
			}
			return fileList.Files[i].Name < fileList.Files[j].Name
		})
	}
}

func formatPerms(mode os.FileMode) string {
	var perms [10]rune

	perms[0] = '-'

	if mode&os.ModeDir != 0 {
		perms[0] = 'd'
	} else if mode&os.ModeSymlink != 0 {
		perms[0] = 'l'
	}

	if mode&0400 != 0 {
		perms[1] = 'r'
	} else {
		perms[1] = '-'
	}
	if mode&0200 != 0 {
		perms[2] = 'w'
	} else {
		perms[2] = '-'
	}
	if mode&0100 != 0 {
		if mode&04000 != 0 {
			perms[3] = 's'
		} else {
			perms[3] = 'x'
		}
	} else {
		if mode&04000 != 0 {
			perms[3] = 'S'
		} else {
			perms[3] = '-'
		}
	}

	if mode&0040 != 0 {
		perms[4] = 'r'
	} else {
		perms[4] = '-'
	}
	if mode&0020 != 0 {
		perms[5] = 'w'
	} else {
		perms[5] = '-'
	}
	if mode&0010 != 0 {
		if mode&02000 != 0 {
			perms[6] = 's'
		} else {
			perms[6] = 'x'
		}
	} else {
		if mode&02000 != 0 {
			perms[6] = 'S'
		} else {
			perms[6] = '-'
		}
	}

	if mode&0004 != 0 {
		perms[7] = 'r'
	} else {
		perms[7] = '-'
	}
	if mode&0002 != 0 {
		perms[8] = 'w'
	} else {
		perms[8] = '-'
	}
	if mode&0001 != 0 {
		if mode&04000 != 0 {
			perms[9] = 't'
		} else {
			perms[9] = 'x'
		}
	} else {
		if mode&04000 != 0 {
			perms[9] = 'T'
		} else {
			perms[9] = '-'
		}
	}

	return string(perms[:])
}

func FormatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
