package filesystem

// A Config regroups configuration options.
type Config struct {
	Profile    string                 `json:"profile"`
	Mountpoint string                 `json:"mount_point"`
	Mount      map[string]interface{} `json:"mount"`
}
