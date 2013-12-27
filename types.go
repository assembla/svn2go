package svn

/*
#include <svn_types.h>
#include <svn_repos.h>
*/
import "C"

type Commit struct {
	Rev    int64 // C.svn_revnum_t
	Author string
	Date   string //Sample "2001-08-31T04:24:14.966996Z"
	Log    string
}

const (
	PropRevAuthor = "svn:author" //Can not use because of linker error: C.SVN_PROP_REVISION_AUTHOR
	PropRevDate   = "svn:date"   //C.SVN_PROP_REVISION_DATE
	PropRevLog    = "svn:log"    //C.SVN_PROP_REVISION_LOG
	EntryTypeNone = C.svn_node_none
	EntryTypeDir  = C.svn_node_dir
)

type Repo struct {
	Path  string
	repos *C.svn_repos_t
	fs    *C.svn_fs_t
	pool  *C.apr_pool_t
	rev   int64
}

type DirEntry struct {
	Name string
	Kind int
}

type CommitCollector struct {
	commits []*Commit
	limit   int
	r       *Repo
}
