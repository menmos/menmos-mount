package mountpoint

import (
	"errors"

	"github.com/menmos/menmos-go"
	"github.com/menmos/menmos-go/payload"
	"github.com/mitchellh/mapstructure"
	"github.com/rclone/rclone/fs"
)

type MountBuilder interface {
	IntoMount(client *menmos.Client, fs fs.Info) (MountPoint, error)
}

type rawQueryMount struct {
	Expression      map[string]interface{} `json:"expression"`
	GroupByTags     bool                   `json:"group_by_tags,omitempty"`
	GroupByMetaKeys []string               `json:"group_by_meta_keys,omitempty"`
}

func (r rawQueryMount) IntoMount(client *menmos.Client, fs fs.Info) (MountPoint, error) {
	parsedExpression, err := payload.ParseExpression(r.Expression)
	if err != nil {
		return nil, err
	}

	return NewQueryMount(parsedExpression, r.GroupByTags, r.GroupByMetaKeys, client, fs), nil
}

type rawBlobMount struct {
	BlobID string `json:"blob_id"`
}

func (r rawBlobMount) IntoMount(client *menmos.Client, fs fs.Info) (MountPoint, error) {
	return NewBlobMount(r.BlobID, client, fs), nil
}

func Load(rawDict map[string]interface{}, client *menmos.Client, fs fs.Info) (MountPoint, error) {
	var mountData MountBuilder
	if _, ok := rawDict["expression"]; ok {
		mountData = rawQueryMount{}
	} else if _, ok := rawDict["blob_id"]; ok {
		mountData = rawBlobMount{}
	} else {
		// We assume virtual mount.
		subMounts := make(map[string]MountPoint)
		for mountName, data := range rawDict {
			if dataMap, ok := data.(map[string]interface{}); ok {
				subMount, err := Load(dataMap, client, fs)
				if err != nil {
					return nil, err
				}
				subMounts[mountName] = subMount
			} else {
				return nil, errors.New("unknown mount type")
			}
		}
		return NewVirtualMount(subMounts), nil
	}

	decoderConfig := mapstructure.DecoderConfig{TagName: "json", Result: &mountData}
	decoder, err := mapstructure.NewDecoder(&decoderConfig)
	if err != nil {
		return nil, err
	}

	if err := decoder.Decode(rawDict); err != nil {
		return nil, err
	}

	return mountData.IntoMount(client, fs)
}
