package checksum

import (
	"bufio"
	"crypto"
	"io"
	"os"

	"path/filepath"
	"sync"
)

var (
	//BufferSize is used to buffer the file before hashing.
	BufferSize = 4096
	//Limit is the limit of how many open files to process at one time.
	Limit = 100
)

type fileParams struct {
	path string
	h    crypto.Hash
}

//File gets the checksum of the file and responds on the channel passed in.
func File(path string, h crypto.Hash, response chan FileChecksumResponse) {
	if !h.Available() {
		response <- NewResponse(nil, NewError(path, ErrHashBinary, nil))
		return
	}
	if info, err := os.Stat(path); err != nil {
		response <- NewResponse(nil, NewError(path, ErrCannotBeRead, err))
		return
	} else if info.IsDir() {
		response <- NewResponse(nil, NewError(path, ErrFileIsDirectory, nil))
		return
	} else if !info.Mode().IsRegular() {
		response <- NewResponse(nil, NewError(path, ErrNotRegular, nil))
		return
	}

	file, err := os.Open(path)
	if err != nil {
		response <- NewResponse(nil, NewError(path, ErrFileNotOpen, err))
		return
	}
	defer file.Close()
	buf := make([]byte, BufferSize)
	fileReader := bufio.NewReader(file)
	hash := h.New()
	for {
		if n, err := fileReader.Read(buf); err == nil {
			nw, e := hash.Write(buf[:n])
			if e != nil {
				response <- NewResponse(nil, NewError(path, ErrWritingHash, err))
				return
			}
			if nw != n {
				response <- NewResponse(nil, NewError(path, ErrImproperLength, nil))
				return
			}
		} else if err == io.EOF {
			break
		} else {
			response <- NewResponse(nil, NewError(path, ErrFileReading, err))
			return
		}
	}
	response <- NewResponse(&FileChecksum{
		FilePath: path,
		Checksum: hash.Sum(nil),
	}, nil)
}

//Directory gets the checksum of all files in the directory and responds on the channel.
func Directory(root string, h crypto.Hash, response chan FileChecksumResponse) {
	if !h.Available() {
		response <- NewResponse(nil, NewError(root, ErrHashBinary, nil))
		return
	}
	if _, err := os.Stat(root); err != nil {
		response <- NewResponse(nil, NewError(root, ErrCannotBeRead, err))
		return
	}
	wg := new(sync.WaitGroup)
	files := make(chan fileParams)
	defer close(files)
	//These are the workers and are limited to Limit
	for i := 0; i < Limit; i++ {
		go func() {
			for {
				select {
				case file, ok := <-files:
					if !ok {
						continue
					}
					File(file.path, file.h, response)
					wg.Done()
				}
			}
		}()
	}

	err := filepath.Walk(root, func(path string, info os.FileInfo, wlkErr error) error {
		if wlkErr != nil {
			response <- NewResponse(nil, NewError(path, ErrWalking, wlkErr))
			return nil
		}
		if !info.IsDir() {
			wg.Add(1)
			files <- fileParams{
				path: path,
				h:    h,
			}
		}
		return nil
	})
	wg.Wait()
	if err != nil {
		response <- NewResponse(nil, err)
	}
}
