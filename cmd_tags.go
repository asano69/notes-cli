package notes

import (
	"fmt"
	"github.com/pkg/errors"
	"io"
	"sort"
	"strings"
)

// TagsCmd represents `notes tags` command. Each public fields represent options of the command
// Out field represents where this command should output.
type TagsCmd struct {
	Config *Config
	// Section is a section name of tags. If this value is empty, tags of all sections will be output
	Section string
	// Out is a writer to write output of this command. Kind of stdout is expected
	Out io.Writer
}

// Do runs `notes tags` command and returns an error if occurs
func (cmd *TagsCmd) Do() error {
	saw := map[string]struct{}{}
	tags := []string{}

	cats, err := CollectSections(cmd.Config, 0)
	if err != nil {
		return err
	}

	if cmd.Section != "" {
		// Even if section is specified, we fetch all sections since error message requires
		// all section names for suggestion.
		cat, ok := cats[cmd.Section]
		if !ok {
			ns := cats.Names()
			return errors.Errorf("Section '%s' does not exist. All sections are %s", cmd.Section, strings.Join(ns, ", "))
		}
		cats = Sections{cmd.Section: cat}
	}

	for _, cat := range cats {
		notes, err := cat.Notes(cmd.Config)
		if err != nil {
			return err
		}
		for _, n := range notes {
			for _, tag := range n.Tags {
				if _, ok := saw[tag]; !ok {
					tags = append(tags, tag)
					saw[tag] = struct{}{}
				}
			}
		}
	}

	sort.Strings(tags)

	_, err = fmt.Fprintln(cmd.Out, strings.Join(tags, "\n"))
	return err
}
