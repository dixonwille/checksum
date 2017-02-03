package main

import (
	"crypto"
	"os"

	"github.com/dixonwille/checksum"
)

func getChecksums(hash crypto.Hash, srcs []string) <-chan checksumResponse {
	var channels []<-chan checksumResponse
	for _, src := range srcs {
		channels = append(channels, getChecksum(hash, src))
	}
	return channelHandler(channels)
}

func getChecksum(hash crypto.Hash, src string) <-chan checksumResponse {
	csr := make(chan checksumResponse)
	go func() {
		defer close(csr)
		var checksums []*checksum.FileChecksum
		if info, err := os.Stat(src); err != nil {
			csr <- newChecksumResponse(nil, err)
			return
		} else if info.IsDir() {
			css, err := checksum.Directory(src, hash)
			if err != nil {
				csr <- newChecksumResponse(nil, err)
				return
			}
			checksums = append(checksums, css...)
		} else {
			cs, err := checksum.File(src, hash)
			if err != nil {
				csr <- newChecksumResponse(nil, err)
				return
			}
			checksums = append(checksums, cs)
		}
		csr <- newChecksumResponse(checksums, nil)
	}()
	return csr
}
