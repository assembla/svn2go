package svn

/*
#include <svn_repos.h>
#include <svn_compat.h>
#include <svn_io.h>
#include <svn_string.h>
*/
import "C"
import (
	"unsafe"
)

// LogEntryReceiverCallback implements svn_repos_get_logs5 callback:
/*
    static svn_error_t *
	log_entry_receiver(void *baton,
                   svn_repos_log_entry_t *log_entry,
                   apr_pool_t *pool)
*/
//export LogEntryReceiverCallback
func LogEntryReceiverCallback(_obj, _entry, _pool unsafe.Pointer) unsafe.Pointer {
	baton := (*CommitCollector)(_obj)
	entry := (*C.svn_repos_log_entry_t)(_entry)

	var author, date, msg *C.char
	C.svn_compat_log_revprops_out(&author, &date, &msg, entry.revprops)
	commit := Commit{int64(entry.revision), C.GoString(author), C.GoString(date), C.GoString(msg)}
	baton.commits = append(baton.commits, commit)
	return nil
}

// HistoryReceiverCallback implements svn_repos_get_history2 callback
/* static svn_error_t *
	svn_repos_history_func_t(void *baton,
					const char *path,
                    svn_revnum_t *log_entry,
                    apr_pool_t *pool)
*/
//export HistoryReceiverCallback
func HistoryReceiverCallback(_obj unsafe.Pointer, path *C.char, revision C.svn_revnum_t, _pool unsafe.Pointer) *C.svn_error_t {
	baton := (*CommitCollector)(_obj)

	baton.revisions = append(baton.revisions, int64(revision))
	if len(baton.revisions) >= baton.limit {
		return C.svn_error_create(C.SVN_ERR_CEASE_INVOCATION, nil, nil)
	}

	return nil
}
