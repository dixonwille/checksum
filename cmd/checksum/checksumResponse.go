package main

import (
	"github.com/dixonwille/checksum"
)

type checksumResponse struct {
	err       error
	checksums []*checksum.FileChecksum
}

func newChecksumResponse(checksums []*checksum.FileChecksum, err error) checksumResponse {
	return checksumResponse{
		checksums: checksums,
		err:       err,
	}
}
