// +build integration

package config

import "github.com/menmos/menmos-go"

type ServerConfig struct {
	Client *menmos.Client
}

var MenmosConfig = ServerConfig{
	Client: nil,
}
