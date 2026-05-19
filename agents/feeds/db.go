package feeds

import (
	"sync"

	"github.com/cockroachdb/pebble"
)

var DB = sync.OnceValues(func() (*pebble.DB, error) {
	db, err := pebble.Open("pebble_data", &pebble.Options{})
	if err != nil {
		return nil, err
	}

	return db, nil
})
