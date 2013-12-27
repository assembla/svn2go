package svn

/*
#include <svn_types.h>
#include <svn_compat.h>
*/
import "C"
import (
	"log"
	"unsafe"
)

/* static svn_error_t *
	log_entry_receiver(void *baton,
                   svn_log_entry_t *log_entry,
                   apr_pool_t *pool)
*/
//export LogEntryReceiverCallback
func LogEntryReceiverCallback(_obj, _entry, _pool unsafe.Pointer) unsafe.Pointer {
	baton := (*CommitCollector)(_obj)
	entry := (*C.svn_log_entry_t)(_entry)

	var author, date, msg *C.char
	C.svn_compat_log_revprops_out(&author, &date, &msg, entry.revprops)
	commit := &Commit{int64(entry.revision), C.GoString(author), C.GoString(date), C.GoString(msg)}
	baton.commits = append(baton.commits, commit)
	return nil
}

//export HistoryReceiverCallback
func HistoryReceiverCallback(_obj unsafe.Pointer, path *C.char, revision C.svn_revnum_t, _pool unsafe.Pointer) *C.svn_error_t {
	baton := (*CommitCollector)(_obj)

	if commit, err := baton.r.CommitInfo(int64(revision)); err != nil {
		log.Println(err)
		return nil // ignore
	} else {
		baton.commits = append(baton.commits, commit)

		if len(baton.commits) >= baton.limit {
			return C.svn_error_create(C.SVN_ERR_CEASE_INVOCATION, nil, nil)
		}
	}

	return nil
}

//export StreamWrite
func StreamWrite(_obj unsafe.Pointer, data *C.char, len *C.apr_size_t) *C.svn_error_t {
	out := (*stringBuffer)(_obj)

	_, err := out.Write([]byte(C.GoString(data)))

	if err != nil {
		log.Println(err)
	}

	return nil
}
