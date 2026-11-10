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
	Profiles map[string]*Profile

	Current *Profile
}

func (sw Switcheroo) Run() {
	//sw.init()

	sw.ghAuthStatus()
	slog.Info("main.go", "Run", "init complete")

	if len(os.Args) > 1 {
		if os.Args[1] == "list" {
			sw.ListProfiles()
		}
		if os.Args[1] == "current" {
			sw.ListCurrentProfile()
		}
		if os.Args[1] == "use" {
			sw.ghAuthSwitch(os.Args[2])
		}

	}
}

/*
sw list		   // show curents of ~/.local/shared/switcheroo
sw use x       // sets sw.Current
sw use system  // sets sw.Current to system ~/.gitconfig
*/

type Host struct {
	State       string `json:"state"`
	Active      bool   `json:"active"`
	Host        string `json:"host"`
	Login       string `json:"login"`
	TokenSource string `json:"tokenSource"`
	Scopes      string `json:"scope"`
	GitProtocol string `json:"gitProtocol"`
}

// taken out github biy
type Auth struct {
	Hosts struct {
		Github []Host `json:"github.com"`
	} `json:"hosts"`
}

type Profile struct {
	Name   string
	Config string
	Active bool
}

func ghAuthSwitchProfile(name string) {
	_, _, err := gh.Exec("auth", "switch", "-u", name)
	if err != nil {
		slog.Error("main.go", "ghAuthSwitchProfile", err.Error())
	}
}

func ghAuthStatusReadProfiles() []Host {
	var auth Auth

	stdout, _, err := gh.Exec("auth", "status", "--json", "hosts")
	if err != nil {
		slog.Error("main.go", "ghAuthStatusReadProfiles", err.Error())
	}
	err = json.Unmarshal(stdout.Bytes(), &auth)
	if err != nil {
		slog.Error("main.go", "ghAuthStatusReadProfiles", err.Error())
	}

	return auth.Hosts.Github
}

func (sw *Switcheroo) ghAuthSwitch(name string) {
	if profile, ok := sw.Profiles[name]; ok {
		ghAuthSwitchProfile(name)
		slog.Info("main.go", "ghAuthSwitch", "switch to "+profile.Name)
		sw.Current = profile
		return
	}

	slog.Error("main.go", "ghAuthSwitch", name+" is not a valid profile")
	os.Exit(1)
}

func (sw *Switcheroo) ghAuthStatus() {
	profiles := ghAuthStatusReadProfiles()
	sw.Profiles = make(map[string]*Profile, len(profiles))

	for _, authProfile := range ghAuthStatusReadProfiles() {
		var profile Profile

		profile.Active = authProfile.Active
		profile.Name = authProfile.Login

		sw.Profiles[profile.Name] = &profile

		if profile.Active {
			profile.Name += "*"
			sw.Current = &profile
		}
	}

}
func (sw Switcheroo) ListCurrentProfile() {
	fmt.Println(sw.Current.Name)
}

func (sw Switcheroo) ListProfiles() {
	for _, profile := range sw.Profiles {
		fmt.Println(profile.Name)
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
