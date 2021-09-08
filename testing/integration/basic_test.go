// +build integration

package integration

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/menmos/menmos-go/payload"
	"github.com/menmos/menmos-mount/filesystem"
	"github.com/menmos/menmos-mount/testing/config"
)

const IntegrationTestTag = "integrationtest"

func TestBasicMount(t *testing.T) {
	client := config.MenmosConfig.Client

	mountPath, err := ioutil.TempDir("", "mount-*")
	if err != nil {
		t.Error(err.Error())
		return
	}
	defer os.Remove(mountPath)

	data := []byte("hello world")
	meta := payload.NewBlobMeta("test_file", "File", uint64(len(data)))
	meta.Tags = append(meta.Tags, IntegrationTestTag)
	blobID, err := client.CreateBlob(io.NopCloser(bytes.NewReader(data)), meta)
	if err != nil {
		t.Error("failed to create blob")
		return
	}

	t.Logf("inserted blob ID: %s", blobID)

	defer func() {
		if e := client.Delete(blobID); e != nil {
			t.Errorf("failed to cleanup: %s", e.Error())
			return
		} else {
			t.Logf("cleaned up %s", blobID)
		}
	}()

	config := filesystem.Config{
		Client:     client,
		Mountpoint: mountPath,
		Mount: map[string]interface{}{
			"integration_mount": map[string]interface{}{
				"expression": map[string]string{
					"tag": IntegrationTestTag,
				},
			},
		},
	}

	mountPoint, err := filesystem.Mount(config, false)
	if err != nil {
		t.Error(err.Error())
		return
	}

	// Now that we're mounted, we check that the file we expect exists.
	dirInfo, err := ioutil.ReadDir(path.Join(mountPath, "integration_mount"))
	if err != nil {
		t.Error(err.Error())
		return
	}

	if len(dirInfo) != 1 {
		t.Errorf("expected count=%d, got count=%d", 1, len(dirInfo))
		return
	}

	if err := mountPoint.Unmount(); err != nil {
		t.Errorf("failed to unmount: %s", err.Error())
		return
	}
}
