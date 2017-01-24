package main

import (
	"fmt"
	"os"

	cli "github.com/jawher/mow.cli"
)

func main() {
	app := cli.App("checksum", "Get the checksum of a file(s)")
	app.Version("v version", "0.0.1")
	app.Command("get", "Get the checksum of a file or files", func(cmd *cli.Cmd) {
		cmd.Spec = "[-r] [-o] ALG [SRC...]"
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
	app.Run(os.Args)
}

func getCommand(output, alg *string, recursive *bool, srcs *[]string) func() {
	return func() {
		fmt.Printf("Checksum of %v in %s [output: %v] [recursively: %v]\n", *srcs, *alg, *output, *recursive)
	}
}

func checkCommand(recursive *bool, checksumFile *string, srcs *[]string) func() {
	return func() {
		fmt.Printf("Checking %v against %s [recursive: %v]\n", *srcs, *checksumFile, *recursive)
	}
}
