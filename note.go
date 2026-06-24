package notes

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// MismatchSectionError represents an error caused when a user specifies mismatched section
type MismatchSectionError struct {
	cat, pathcat, path string
}

func (e *MismatchSectionError) Error() string {
	return fmt.Sprintf("Section does not match to file path. Section is '%s' but it should be '%s' from its file path. File path is '%s'", e.cat, e.pathcat, e.path)
}

// Is returns if given error is a MismatchSectionError or not
func (e *MismatchSectionError) Is(target error) bool {
	_, ok := target.(*MismatchSectionError)
	return ok
}

// Note represents a note stored on filesystem or will be created
type Note struct {
	// Config is a configuration of notes command which was created by NewConfig()
	Config *Config
	// Section is a section string. It must not be empty
	Section string
	// Tags is tags of note. It can be empty and cannot contain comma
	Tags []string
	// Created is a datetime when note was created
	Created time.Time
	// File is a file name of the note
	File string
	// Title is a title string of the note
	Title string
	// Summary is an optional short description of the note
	Summary string
	// Code is an optional code identifier or snippet associated with the note
	Code string
}

// DirPath returns the absolute section directory path of the note
func (note *Note) DirPath() string {
	return filepath.Join(note.Config.HomePath, filepath.FromSlash(note.Section))
}

// FilePath returns the absolute file path of the note
func (note *Note) FilePath() string {
	return filepath.Join(note.Config.HomePath, filepath.FromSlash(note.Section), note.File)
}

// RelFilePath returns the relative file path of the note from home directory
func (note *Note) RelFilePath() string {
	return filepath.Join(filepath.FromSlash(note.Section), note.File)
}

// TemplatePath resolves a path to template file of the note. If no template is found, it returns
// false as second return value
func (note *Note) TemplatePath() (string, bool) {
	p := note.DirPath()
	for {
		f := filepath.Join(p, ".template.md")
		if s, err := os.Stat(f); err == nil && !s.IsDir() {
			return f, true
		}
		if p == note.Config.HomePath {
			return "", false
		}
		p = filepath.Dir(p)
	}
}

// Create creates a file of the note. When title is empty, file name omitting file extension is used
// for it. This function will fail when the file is already existing.
func (note *Note) Create() error {
	var template []byte
	if p, ok := note.TemplatePath(); ok {
		b, err := os.ReadFile(p)
		if err != nil {
			return errors.Wrapf(err, "Cannot read template file %q", p)
		}
		template = b
	}

	title := note.Title
	if title == "" {
		title = strings.TrimSuffix(note.File, filepath.Ext(note.File))
	}

	var b bytes.Buffer

	// Frontmatter is rendered from frontmatterSchema (see frontmatter.go), so
	// the field set, order, and formatting (e.g. title/summary being quoted)
	// stay in sync with what FixCmd produces.
	// sections is intentionally omitted: it is always derivable from the
	// file's location relative to NOTES_CLI_HOME, so storing it in the
	// frontmatter is redundant. LoadNote derives it from the path when absent.
	b.WriteString(renderFrontmatter(
		map[string]string{
			"title":   title,
			"summary": "",
			"draft":   "",
			"date":    note.Created.Format(time.RFC3339),
			"lastmod": "",
		},
		map[string][]string{
			"tags": note.Tags,
			// "sections" is deliberately omitted.
		},
	))
	b.WriteRune('\n')

	if len(template) > 0 {
		b.Write(template)
	}

	d := note.DirPath()
	if err := os.MkdirAll(d, 0755); err != nil {
		return errors.Wrapf(err, "Could not create section directory '%s'", d)
	}

	p := filepath.Join(d, note.File)
	if _, err := os.Stat(p); err == nil {
		return errors.Errorf("Cannot create new note since file '%s' already exists. Please edit it", note.RelFilePath())
	}

	f, err := os.Create(p)
	if err != nil {
		return errors.Wrap(err, "Cannot create note file")
	}
	defer f.Close()

	_, err = f.Write(b.Bytes())
	return err
}


// Open opens the note using an editor command user set. When user did not set any editor command
// with $NOTES_CLI_EDITOR, this method fails. Otherwise, an editor process is spawned with argument
// of path to the note file
func (note *Note) Open() error {
	return openEditor(note.Config, note.FilePath())
}

