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
		t.Fatal(err)
	}

	defer r.Close()

	rev, err := r.LatestRevision()

	if err != nil {
		t.Fatal(err)
	}

	if rev <= 0 {
		t.Errorf("Latest revision should be >= 0, but it is %d", rev)
	}

	c, err := r.CommitInfo(1)

	if err != nil {
		t.Fatal(err)
	}

	if c.Author != "lz" {
		t.Fatalf("Wrong author: '%s'", c.Author)
	}

	commits, err := r.Commits(1, 2)

	if err != nil {
		t.Fatal(err)
	}

	if len(commits) != 2 {
		t.Error("it should return 2 commits")
	}

	for i := 0; i < len(commits); i++ {
		commit := commits[i]

		if commit.Author != "lz" {
			t.Errorf("Wrong author: '%s'", commit.Author)
		}
	}

	rev, err = r.LastPathRev("trunk/Makefile", 6)

	if err != nil {
		t.Error("failed to get rev", err)
	} else if rev != 5 {
		t.Error("Extected rev 5, got", rev)
	}

	commits, err = r.History("trunk", 0, rev, 2)

	if err != nil {
		t.Fatal(err)
	}

	if len(commits) != 2 {
		t.Error("#History should return 2 commits")
	}

	entries, err := r.Tree("trunk", 5)

	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 2 {
		t.Error("Only two entries should be in trunk/ folder at rev 5")
	}

	if reader, err := r.FileContent("trunk/TODO", 6); err != nil {
		t.Fatal(err)
	} else {
		data, e := ioutil.ReadAll(reader)

		if e != nil {
			t.Fatal(e)
		}

		if string(data) != "Readme\n" {
			t.Error("Wrong trunk/TODO content", string(data))
		}
	}

	ci, err := r.Changeset(5)

	if err != nil {
		t.Fatal(err)
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
		t.Errorf("Wrong file diff:\n%s\nExpected:\n%s", ci.ChangedPaths["trunk/Makefile"].Diff, diff)
	}

	// commit := ci.Commit
	// log.Println("Rev:", commit.Rev, "Date:", commit.Date, "Author:", commit.Author)
	// log.Println(commit.Log)

	// for key, value := range ci.ChangedPaths {
	// 	log.Printf("%s %s %d\n", value.Action, key, value.Kind)
	// 	log.Println(value.Diff)
	// }
}

func TestCreateRepo(t *testing.T) {
	path := filepath.Join(os.TempDir(), "svn-repo")
	os.RemoveAll(path)
	err := Create(path)

	if err != nil {
		t.Fatal(err)
	}
}
