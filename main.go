package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"mount/filemodel"
	"os"

	"github.com/jacobsa/fuse"
	"github.com/jacobsa/fuse/fuseops"
	"github.com/jacobsa/fuse/fuseutil"
	"github.com/jacobsa/timeutil"
)

var fType = flag.String("type", "", "The name of the samples/ sub-dir.")
var fMountPoint = flag.String("mount_point", "", "Path to mount point.")
var fReadyFile = flag.Uint64("ready_file", 0, "FD to signal when ready.")

var fReadOnly = flag.Bool("read_only", false, "Mount in read-only mode.")
var fDebug = flag.Bool("debug", false, "Enable debug logging.")

func NewSQLFS(clock timeutil.Clock) (fuse.Server, error) {
	fs := &sqlFS{
		Clock: clock,
	}

	return fuseutil.NewFileSystemServer(fs), nil
}

type sqlFS struct {
	fuseutil.NotImplementedFileSystem

	Clock timeutil.Clock
}

const (
	rootInode fuseops.InodeID = fuseops.RootInodeID + iota
	helloInode
	dirInode
	worldInode
)

// We have a fixed directory structure.
var gInodeInfo map[fuseops.InodeID]filemodel.InodeInfo

func findChildInode(
	name string,
	children []fuseutil.Dirent) (fuseops.InodeID, error) {
	for _, e := range children {
		if e.Name == name {
			return e.Inode, nil
		}
	}

	return 0, fuse.ENOENT
}

func (fs *sqlFS) patchAttributes(
	attr *fuseops.InodeAttributes) {
	now := fs.Clock.Now()
	attr.Atime = now
	attr.Mtime = now
	attr.Crtime = now
}

func (fs *sqlFS) StatFS(
	ctx context.Context,
	op *fuseops.StatFSOp) error {
	return nil
}

func (fs *sqlFS) LookUpInode(
	ctx context.Context,
	op *fuseops.LookUpInodeOp) error {
	// Find the info for the parent.
	parentInfo, ok := gInodeInfo[op.Parent]
	if !ok {
		return fuse.ENOENT
	}

	// Find the child within the parent.
	childInode, err := findChildInode(op.Name, parentInfo.Children)
	if err != nil {
		return err
	}

	// Copy over information.
	op.Entry.Child = childInode
	op.Entry.Attributes = gInodeInfo[childInode].Attributes

	// Patch attributes.
	fs.patchAttributes(&op.Entry.Attributes)

	return nil
}

func (fs *sqlFS) GetInodeAttributes(
	ctx context.Context,
	op *fuseops.GetInodeAttributesOp) error {
	// Find the info for this inode.
	info, ok := gInodeInfo[op.Inode]
	if !ok {
		return fuse.ENOENT
	}

	// Copy over its attributes.
	op.Attributes = info.Attributes

	// Patch attributes.
	fs.patchAttributes(&op.Attributes)

	return nil
}

func (fs *sqlFS) OpenDir(
	ctx context.Context,
	op *fuseops.OpenDirOp) error {
	// Allow opening any directory.
	return nil
}

func (fs *sqlFS) ReadDir(
	ctx context.Context,
	op *fuseops.ReadDirOp) error {
	// Find the info for this inode.
	log.Println(op.Inode)
	info, ok := gInodeInfo[op.Inode]
	if !ok {
		return fuse.ENOENT
	}

	if !info.Dir {
		return fuse.EIO
	}
	//log.Println(op.Inode)

	entries := info.Children

	// Grab the range of interest.
	if op.Offset > fuseops.DirOffset(len(entries)) {
		return fuse.EIO
	}

	entries = entries[op.Offset:]
	//log.Println(op.Inode)

	// Resume at the specified offset into the array.
	for _, e := range entries {
		n := fuseutil.WriteDirent(op.Dst[op.BytesRead:], e)
		if n == 0 {
			break
		}

		op.BytesRead += n
	}

	return nil
}

func (fs *sqlFS) OpenFile(
	ctx context.Context,
	op *fuseops.OpenFileOp) error {
	// Allow opening any file.
	return nil
}

func (fs *sqlFS) ReadFile(
	ctx context.Context,
	op *fuseops.ReadFileOp) error {
	// Let io.ReaderAt deal with the semantics.
	node, ok := gInodeInfo[op.Inode]
	if !ok {
		return fmt.Errorf("不存在的InodeID")
	}
	model, err := filemodel.GetModelByID(filemodel.GetDB(), node.RowID)
	if err != nil {
		return fmt.Errorf("读取数据错误")
	}

	reader := bytes.NewReader(model.Script)

	op.BytesRead, err = reader.ReadAt(op.Dst, op.Offset)

	// Special case: FUSE doesn't expect us to return io.EOF.
	if err == io.EOF {
		return nil
	}

	return err
}
func (fs *sqlFS) WriteFile(
	ctx context.Context,
	op *fuseops.WriteFileOp) error {

	node, ok := gInodeInfo[op.Inode]
	if !ok {
		return fmt.Errorf("不存在的InodeID")
	}
	model, err := filemodel.GetModelByID(filemodel.GetDB(), node.RowID)
	if err != nil {
		return fmt.Errorf("读取数据错误")
	}
	model.Script = op.Data
	err = filemodel.SaveModel(filemodel.GetDB(), model)
	return err
}
func (fs *sqlFS) SetInodeAttributes(
	ctx context.Context,
	op *fuseops.SetInodeAttributesOp) error {
	return nil
}
func (fs *sqlFS) FlushFile(
	ctx context.Context,
	op *fuseops.FlushFileOp) (err error) {
	return nil
}
func (fs *sqlFS) ReleaseDirHandle(
	ctx context.Context,
	op *fuseops.ReleaseDirHandleOp) error {
	return nil
}
func (fs *sqlFS) SyncFile(
	ctx context.Context,
	op *fuseops.SyncFileOp) error {
	fmt.Println("sync")
	return nil
}

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	var err error
	err = filemodel.OpenDB("sqlite://test.db")
	if err != nil {
		log.Fatal(err)
	}
	gInodeInfo, err = filemodel.GetNodeInfos()
	if err != nil {
		log.Fatal(err)
	}
	bs, _ := json.Marshal(gInodeInfo)
	log.Println(string(bs))
	// Create an appropriate file system.
	server, err := NewSQLFS(timeutil.RealClock())
	if err != nil {
		log.Fatalf("makeFS: %v", err)
	}

	// Mount the file system.
	if *fMountPoint == "" {
		log.Fatalf("You must set --mount_point.")
	}

	cfg := &fuse.MountConfig{
		ReadOnly: *fReadOnly,
	}

	if *fDebug {
		cfg.DebugLogger = log.New(os.Stderr, "fuse: ", 0)
	}

	mfs, err := fuse.Mount(*fMountPoint, server, cfg)
	if err != nil {
		log.Fatalf("Mount: %v", err)
	}

	// Wait for it to be unmounted.
	if err = mfs.Join(context.Background()); err != nil {
		log.Fatalf("Join: %v", err)
	}
}