// ReadBodyLines reads body of note until maxLines lines and returns it as string and number of lines as int
func (note *Note) ReadBodyLines(maxLines int) (string, int, error) {
	path := note.FilePath()
	f, err := os.Open(path)
	if err != nil {
		return "", 0, errors.Wrap(err, "Cannot open note file")
	}
	defer f.Close()

	r := bufio.NewReader(f)

	// Skip YAML frontmatter (--- ... ---)
	firstLine, err := r.ReadString('\n')
	if err != nil || strings.TrimSpace(firstLine) != "---" {
		return "", 0, errors.Errorf("Cannot read frontmatter of note file. Some metadata may be missing in '%s'", note.RelFilePath())
	}
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return "", 0, errors.Wrapf(err, "Cannot read metadata of note file. Some metadata may be missing in '%s'", note.RelFilePath())
		}
		if strings.TrimSpace(line) == "---" {
			break
		}
	}

	// Skip blank lines and legacy H1 title lines to reach body content.
	// New-format notes have no H1, but old notes do; skipping "# ..." lines
	// preserves backward compatibility without showing the title twice.
	var firstBodyLine []byte
	for {
		b, err := r.ReadBytes('\n')
		if err != nil {
			return "", 0, nil // no body
		}
		s := strings.TrimSpace(string(b))
		if s == "" || strings.HasPrefix(s, "# ") {
			continue
		}
		firstBodyLine = b
		break
	}

	if len(firstBodyLine) == 0 {
		return "", 0, nil
	}

	var buf bytes.Buffer
	buf.Write(firstBodyLine)
	readLines := 1

	for numLines := 1; numLines < maxLines; numLines++ {
		b, err := r.ReadBytes('\n')
		if err != nil {
			break
		}
		readLines++
		buf.Write(b)
	}

	return buf.String(), readLines, nil
}

// NewNote creates a new note instance with given parameters and configuration. Section and file name
// cannot be empty. If given file name lacks file extension, it automatically adds ".md" to file name.
// The created file name is also prefixed with the creation date as "YYYY-MM-DD-".
func NewNote(cat, tags, file, title string, cfg *Config) (*Note, error) {
	cat = strings.TrimSpace(cat)
	file = strings.TrimSpace(file)
	title = strings.TrimSpace(title)

	for _, part := range strings.Split(cat, "/") {
		if err := validateDirname(part); err != nil {
			return nil, errors.Wrapf(err, "Invalid section part '%s' as directory name", part)
		}
	}

	if file == "" || strings.HasPrefix(file, ".") {
		return nil, errors.New("File name cannot be empty and cannot start with '.'")
	}

	ts := []string{}
	for _, t := range strings.Split(tags, ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			ts = append(ts, t)
		}
	}

	if !strings.HasSuffix(file, ".md") {
		file += ".md"
	}

	// Default title is derived from the file name before the date prefix is added below.
	if title == "" {
		title = strings.TrimSuffix(file, ".md")
	}

	now := time.Now()
	file = now.Format("2006-01-02") + "-" + file

	return &Note{cfg, cat, ts, now, file, title, "", ""}, nil
}

