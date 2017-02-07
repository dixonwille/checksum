/*Package checksum gets the checksum of a file or files withing a directory.

You will need to import the appropriate package in you package to be able to get the checksum of the files.
Check out https://golang.org/pkg/crypto/#Hash to get a list of possable checksums and which imports to include in your package.
You will need to do something like:
	import _ "crypto/md5"
to import the appropriate hash into your package.
*/
package checksum

//FileChecksum is the checksum of a file.
type FileChecksum struct {
	FilePath string
	Checksum []byte
}

//FileChecksumResponse is a response from this library over the channel.
type FileChecksumResponse struct {
	Checksum *FileChecksum
	Err      error
}

//NewResponse creates a new response for the channel.
func NewResponse(checksum *FileChecksum, err error) FileChecksumResponse {
	return FileChecksumResponse{
		Checksum: checksum,
		Err:      err,
	}
}
