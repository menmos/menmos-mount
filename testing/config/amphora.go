// +build integration

package config

import (
	_ "embed"
	"os/exec"
)

//go:embed data/amphora.toml
var amphoraConfigTemplate string

type amphoraHandle struct {
	cmd *exec.Cmd
}

func newAmphora(cwd string, binPath string) (*amphoraHandle, error) {
	cmd, err := startProcess(binPath, cwd, "amphora.toml", amphoraConfigTemplate, "http://localhost:30301/health")
	if err != nil {
		return nil, err
	}
	return &amphoraHandle{cmd}, nil
}

func (h *amphoraHandle) Stop() error {
	if err := h.cmd.Process.Kill(); err != nil {
		return err
	}

	return h.cmd.Wait()
}
