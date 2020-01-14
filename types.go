package svn

/*
#include <svn_types.h>
#include <svn_repos.h>
#include <svn_pools.h>
#include <svn_props.h>
*/
import "C"

import "runtime"

// Commit stores commit details
type Commit struct {
	Rev    int64 // C.svn_revnum_t
	Author string
	Date   string //Sample "2001-08-31T04:24:14.966996Z"
	Log    string
}

const (
	// PropRevAuthor is for author
	PropRevAuthor = C.SVN_PROP_REVISION_AUTHOR
	// PropRevDate is for date
	PropRevDate   = C.SVN_PROP_REVISION_DATE
	// PropRevLog is for log
	PropRevLog    = C.SVN_PROP_REVISION_LOG
	// EntryTypeNone is for unknown entry types
	EntryTypeNone = C.svn_node_none
	// EntryTypeDir is for dirs
	EntryTypeDir  = C.svn_node_dir
)

// Repo wraps a SVN repo
type Repo struct {
	Path  string
	repos *C.svn_repos_t
	fs    *C.svn_fs_t
	pool  *C.apr_pool_t
	rev   int64
}

// DirEntry wraps a dir entry
type DirEntry struct {
	Name string
	Kind int
}

// NewCommitCollector returns a commit collector for SVN callbacks
func NewCommitCollector(pool *C.apr_pool_t) *CommitCollector {
	c := &CommitCollector{commits: make([]Commit, 0), pool: pool}
	runtime.SetFinalizer(c, (*CommitCollector).Free)
	return c
}

// CommitCollector is used for collecting commit data
type CommitCollector struct {
	commits []Commit
	revisions []int64
	limit   int
	r       *Repo
	pool    *C.apr_pool_t
}

// Free frees C objects memory from commit collector
func (c *CommitCollector) Free() {
	runtime.SetFinalizer(c, nil)
	C.svn_pool_destroy(c.pool)
}
