package fuse

import "sync"

// #include "wrapper.h"
import "C"

type MountCfg struct {
	fs FileSystem
	se *C.struct_fuse_session
	ch *C.struct_fuse_chan
}

// State which tracks instances of FileSystem, with a unique identifier used
// by C code.  This avoids passing Go pointers into C code.
var fsMapLock sync.RWMutex
var rawFSMap = make(map[string]MountCfg)

// registerFS registers a filesystem with the bridge layer.
// Returns an integer id, which identifies the filesystem instance.
//
// When calling the FUSE lowlevel initialization method (eg fuse_lowlevel_new), the userdata
// argument must be a pointer to an integer holding this id value.  The bridge methods use this to
// determine which filesystem will handle FUSE callbacks.
//
// When the filesystem is no longer active, DeregisterFS can be called to release resources.
func registerFS(mountpoint string, fs FileSystem, se *C.struct_fuse_session, ch *C.struct_fuse_chan) {
	fsMapLock.Lock()
	defer fsMapLock.Unlock()

	rawFSMap[mountpoint] = MountCfg{
		fs: fs,
		se: se,
		ch: ch,
	}
}

// deregisterFS releases a previously allocated filesystem from RegisterRawFs.
func deregisterFS(mountpoint string) {
	fsMapLock.Lock()
	defer fsMapLock.Unlock()

	delete(rawFSMap, mountpoint)
}

// getFS returns the filesystem for the given mountpoint.
func getFS(mountpoint string) FileSystem {
	fsMapLock.RLock()
	defer fsMapLock.RUnlock()

	mi, found := rawFSMap[mountpoint]

	if !found {
		return nil
	}

	return mi.fs
}

// getMountCfg returns the MountCfg for given mountpoint.
func getMountCfg(mountpoint string) MountCfg {
	fsMapLock.RLock()
	defer fsMapLock.RUnlock()

	return rawFSMap[mountpoint]
}