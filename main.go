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
	slog.Info("main.go", "Run", "init complete")

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
		slog.Info("main.go", "ghAuthSwitch", "switch to "+profile.Name)
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
		}

		sw.Profiles[profile.Name] = profile

		if profile.Active {
			profile.Name += "*"
			sw.Current = profile
			sw.setCurrentGitConfig()
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

func (sw *Switcheroo) setCurrentGitConfig() {
	sw.Current.Config = sw.APPDIR + filepath.Join("/", sw.Current.Name[:len(sw.Current.Name)-1])
	err := os.Setenv("GIT_CONFIG_GLOBAL", sw.Current.Config)
	if err != nil {
		slog.Error("main.go", "setCurrentGitConfig", err.Error())
		os.Exit(1)
	}
	slog.Info("main.go", "setCurrentGitConfig", "current profile: "+sw.Current.Config)
}

func createGitConfigDir(configDir string) {
	var err error
	err = os.Mkdir(configDir, 0755)
	if err != nil {
		slog.Warn("main.go", "init", "app dir failed to create: "+err.Error())
		return
	}
	slog.Info("main.go", "init", "app dir created: "+configDir)
}

// APPDIR will be set via var or config file
const gitConfigs = "~/.local/share/switcheroo"

func main() {
	slog.Info("main.go", "main", "start switcheroo")
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
