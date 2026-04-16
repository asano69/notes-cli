package notes

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kballard/go-shellquote"
	"github.com/pkg/errors"
)

// fzfOptions configures a single fzf invocation.
type fzfOptions struct {
	Multi  bool
	Prompt string
	Header string
}

// buildFzfInput builds a tab-separated string to feed to fzf from a list of notes.
// Each line is: "<rel-path>\t<rel-path> <tags> <title>"
// fzf is configured with --with-nth=2 so users see the second field, while the
// first field (rel-path) is used internally to resolve the absolute path.
func buildFzfInput(notes []*Note) string {
	var b strings.Builder
	for _, n := range notes {
		rel := n.RelFilePath()
		tags := strings.Join(n.Tags, ",")
		fmt.Fprintf(&b, "%s\t%s %s %s\n", rel, rel, tags, n.Title)
	}
	return b.String()
}

// runFzf runs fzf with the given note-list input and returns the relative paths
// of the selected notes. If the user cancels (exit 1 or 130), nil is returned
// with no error.
func runFzf(cfg *Config, input string, opts fzfOptions) ([]string, error) {
	cmdline, err := shellquote.Split(cfg.FzfCmd)
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse fzf command ($NOTES_CLI_FZF / fzf_cmd)")
	}

	preview := fmt.Sprintf("%s --color=always %s/{1}", cfg.BatCmd, cfg.HomePath)

	args := append(cmdline[1:],
		"--delimiter=\t",
		"--with-nth=2",
		"--preview="+preview,
		"--preview-window="+cfg.FzfPreviewWindow,
	)
	if opts.Multi {
		args = append(args, "--multi")
	}
	if opts.Prompt != "" {
		args = append(args, "--prompt="+opts.Prompt)
	}
	if opts.Header != "" {
		args = append(args, "--header="+opts.Header)
	}

	return runFzfRaw(cmdline[0], args, input)
}

// runFzfLines runs fzf over arbitrary line input (e.g. a list of tag names)
// and returns the selected lines. Useful for the tag selection step.
func runFzfLines(cfg *Config, lines []string, opts fzfOptions) ([]string, error) {
	cmdline, err := shellquote.Split(cfg.FzfCmd)
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse fzf command ($NOTES_CLI_FZF / fzf_cmd)")
	}

	args := cmdline[1:]
	if opts.Multi {
		args = append(args, "--multi")
	}
	if opts.Prompt != "" {
		args = append(args, "--prompt="+opts.Prompt)
	}
	if opts.Header != "" {
		args = append(args, "--header="+opts.Header)
	}

	return runFzfRaw(cmdline[0], args, strings.Join(lines, "\n"))
}

// runFzfRaw is the low-level fzf runner. It pipes input to fzf, captures
// stdout, and returns non-empty selected lines. Exit codes 1 and 130 (no
// match / user cancelled) are treated as empty selection, not errors.
func runFzfRaw(bin string, args []string, input string) ([]string, error) {
	cmd := exec.Command(bin, args...)
	cmd.Stdin = strings.NewReader(input)
	cmd.Stderr = os.Stderr

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			if ee.ExitCode() == 1 || ee.ExitCode() == 130 {
				return nil, nil
			}
		}
		return nil, errors.Wrap(err, "fzf failed")
	}

	raw := strings.TrimRight(out.String(), "\n")
	if raw == "" {
		return nil, nil
	}
	all := strings.Split(raw, "\n")
	result := make([]string, 0, len(all))
	for _, l := range all {
		if l != "" {
			result = append(result, l)
		}
	}
	return result, nil
}

// relPathsToAbsPaths converts fzf-returned relative paths to absolute paths
// under the notes home directory.
func relPathsToAbsPaths(cfg *Config, relPaths []string) []string {
	abs := make([]string, len(relPaths))
	for i, rel := range relPaths {
		abs[i] = filepath.Join(cfg.HomePath, filepath.FromSlash(rel))
	}
	return abs
}

// collectFilteredNotes returns all notes matching optional category and tag
// regular expression patterns. Empty patterns match everything.
func collectFilteredNotes(cfg *Config, categoryRe, tagRe string) ([]*Note, error) {
	cats, err := CollectCategories(cfg, 0)
	if err != nil {
		return nil, err
	}

	var catReg *regexp.Regexp
	if categoryRe != "" {
		if catReg, err = regexp.Compile(categoryRe); err != nil {
			return nil, errors.Wrap(err, "invalid category regular expression")
		}
	}

	var tagReg *regexp.Regexp
	if tagRe != "" {
		if tagReg, err = regexp.Compile(tagRe); err != nil {
			return nil, errors.Wrap(err, "invalid tag regular expression")
		}
	}

	var notes []*Note
	for name, cat := range cats {
		if catReg != nil && !catReg.MatchString(name) {
			continue
		}
		for _, p := range cat.NotePaths {
			note, err := LoadNote(p, cfg)
			if err != nil {
				return nil, err
			}
			if tagReg == nil {
				notes = append(notes, note)
				continue
			}
			for _, t := range note.Tags {
				if tagReg.MatchString(t) {
					notes = append(notes, note)
					break
				}
			}
		}
	}
	return notes, nil
}
