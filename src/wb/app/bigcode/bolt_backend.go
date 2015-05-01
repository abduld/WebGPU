package bigcode

import (
	"strconv"
	"time"
	"wb/app/models"

	"github.com/boltdb/bolt"
)

var boltdb *bolt.DB

func getSignature(mpNum int, progId int64) (string, error) {
	var err error = nil
	var sig string = ""
	bucketName := strconv.Itoa(mpNum)
	boltdb.View(func(tx *bolt.Tx) error {
		sig = string(tx.Bucket([]byte(bucketName)).Get([]byte(strconv.Itoa(int(progId)))))
		return nil
	})
	return sig, err
}

func addSignature(mpNum int, p models.Program, sig string) error {
	var err error = nil
	bucketName := strconv.Itoa(mpNum)
	boltdb.Update(func(tx *bolt.Tx) error {
		var bucket *bolt.Bucket
		if bucket, err = tx.CreateBucketIfNotExists([]byte(bucketName)); err == nil {
			bucket.Put([]byte(strconv.Itoa(int(p.Id))), []byte(sig))
			return nil
		} else {
			return err
		}
	})
	return err
}

func initBoldBackend() error {
	var err error
	boltdb, err = bolt.Open("bigcode.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}

	return nil
}
