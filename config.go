package notes

import (
	"github.com/pkg/errors"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
)

// Config represents user configuration of notes command
type Config struct {
	// HomePath is a file path to directory of home of notes command. If $NOTES_CLI_HOME is set, it is used.
	// Otherwise, notes-cli directory in XDG data directory is used. This directory is automatically created
	// when config is created
	HomePath string
	// GitPath is a file path to `git` executable. If $NOTES_CLI_GIT is set, it is used.
	// Otherwise, `git` is used by default. This is optional and can be empty. When empty, some command
	// and functionality which require Git don't work
	GitPath string
	// EditorCmd is a command of your favorite editor. If $NOTES_CLI_EDITOR is set, it is used. This value is
	// similar to $EDITOR environment variable and can contain command arguments like "vim -g". Otherwise,
	// this value will be empty. When empty, some functionality which requires an editor to open note doesn't
	// work
	EditorCmd string
}

func homePath() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", errors.Wrap(err, "Cannot locate home directory. Please set $NOTES_CLI_HOME")
	}

	if env := os.Getenv("NOTES_CLI_HOME"); env != "" {
		if strings.HasPrefix(env, "~"+string(filepath.Separator)) {
			env = filepath.Join(u.HomeDir, env[2:])
		}
		return filepath.Clean(env), nil
	}

	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "notes-cli"), nil
	}

	if runtime.GOOS == "windows" {
		if env := os.Getenv("APPLOCALDATA"); env != "" {
			return filepath.Join(env, "notes-cli"), nil
		}
	}

	return filepath.Join(u.HomeDir, ".local", "share", "notes-cli"), nil
}

func gitPath() string {
	c := "git"
	if env, ok := os.LookupEnv("NOTES_CLI_GIT"); ok {
		c = env // Removed filepath.Clean here — avoid unnecessary path normalization
	}

	// Use LookPath only to check availability. Return the command name as-is
	// so the OS resolves it via PATH at exec time. Storing the resolved absolute
	// path would break on NixOS and similar systems where store paths change on
	// every upgrade.
	if _, err := exec.LookPath(c); err != nil {
		// Git is optional
		return ""
	}

	return c // Return c (command name), not the resolved absolute path
}

func editorCmd() string {
	if env, ok := os.LookupEnv("NOTES_CLI_EDITOR"); ok {
		return env
	}
	if env, ok := os.LookupEnv("EDITOR"); ok {
		return env
	}
	return ""
}

// NewConfig creates a new Config instance by looking the user's environment. GitPath and EditorPath
// may be empty when proper configuration is not found. When home directory path cannot be located,
// this function returns an error
func NewConfig() (*Config, error) {
	h, err := homePath()
	if err != nil {
		return nil, err
	}

	// Ensure home directory exists
	if err := os.MkdirAll(h, 0755); err != nil {
		return nil, errors.Wrapf(err, "Could not create home '%s'", h)
	}

	return &Config{
		HomePath:  h,
		GitPath:   gitPath(),
		EditorCmd: editorCmd(),
	}, nil
}
