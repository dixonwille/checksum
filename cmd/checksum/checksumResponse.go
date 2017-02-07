package main

import (
	"github.com/dixonwille/checksum"
)

type checksumResponse struct {
	errs      []error
	checksums []*checksum.FileChecksum
}

func newChecksumResponse(checksums []*checksum.FileChecksum, err ...error) checksumResponse {
	return checksumResponse{
		checksums: checksums,
		errs:      err,
	}
}
