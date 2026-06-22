package notes

import (
	"io"
	"os"

	"github.com/alecthomas/kong"
	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
)

// Cmd is an interface for subcommands of notes command
type Cmd interface {
	Do() error
}

// Version is version string of notes command. It conforms semantic versioning
var Version = "2.0.0"
var description = `Simple note taking tool for command line with your favorite editor.

You can manage (create/open/list) notes via this tool on terminal. notes also
optionally can save your notes thanks to Git to avoid losing your notes.

notes is intended to be used nicely with other commands such as grep (or ag, rg),
rm, filtering tools such as fzf or peco and editors which can be started from
command line.

notes is developed at https://github.com/asano69/notes-cli. If you're seeing a bug or having a feature request,
please create a new issue. Pull requests are more than welcome.`

type cliOptions struct {
	NoColor     bool             `help:"Disable color output"`
	ColorAlways bool             `short:"A" help:"Enable color output always"`
	Version     kong.VersionFlag `help:"Show version"`

	New        newCommand        `cmd:"" help:"Create a new note with given category and file name"`
	List       listCommand       `cmd:"" aliases:"ls" help:"List notes with filtering by categories and/or tags with regular expressions. By default, it shows full path of notes"`
	Categories categoriesCommand `cmd:"" aliases:"cats" help:"List all categories to stdout"`
	Tags       tagsCommand       `cmd:"" help:"List all tags"`
	TagMod     tagModCommand     `cmd:"" name:"tagmod" help:"Rename or delete a tag across all notes"`
	TagAdd     tagAddCommand     `cmd:"" name:"tagadd" help:"Add a tag to a note file, or to all notes in a category"`
	Save       saveCommand       `cmd:"" help:"Save notes using Git. It adds all notes and creates a commit to Git repository at home directory"`
	Config     configCommand     `cmd:"" help:"Output config values to stdout. By default output all values with KEY=VALUE style"`
	Fix        fixCommand        `cmd:"" help:"Repair note YAML frontmatter into the format notes-cli expects"`
	Edit       editCommand       `cmd:"" help:"Interactively select a note with fzf and open it in the editor"`
}

type newCommand struct {
	Category string `arg:"" help:"Category of note. Note must belong to one category"`
	Filename string `arg:"" help:"File name of note. It automatically adds '.md' file extension if omitted"`
	Tags     string `arg:"" optional:"" help:"Comma-separated tags of note. Zero or more tags can be specified to note"`
	NoInline bool   `name:"no-inline-input" help:"Does not request inline input even if no editor command is set to $NOTES_CLI_EDITOR"`
	NoEdit   bool   `name:"no-edit" help:"Does not open an editor even if an editor command is set to $NOTES_CLI_EDITOR"`
}

type listCommand struct {
	Full     bool   `short:"f" help:"Show list of full information of note (full path, metadata, title, body (up to 10 lines)) instead of file path"`
	Category string `short:"c" help:"Filter list by category name with regular expression"`
	Tag      string `short:"t" help:"Filter list by tag name with regular expression"`
	Relative bool   `short:"r" help:"Show relative paths from $NOTES_CLI_HOME directory"`
	Oneline  bool   `short:"o" help:"Show oneline information of note (relative path, category, tags, title) instead of file path"`
	SortBy   string `name:"sort" short:"s" enum:"modified,created,filename,category" default:"created" help:"Sort list by 'modified', 'created', 'filename' or 'category'. Default is 'created'"`
	Edit     bool   `short:"e" help:"Open listed notes with your favorite editor. $NOTES_CLI_EDITOR must be set. Paths of listed notes are passed to the editor command's arguments"`
}

type categoriesCommand struct{}

type tagsCommand struct {
	Category string `arg:"" optional:"" help:"Show tags of specified category. If not specified, all tags are output"`
}

type tagModCommand struct {
	From  string `arg:"" help:"Tag name to rename or delete"`
	To    string `arg:"" optional:"" help:"New tag name. When omitted, the 'from' tag is deleted instead of renamed"`
	Force bool   `short:"f" help:"Required to actually delete a tag when 'to' is omitted"`
}

type tagAddCommand struct {
	Tag    string `arg:"" help:"Tag name to add"`
	Target string `arg:"" help:"Path to a note file, or a category name (the 'categories' value in frontmatter, e.g. 'animal/dog')"`
}

type saveCommand struct {
	Message string `short:"m" help:"Commit message on save. If omitted, an automatic message will be used"`
}

