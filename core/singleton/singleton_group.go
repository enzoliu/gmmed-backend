package singleton

import (
	"sync"

	"breast-implant-warranty-system/pkg/dbutil"
)

type Group struct {
	readDB     dbutil.PgxReaderItf
	readDBOnce sync.Once

	writeDB     dbutil.PgxClientItf
	writeDBOnce sync.Once
}
