package core

import (
	"Gel/internal/domain"
	"Gel/internal/storage"
	"errors"
	"os"
)

// IndexService manages loading and saving repository index state.
type IndexService struct {
	indexStorage *storage.IndexStorage
}

// NewIndexService creates an index service backed by the provided index storage.
func NewIndexService(indexStorage *storage.IndexStorage) *IndexService {
	return &IndexService{
		indexStorage: indexStorage,
	}
}

// Read loads the repository index from storage.
// If the index file does not exist yet, it returns an empty index.
func (i *IndexService) Read() (*domain.Index, error) {
	data, err := i.indexStorage.Read()
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return domain.DecodeIndex(data)
}

// Write serializes the given index and persists it to storage.
func (i *IndexService) Write(index *domain.Index) error {
	serializedData, err := domain.EncodeIndex(index)
	if err != nil {
		return err
	}
	return i.indexStorage.Write(serializedData)
}

// GetEntries returns the current index entries from storage.
func (i *IndexService) GetEntries() ([]domain.IndexEntry, error) {
	index, err := i.Read()
	if err != nil {
		return nil, err
	}
	return index.Entries(), nil
}

// WriteEntries replaces the current index entries and persists the updated index.
func (i *IndexService) WriteEntries(entries []domain.IndexEntry) error {
	index, err := i.Read()
	if err != nil {
		return err
	}
	for _, entry := range entries {
		// TODO: handle error
		index.SetEntry(entry)
	}
	return i.Write(index)
}
