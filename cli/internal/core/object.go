package core

import (
	"Gel/internal/domain"
	"Gel/internal/storage"
	"fmt"
	"os"
)

type ObjectService struct {
	objectStorage *storage.ObjectStorage
}

func NewObjectService(objectStorage *storage.ObjectStorage) *ObjectService {
	return &ObjectService{
		objectStorage: objectStorage,
	}
}

func (o *ObjectService) GetObjectSize(hash domain.Hash) (uint32, error) {
	compressedData, err := o.objectStorage.Read(hash)
	if err != nil {
		return 0, err
	}

	data, err := Decompress(compressedData)
	if err != nil {
		return 0, err
	}

	object, err := domain.DecodeObject(data)
	if err != nil {
		return 0, err
	}
	return uint32(object.Size()), nil
}

func (o *ObjectService) Write(hash domain.Hash, data []byte) error {
	compressedData, err := Compress(data)
	if err != nil {
		return err
	}
	return o.objectStorage.Write(hash, compressedData)
}

func (o *ObjectService) Read(hash domain.Hash) (domain.Object, error) {
	compressedData, err := o.objectStorage.Read(hash)
	if err != nil {
		return nil, err
	}

	data, err := Decompress(compressedData)
	if err != nil {
		return nil, err
	}

	object, err := domain.DecodeObject(data)
	if err != nil {
		return nil, err
	}
	return object, nil
}

func (o *ObjectService) ReadBlob(hash domain.Hash) (*domain.Blob, error) {
	object, err := o.Read(hash)
	if err != nil {
		return nil, err
	}
	blob, ok := object.(*domain.Blob)
	if !ok {
		// TODO: return proper error
		return nil, fmt.Errorf("object is not a blob")
	}
	return blob, nil
}

func (o *ObjectService) ReadTree(hash domain.Hash) (*domain.Tree, error) {
	object, err := o.Read(hash)
	if err != nil {
		return nil, err
	}

	tree, ok := object.(*domain.Tree)
	if !ok {
		// TODO: return proper error
		return nil, fmt.Errorf("object is not a tree")
	}
	return tree, nil
}

func (o *ObjectService) ReadCommit(hash domain.Hash) (*domain.Commit, error) {
	object, err := o.Read(hash)
	if err != nil {
		return nil, err
	}

	commit, ok := object.(*domain.Commit)
	if !ok {
		// TODO: return proper error
		return nil, fmt.Errorf("object is not a commit")
	}
	return commit, nil
}

func (o *ObjectService) Exists(hash domain.Hash) (bool, error) {
	return o.objectStorage.Exists(hash)
}

func (o *ObjectService) ComputeObjectHash(path domain.AbsolutePath) (domain.Hash, []byte, error) {
	data, err := os.ReadFile(path.String())
	if err != nil {
		return domain.Hash{}, nil, fmt.Errorf("failed to read file at '%s': %w", path, err)
	}

	blob := domain.NewBlob(data)
	encodedData, err := domain.EncodeObject(blob)
	if err != nil {
		return domain.Hash{}, nil, fmt.Errorf("failed to encode object: %w", err)
	}
	hexHash := ComputeSHA256(encodedData)
	hash, err := domain.NewHashFromHex(hexHash)
	if err != nil {
		return domain.Hash{}, nil, err
	}
	return hash, encodedData, nil
}
