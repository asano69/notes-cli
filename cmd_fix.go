package notes

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"
)

// FixCmd represents `notes fix` command.
type FixCmd struct {
	cli    *kingpin.CmdClause
	Config *Config
	// DryRun prints what would be changed without modifying files
	DryRun bool
	// Out is a writer to write output of this command
	Out io.Writer
}

func (cmd *FixCmd) defineCLI(app *kingpin.Application) {
	cmd.cli = app.Command("fix", "Fix notes whose category property does not match their parent directory")
	cmd.cli.Flag("dry-run", "Print what would be changed without modifying files").Short('n').BoolVar(&cmd.DryRun)
}

func (cmd *FixCmd) matchesCmdline(cmdline string) bool {
	return cmd.cli.FullCommand() == cmdline
}

// fixCategoryInFile rewrites the category field in the YAML frontmatter of the given file.
func fixCategoryInFile(path, correctCategory string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	f.Close()
	if err := scanner.Err(); err != nil {
		return err
	}

	// Rewrite the category line inside the YAML frontmatter (between the two --- markers).
	inFrontmatter := false
	fixed := false
	for i, line := range lines {
		if line == "---" {
			if !inFrontmatter {
				inFrontmatter = true
				continue
			}
			// Closing --- reached without finding the category line; give up.
			break
		}
		if inFrontmatter && strings.HasPrefix(line, "category: ") {
			lines[i] = "category: " + correctCategory
			fixed = true
			break
		}
	}

	if !fixed {
		return fmt.Errorf("category field not found in frontmatter of %q", path)
	}

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	w := bufio.NewWriter(out)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}

// Do runs `notes fix` command and returns an error if any occurs.
func (cmd *FixCmd) Do() error {
	cats, err := CollectCategories(cmd.Config, 0)
	if err != nil {
		return err
	}

	fixed := 0
	for _, cat := range cats {
		for _, p := range cat.NotePaths {
			_, err := LoadNote(p, cmd.Config)
			if err == nil {
				continue
			}

			var mismatch *MismatchCategoryError
			if !errors.As(err, &mismatch) {
				return err
			}

			fmt.Fprintf(cmd.Out, "%s\n  category: %q -> %q\n", p, mismatch.cat, mismatch.pathcat)

			if cmd.DryRun {
				continue
			}

			if err := fixCategoryInFile(p, mismatch.pathcat); err != nil {
				return err
			}
			fixed++
		}
	}

	if !cmd.DryRun && fixed > 0 {
		fmt.Fprintf(cmd.Out, "\nFixed %d note(s).\n", fixed)
	}
	if cmd.DryRun {
		fmt.Fprintln(cmd.Out, "\n(dry-run: no files were modified)")
	}

	return nil
}
