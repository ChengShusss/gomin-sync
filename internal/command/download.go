package command

import (
	"fmt"
	"gomin-sync/internal/config"
	"os"

	"github.com/spf13/pflag"
)

func downloadDir(localPath, remotePrefix string) {

}

func Download() {
	config.LoadConfig("")

	var remotePrefix string
	pflag.StringVarP(&remotePrefix, "remotePrefix", "p", "", "remote prefix add to path")
	pflag.BoolVarP(&config.Force, "forceUpload", "f", false, "force to upload files")
	pflag.CommandLine.Parse(os.Args[2:])

	left := pflag.Args()
	if len(left) == 0 {
		fmt.Printf("Please specific local dir download to\n")
		os.Exit(1)
	}
	if len(left) > 1 {
		fmt.Printf("too many path is given\n")
		os.Exit(1)
	}

	// downloadDir(left[0], remotePrefix)
	config.GetBool("1")
}
