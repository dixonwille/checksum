package checksum

import "errors"
import "fmt"

//CSError is any checksum Error.
type CSError struct {
	Path  string
	Err   error
	Inner error
}

var (
	//ErrCannotBeRead is returned when file or folder could not be read for meta information.
	ErrCannotBeRead = errors.New("This file/folder could not be read")
	//ErrFileReading is returned when while reading the file an error occured.
	ErrFileReading = errors.New("This file had an error while reading")
	//ErrFileNotOpen is returned when this package could not open the file for reading.
	ErrFileNotOpen = errors.New("This file could not be opened")
	//ErrFileIsDirectory is returned when using the wrong method when passing in a directory.
	ErrFileIsDirectory = errors.New("File is a directory and should use the Directory function")
	//ErrHashBinary is returned when the proper binary is not included in the executable for the given has.
	ErrHashBinary = errors.New("Supplied hash's binary is not included in executable")
	//ErrImproperLength is returned when the expected number of bytes is not written to the hashing function.
	ErrImproperLength = errors.New("Did not write expected number of bytes to the hashing function")
	//ErrWritingHash is returned when there was a problem writing bytes to the hashing function.
	ErrWritingHash = errors.New("Could not write bytes to the hashubg function")
	//ErrWalking is returned when there was an issue walking to the specified file or folder.
	ErrWalking = errors.New("There was a problem walking to this file/folder")
)

func NewError(path string, err, inner error) *CSError {
	return &CSError{
		Path:  path,
		Err:   err,
		Inner: inner,
	}
}

//Error is used to implement the Error interface.
func (e *CSError) Error() string {
	err := fmt.Sprintf("%s: %v", e.Path, e.Err)
	if e.Inner != nil {
		err += fmt.Sprintf(" Inner: %v", e.Inner)
	}
	return err
}
