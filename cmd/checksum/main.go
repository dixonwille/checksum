package main

import (
	"bufio"
	"fmt"
	"os"
	"reflect"

	"crypto"
	_ "crypto/md5"
	_ "crypto/sha1"
	_ "crypto/sha512"

	"strings"

	"github.com/dixonwille/checksum"
	cli "github.com/jawher/mow.cli"
)

var (
	supported = make(map[string]crypto.Hash)
)

func init() {
	supported["sha1"] = crypto.SHA1
	supported["md5"] = crypto.MD5
	supported["sha512"] = crypto.SHA512
}

func main() {
	app := cli.App("checksum", "Get the checksum of a file, files, or directory")
	app.Version("v version", "0.0.1")
	app.Command("get", "Get the checksum of a file or files", func(cmd *cli.Cmd) {
		cmd.Spec = "[-r] [-o] HASH SRC..."
		hash := cmd.StringArg("HASH", "", "The hash to use for the checksum")
		recursive := cmd.BoolOpt("r recursive", false, "Get the checksum of folders and files recursively")
		output := cmd.StringOpt("o output", "", "The output file of the checksum(s)")
		srcs := cmd.StringsArg("SRC", nil, "File(s) or directory to get the checksum of")
		cmd.Action = getCommand(output, hash, recursive, srcs)
	})
	app.Command("check", "Check the checksum of local files to a output file and return differing files", func(cmd *cli.Cmd) {
		cmd.Spec = "[-r] CHECKSUMFILE [SRC...]"
		recursive := cmd.BoolOpt("r recursive", false, "Compare the checksum of folders and files recursively")
		checksum := cmd.StringArg("CHECKSUMFILE", "", "The checksum file to check your files against")
		srcs := cmd.StringsArg("SRC", nil, "The files to check against the Checksum file")
		cmd.Action = checkCommand(recursive, checksum, srcs)
	})
	app.Command("list", "List all the possable hashes to use", func(cmd *cli.Cmd) {
		cmd.Spec = ""
		cmd.Action = listCommand()
	})
	app.Run(os.Args)
}

func getCommand(output, hash *string, recursive *bool, srcs *[]string) func() {
	return func() {
		err := validateSources(*recursive, *srcs)
		if err != nil {
			fmt.Println(err)
			cli.Exit(1)
		}
		hashMethod, ok := supported[*hash]
		if !ok {
			fmt.Printf("The hash provided is not supported: %s\n", *hash)
			cli.Exit(1)
		}
		getChannel := getChecksums(hashMethod, *srcs)
		checksums, ok := channelResponse(getChannel)
		if !ok {
			fmt.Println("Could not get the checksum of all the files")
			cli.Exit(1)
		}
		var w *bufio.Writer
		if *output == "" {
			w = bufio.NewWriter(os.Stdout)
		} else {
			file, err := os.Create(*output)
			if err != nil {
				fmt.Println(err)
				cli.Exit(1)
			}
			w = bufio.NewWriter(file)
		}
		err = iniFormatChecksums(w, *hash, checksums)
		if err != nil {
			fmt.Println(err)
			cli.Exit(1)
		}
	}
}

func checkCommand(recursive *bool, checksumFile *string, srcs *[]string) func() {
	return func() {
		tmpSrcs := append(*srcs, *checksumFile) //To make sure the checksumFile exists
		err := validateSources(*recursive, tmpSrcs)
		if err != nil {
			fmt.Println(err)
			cli.Exit(1)
		}
		fmt.Printf("Checking %v against %s [recursive: %v]\n", *srcs, *checksumFile, *recursive)
	}
}

func listCommand() func() {
	return func() {
		var help []string
		for k := range supported {
			help = append(help, k)
		}
		fmt.Printf("* %s\n", strings.Join(help, "\n* "))
	}
}

func validateSources(recursive bool, srcs []string) error {
	for _, src := range srcs {
		if info, err := os.Stat(src); os.IsNotExist(err) {
			return fmt.Errorf("%s: File not found\nInner Error: %v", src, err)
		} else if err != nil {
			return fmt.Errorf("%s: Something went wrong\nInner Error: %v", src, err)
		} else if info.IsDir() && !recursive {
			return fmt.Errorf("%s: Must include recursive if you want to get/check the checksum of files inside a folder", src)
		}
	}
	return nil
}

func channelResponse(resChan <-chan checksumResponse) ([]*checksum.FileChecksum, bool) {
	wasErr := false
	var checksums []*checksum.FileChecksum
	for response := range resChan {
		if response.err != nil {
			wasErr = true
			fmt.Println(response.err)
		}
		if response.checksums != nil {
			checksums = append(checksums, response.checksums...)
		}
	}
	return checksums, !wasErr
}

func channelHandler(resChan []<-chan checksumResponse) <-chan checksumResponse {
	mainChan := make(chan checksumResponse)
	go func() {
		defer close(mainChan)
		cases := make([]reflect.SelectCase, len(resChan))
		for i, ch := range resChan {
			cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
		}

		remaining := len(cases)
		for remaining > 0 {
			chosen, value, ok := reflect.Select(cases)
			if !ok {
				cases[chosen].Chan = reflect.ValueOf(nil)
				remaining--
				continue
			}
			var response checksumResponse
			if value.CanInterface() {
				var ok bool
				resInterface := value.Interface()
				response, ok = resInterface.(checksumResponse)
				if !ok {
					response = newChecksumResponse(nil, fmt.Errorf("Could not convert channel to checksum channel"))
				}
			} else {
				response = newChecksumResponse(nil, fmt.Errorf("Channel value can not impliment an interface"))
			}
			mainChan <- response
		}
	}()
	return mainChan
}
