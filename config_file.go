package notes

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// FileConfig holds per-user settings loaded from $NOTES_CLI_HOME/.notes-cli.toml.
// All fields have defaults; a missing file is not an error.
// Environment variables (NOTES_CLI_FZF, NOTES_CLI_BAT, NOTES_CLI_FZF_PREVIEW_WINDOW)
// take precedence over values in this file.
type FileConfig struct {
	FzfCmd           string `toml:"fzf_cmd"`
	BatCmd           string `toml:"bat_cmd"`
	FzfPreviewWindow string `toml:"fzf_preview_window"`
}

func defaultFileConfig() FileConfig {
	return FileConfig{
		FzfCmd:           "fzf",
		BatCmd:           "bat",
		FzfPreviewWindow: "up:60%",
	}
}

// loadFileConfig reads the TOML config file from homePath/.notes-cli.toml.
// Missing file or unrecognised fields are silently ignored; defaults are preserved.
func loadFileConfig(homePath string) FileConfig {
	cfg := defaultFileConfig()
	p := filepath.Join(homePath, ".notes-cli.toml")
	if _, err := os.Stat(p); err != nil {
		return cfg
	}
	// Ignore decode errors so that partially-valid files still apply valid fields.
	toml.DecodeFile(p, &cfg) //nolint:errcheck
	return cfg
}
