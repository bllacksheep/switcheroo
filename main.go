package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	_ "strconv"
	"strings"

	"github.com/cli/go-gh/v2"
	_ "github.com/cli/go-gh/v2/pkg/api"
)

type Switcheroo struct {
	APPDIR   string
	Profiles []Profile

	Current *Profile
}

func (sw Switcheroo) Run() {
	//sw.init()

	slog.Info("main.go", "Run", "init complete")

	if len(os.Args) > 1 {
		if os.Args[1] == "list" {
			sw.ListProfiles()
		}

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

type Auth struct {
	Hosts struct {
		Host []struct {
			State       string `json:"state"`
			Active      bool   `json:"active"`
			Host        string `json:"host"`
			Login       string `json:"login"`
			TokenSource string `json:"tokenSource"`
			Scopes      string `json:"scope"`
			GitProtocol string `json:"gitProtocol"`
		} `json:"github.com"`
	} `json:"hosts"`
}

type Profile struct {
	Name   string
	Config string
	Active bool
}

func (sw *Switcheroo) readProfiles() {
	var auth Auth
	stdout, _, err := gh.Exec("auth", "status", "--json", "hosts")
	if err != nil {
		slog.Error("main.go", "listProfiles", err.Error())
	}
	err = json.Unmarshal(stdout.Bytes(), &auth)
	if err != nil {
		slog.Error("main.go", "listProfiles", err.Error())

	}

	for _, host := range auth.Hosts.Host {
		var profile Profile

		profile.Active = host.Active
		profile.Name = host.Login

		sw.Profiles = append(sw.Profiles, profile)
		if profile.Active {
			sw.Current = &profile
		}
	}

}
func (sw Switcheroo) ListCurrentProfile() {
	sw.readProfiles()
	fmt.Println(sw.Current)
}

func (sw Switcheroo) ListProfiles() {
	sw.readProfiles()

	for _, profile := range sw.Profiles {
		var current string
		if profile.Active {
			current = "*"
		}
		fmt.Println(profile.Name + current)
	}
}

/*
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
*/

/*
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
*/

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

	Switcheroo{
		APPDIR: appdir,
	}.Run()
}
