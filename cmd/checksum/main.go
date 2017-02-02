package main

import (
	"fmt"
	"os"

	"crypto"
	_ "crypto/md5"
	_ "crypto/sha1"

	"strings"

	cli "github.com/jawher/mow.cli"
)

var (
	supported = make(map[string]crypto.Hash)
)

func init() {
	supported["sha1"] = crypto.SHA1
	supported["md5"] = crypto.MD5
}

func main() {
	app := cli.App("checksum", "Get the checksum of a file, files, or directory")
	app.Version("v version", "0.0.1")
	app.Command("get", "Get the checksum of a file or files", func(cmd *cli.Cmd) {
		cmd.Spec = "[-r] [-o] ALG SRC..."
		alg := cmd.StringArg("ALG", "", "The algorithm to use for the checksum")
		recursive := cmd.BoolOpt("r recursive", false, "Get the checksum of folders and files recursively")
		output := cmd.StringOpt("o output", "", "The output file of the checksum(s)")
		srcs := cmd.StringsArg("SRC", nil, "File(s) or directory to get the checksum of")
		cmd.Action = getCommand(output, alg, recursive, srcs)
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

func getCommand(output, alg *string, recursive *bool, srcs *[]string) func() {
	return func() {
		err := validateSources(*recursive, *srcs)
		if err != nil {
			fmt.Println(err)
			cli.Exit(1)
		}
		fmt.Printf("Checksum of %v in %s [output: %v] [recursively: %v]\n", *srcs, *alg, *output, *recursive)
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
