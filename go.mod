module github.com/menmos/menmos-mount

go 1.16

replace github.com/rclone/rclone => github.com/menmos/rclone v1.54.1-0.20210312030826-dcedb43608f8

require (
	bazil.org/fuse v0.0.0-20200524192727-fb710f7dfd05
	github.com/billziss-gh/cgofuse v1.4.0 // indirect
	github.com/menmos/menmos-go v0.0.0-20210327184101-ecf99c567e7a
	github.com/mitchellh/mapstructure v1.4.1
	github.com/pkg/errors v0.9.1
	github.com/rclone/rclone v1.54.1
	github.com/urfave/cli/v2 v2.3.0
	golang.org/x/sys v0.0.0-20210124154548-22da62e12c0c
)
