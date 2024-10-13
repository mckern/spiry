package spiry

import "time"

type ExpiringResource interface {
	Name() string
	Expiry() (time.Time, error)
}
