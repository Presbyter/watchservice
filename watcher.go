package watchserver

import (
	"context"
	"errors"
	"fmt"
	"github.com/Presbyter/watchserver/repository"
	"github.com/Presbyter/watchserver/unit"
	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"os"
	"path/filepath"
)

type WatchServer struct {
	folderPath         string
	watcher            *fsnotify.Watcher
	createEventHandler func(event *fsnotify.Event)
	updateEventHandler func(event *fsnotify.Event)
	deleteEventHandler func(event *fsnotify.Event)
	chmodEventHandler  func(event *fsnotify.Event)
}

func (s *WatchServer) init() {
	fiList, err := os.ReadDir(s.folderPath)
	if err != nil {
		logrus.Errorf("read dir fail. error: %s. %s", err, s.folderPath)
		return
	}

	repo := repository.NewFileVersionRepo()

	for _, f := range fiList {
		if f.IsDir() {
			continue
		}

		fileName := filepath.Join(s.folderPath, f.Name())
		hash, err := unit.Md5Hash(fileName)
		if err != nil {
			logrus.Errorf("md5 sum file fail. error: %s", err)
			return
		}

		entity := &repository.FileVersion{
			FileName: fileName,
			Hash:     fmt.Sprintf("%x", hash),
		}

		if err := repo.SaveEntity(entity); err != nil {
			logrus.Errorf("save file version fail. error: %s. %s", err, fileName)
			continue
		}
	}
}

func (s *WatchServer) Run(ctx context.Context) error {
	s.init()

	errCh := make(chan error)
	go func() {
		for {
			select {
			case event := <-s.watcher.Events:
				switch event.Op {
				case fsnotify.Create:
					if s.createEventHandler != nil {
						s.createEventHandler(&event)
					}
				case fsnotify.Write:
					if s.updateEventHandler != nil {
						s.updateEventHandler(&event)
					}
				case fsnotify.Remove:
					if s.deleteEventHandler != nil {
						s.deleteEventHandler(&event)
					}
				case fsnotify.Rename:
					if s.deleteEventHandler != nil {
						s.deleteEventHandler(&event)
					}
				case fsnotify.Chmod:
					if s.chmodEventHandler != nil {
						s.chmodEventHandler(&event)
					}
				}
			case err := <-s.watcher.Errors:
				errCh <- err
			}
		}
	}()

	if err := s.watcher.Add(s.folderPath); err != nil {
		logrus.Errorf("watcher add folder fail. error: %s", err)
		return err
	}

	select {
	case err := <-errCh:
		logrus.Errorf("watcher received error: %s", err)
		return err
	case <-ctx.Done():
		return nil
	}
}

func (s *WatchServer) SetCreateEventHandler(foo func(event *fsnotify.Event)) {
	s.createEventHandler = foo
}

func (s *WatchServer) UpdateEventHandler(foo func(event *fsnotify.Event)) {
	s.updateEventHandler = foo
}

func (s *WatchServer) DeleteEventHandler(foo func(event *fsnotify.Event)) {
	s.deleteEventHandler = foo
}

func (s *WatchServer) Close() error {
	return s.watcher.Close()
}

func NewWatcher(path string) *WatchServer {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	path, _ = filepath.Abs(path)
	return &WatchServer{
		folderPath:         path,
		watcher:            watcher,
		createEventHandler: DefaultCreateEventHandler,
		updateEventHandler: DefaultUpdateEventHandler,
		deleteEventHandler: DefaultDeleteEventHandler,
	}
}

func DefaultCreateEventHandler(event *fsnotify.Event) {
	fileName, _ := filepath.Abs(event.Name)
	logrus.Debugf("create file: %s", fileName)

	hash, err := unit.Md5Hash(fileName)
	if err != nil {
		logrus.Errorf("md5 sum file fail. error: %s", err)
		return
	}

	repo := repository.NewFileVersionRepo()
	entity := &repository.FileVersion{
		FileName: fileName,
		Hash:     fmt.Sprintf("%s", hash),
	}
	if err := repo.SaveEntity(entity); err != nil {
		logrus.Errorf("save file verison fail. error: %s", err)
		return
	}
}

func DefaultUpdateEventHandler(event *fsnotify.Event) {
	fileName, _ := filepath.Abs(event.Name)
	logrus.Debugf("update file: %s", fileName)

	hash, err := unit.Md5Hash(fileName)
	if err != nil {
		logrus.Errorf("md5 sum file fail. error: %s", err)
		return
	}

	repo := repository.NewFileVersionRepo()
	entity, err := repo.GetByFileName(fileName)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		entity = &repository.FileVersion{
			FileName: fileName,
			Hash:     fmt.Sprintf("%s", hash),
		}
	} else if err != nil {
		logrus.Errorf("get file version fail. error: %s", err)
		return
	} else {
		entity.Hash = fmt.Sprintf("%x", hash)
	}

	if err := repo.SaveEntity(entity); err != nil {
		logrus.Errorf("save file version fail. error: %s", err)
		return
	}
}

func DefaultDeleteEventHandler(event *fsnotify.Event) {
	fileName, _ := filepath.Abs(event.Name)
	logrus.Debugf("delete file: %s", fileName)

	repo := repository.NewFileVersionRepo()
	if err := repo.RemoveByName(fileName); err != nil {
		logrus.Errorf("delete file version fail. error: %s", err)
		return
	}
}
