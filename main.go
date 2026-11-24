package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/cli/go-gh/v2"
	_ "github.com/cli/go-gh/v2/pkg/api"
)

type Profile struct {
	Name    string
	Config  string
	Active  bool
	Default bool
}

type Switcheroo struct {
	APPDIR   string
	Profiles map[string]*Profile
	Aliases  map[string]string

	Verbose *bool
	Current *Profile

	Config *Config
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
	sw.Verbose = flag.Bool("v", false, "verbose")
	flag.Parse()

	logLevel := slog.LevelInfo

	if *sw.Verbose {
		logLevel = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	handler := slog.NewTextHandler(os.Stderr, opts)
	slog.SetDefault(slog.New(handler))

	if err := createGitConfigDir(sw.APPDIR); err != nil {
		slog.Error("failed to create appdir", "error:", err.Error())
		os.Exit(1)
	}
	slog.Debug("main.go", "Run", "created git config dir")

	if err := sw.ghAuthStatus(); err != nil {
		slog.Error(err.Error())
	}
	slog.Debug("main.go", "Run", "got gh status")
	slog.Debug("main.go", "Run", "init complete")

	if len(os.Args) > 1 {
		if flag.Arg(0) == "list" {
			sw.ListProfiles()
		}
		if flag.Arg(0) == "current" {
			sw.ListCurrentProfile()
			return
		}
		if flag.Arg(0) == "write" {
			if err := writeGitConfigFile("yolo"); err != nil {
				slog.Error("main.go", "error:", err.Error())
			}
		}
		if flag.Arg(0) == "use" {
			if flag.Arg(1) != "" {
				if profile, ok := sw.Aliases[flag.Arg(1)]; ok {
					sw.ghAuthSwitch(profile)
					return
				}
				sw.ghAuthSwitch(flag.Arg(1))
			} else {
				slog.Error("main.go", "Run", "'use' requires an arg")
			}
		}
	}
}

func ghAuthStatusReadProfiles() ([]Host, error) {
	var auth Auth

	if stdout, _, err := gh.Exec("auth", "status", "--json", "hosts"); err == nil {
		err = json.Unmarshal(stdout.Bytes(), &auth)
		if err != nil {
			return nil, fmt.Errorf("failed to get auth status %s", err)
		}
	}
	return auth.Hosts.Github, nil
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

func (sw *Switcheroo) ghAuthStatus() error {
	profiles, err := ghAuthStatusReadProfiles()
	if err != nil {
		return err
	}

	sw.Profiles = make(map[string]*Profile, len(profiles))

	for _, authProfile := range profiles {
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
		}
	}
	return nil
}

// isn't writing anything right now
func writeGitConfigFile(path string) error {
	cf, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE|os.O_EXCL, 0755)
	if errors.Is(err, os.ErrExist) {
		slog.Debug("main.go", "writeGitConfigFile", "skipping "+err.Error())
		return nil
	}
	if err != nil {
		return err
	}
	defer cf.Close()
	return nil
}

func (sw Switcheroo) ListCurrentProfile() {
	slog.Debug("main.go", "ListCurrentProfile", "current: "+sw.Current.Name)
	fmt.Println(sw.Current.Name)
}

func (sw Switcheroo) ListProfiles() {
	for _, profile := range sw.Profiles {
		slog.Debug("main.go", "ListProfiles", "profile: "+profile.Name)
		fmt.Println(profile.Name)
	}
}

func (sw *Switcheroo) setCurrentGitConfig() {
	slog.Debug("main.go", "setCurrentGitConfig", "current profile: "+sw.Current.Config)
	fmt.Printf("export GIT_CONFIG_GLOBAL=%s\n", sw.Current.Config)
}

func createGitConfigDir(configDir string) error {
	var err error
	err = os.Mkdir(configDir, 0755)

	if errors.Is(err, os.ErrExist) {
		slog.Debug("main.go", "createGitConfigDir", "skipping: "+err.Error())
		return nil
	}

	return err
}

func main() {
	slog.Debug("main.go", "main", "start switcheroo")
	config, err := validateConfig()
	if err != nil {
		slog.Error("main.go", "main", err.Error())
		os.Exit(1)
	}

	aliases := make(map[string]string, len(config.Switcheroo.GitConfigs))

	for _, alias := range config.Switcheroo.Aliases {
		aliases[alias.Name] = alias.Account
	}

	Switcheroo{
		APPDIR:  config.Switcheroo.GitConfigs,
		Aliases: aliases,
	}.Run()
}
