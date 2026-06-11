package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"roub-crt/internal/file_transfer"
	"roub-crt/internal/connection"
)

var transferSession string
var transferLocalPath string
var transferRemotePath string

func init() {
	rootCmd.AddCommand(transferCmd)

	transferCmd.Flags().StringVarP(&transferSession, "session", "s", "", "Session ID to use for file transfer")
	transferCmd.Flags().StringVarP(&transferLocalPath, "local", "l", ".", "Local path")
	transferCmd.Flags().StringVarP(&transferRemotePath, "remote", "r", ".", "Remote path")
}

var transferCmd = &cobra.Command{
	Use:   "transfer",
	Short: "File transfer with SFTP/SCP",
	Long:  `Transfer files between local and remote systems using SFTP or SCP protocol.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("File transfer functionality")
		fmt.Println("Use 'roub-crt interactive' for the full dual-pane file manager interface")
	},
}

func NewSFTPTransfer(sshConn *connection.SSHConn) (*file_transfer.SFTPClient, error) {
	return file_transfer.NewSFTPClient(sshConn.Client)
}

func ListLocalFiles(path string, showHidden bool) (*file_transfer.FileList, error) {
	fs := file_transfer.NewLocalFileSystem()
	return fs.ListDir(path)
}

func ListRemoteFiles(sftp *file_transfer.SFTPClient, path string) (*file_transfer.FileList, error) {
	return sftp.ListDir(path)
}

func DownloadFile(sftp *file_transfer.SFTPClient, remotePath, localPath string) error {
	progress := make(chan file_transfer.TransferProgress, 1)

	go func() {
		for p := range progress {
			percent := float64(p.TransferredBytes) / float64(p.TotalBytes) * 100
			fmt.Printf("\rDownloading: %.1f%% (%d / %d bytes)",
				percent, p.TransferredBytes, p.TotalBytes)
		}
	}()

	err := sftp.DownloadFile(remotePath, localPath, progress)
	close(progress)
	if err != nil {
		return err
	}

	fmt.Println()
	return nil
}

func UploadFile(sftp *file_transfer.SFTPClient, localPath, remotePath string) error {
	progress := make(chan file_transfer.TransferProgress, 1)

	go func() {
		for p := range progress {
			percent := float64(p.TransferredBytes) / float64(p.TotalBytes) * 100
			fmt.Printf("\rUploading: %.1f%% (%d / %d bytes)",
				percent, p.TransferredBytes, p.TotalBytes)
		}
	}()

	err := sftp.UploadFile(localPath, remotePath, progress)
	close(progress)
	if err != nil {
		return err
	}

	fmt.Println()
	return nil
}

func MakeRemoteDir(sftp *file_transfer.SFTPClient, path string) error {
	return sftp.MakeDir(path)
}

func MakeLocalDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func DeleteRemoteFile(sftp *file_transfer.SFTPClient, path string) error {
	return sftp.DeleteFile(path)
}

func DeleteLocalFile(path string) error {
	return os.Remove(path)
}

func RenameRemoteFile(sftp *file_transfer.SFTPClient, oldPath, newPath string) error {
	return sftp.Rename(oldPath, newPath)
}

func RenameLocalFile(oldPath, newPath string) error {
	return os.Rename(oldPath, newPath)
}
