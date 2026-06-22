package notes

import (
	"reflect"
	"testing"
)

func TestParseCmdReturnsKongBackedCommands(t *testing.T) {
	t.Setenv("NOTES_CLI_HOME", t.TempDir())
	t.Setenv("NOTES_CLI_GIT", "")
	t.Setenv("NOTES_CLI_EDITOR", "")
	t.Setenv("NOTES_CLI_PAGER", "")

	tests := []struct {
		name string
		args []string
		want any
	}{
		{"new", []string{"new", "cat", "file", "tag"}, &NewCmd{}},
		{"list", []string{"list", "--oneline"}, &ListCmd{}},
		{"list alias", []string{"ls", "--oneline"}, &ListCmd{}},
		{"categories", []string{"categories"}, &CategoriesCmd{}},
		{"categories alias", []string{"cats"}, &CategoriesCmd{}},
		{"tags", []string{"tags", "cat"}, &TagsCmd{}},
		{"tagmod", []string{"tagmod", "old", "new"}, &TagModCmd{}},
		{"tagadd", []string{"tagadd", "tag", "cat/file.md"}, &AddTagCmd{}},
		{"save", []string{"save", "--message", "msg"}, &SaveCmd{}},
		{"config", []string{"config", "home"}, &ConfigCmd{}},
		{"fix", []string{"fix", "--dry-run"}, &FixCmd{}},
		{"edit", []string{"edit", "--category", "cat", "--tag", "tag"}, &EditCmd{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCmd(tt.args)
			if err != nil {
				t.Fatalf("ParseCmd(%v) returned error: %v", tt.args, err)
			}
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Fatalf("got %T, want %T", got, tt.want)
			}
		})
	}
}
