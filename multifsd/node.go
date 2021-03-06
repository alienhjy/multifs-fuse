package main

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"syscall"

	"bazil.org/fuse"
	"golang.org/x/net/context"
)

// Node means File or Dir or other type of file
type Node struct {
	Path string
}

// GetFullPath return the real path of File in backend.
func (nd *Node) getFullPath() (string, error) {
	// check if this node is marked as deleted
	if nd.checkDeleted("") != nil {
		return "", os.ErrNotExist
	}

	fullPath := filepath.Join(fusefs.master, nd.Path)
	_, err := os.Lstat(fullPath)
	if err != nil {
	GetSlaves:
		for _, slave := range fusefs.slaves {
			fullPath = filepath.Join(slave, nd.Path)
			_, err = os.Lstat(fullPath)
			if err != nil {
				continue GetSlaves
			}
			break GetSlaves
		}
		if err != nil {
			fullPath = ""
		}
	}

	return fullPath, err
}

// CheckDeleted check if the file/dir is mark as deleted.
func (nd *Node) checkDeleted(name string) error {
	log.Println("Dir.checkDeleted: ", filepath.Join(nd.Path, name))
	var (
		fullCheckPath string
	)
	fullCheckPath = filepath.Join(fusefs.master, nd.Path, name)
	fInfo, err := os.Lstat(fullCheckPath)
	if err != nil {
		return err
	}
	if (fInfo.Mode() & os.ModeSymlink) == os.ModeSymlink {
		linkName, err := os.Readlink(fullCheckPath)
		if err == nil && linkName == deletedMark {
			return os.ErrNotExist
		}
	}
	return nil
}

// Attr for get attr of Node
func (nd *Node) Attr(ctx context.Context, a *fuse.Attr) error {
	log.Println("Node.Attr:", nd.Path)
	fInfo, err := os.Lstat(filepath.Join(fusefs.master, nd.Path))
	if err != nil {
	GetSlaves:
		for _, slave := range fusefs.slaves {
			fInfo, err = os.Lstat(filepath.Join(slave, nd.Path))
			if err != nil {
				continue GetSlaves
			}
			break GetSlaves
		}
	}

	// Get file attr from backend
	sysStat, ok := fInfo.Sys().(*syscall.Stat_t)
	if !ok {
		return errors.New("FileInfo.sys() is not a syscall.Stat_t")
	}

	// Fill attr
	a.Inode = sysStat.Ino
	a.Size = (uint64)(sysStat.Size)
	a.Blocks = (uint64)(sysStat.Blocks)
	// Atime == Mtime == Ctime
	a.Atime = fInfo.ModTime()
	a.Mtime = fInfo.ModTime()
	a.Ctime = fInfo.ModTime()
	if fusefs.readOnly {
		a.Mode = fInfo.Mode() & 037777777555
	} else {
		a.Mode = fInfo.Mode()
	}
	a.Nlink = (uint32)(sysStat.Nlink)
	a.Uid = sysStat.Uid
	a.Gid = sysStat.Gid
	a.Rdev = (uint32)(sysStat.Rdev)
	a.BlockSize = (uint32)(sysStat.Blksize)

	return nil
}

// Access checks wheather operation has permission
func (nd *Node) Access(ctx context.Context, req *fuse.AccessRequest) error {
	// TODO: check permission
	return nil
}

// Readlink reads a symbolic link
func (nd *Node) Readlink(ctx context.Context, req *fuse.ReadlinkRequest) (string, error) {
	fullPath, err := nd.getFullPath()
	if err != nil {
		return "", err
	}
	return os.Readlink(fullPath)
}
