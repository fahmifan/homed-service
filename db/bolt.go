package db

import (
	"time"

	"gitlab.com/homed/homed-service/model"

	"gitlab.com/homed/homed-service/config"

	"github.com/boltdb/bolt"
	log "github.com/sirupsen/logrus"
)

// NewBoltDB :nodoc:
func NewBoltDB() *bolt.DB {
	db, err := bolt.Open(config.BoltDBName(), 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(err)
	}

	buckets := [][]byte{model.VideoBucket()}
	err = db.Update(func(tx *bolt.Tx) error {
		for _, b := range buckets {
			_, err := tx.CreateBucketIfNotExists(b)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	return db
}
