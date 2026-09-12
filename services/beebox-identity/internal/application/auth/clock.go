package auth

import "time"

type Clock interface {
	Now() time.Time
}
