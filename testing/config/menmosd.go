// +build integration

package config

import (
	_ "embed"
	"os/exec"
)

//go:embed data/menmosd.toml
var menmosdConfigTemplate string

type menmosdHandle struct {
	cmd *exec.Cmd
}

func newMenmosd(cwd string, binPath string) (*menmosdHandle, error) {
	cmd, err := startProcess(binPath, cwd, "menmosd.toml", menmosdConfigTemplate, "http://localhost:30300/health")
	if err != nil {
		return nil, err
	}
	return &menmosdHandle{cmd}, nil
}

func (h *menmosdHandle) Stop() error {
	if err := h.cmd.Process.Kill(); err != nil {
		return err
	}

	return h.cmd.Wait()
}
