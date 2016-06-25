package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"

	"github.com/docker/go-plugins-helpers/volume"
	zfs "github.com/mistifyio/go-zfs"
)

const (
	stateDir  = "/var/lib/docker/plugin-data/"
	stateFile = "zfs-freight.json"
)

type ZfsVolumeDriver struct {
	volumes map[string]string
	mutex   *sync.Mutex
	config  DriverConfig
	name    string `string:"zfs-freight"`
}

type saveData struct {
	State map[string]string `json:"state"`
}

func newZfsVolumeDriver(c DriverConfig) ZfsVolumeDriver {
	fmt.Printf("Starting ZFS Freight Service...\n")

	driver := ZfsVolumeDriver{
		volumes: map[string]string{},
		mutex:   &sync.Mutex{},
		config:  c,
		name:    "zfs-freight",
	}

	os.Mkdir(stateDir, 0700)

	_, driver.volumes = driver.findExistingVolumesFromStateFile()
	fmt.Printf("Found %d volumes on startup\n", len(driver.volumes))

	return driver
}

func (driver ZfsVolumeDriver) Get(req volume.Request) volume.Response {
	if driver.exists(req.Name) {
		return volume.Response{Volume: driver.volume(req.Name)}
	}

	return volume.Response{
		Err: fmt.Sprintf("No volume found with the name %s", req.Name),
	}
}

func (driver ZfsVolumeDriver) List(req volume.Request) volume.Response {
	var volumes []*volume.Volume
	for name, _ := range driver.volumes {
		volumes = append(volumes, driver.volume(name))
	}

	return volume.Response{Volumes: volumes}
}

func (driver ZfsVolumeDriver) Create(req volume.Request) volume.Response {
	fmt.Printf("Create volume %s\n", req.Name)

	driver.mutex.Lock()
	defer driver.mutex.Unlock()

	if driver.exists(req.Name) {
		return volume.Response{Err: fmt.Sprintf("The volume %s already exists", req.Name)}
	}

	vol, err := zfs.CreateFilesystem(fmt.Sprintf("%s/%s", driver.config.Zpool, req.Name), req.Options)
	if err != nil {
		fmt.Println(err.Error())
		return volume.Response{Err: err.Error()}
	}

	driver.volumes[req.Name] = vol.Name
	e := driver.saveState(driver.volumes)
	if e != nil {
		fmt.Println(e.Error())
		return volume.Response{Err: fmt.Sprintf("Failed to save plugin state.")}
	}

	return volume.Response{}
}

func (driver ZfsVolumeDriver) Remove(req volume.Request) volume.Response {
	fmt.Printf("Remove volume %s\n", req.Name)

	driver.mutex.Lock()
	defer driver.mutex.Unlock()

	// delete/archive zfs slice?

	delete(driver.volumes, req.Name)

	err := driver.saveState(driver.volumes)
	if err != nil {
		fmt.Println(err.Error())
		return volume.Response{Err: fmt.Sprintf("Failed to save plugin state.")}
	}

	return volume.Response{}
}

func (driver ZfsVolumeDriver) Mount(req volume.Request) volume.Response {
	res := driver.Path(req)

	fmt.Printf("Mount volume %s on %s\n", req.Name, res.Mountpoint)

	return res
}

func (driver ZfsVolumeDriver) Unmount(req volume.Request) volume.Response {
	fmt.Printf("Unmount volume %s\n", req.Name)

	return driver.Path(req)
}

func (driver ZfsVolumeDriver) Path(req volume.Request) volume.Response {
	vol, err := zfs.GetDataset(fmt.Sprintf("%s/%s", driver.config.Zpool, req.Name))
	if err != nil {
		fmt.Println(err.Error())
		return volume.Response{Err: err.Error()}
	}

	return volume.Response{Mountpoint: vol.Mountpoint}
}

func (driver ZfsVolumeDriver) Capabilities(req volume.Request) volume.Response {
	return volume.Response{Capabilities: volume.Capability{Scope: "local"}}
}

func (driver ZfsVolumeDriver) exists(name string) bool {
	return driver.volumes[name] != ""
}

func (driver ZfsVolumeDriver) volume(name string) *volume.Volume {
	vol, err := zfs.GetDataset(fmt.Sprintf("%s/%s", driver.config.Zpool, name))
	if err != nil {
		fmt.Println(err.Error())
	}

	return &volume.Volume{
		Name:       name,
		Mountpoint: vol.Mountpoint,
	}
}

func (driver ZfsVolumeDriver) findExistingVolumesFromStateFile() (error, map[string]string) {
	path := path.Join(stateDir, stateFile)
	fileData, err := ioutil.ReadFile(path)
	if err != nil {
		return err, map[string]string{}
	}

	var data saveData
	e := json.Unmarshal(fileData, &data)
	if e != nil {
		return e, map[string]string{}
	}

	return nil, data.State
}

func (driver ZfsVolumeDriver) saveState(volumes map[string]string) error {
	data := saveData{State: volumes}

	fileData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	path := path.Join(stateDir, stateFile)
	return ioutil.WriteFile(path, fileData, 0600)
}
