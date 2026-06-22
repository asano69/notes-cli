package notes

import (
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

notes is developed at https://github.com/rhysd/notes-cli. If you're seeing a bug or having a feature request,
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
	SortBy   string `name:"sort" short:"s" enum:"modified,created,filename,category" help:"Sort list by 'modified', 'created', 'filename' or 'category'. Default is 'created'"`
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

	switch ctx.Command() {
	case "new <category> <filename> [<tags>]":
		return &NewCmd{Config: c, Category: cli.New.Category, Filename: cli.New.Filename, Tags: cli.New.Tags, NoInline: cli.New.NoInline, NoEdit: cli.New.NoEdit}, nil
	case "list", "ls", "list ls":
		return &ListCmd{Config: c, Out: colorStdout, Full: cli.List.Full, Category: cli.List.Category, Tag: cli.List.Tag, Relative: cli.List.Relative, Oneline: cli.List.Oneline, SortBy: cli.List.SortBy, Edit: cli.List.Edit}, nil
	case "categories", "cats", "categories cats":
		return &CategoriesCmd{Config: c, Out: os.Stdout}, nil
	case "tags [<category>]":
		return &TagsCmd{Config: c, Out: os.Stdout, Category: cli.Tags.Category}, nil
	case "tagmod <from> [<to>]":
		return &TagModCmd{Config: c, From: cli.TagMod.From, To: cli.TagMod.To, Force: cli.TagMod.Force}, nil
	case "tagadd <tag> <target>":
		return &AddTagCmd{Config: c, Tag: cli.TagAdd.Tag, Target: cli.TagAdd.Target}, nil
	case "save":
		return &SaveCmd{Config: c, Message: cli.Save.Message}, nil
	case "config [<name>]":
		return &ConfigCmd{Config: c, Out: os.Stdout, Name: cli.Config.Name}, nil
	case "fix":
		return &FixCmd{Config: c, Out: os.Stdout, DryRun: cli.Fix.DryRun}, nil
	case "edit":
		return &EditCmd{Config: c, Category: cli.Edit.Category, Tag: cli.Edit.Tag}, nil
	default:
		panic("FATAL: Unknown command: " + ctx.Command())
	}
}
