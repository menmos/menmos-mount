// +build integration

package integration

import (
	"flag"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/menmos/menmos-go"
	"github.com/menmos/menmos-mount/testing/config"
)

var integrationLocal = flag.Bool("integration-local", false, "spin up a local instance of menmos within the test")
var menmosdBuildDirectory = flag.String("menmosd-target", "", "path to the directory containing release builds for menmosd and amphora")

var menmosdAddress = flag.String("menmosd-address", "localhost", "menmosd address")
var menmosdPort = flag.Int("menmosd-port", 30300, "menmosd port")
var menmosdUsername = flag.String("menmos-username", "admin", "menmosd username")
var menmosdPassword = flag.String("menmos-password", "integration", "menmosd password")

func TestMain(m *testing.M) {
	flag.Parse()

	var exitCode int
	if *integrationLocal {
		log.Println(">>> starting our local menmos cluster")
		h, err := config.Menmos(*menmosdBuildDirectory)
		if err != nil {
			panic(err)
		}
		config.MenmosConfig.Client = h.Client()
		exitCode = m.Run()
		h.Stop()
	} else {
		log.Println(">>> running in integration mode")
		client, err := menmos.New(fmt.Sprintf("http://%s:%d", *menmosdAddress, *menmosdPort), *menmosdUsername, *menmosdPassword)
		if err != nil {
			panic(err)
		}

		config.MenmosConfig = config.ServerConfig{Client: client}

		exitCode = m.Run()
	}

	// TODO: Add a cleanup routine here that queries for all items with the "integration" tag and deletes them all, just in case.

	os.Exit(exitCode)
}
