package checksum

import (
	"bufio"
	"crypto"
	"fmt"
	"io"
	"os"

	"sync"

	"github.com/stretchr/powerwalk"
)

const bufferSize = 4096

//File is used to get the checksum of the file.
func File(path string, h crypto.Hash) (*FileChecksum, error) {
	if !h.Available() {
		return nil, newError(path, ErrHashBinary, nil)
	}
	hash := h.New()
	if info, err := os.Stat(path); err != nil {
		return nil, newError(path, ErrCannotBeRead, err)
	} else if info.IsDir() {
		return nil, newError(path, ErrFileIsDirectory, nil)
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, newError(path, ErrFileNotOpen, err)
	}
	defer file.Close()
	buf := make([]byte, bufferSize)
	fileReader := bufio.NewReader(file)
	for {
		if n, err := fileReader.Read(buf); err == nil {
			nw, err := hash.Write(buf[:n])
			if err != nil {
				return nil, newError(path, ErrWritingHash, err)
			}
			if nw != n {
				fmt.Printf("ReadBits: %d\nWriteBits: %d\n", n, nw)
				return nil, newError(path, ErrImproperLength, nil)
			}
		} else if err == io.EOF {
			break
		} else {
			return nil, newError(path, ErrFileReading, err)
		}
	}
	return &FileChecksum{
		FilePath: path,
		Checksum: hash.Sum(nil),
	}, nil
}

//Directory walks the given directory and returning the checksum of all the files in it.
func Directory(root string, h crypto.Hash) ([]*FileChecksum, []error) {
	if _, err := os.Stat(root); err != nil {
		return nil, []error{newError(root, ErrCannotBeRead, err)}
	}
	padlock := new(sync.Mutex)
	var folderSums []*FileChecksum
	var allErrors []error
	walkErr := powerwalk.Walk(root, func(path string, info os.FileInfo, intErr error) error {
		if intErr != nil {
			allErrors = append(allErrors, newError(path, ErrWalking, intErr))
			return nil
		}
		if !info.IsDir() {
			fileChecksum, fileErr := File(path, h)
			if fileErr != nil {
				allErrors = append(allErrors, fileErr)
				return nil
			}
			padlock.Lock()
			defer padlock.Unlock()
			folderSums = append(folderSums, fileChecksum)
		}
		return nil
	})
	if walkErr != nil {
		allErrors = append(allErrors, newError(root, ErrWalking, walkErr))
	}
	return folderSums, allErrors
}
