package checksum

import (
	"bufio"
	"crypto"
	"fmt"
	"io"
	"os"

	"github.com/stretchr/powerwalk"

	"path/filepath"
	"sync"
)

const bufferSize = 4096

//File is used to get the checksum of the file.
func File(path string, h crypto.Hash) (*FileChecksum, error) {
	if !h.Available() {
		return nil, NewError(path, ErrHashBinary, nil)
	}
	hash := h.New()
	if info, err := os.Stat(path); err != nil {
		return nil, NewError(path, ErrCannotBeRead, err)
	} else if info.IsDir() {
		return nil, NewError(path, ErrFileIsDirectory, nil)
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, NewError(path, ErrFileNotOpen, err)
	}
	defer file.Close()
	buf := make([]byte, bufferSize)
	fileReader := bufio.NewReader(file)
	for {
		if n, err := fileReader.Read(buf); err == nil {
			nw, err := hash.Write(buf[:n])
			if err != nil {
				return nil, NewError(path, ErrWritingHash, err)
			}
			if nw != n {
				fmt.Printf("ReadBits: %d\nWriteBits: %d\n", n, nw)
				return nil, NewError(path, ErrImproperLength, nil)
			}
		} else if err == io.EOF {
			break
		} else {
			return nil, NewError(path, ErrFileReading, err)
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
		return nil, []error{NewError(root, ErrCannotBeRead, err)}
	}
	var folderSums []*FileChecksum
	var allErrors []error
	errLock := new(sync.Mutex)
	sumLock := new(sync.Mutex)
	walkErr := powerwalk.Walk(root, func(path string, info os.FileInfo, intErr error) error {
		if intErr != nil {
			errLock.Lock()
			allErrors = append(allErrors, NewError(path, ErrWalking, intErr))
			errLock.Unlock()
			return nil
		}
		if !info.IsDir() {
			fileChecksum, fileErr := File(path, h)
			if fileErr != nil {
				errLock.Lock()
				allErrors = append(allErrors, fileErr)
				errLock.Unlock()
				return filepath.SkipDir
			}
			sumLock.Lock()
			folderSums = append(folderSums, fileChecksum)
			sumLock.Unlock()
		}
		return nil
	})
	if walkErr != nil {
		allErrors = append(allErrors, NewError(root, ErrWalking, walkErr))
	}
	return folderSums, allErrors
}