// LoadNote reads note file from given path, parses it and creates Note instance. When given file path
// does not exist or when the file does not contain mandatory metadata ('tags' and 'created'/'date'),
// this function returns an error.
// Section is derived from the file's location relative to HomePath when absent from frontmatter,
// so 'section'/'sections' in frontmatter is optional.
func LoadNote(path string, cfg *Config) (*Note, error) {
	// This is necessary for macOS, where path contains NFD format
	path = normPathNFD(path)

	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot open note file")
	}
	defer f.Close()

	note := &Note{Config: cfg}
	note.File = filepath.Base(path)

	s := bufio.NewScanner(f)

	// Expect opening ---
	if !s.Scan() || s.Text() != "---" {
		return nil, errors.Errorf("Note '%s' must start with YAML frontmatter '---'", canonPath(path))
	}

	// Parse frontmatter fields.
	// Supports both the new format (title, date, sections, summary, code) and the
	// legacy format (section, created) for backward compatibility.
	inTagsList := false
	inSectionsList := false
	for s.Scan() {
		line := s.Text()
		if line == "---" {
			break
		}

		// Handle block-style tags (e.g. Obsidian format):
		//   tags:
		//     - foo
		//     - bar
		if inTagsList {
			if trimmed := strings.TrimSpace(line); strings.HasPrefix(trimmed, "- ") {
				if tag := strings.TrimSpace(trimmed[2:]); tag != "" {
					note.Tags = append(note.Tags, tag)
				}
				continue
			}
			inTagsList = false
		}

		// Handle block-style sections (e.g. Hugo/Obsidian format):
		//   sections:
		//     - foo
		//     - bar
		// Only the first item is used, same as the inline format below.
		if inSectionsList {
			if trimmed := strings.TrimSpace(line); strings.HasPrefix(trimmed, "- ") {
				if c := strings.TrimSpace(trimmed[2:]); c != "" && note.Section == "" {
					note.Section = c
				}
				continue
			}
			inSectionsList = false
		}

		switch {
		case strings.HasPrefix(line, "title: "):
			note.Title = strings.TrimSpace(line[7:])

		// New format: sections as an inline list; the first element is used as the section.
		case strings.HasPrefix(line, "sections: "):
			raw := strings.TrimSpace(strings.Trim(strings.TrimSpace(line[12:]), "[]"))
			for _, c := range strings.Split(raw, ",") {
				if c = strings.TrimSpace(c); c != "" {
					note.Section = c
					break
				}
			}

		// New format: sections as a block list (see inSectionsList above).
		case line == "sections:":
			inSectionsList = true

		// Legacy format: single section string.
		case strings.HasPrefix(line, "section: "):
			note.Section = strings.TrimSpace(line[10:])

		case strings.HasPrefix(line, "tags: "):
			// Inline format: tags: [foo, bar]
			raw := strings.TrimSpace(strings.Trim(strings.TrimSpace(line[6:]), "[]"))
			note.Tags = make([]string, 0)
			for _, t := range strings.Split(raw, ",") {
				if t = strings.TrimSpace(t); t != "" {
					note.Tags = append(note.Tags, t)
				}
			}
		case line == "tags:":
			// Block format: tags followed by "- item" lines
			note.Tags = make([]string, 0)
			inTagsList = true

		// New format: date field.
		case strings.HasPrefix(line, "date: "):
			raw := strings.TrimSpace(line[6:])
			if raw == "" {
				break
			}
			t, err := time.Parse(time.RFC3339, raw)
			if err != nil {
				t, err = time.ParseInLocation("2006-01-02T15:04:05", raw, time.Local)
				if err != nil {
					return nil, errors.Wrapf(err, "Cannot parse date as RFC3339 format: %s", line)
				}
			}
			note.Created = t

		// Legacy format: created field.
		case strings.HasPrefix(line, "created: "):
			raw := strings.TrimSpace(line[9:])
			if raw == "" {
				break
			}
			t, err := time.Parse(time.RFC3339, raw)
			if err != nil {
				// Fall back to old format without timezone offset for existing notes
				t, err = time.ParseInLocation("2006-01-02T15:04:05", raw, time.Local)
				if err != nil {
					return nil, errors.Wrapf(err, "Cannot parse created date time as RFC3339 format: %s", line)
				}
			}
			note.Created = t

		case strings.HasPrefix(line, "summary:"):
			note.Summary = strings.TrimSpace(line[8:])

		case strings.HasPrefix(line, "code:"):
			note.Code = strings.TrimSpace(line[5:])
		}
	}

	if err := s.Err(); err != nil {
		return nil, errors.Wrapf(err, "Cannot read note file '%s'", canonPath(path))
	}

	// Section is optional in frontmatter: derive it from the file's location
	// relative to HomePath when absent or empty.
	if note.Section == "" {
		parent := filepath.Dir(path)
		if rel, err := filepath.Rel(cfg.HomePath, parent); err == nil {
			note.Section = filepath.ToSlash(rel)
		}
	}

	if note.Tags == nil || note.Created.IsZero() {
		return nil, errors.Errorf("Missing metadata in file '%s'. 'tags' and 'created'/'date' are mandatory", canonPath(path))
	}

	// For old-format notes, title lives in the H1 heading after the closing ---.
	// For new-format notes, title was already set from the frontmatter above.
	if note.Title == "" {
		for s.Scan() {
			line := s.Text()
			if strings.HasPrefix(line, "# ") {
				note.Title = strings.TrimSpace(line[2:])
				break
			}
			if strings.TrimSpace(line) != "" {
				// Non-blank line that is not an H1 heading — treat as no title
				break
			}
		}
	}

	if note.Title == "" {
		note.Title = "(no title)"
	}

	parent := filepath.Dir(path)
	rel, err := filepath.Rel(cfg.HomePath, parent)
	name := filepath.ToSlash(rel)
	if err != nil || filepath.ToSlash(rel) != note.Section {
		return note, &MismatchSectionError{note.Section, name, path}
	}

	return note, nil
}
