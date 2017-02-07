package main

import (
	"bufio"
	"fmt"

	"github.com/dixonwille/checksum"
)

func iniFormatChecksums(w *bufio.Writer, hash string, checksums []*checksum.FileChecksum, errs []error) error {
	defer w.Flush()
	_, err := fmt.Fprintf(w, "[Config]\nhash=%s\n[Files]\n", hash)
	if err != nil {
		return err
	}
	for _, cs := range checksums {
		_, err = fmt.Fprintf(w, "%s=%x\n", cs.FilePath, cs.Checksum)
		if err != nil {
			return err
		}
	}
	if errs != nil && len(errs) > 0 {
		_, err := fmt.Fprintf(w, "[Errors]\n")
		if err != nil {
			return err
		}
	}
	for i, e := range errs {
		_, err = fmt.Fprintf(w, "%d=%v", i, e)
	}
	return nil
}
