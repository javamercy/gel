package core

import (
	"Gel/internal/domain"
	"fmt"
)

type UpdateRefService struct {
	refService *RefService
}

// NewUpdateRefService creates a service for direct ref updates/deletes.
func NewUpdateRefService(refService *RefService) *UpdateRefService {
	return &UpdateRefService{
		refService: refService,
	}
}

// Update performs an unconditional write of newHash to ref.
func (u *UpdateRefService) Update(ref string, newHash domain.Hash) error {
	return u.refService.Write(ref, newHash)
}

// UpdateSafe performs compare-and-swap semantics: it writes newHash
// only if the current ref value matches oldHash.
func (u *UpdateRefService) UpdateSafe(ref string, newHash, oldHash domain.Hash) error {
	currentHash, err := u.refService.Read(ref)
	if err != nil {
		return err
	}
	if currentHash != oldHash {
		return fmt.Errorf("update-ref: '%s': %w", ref, ErrRefUpdateConflict)
	}
	return u.refService.Write(ref, newHash)
}

// Delete removes ref unconditionally.
func (u *UpdateRefService) Delete(ref string) error {
	return u.refService.Delete(ref)
}

// DeleteSafe removes ref only when its current value equals oldHash.
func (u *UpdateRefService) DeleteSafe(ref string, oldHash domain.Hash) error {
	currentHash, err := u.refService.Read(ref)
	if err != nil {
		return err
	}
	if currentHash != oldHash {
		return fmt.Errorf("update-ref: '%s': %w", ref, ErrRefUpdateConflict)
	}
	return u.refService.Delete(ref)
}
