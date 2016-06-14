package main

import (
	"fmt"
	"testing"

	"github.com/docker/go-plugins-helpers/volume"
	zfs "github.com/mistifyio/go-zfs"
)

var (
	c = DriverConfig{
		Listen: ":9732",
		Zpool: "test",
	}
	testVolumeName    = "test-volume"
	testVolumeOptions = map[string]string{
		"compression": "on",
	}
)

func TestCreate(t *testing.T) {
	driver := newZfsVolumeDriver(c)

	createHelper(driver, t, testVolumeName, testVolumeOptions)

	// test that a volume was created
	vol, err := zfs.GetDataset(fmt.Sprintf("%s/%s", driver.config.Zpool, testVolumeName))
	if err != nil {
		t.Error("!!! Volume was not created", err.Error())
	}

	// test that the volume's options are correct
	if vol.Compression != "on" {
		t.Error("!!! Volume options are not correct")
	}

	// test that plugin state store has one entry
	if len(driver.volumes) != 1 {
		t.Error("!!! State store should report exactly 1 volume")
	}

	cleanupHelper(driver, t, testVolumeName)
}

func TestGet(t *testing.T) {
	driver := newZfsVolumeDriver(c)

	createHelper(driver, t, testVolumeName, testVolumeOptions)

	// test that we can find the volume
	res := driver.Get(volume.Request{ Name: testVolumeName })
	if res.Err != "" {
		t.Error("!!! Failed to find volume!")
	}

	cleanupHelper(driver, t, testVolumeName)
}

func TestList(t *testing.T) {
	driver := newZfsVolumeDriver(c)

	createHelper(driver, t, testVolumeName, testVolumeOptions)

	// test that we can list volumes
	res := driver.List(volume.Request{})
	if len(res.Volumes) != 1 {
		t.Error("!!! Failed to find single volume")
	}

	name := testVolumeName + "2"
	createHelper(driver, t, name, testVolumeOptions)

	// test that we can list multiple volumes
	res2 := driver.List(volume.Request{})
	if len(res2.Volumes) != 2 {
		t.Error("!!! Failed to find multiple volumes")
	}

	cleanupHelper(driver, t, testVolumeName)
	cleanupHelper(driver, t, name)
}

func TestMountUnmountPath(t *testing.T) {
	driver := newZfsVolumeDriver(c)

	createHelper(driver, t, testVolumeName, testVolumeOptions)

	// mount, mount and path should have same output (they all use Path under the hood)
	pathRes := driver.Path(volume.Request{Name: testVolumeName})
	mountRes := driver.Mount(volume.Request{Name: testVolumeName})
	unmountRes := driver.Unmount(volume.Request{Name: testVolumeName})

	// test that the zfs volume exists
	vol, err := zfs.GetDataset(fmt.Sprintf("%s/%s", driver.config.Zpool, testVolumeName))
	if err != nil {
		t.Error("!!! Could not find ZFS volume", err.Error())
	}

	// test that mountpoints are correct
	if !(pathRes.Mountpoint == mountRes.Mountpoint &&
		mountRes.Mountpoint == unmountRes.Mountpoint &&
		unmountRes.Mountpoint == vol.Mountpoint) {
		t.Error("!!! Mount, Unmount and Path do not return the same Mountpoint")
	}

	cleanupHelper(driver, t, testVolumeName)
}

func createHelper(driver ZfsVolumeDriver, t *testing.T, name string, options map[string]string) {
	res := driver.Create(volume.Request{
		Name:    name,
		Options: options,
	})

	if res.Err != "" {
		t.Error("[createHelper]", res.Err)
	}
}

func cleanupHelper(driver ZfsVolumeDriver, t *testing.T, name string) {
	vol, err := zfs.GetDataset(fmt.Sprintf("%s/%s", driver.config.Zpool, name))
	if err != nil {
		t.Error("[Cleanup] ZFS volume could not be found:", err.Error())
	}

	err2 := vol.Destroy(zfs.DestroyDefault)
	if err2 != nil {
		t.Error("[Cleanup] ZFS volume was not destroyed:", err2.Error())
	}

	res := driver.Remove(volume.Request{Name: name})
	if res.Err != "" {
		t.Error("[Cleanup] Volume was not removed from state storage:", res.Err)	
	}
}
