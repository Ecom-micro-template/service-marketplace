package shared

import (
	"errors"
	"fmt"
)

// SellerStatus represents the status of a marketplace seller.
type SellerStatus string

// Seller status constants
const (
	SellerPending   SellerStatus = "pending"
	SellerApproved  SellerStatus = "approved"
	SellerActive    SellerStatus = "active"
	SellerSuspended SellerStatus = "suspended"
	SellerRejected  SellerStatus = "rejected"
)

// ErrInvalidSellerStatus is returned for invalid status values.
var ErrInvalidSellerStatus = errors.New("invalid seller status")

// AllSellerStatuses returns all valid statuses.
func AllSellerStatuses() []SellerStatus {
	return []SellerStatus{SellerPending, SellerApproved, SellerActive, SellerSuspended, SellerRejected}
}

// IsValid returns true if the status is valid.
func (s SellerStatus) IsValid() bool {
	switch s {
	case SellerPending, SellerApproved, SellerActive, SellerSuspended, SellerRejected:
		return true
	default:
		return false
	}
}

// String returns the string representation.
func (s SellerStatus) String() string {
	return string(s)
}

// CanSell returns true if seller can create listings.
func (s SellerStatus) CanSell() bool {
	return s == SellerActive
}

// CanReceivePayments returns true if seller can receive payments.
func (s SellerStatus) CanReceivePayments() bool {
	return s == SellerActive
}

// IsApproved returns true if seller is approved or active.
func (s SellerStatus) IsApproved() bool {
	return s == SellerApproved || s == SellerActive
}

// ParseSellerStatus parses a string into a SellerStatus.
func ParseSellerStatus(str string) (SellerStatus, error) {
	s := SellerStatus(str)
	if !s.IsValid() {
		return "", fmt.Errorf("%w: %s", ErrInvalidSellerStatus, str)
	}
	return s, nil
}
