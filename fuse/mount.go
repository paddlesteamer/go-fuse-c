package fuse

// #include "wrapper.h"
// #include <stdlib.h>
import "C"
import (
	"fmt"
	"os"
	"path/filepath"
)

// MountAndRun mounts the filesystem and enters the Fuse event loop.
// The argumenst are passed to libfuse to mount the filesystem.  Any flags supported by libfuse are
// allowed. The call returns immediately on error, or else blocks until the filesystem is
// unmounted.
//
// Example:
//
//   fs := &MyFs{}
//   err := fuse.MountAndRun(os.Args, fs)
func MountAndRun(args []string, fs FileSystem) int {

	// Make args available to C code.
	argv := make([]*C.char, 0, len(args)+1)
	for _, s := range args {
		p := C.CString(s)
		argv = append(argv, p)
	}
	if len(args) < 2 {
		argv = append(argv, C.CString("-h"))
	}
	argc := C.int(len(argv))

	fuseArgs := C.InitArgs(argc, &argv[0])
	mountpoint := C.ParseMountpoint(fuseArgs)

	ch := C.Mount(mountpoint, fuseArgs)
	if ch == nil {
		return -1
	}

	se := C.NewSession(mountpoint, fuseArgs, ch)
	if se == nil {
		return -1
	}

	mp := C.GoString(mountpoint)

	registerFS(mp, fs, se, ch)
	defer deregisterFS(mp)

	return int(C.Run(mountpoint, se, ch))
}

func UMount(mountpoint string) {
	if !filepath.IsAbs(mountpoint) {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Printf("couldn't get current working directory: %v\n", err)
			return
		}

		mountpoint = filepath.Join(cwd, mountpoint)
	}

	mountpoint, err := filepath.Abs(mountpoint)
	if err != nil {
		fmt.Printf("couldn't get absolute path of %s: %v\n", mountpoint, err)
		return
	}

	mcfg := getMountCfg(mountpoint)

	C.Exit(C.CString(mountpoint), mcfg.se, mcfg.ch)
}
