package main

import (
	"crypto"
	"os"

	"sync"

	"github.com/dixonwille/checksum"
)

func getChecksums(hash crypto.Hash, srcs []string, response chan checksum.FileChecksumResponse) {
	wg := new(sync.WaitGroup)
	for _, src := range srcs {
		wg.Add(1)
		go func(s string) {
			defer wg.Done()
			getChecksum(hash, s, response)
		}(src)
	}
	wg.Wait()
}

func getChecksum(hash crypto.Hash, src string, response chan checksum.FileChecksumResponse) {
	if info, err := os.Stat(src); err != nil {
		response <- checksum.NewResponse(nil, checksum.NewError(src, checksum.ErrCannotBeRead, err))
	} else if info.IsDir() {
		checksum.Directory(src, hash, response)
	} else {
		checksum.File(src, hash, response)
	}
}
