package notes

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/mattn/go-runewidth"
	"github.com/pkg/errors"
)

var (
	yellow = color.New(color.FgYellow)
	bold   = color.New(color.Bold)
	green  = color.New(color.FgGreen)
)

// ListCmd represents `notes list` command. Each public fields represent options of the command
// Out field represents where this command should output.
type ListCmd struct {
	Config *Config
	// Full is a flag equivalent to --full
	Full bool
	// Category is a regex string equivalent to --category
	Category string
	// Tag is a regex string equivalent to --tag
	Tag string
	// Relative is a flag equivalent to --relative
	Relative bool
	// Oneline is a flag equivalent to --oneline
	Oneline bool
	// SortBy is a string indicating how to sort the list. This value is equivalent to --sort option
	SortBy string
	// Edit is a flag equivalent to --edit
	Edit bool
	// Out is a writer to write output of this command. Kind of stdout is expected
	Out io.Writer
}

func (cmd *ListCmd) printNoteFullTo(out *bufio.Writer, note *Note) {
	green.Fprintln(out, note.FilePath())
	yellow.Fprint(out, "Category: ")
	fmt.Fprintln(out, note.Category)
	yellow.Fprint(out, "Tags:     ")
	fmt.Fprintln(out, strings.Join(note.Tags, ", "))
	yellow.Fprint(out, "Created:  ")
	fmt.Fprintln(out, note.Created.Format(time.RFC3339))
	if note.Title != "" {
		bold.Fprintf(out, "\n%s\n%s\n\n", note.Title, strings.Repeat("=", runewidth.StringWidth(note.Title)))
	}

	body, size, err := note.ReadBodyLines(10)
	if err != nil || len(body) == 0 {
		return
	}

	fmt.Fprint(out, body)

	// Ensure body ends with newline
	if !strings.HasSuffix(body, "\n") {
		fmt.Fprintln(out)
	}

	// Body text was truncated. To indicate it, add ellipsis at the end
	if size == 10 {
		fmt.Fprintln(out, "...")
	}

	// Finally separate each note with blank line
	fmt.Fprintln(out)
}

// printOnelineNotes outputs notes as TSV with a header row. No color or
// alignment is applied so the output is easy to consume in tools like nushell.
func (cmd *ListCmd) printOnelineNotes(notes []*Note) error {
	out := bufio.NewWriter(cmd.Out)
	fmt.Fprintln(out, "categories\tfilename\ttags\ttitle")
	for _, note := range notes {
		fmt.Fprintf(out, "%s\t%s\t%s\t%s\n",
			note.Category,
			note.File,
			strings.Join(note.Tags, ","),
			note.Title,
		)
	}
	return out.Flush()
}

func (cmd *ListCmd) printNotes(notes []*Note) error {
	switch strings.ToLower(cmd.SortBy) {
	case "filename":
		sortByFilename(notes)
	case "category":
		sortByCategory(notes)
	case "modified":
		if err := sortByModified(notes); err != nil {
			return err
		}
	default:
		sortByCreated(notes)
	}

	if cmd.Full {
		out := bufio.NewWriter(cmd.Out)
		for _, note := range notes {
			cmd.printNoteFullTo(out, note)
		}
		return out.Flush()
	}

	if cmd.Oneline {
		return cmd.printOnelineNotes(notes)
	}

	if cmd.Edit {
		args := make([]string, 0, len(notes))
		for _, n := range notes {
			args = append(args, n.FilePath())
		}
		return openEditor(cmd.Config, args...)
	}

	var b bytes.Buffer
	if cmd.Relative {
		for _, note := range notes {
			b.WriteString(note.RelFilePath())
			b.WriteRune('\n')
		}
	} else {
		for _, note := range notes {
			b.WriteString(note.FilePath())
			b.WriteRune('\n')
		}
	}

	_, err := cmd.Out.Write(b.Bytes())
	return err
}

// Do runs `notes list` command and returns an error if occurs
func (cmd *ListCmd) Do() error {
	cats, err := CollectCategories(cmd.Config, 0)
	if err != nil {
		return err
	}

	var catReg *regexp.Regexp
	if cmd.Category != "" {
		if catReg, err = regexp.Compile(cmd.Category); err != nil {
			return errors.Wrap(err, "Regular expression for filtering categories is invalid")
		}
	}

	numNotes := 0
	for n, c := range cats {
		if catReg != nil && !catReg.MatchString(n) {
			delete(cats, n)
			continue
		}
		numNotes += len(c.NotePaths)
	}

	var tagReg *regexp.Regexp
	if cmd.Tag != "" {
		if tagReg, err = regexp.Compile(cmd.Tag); err != nil {
			return errors.Wrap(err, "Regular expression for filtering tags is invalid")
		}
	}

	notes := make([]*Note, 0, numNotes)
	for _, cat := range cats {
		for _, p := range cat.NotePaths {
			note, err := LoadNote(p, cmd.Config)
			if err != nil {
				return err
			}
			if tagReg == nil {
				notes = append(notes, note)
				continue
			}
			for _, tag := range note.Tags {
				if tagReg.MatchString(tag) {
					notes = append(notes, note)
					break
				}
			}
			// When no tag is matched to tag regex, the note is ignored
		}
	}

	if len(notes) == 0 {
		return nil
	}

	return cmd.printNotes(notes)
}
