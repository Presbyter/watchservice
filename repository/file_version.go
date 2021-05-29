package repository

import (
	"time"
)

type FileVersion struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	FileName  string `gorm:"uniqueIndex"`
	Hash      string
}

type FileVersionRepo struct {
	*localDB
}

func NewFileVersionRepo() *FileVersionRepo {
	return &FileVersionRepo{lDB}
}

func (r *FileVersionRepo) Get(id uint) (*FileVersion, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	var entity FileVersion
	err := r.First(&entity, id).Error
	return &entity, err
}

func (r *FileVersionRepo) GetByFileName(name string) (*FileVersion, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	var entity FileVersion
	err := r.First(&entity, FileVersion{FileName: name}).Error
	return &entity, err
}

func (r *FileVersionRepo) Remove(id uint) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	return r.Delete(&FileVersion{}, id).Error
}

func (r *FileVersionRepo) RemoveByName(name string) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	return r.Delete(&FileVersion{}, FileVersion{FileName: name}).Error
}

func (r *FileVersionRepo) SaveEntity(entity *FileVersion) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	return r.Save(entity).Error
}
