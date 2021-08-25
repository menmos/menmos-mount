module github.com/menmos/menmos-mount

go 1.16

replace github.com/rclone/rclone => github.com/menmos/rclone v1.54.1-0.20210824133103-2a68cccd1d16

require (
	bazil.org/fuse v0.0.0-20200524192727-fb710f7dfd05
	github.com/menmos/menmos-go v0.0.0-20210825004229-a681775a628c
	github.com/mitchellh/mapstructure v1.4.1
	github.com/pkg/errors v0.9.1
	github.com/rclone/rclone v1.56.0
	github.com/urfave/cli/v2 v2.3.0
	golang.org/x/sys v0.0.0-20210823070655-63515b42dcdf
)
