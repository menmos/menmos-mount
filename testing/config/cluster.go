// +build integration

package config

import (
	"errors"
	"io/ioutil"
	"os"
	"path"

	"github.com/menmos/menmos-go"
)

type Cluster struct {
	menmosd *menmosdHandle
	amphora *amphoraHandle

	client *menmos.Client

	workingDirectory string
}

func Menmos(buildDirectory string) (*Cluster, error) {
	if len(buildDirectory) == 0 {
		return nil, errors.New("build directory is not set")
	}

	workingDirectory, err := ioutil.TempDir("", "menmos-integration-cluster-*")
	if err != nil {
		return nil, err
	}

	menmosdPath := path.Join(buildDirectory, "menmosd")
	amphoraPath := path.Join(buildDirectory, "amphora")

	menmosd, err := newMenmosd(workingDirectory, menmosdPath)
	if err != nil {
		return nil, err
	}

	client, err := menmos.New("http://localhost:30300", "admin", "integration")
	if err != nil {
		return nil, err
	}

	amphora, err := newAmphora(workingDirectory, amphoraPath)
	if err != nil {
		return nil, err
	}

	if err := waitUntil(func() bool {
		nodes, err := client.ListStorageNodes()
		if err != nil {
			return false
		}
		return len(nodes) > 0
	}, 5); err != nil {
		return nil, err
	}

	return &Cluster{menmosd: menmosd, amphora: amphora, workingDirectory: workingDirectory, client: client}, nil
}

func (m *Cluster) Client() *menmos.Client {
	return m.client
}

func (m *Cluster) Stop() {
	m.menmosd.Stop()
	m.amphora.Stop()
	os.Remove(m.workingDirectory)
}
