package notes

import (
	"errors"
	"fmt"
	"io"

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

			if err := updateCategoryInFile(p, mismatch.pathcat); err != nil {
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
