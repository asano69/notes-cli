package notes

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// SectionCollectMode customizes the behavior of how to collect sections
type SectionCollectMode uint

const (
	// OnlyFirstSection is a flag to stop collecting sections earlier. If this flag is included
	// in mode parameter of CollectSections(), it collects only first section and only first
	// note and stops finding anymore.
	OnlyFirstSection SectionCollectMode = 1 << iota
)

// Section represents a section directory which contains some notes
type Section struct {
	// Path is a path to the section directory
	Path string
	// Name is a name of section
	Name string
	// NotePaths are paths to notes of the section
	NotePaths []string
}

// Notes returns all Note instances which belong to the section
func (cat *Section) Notes(c *Config) ([]*Note, error) {
	notes := make([]*Note, 0, len(cat.NotePaths))
	for _, p := range cat.NotePaths {
		n, err := LoadNote(p, c)
		if err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, nil
}

// Sections is a map from section name to Section instance
type Sections map[string]*Section

// Names returns all section names as slice
func (cats Sections) Names() []string {
	ss := make([]string, 0, len(cats))
	for n := range cats {
		ss = append(ss, n)
	}
	return ss
}

// Notes returns all Note instances which belong to the sections
func (cats Sections) Notes(cfg *Config) ([]*Note, error) {
	numNotes := 0
	for _, c := range cats {
		numNotes += len(c.NotePaths)
	}

	notes := make([]*Note, 0, numNotes)
	for _, c := range cats {
		for _, p := range c.NotePaths {
			n, err := LoadNote(p, cfg)
			if err != nil {
				return nil, err
			}
			notes = append(notes, n)
		}
	}
	return notes, nil
}

// CollectSections collects all sections under home by default. The behavior of collecting sections
// can be customized with mode parameter. Default mode value is 0 (nothing specified).
func CollectSections(cfg *Config, mode SectionCollectMode) (Sections, error) {
	cats := Sections(map[string]*Section{})

	fs, err := os.ReadDir(cfg.HomePath)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot read home")
	}

	stopWalking := false
	for _, f := range fs {
		name := f.Name()
		if !f.IsDir() || strings.HasPrefix(name, ".") {
			continue
		}

		root := filepath.Join(cfg.HomePath, name)
		if err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if stopWalking {
				return filepath.SkipDir
			}

			if err != nil {
				return err
			}

			path = normPathNFD(path)
			name := info.Name()

			if info.IsDir() {
				if strings.HasPrefix(name, ".") {
					return filepath.SkipDir
				}
				rel := strings.TrimPrefix(path, cfg.HomePath)
				n := strings.TrimPrefix(filepath.ToSlash(rel), "/")
				cats[n] = &Section{Name: n, Path: path}
				return nil
			}

			if strings.HasPrefix(name, ".") || !strings.HasSuffix(name, ".md") {
				return nil
			}

			rel := strings.TrimPrefix(filepath.Dir(path), cfg.HomePath)
			cat := cats[strings.TrimPrefix(filepath.ToSlash(rel), "/")]
			cat.NotePaths = append(cat.NotePaths, path)

			if mode&OnlyFirstSection != 0 {
				stopWalking = true
				return filepath.SkipDir
			}

			return nil
		}); err != nil {
			return nil, errors.Wrapf(err, "Cannot walk on directory for section %q", name)
		}
		if stopWalking {
			break
		}
	}

	// Remove section which has no note
	for n, c := range cats {
		if len(c.NotePaths) == 0 {
			delete(cats, n)
		}
	}

	return cats, nil
}
