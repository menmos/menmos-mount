// +build integration

package config

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"time"
)

func waitUntil(fn func() bool, timeoutSeconds int) error {
	for {
		if fn() {
			return nil
		} else {
			time.Sleep(50 * time.Millisecond)
		}

		// TODO if waited too much return error
	}
}

func waitUntilOk(url string, timeoutSeconds int) error {
	return waitUntil(func() bool {
		resp, err := http.Get(url)

		if err != nil || resp.StatusCode >= 300 || resp.StatusCode < 200 {
			return false
		} else {
			return true
		}
	}, timeoutSeconds)
}

func startProcess(executable string, cwd string, configName string, configTemplate string, healthURL string) (*exec.Cmd, error) {
	// Format configuration
	formattedConfig := fmt.Sprintf(configTemplate, cwd)

	// Write config on disk.
	configPath := path.Join(cwd, configName)
	if err := ioutil.WriteFile(configPath, []byte(formattedConfig), fs.ModePerm); err != nil {
		return nil, err
	}

	// Start the process
	cmd := exec.Command(executable, "--cfg", configPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	waitUntilOk(healthURL, 5)

	return cmd, nil
}
