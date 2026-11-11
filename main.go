package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/cli/go-gh/v2"
	_ "github.com/cli/go-gh/v2/pkg/api"
)

/*
sw list		   // show curents of ~/.local/shared/switcheroo
sw use x       // sets sw.Current
sw use system  // sets sw.Current to system ~/.gitconfig
*/

type Profile struct {
	Name   string
	Config string
	Active bool
}

type Switcheroo struct {
	APPDIR   string
	Profiles map[string]*Profile

	Verbose *bool
	Current *Profile
}

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

func (sw Switcheroo) Run() {
	createGitConfigDir(sw.APPDIR)

	sw.ghAuthStatus()

	sw.Verbose = flag.Bool("v", false, "verbose")
	flag.Parse()

	slog.Debug("main.go", "Run", "init complete")

	if len(os.Args) > 1 {
		if os.Args[1] == "list" {
			sw.ListProfiles()
		}
		if os.Args[1] == "current" {
			sw.ListCurrentProfile()
		}
		if os.Args[1] == "use" {
			if len(os.Args[2:]) > 0 {
				sw.ghAuthSwitch(os.Args[2])
			} else {
				slog.Error("main.go", "Run", "'use' requires an arg")
			}
		}
	}
}

func ghAuthStatusReadProfiles() []Host {
	var auth Auth

	if stdout, _, err := gh.Exec("auth", "status", "--json", "hosts"); err == nil {
		err = json.Unmarshal(stdout.Bytes(), &auth)
		if err != nil {
			slog.Error("main.go", "ghAuthStatusReadProfiles", err.Error())
			os.Exit(1)
		}
	} else {
		slog.Error("main.go", "ghAuthStatusReadProfiles", err.Error())
	}
	return auth.Hosts.Github
}

func ghAuthSwitchProfile(name string) {
	if _, _, err := gh.Exec("auth", "switch", "-u", name); err != nil {
		slog.Error("main.go", "ghAuthSwitchProfile", err.Error())
	}
	if _, _, err := gh.Exec("auth", "setup-git"); err != nil {
		slog.Error("main.go", "ghAuthSwitchProfile", err.Error())
	}
}

func (sw *Switcheroo) ghAuthSwitch(name string) {
	if profile, ok := sw.Profiles[name]; ok {
		ghAuthSwitchProfile(name)
		slog.Debug("main.go", "ghAuthSwitch", "switch to "+profile.Name)
		sw.Current = profile
		sw.setCurrentGitConfig()
		return
	}
	slog.Error("main.go", "ghAuthSwitch", name+" is not a valid profile")
	os.Exit(1)
}

func (sw *Switcheroo) ghAuthStatus() {
	profiles := ghAuthStatusReadProfiles()
	sw.Profiles = make(map[string]*Profile, len(profiles))

	for _, authProfile := range ghAuthStatusReadProfiles() {
		profile := &Profile{
			Active: authProfile.Active,
			Name:   authProfile.Login,
			Config: sw.APPDIR + filepath.Join("/", authProfile.Login),
		}

		sw.Profiles[profile.Name] = profile

		writeGitConfigFile(profile.Config)

		if profile.Active {
			profile.Name += "*"
			sw.Current = profile
			sw.setCurrentGitConfig()
		}
	}
}

func writeGitConfigFile(path string) {
	cf, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE|os.O_EXCL, 0755)
	if errors.Is(err, os.ErrExist) {
		slog.Debug("main.go", "writeGitConfigFile", "skipping "+err.Error())
		return
	}
	if err != nil {
		slog.Error("main.go", "writeGitConfigFile", err.Error())
		return
	}
	defer cf.Close()
}

func (sw Switcheroo) ListCurrentProfile() {
	fmt.Println(sw.Current.Name)
}

func (sw Switcheroo) ListProfiles() {
	for _, profile := range sw.Profiles {
		fmt.Println(profile.Name)
	}
}

func (sw *Switcheroo) setCurrentGitConfig() {
	fmt.Printf("export GIT_CONFIG_GLOBAL=%s\n", sw.Current.Config)
	slog.Debug("main.go", "setCurrentGitConfig", "current profile: "+sw.Current.Config)
}

func createGitConfigDir(configDir string) {
	var err error
	err = os.Mkdir(configDir, 0755)
	if errors.Is(err, os.ErrExist) {
		slog.Debug("main.go", "createGitConfigDir", "skipping: "+err.Error())
		return
	}
	if err != nil {
		slog.Error("main.go", "createGitConfigDir", "app dir failed to create: "+err.Error())
		return
	}
	slog.Debug("main.go", "createGitConfigDir", "app dir created: "+configDir)
}

// APPDIR will be set via var or config file
const gitConfigs = "~/.local/share/switcheroo"

func main() {
	slog.Debug("main.go", "main", "start switcheroo")
	appDir := gitConfigs

	if strings.HasPrefix(gitConfigs, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			slog.Error("main.go", "main", "failed to get homedir")
			os.Exit(1)
		}
		appDir = filepath.Join(home, gitConfigs[2:])
	}

	Switcheroo{
		APPDIR: appDir,
	}.Run()
}
