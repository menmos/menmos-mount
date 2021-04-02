package mountpoint

import (
	"sync"

	"github.com/menmos/menmos-go/payload"
)

type pathCache struct {
	mutex sync.Mutex
	data  map[string]string
}

func newPathCache() *pathCache {
	return &pathCache{
		mutex: sync.Mutex{},
		data:  make(map[string]string),
	}
}

func (c *pathCache) GetBlobID(pathSegment string) (string, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if blobID, ok := c.data[pathSegment]; ok {
		return blobID, ok
	}

	return "", false
}

func (c *pathCache) SetBlobID(pathSegment string, blobID string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data[pathSegment] = blobID
}

func (c *pathCache) SetQuery(pathSegment string, query *payload.Query) {}
