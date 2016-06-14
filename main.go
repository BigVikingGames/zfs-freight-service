package main

import (
	"fmt"
	"log"

	"github.com/docker/go-plugins-helpers/volume"
	"github.com/kelseyhightower/envconfig"
)

type DriverConfig struct {
	Listen string `default:":2379"`
	Zpool  string `default:"tank"`
}

func main() {
	var c DriverConfig
	err := envconfig.Process("freight", &c)
	if err != nil {
		log.Fatal(err.Error())
	}

	driver := newZfsVolumeDriver(c)
	handler := volume.NewHandler(driver)

	fmt.Println(handler.ServeTCP(driver.name, c.Listen))
}
