package repository

import (
	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"sync"
)

type localDB struct {
	*gorm.DB
	lock sync.RWMutex
}

var lDB *localDB

func init() {
	dsn := "file::memory:?cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn))
	if err != nil {
		logrus.Errorf("open sqlite fail. error: %s", err)
		panic(err)
	}
	logrus.Debug("open sqlite db")

	db.AutoMigrate(&FileVersion{})

	lDB = &localDB{
		DB:   db,
		lock: sync.RWMutex{},
	}
}
