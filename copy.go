package main

import (
	"fmt"
	"io"
	"os"
	"path"

	"github.com/pkg/sftp"
)

func copyLocalDirToRemote(src string, dst string, sftpClient *sftp.Client) error {
	var err error
	var fds []os.FileInfo

	if err = sftpClient.MkdirAll(dst); err != nil {
		return err
	}

	if fds, err = sftpClient.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = copyLocalDirToRemote(srcfp, dstfp, sftpClient); err != nil {
				fmt.Println(err)
			}
		} else {
			if err = copyRemoteFileToLocal(srcfp, dstfp, sftpClient); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}

// copy dir from remote host
func copyRemoteDirToLocal(src string, dst string, sftpClient *sftp.Client) error {
	var err error
	var fds []os.FileInfo
	var srcinfo os.FileInfo

	if srcinfo, err = sftpClient.Stat(src); err != nil {
		return err
	}

	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}

	if fds, err = sftpClient.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = copyRemoteDirToLocal(srcfp, dstfp, sftpClient); err != nil {
				fmt.Println(err)
			}
		} else {
			if err = copyRemoteFileToLocal(srcfp, dstfp, sftpClient); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}

func copyLocalFileToRemote(src, dst string, sftpClient *sftp.Client) error {
	var err error
	var srcfd *os.File
	var dstfd *sftp.File
	var srcinfo os.FileInfo

	if srcfd, err = os.Open(src); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = sftpClient.Create(dst); err != nil {
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	return sftpClient.Chmod(dst, srcinfo.Mode())
}

// copy file from remote
func copyRemoteFileToLocal(src, dst string, sftpClient *sftp.Client) error {
	var err error
	var srcfd *sftp.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if srcfd, err = sftpClient.Open(src); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.Create(dst); err != nil {
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if srcinfo, err = sftpClient.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}
