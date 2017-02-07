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
		checksums, errs := channelResponse(getChannel) //ignore errs so we can put it in the output
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
		err = iniFormatChecksums(w, *hash, checksums, cleanErrors(errs))
		if err != nil {
			fmt.Println(err)
			cli.Exit(1)
		}
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

//Parse the channel for errors and responses and return the summation of it all
func channelResponse(resChan <-chan checksumResponse) ([]*checksum.FileChecksum, []error) {
	var checksums []*checksum.FileChecksum
	var allErrors []error
	for response := range resChan {
		if response.errs != nil && len(response.errs) > 0 {
			allErrors = append(allErrors, response.errs...)
		}
		if response.checksums != nil && len(response.checksums) > 0 {
			checksums = append(checksums, response.checksums...)
		}
	}
	return checksums, allErrors
}

//Basically takes a bunch of channels and outputs them to a single channel
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

func cleanErrors(errs []error) []error {
	var newErrs []error
	for _, e := range errs {
		if e != nil {
			newErrs = append(newErrs, e)
		}
	}
	return newErrs
}
