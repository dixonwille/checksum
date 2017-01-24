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

//File is used to get the checksum of the file
func File(path string, h crypto.Hash) (*FileChecksum, error) {
	hash := h.New()
	if info, err := os.Stat(path); err != nil {
		return nil, err
	} else if info.IsDir() {
		return nil, fmt.Errorf("Cannot get checksum of a directory. Please use the Directory function")
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	buf := make([]byte, bufferSize)
	fileReader := bufio.NewReader(file)
	for {
		if n, err := fileReader.Read(buf); err == nil {
			nw, err := hash.Write(buf[:n])
			if err != nil {
				return nil, err
			}
			if nw != n {
				fmt.Printf("ReadBits: %d\nWriteBits: %d\n", n, nw)
				return nil, fmt.Errorf("Did not write all the bytes to the hash function")
			}
		} else if err == io.EOF {
			break
		} else {
			return nil, err
		}
	}
	return &FileChecksum{
		FilePath: path,
		Checksum: hash.Sum(nil),
	}, nil
}

//Directory walks the given directory and returning the checksum of all the files in it
func Directory(root string, h crypto.Hash) ([]*FileChecksum, error) {
	if _, err := os.Stat(root); err != nil {
		return nil, err
	}
	padlock := new(sync.Mutex)
	var folderSums []*FileChecksum
	err := powerwalk.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			fileChecksum, err := File(path, h)
			if err != nil {
				return err
			}
			padlock.Lock()
			defer padlock.Unlock()
			folderSums = append(folderSums, fileChecksum)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return folderSums, nil
}
