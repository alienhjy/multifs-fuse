package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type strSlice []string

func (slf *strSlice) String() string {
	return fmt.Sprint(*slf)
}

func (slf *strSlice) Set(value string) error {
	for _, tmp := range *slf {
		if value == tmp {
			return errors.New("Duplicate flags")
		}
	}
	*slf = append(*slf, strings.Split(value, ",")...)
	return nil
}

type mountFlags struct {
	src string
	des string
	opt []string
}

// func optHelpText() string {
// 	var (
// 		text string
// 	)
// 	text = `option:
// 	ro
// 		Mount filesystem read-only.
// 	`
// 	return text
// }

func usage() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, filepath.Base(os.Args[0]), "[-o option[,...]] <master_path> <mount_point>")
	fmt.Fprintln(os.Stderr, `
Option:
	ro
		Mount filesystem as read-only.
	slave=dir1,slave=dir2,slave=dir3,...
		Slave backends. If there is more than one slave backend, use this option
		multiple times. This filesystem will read from those backends as the order
		of those slave option.
	`)
	return
}

// flagParse() parse the flag of mount util.
func flagParse() error {
	var (
		strMntOpt   strSlice
		nonFlagArgs []string
		tmpOSArgs   []string
		err         error
	)
	tmpOSArgs = os.Args
	flag.Usage = usage
	flag.Var(&strMntOpt, "o", "Assign mount options.")
	flag.Parse()
	os.Args = append(os.Args[:1], flag.Args()...)
	flag.Parse()
	for len(flag.Args()) != 0 {
		nonFlagArgs = append(nonFlagArgs, flag.Arg(0))
		os.Args = append(os.Args[:1], os.Args[2:]...)
		flag.Parse()
		os.Args = append(os.Args[:1], flag.Args()...)
		flag.Parse()
	}

	if len(nonFlagArgs) == 2 {
		mntFlags.src = nonFlagArgs[0]
		mntFlags.des = nonFlagArgs[1]
		mntFlags.opt = strMntOpt
	} else {
		flag.Usage()
		log.Println("flag.Args():", flag.Args())
		log.Println("strMntOpt:", strMntOpt)
	}

	os.Args = tmpOSArgs

	return err
}

func optToCmd() []string {
	var (
		cmd []string
	)
	cmd = append(cmd, "./bin/multifsd")

	cmd = append(cmd, "-master="+mntFlags.src)
	cmd = append(cmd, "-target="+mntFlags.des)
	for _, opt := range mntFlags.opt {
		opts := strings.SplitN(opt, "=", 2)
		switch opts[0] {
		case "ro":
			cmd = append(cmd, "-ro")
		case "cor":
			cmd = append(cmd, "-cor")
		case "slave":
			cmd = append(cmd, "-slaves="+opts[1])
		default:
			log.Println("WARING: mount option", opt, "not supported.")
		}
	}

	return cmd
}
