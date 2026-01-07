// Package shared provides shared value objects for the marketplace domain.
package shared

import (
	"errors"
	"fmt"
)

// ListingStatus represents the status of a marketplace listing.
type ListingStatus string

// Listing status constants
const (
	ListingDraft    ListingStatus = "draft"
	ListingPending  ListingStatus = "pending"
	ListingActive   ListingStatus = "active"
	ListingPaused   ListingStatus = "paused"
	ListingRejected ListingStatus = "rejected"
	ListingSoldOut  ListingStatus = "sold_out"
)

// ErrInvalidListingStatus is returned for invalid status values.
var ErrInvalidListingStatus = errors.New("invalid listing status")

// AllListingStatuses returns all valid statuses.
func AllListingStatuses() []ListingStatus {
	return []ListingStatus{ListingDraft, ListingPending, ListingActive, ListingPaused, ListingRejected, ListingSoldOut}
}

// IsValid returns true if the status is valid.
func (s ListingStatus) IsValid() bool {
	switch s {
	case ListingDraft, ListingPending, ListingActive, ListingPaused, ListingRejected, ListingSoldOut:
		return true
	default:
		return false
	}
}

// String returns the string representation.
func (s ListingStatus) String() string {
	return string(s)
}

// IsVisible returns true if listing is visible to buyers.
func (s ListingStatus) IsVisible() bool {
	return s == ListingActive
}

// IsPurchasable returns true if listing can be purchased.
func (s ListingStatus) IsPurchasable() bool {
	return s == ListingActive
}

// CanBeEdited returns true if listing can be edited.
func (s ListingStatus) CanBeEdited() bool {
	return s == ListingDraft || s == ListingPending || s == ListingPaused || s == ListingRejected
}

// ParseListingStatus parses a string into a ListingStatus.
func ParseListingStatus(str string) (ListingStatus, error) {
	s := ListingStatus(str)
	if !s.IsValid() {
		return "", fmt.Errorf("%w: %s", ErrInvalidListingStatus, str)
	}
	return s, nil
}