type configCommand struct {
	Name string `arg:"" optional:"" help:"Key name. One of 'home', 'git', 'editor', 'fzf', 'bat'. Only value will be output"`
}

type fixCommand struct {
	DryRun bool `name:"dry-run" short:"n" help:"Print what would be changed without modifying files"`
}

type editCommand struct {
	Category string `short:"c" help:"Filter by category name with regular expression"`
	Tag      string `short:"t" help:"Filter by tag name with regular expression"`
}

func (cmd *newCommand) runtimeCmd(c *Config, _ io.Writer) Cmd {
	return &NewCmd{Config: c, Category: cmd.Category, Filename: cmd.Filename, Tags: cmd.Tags, NoInline: cmd.NoInline, NoEdit: cmd.NoEdit}
}

func (cmd *listCommand) runtimeCmd(c *Config, out io.Writer) Cmd {
	return &ListCmd{Config: c, Out: out, Full: cmd.Full, Category: cmd.Category, Tag: cmd.Tag, Relative: cmd.Relative, Oneline: cmd.Oneline, SortBy: cmd.SortBy, Edit: cmd.Edit}
}

func (cmd *categoriesCommand) runtimeCmd(c *Config, _ io.Writer) Cmd {
	return &CategoriesCmd{Config: c, Out: os.Stdout}
}

func (cmd *tagsCommand) runtimeCmd(c *Config, _ io.Writer) Cmd {
	return &TagsCmd{Config: c, Out: os.Stdout, Category: cmd.Category}
}

func (cmd *tagModCommand) runtimeCmd(c *Config, _ io.Writer) Cmd {
	return &TagModCmd{Config: c, From: cmd.From, To: cmd.To, Force: cmd.Force}
}

func (cmd *tagAddCommand) runtimeCmd(c *Config, _ io.Writer) Cmd {
	return &AddTagCmd{Config: c, Tag: cmd.Tag, Target: cmd.Target}
}

func (cmd *saveCommand) runtimeCmd(c *Config, _ io.Writer) Cmd {
	return &SaveCmd{Config: c, Message: cmd.Message}
}

func (cmd *configCommand) runtimeCmd(c *Config, _ io.Writer) Cmd {
	return &ConfigCmd{Config: c, Out: os.Stdout, Name: cmd.Name}
}

func (cmd *fixCommand) runtimeCmd(c *Config, _ io.Writer) Cmd {
	return &FixCmd{Config: c, Out: os.Stdout, DryRun: cmd.DryRun}
}

func (cmd *editCommand) runtimeCmd(c *Config, _ io.Writer) Cmd {
	return &EditCmd{Config: c, Category: cmd.Category, Tag: cmd.Tag}
}

type commandConfig interface {
	runtimeCmd(*Config, io.Writer) Cmd
}

// ParseCmd parses given arguments as command line options and returns corresponding subcommand instance.
// When no subcommand matches or argus contains invalid argument, it returns an error
func ParseCmd(args []string) (Cmd, error) {
	c, err := NewConfig()
	if err != nil {
		return nil, err
	}

	colorStdout := colorable.NewColorableStdout()

	// When `notes` command is run with no argument,
	//   - if there is no note, show usage help
	//   - if there is one or more notes, show the list with `list --oneline`
	// ref: #2
	if len(args) == 0 {
		if cats, err := CollectCategories(c, OnlyFirstCategory); err == nil && len(cats) > 0 {
			return &ListCmd{
				Config:  c,
				Out:     colorStdout,
				Oneline: true,
			}, nil
		}
	}

	var cli cliOptions
	parser, err := kong.New(&cli,
		kong.Name("notes"),
		kong.Description(description),
		kong.Vars{"version": Version},
	)
	if err != nil {
		return nil, err
	}

	ctx, err := parser.Parse(args)
	if err != nil {
		if ext, ok := NewExternalCmd(err, args); ok {
			return ext, nil
		}
		return nil, err
	}

	if cli.ColorAlways {
		color.NoColor = false
	}

	if cli.NoColor {
		color.NoColor = true
	}

	target := ctx.Selected().Target
	if target.CanInterface() {
		if cmd, ok := target.Interface().(commandConfig); ok {
			return cmd.runtimeCmd(c, colorStdout), nil
		}
	}
	if target.CanAddr() {
		if cmd, ok := target.Addr().Interface().(commandConfig); ok {
			return cmd.runtimeCmd(c, colorStdout), nil
		}
	}
	panic("FATAL: Unknown command: " + ctx.Command())
}
