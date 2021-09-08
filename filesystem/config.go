package filesystem

import "github.com/menmos/menmos-go"

// A Config regroups configuration options.
type Config struct {
	Client     *menmos.Client
	Profile    string                 `json:"profile"`
	Mountpoint string                 `json:"mount_point"`
	Mount      map[string]interface{} `json:"mount"`
}
