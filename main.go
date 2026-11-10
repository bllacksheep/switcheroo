package main

import (
	_ "fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Swicheroo struct {
	APPDIR   string
	Profiles []os.DirEntry

	Current string
}

func (sw Swicheroo) Run() {
	sw.init()
	slog.Info("main.go", "Run", "init complete")

	args := os.Args[1:]

	if args[1] == "list" {
		sw.listProfiles()
	}

}

/*
sw list		   // show curents of ~/.local/shared/switcheroo
sw use x       // sets sw.Current
sw use system  // sets sw.Current to system ~/.gitconfig

sw create x
sw reanme x y
sw delete x
sw info x
*/

func (sw Swicheroo) listProfiles() {
	slog.Info("main.go", "listProfiles", sw.Current)
}

func (sw *Swicheroo) getCurrentProifle() {
	numSymLinks := 0
	for _, profile := range sw.Profiles {
		target := sw.APPDIR + filepath.Join("/", profile.Name())
		f, err := os.Lstat(target)
		if err != nil {
			slog.Error("main.go", "getCurrent", err.Error())
		}
		if f.Mode()&os.ModeSymlink != 0 {
			numSymLinks++
			sw.Current, err = os.Readlink(target)
			if err != nil {
				slog.Error("main.go", "getCurrent", err.Error())
			}
			if numSymLinks > 1 {
				slog.Error("main.go", "getCurrent", "found "+strconv.Itoa(numSymLinks)+" current profiles should only be 1")
				os.Exit(1)
			}
		}
	}
	slog.Info("main.go", "getCurrent", "current profile: "+sw.Current)
}

func (sw *Swicheroo) init() {
	var err error
	slog.Info("main.go", "init", "scan app dir: "+sw.APPDIR)
	sw.Profiles, err = os.ReadDir(sw.APPDIR)

	if err != nil {
		slog.Warn("main.go", "init", err.Error())
		err = os.Mkdir(sw.APPDIR, 0755)
		if err != nil {
			slog.Error("main.go", "init", "app dir failed to create: "+err.Error())
			os.Exit(1)
		}
		slog.Info("main.go", "init", "app dir created: "+sw.APPDIR)
	}

	sw.getCurrentProifle()
}

func main() {
	slog.Info("main.go", "main", "start switcheroo")

	// APPDIR will be set via var or config file
	appdir := "~/.local/share/switcheroo"

	if strings.HasPrefix(appdir, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			slog.Error("main.go", "main", "failed to get homedir")
			os.Exit(1)
		}
		appdir = filepath.Join(home, appdir[2:])
	}

	Swicheroo{
		APPDIR: appdir,
	}.Run()
}
