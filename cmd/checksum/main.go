package main

import (
	"fmt"
	"os"
	"time"

	"crypto"
	_ "crypto/md5"
	_ "crypto/sha1"
	_ "crypto/sha512"

	"strings"

	"sync"

	"bufio"

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
		cmd.Spec = "[-r] [-o] [-e] HASH SRC..."
		hash := cmd.StringArg("HASH", "", "The hash to use for the checksum")
		recursive := cmd.BoolOpt("r recursive", false, "Get the checksum of folders and files recursively")
		output := cmd.StringOpt("o output", "", "The output file of the checksum(s)")
		errOutput := cmd.StringOpt("e errors", "", "The output file for the errors")
		srcs := cmd.StringsArg("SRC", nil, "File(s) or directory to get the checksum of")
		cmd.Action = getCommand(output, errOutput, hash, recursive, srcs)
	})
	app.Command("list", "List all the possable hashes to use", func(cmd *cli.Cmd) {
		cmd.Spec = ""
		cmd.Action = listCommand()
	})
	app.Run(os.Args)
}

func getCommand(output, errors, hash *string, recursive *bool, srcs *[]string) func() {
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

		var outBuff *bufio.Writer
		var errBuff *bufio.Writer
		if *output != "" {
			f, err := os.Create(*output)
			if err != nil {
				fmt.Printf("Could not create the output file: %s Inner: %v\n", *output, err)
				cli.Exit(1)
			}
			defer f.Close()
			outBuff = bufio.NewWriter(f)
			fmt.Fprintf(outBuff, "[Config]\nhash=%s\n[Files]\n", *hash)
		} else {
			outBuff = bufio.NewWriter(os.Stdout)
		}
		defer outBuff.Flush()
		if *errors != "" {
			f, err := os.Create(*errors)
			if err != nil {
				fmt.Printf("Could not create the error file: %s Inner: %v\n", *errors, err)
				cli.Exit(1)
			}
			defer f.Close()
			errBuff = bufio.NewWriter(f)
			fmt.Fprintf(errBuff, "[Errors]\n")
		} else {
			errBuff = bufio.NewWriter(os.Stderr)
		}
		defer errBuff.Flush()

		respWG := new(sync.WaitGroup)
		response := make(chan checksum.FileChecksumResponse)
		defer close(response)
		go func() {
			for {
				select {
				case resp, ok := <-response:
					respWG.Add(1)
					if !ok {
						respWG.Done()
						continue
					}
					if resp.Checksum != nil {
						fmt.Fprintf(outBuff, "%s=%x\n", resp.Checksum.FilePath, resp.Checksum.Checksum)
					}
					if resp.Err != nil {
						if e, ok := resp.Err.(checksum.CSError); ok {
							fmt.Fprintf(errBuff, "%s=%s\n", e.Path, e.Error())
						} else {
							fmt.Fprintf(errBuff, "%s=%s\n", time.Now().Format(time.RFC3339), resp.Err.Error())
						}
					}
					respWG.Done()
				}
			}
		}()
		getChecksums(hashMethod, *srcs, response)
		respWG.Wait()
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
