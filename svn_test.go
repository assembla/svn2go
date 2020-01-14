package svn

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestBasic(t *testing.T) {
	r, err := Open("testdata/sample")

	if err != nil {
		t.Fatalf("Failed to open repo: %s", err.Error())
	}

	defer r.Close()

	rev, err := r.LatestRevision()

	if err != nil {
		t.Fatalf("Failed to get latest revision: %s", err.Error())
	}

	if rev <= 0 {
		t.Errorf("Expected latest revision to be positive, got %d", rev)
	}

	c, err := r.CommitInfo(1)

	if err != nil {
		t.Fatalf("Failed to get commit info for the first revision: %s", err.Error())
	}

	if c.Author != "lz" {
		t.Fatalf("Wrong author: '%s'", c.Author)
	}

	commits, err := r.Commits(1, 2)

	if err != nil {
		t.Fatalf("Failed to get commits: %s", err.Error())
	}

	commitsCount := len(commits)

	if commitsCount != 2 {
		t.Errorf("Expected 2 commits, got %d", commitsCount)
	}

	for _, commit := range commits {
		if commit.Author != "lz" {
			t.Errorf("Wrong author: '%s'", commit.Author)
		}
	}

	rev, err = r.LastPathRev("trunk/Makefile", 6)

	if err != nil {
		t.Errorf("Failed to get rev: %s", err.Error())
	} else if rev != 5 {
		t.Errorf("Expected rev 5, got %d", rev)
	}

	commits, err = r.History("trunk", 0, rev, 2)

	if err != nil {
		t.Fatalf("Failed to get history: %s", err.Error())
	}

	if len(commits) != 2 {
		t.Errorf("Expected history call to return 2 commits, got %d", len(commits))
	}

	entries, err := r.Tree("trunk", 5)

	if err != nil {
		t.Fatalf("Failed to get tree for trunk: %s", err.Error())
	}

	if len(entries) != 2 {
		t.Errorf("Expected the number of entries in trunk/ folder at rev 5 to equal 2, got %d", len(entries))
	}

	if size, err := r.FileSize("trunk/Makefile", 6); err != nil {
		t.Fatalf("Failed to get file size: %s", err.Error())
	} else {
		exp := int64(1279)

		if size != exp {
			t.Errorf("Wrong file size, expected %d, got %d", exp, size)
		}
	}

	mimeExp := "application/octet-stream"

	if mime, err := r.MimeType("trunk/images/play.png", 7); err != nil {
		t.Fatalf("Failed to get mime type: %s", err.Error())
	} else {
		if mimeExp != mime {
			t.Errorf("Wrong file mime type, expected %s, got %s", mimeExp, mime)
		}
	}

	if reader, err := r.FileContent("trunk/TODO", 6); err != nil {
		t.Fatalf("Failed to read file content: %s", err.Error())
	} else {
		data, err := ioutil.ReadAll(reader)

		if err != nil {
			t.Fatalf("Failed to read from reader: %s", err.Error())
		}

		if string(data) != "Readme\n" {
			t.Errorf("Wrong trunk/TODO content: %s", string(data))
		}
	}

	ci, err := r.Changeset(5, false)

	if err != nil {
		t.Errorf("Failed to get changeset: %s", err.Error())
	}

	diff := `Index: trunk/Makefile
===================================================================
--- trunk/Makefile	(revision 4)
+++ trunk/Makefile	(revision 5)
@@ -1,3 +1,4 @@
+# Make file to build newbc project
 export GOPATH := $(CURDIR)
 export LIBGIT_INSTALL_PREFIX := $(CURDIR)/vendor/libgit2_bin
 export LIBGIT_SRC_PATH := $(CURDIR)/vendor/libgit2
`
	if ci.ChangedPaths["trunk/Makefile"].Diff != diff {
		t.Errorf("Wrong file diff:\n%s\n\nExpected:\n%s", ci.ChangedPaths["trunk/Makefile"].Diff, diff)
	}

	ci, err = r.Changeset(6, false)

	if err != nil {
		t.Fatalf("Failed to get changeset 6: %s", err.Error())
	}

	diff = `Index: trunk/TODO
===================================================================
--- trunk/TODO	(revision 0)
+++ trunk/TODO	(revision 6)
@@ -0,0 +1 @@
+Readme
`
	if ci.ChangedPaths["trunk/TODO"].Diff != diff {
		t.Errorf("Wrong file diff:\n%s\n\nExpected:\n%s", ci.ChangedPaths["trunk/TODO"].Diff, diff)
	}

	ci, err = r.Changeset(9, true)

	if err != nil {
		t.Fatalf("Failed to get changeset 9: %s", err.Error())
	}

	if len(ci.ChangedPaths["trunk/Makefile"].Diff) > 0 {
		t.Errorf("Empty diff expected, but got \n%s", ci.ChangedPaths["trunk/Makefile"].Diff)
	}

	log := "White space change"

	if ci.Commit.Log != log {
		t.Errorf("Bad commit log message, expected '%s', got '%s'", log, ci.Commit.Log)
	}

	value, err := r.PropGet("trunk/img", 10, "svn:special")

	if err != nil {
		t.Errorf("Can not get prop: %s", err.Error())
	}

	if value != "*" {
		t.Errorf("Bad prop value for svn:special, got: %s", value)
	}

	props, err := r.PropList("trunk/img", 10)

	if err != nil {
		t.Errorf("Failed to get prop list: %s", err.Error())
	}

	if len(props) != 1 {
		t.Errorf("number of properties should be 1, got %d", len(props))
	}

	testPropKey := "svn:special"
	testPropVal := "*"
	if props[testPropKey] != testPropVal {
		t.Errorf("expected to get:%q, got: %q", testPropVal, props[testPropKey])
	}
}

func TestCreateRepo(t *testing.T) {
	path := filepath.Join(os.TempDir(), "svn-repo")
	os.RemoveAll(path)
	err := Create(path)

	if err != nil {
		t.Fatalf("Failed to create repo: %s", err.Error())
	}
}
