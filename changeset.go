package svn

//#include <svn_types.h>
//#include <svn_pools.h>
//#include <svn_fs.h>
//#include <svn_delta.h>
//#include <svn_repos.h>
//#include <svn_diff.h>
/*
extern svn_error_t * Go_svn_repos_dir_delta2(svn_fs_root_t *src_root,
                     svn_fs_root_t *tgt_root,
                     const svn_delta_editor_t *editor,
                     void *edit_baton,
                     apr_pool_t *pool);
extern svn_stream_t * CreateWriterStream(void *baton, apr_pool_t *pool);
extern char * defaultEncoding();
extern svn_error_t * FileMimeType(svn_string_t **mimetype, svn_fs_root_t *root, const char *path, apr_pool_t *pool);
*/
import "C"

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"unsafe"
)

// ChangesetInfo is used to keep commit info
type ChangesetInfo struct {
	Commit       *Commit
	ChangedPaths map[string]*ChangedEntry
}

// Changeset returns changeset details: author, date, log, change files, dirs, diff of files and properties
func (r *Repo) Changeset(rev int64, diffIgnoreSpace bool) (*ChangesetInfo, error) {
	if rev == 0 {
		return nil, fmt.Errorf("Invalid revision: %d", rev)
	}

	rez := &ChangesetInfo{}
	var err error
	rez.Commit, err = r.CommitInfo(rev)

	if err != nil {
		return nil, err
	}

	pool := C.svn_pool_create_ex(r.pool, nil)
	defer C.svn_pool_destroy(pool)

	var (
		editor    *C.svn_delta_editor_t
		editBaton unsafe.Pointer
	)

	c := &collector{r: r, diffIgnoreSpace: diffIgnoreSpace}
	c.changes = make(map[string]*ChangedEntry)

	if e := C.svn_fs_revision_root(&c.baseRoot, r.fs, C.svn_revnum_t(rev)-1, pool); e != nil {
		return nil, makeError(e)
	}
	defer C.svn_fs_close_root(c.baseRoot)

	if e := C.svn_fs_revision_root(&c.root, r.fs, C.svn_revnum_t(rev), pool); e != nil {
		return nil, makeError(e)
	}
	defer C.svn_fs_close_root(c.root)

	// TODO improvement: rewrite editor #open_file #open_directory to limit entries count that we will process.
	// We had cases with crazy repos with 16k modifications
	if e := C.svn_repos_node_editor(&editor, &editBaton, r.repos, c.baseRoot, c.root, r.pool, pool); e != nil {
		return nil, makeError(e)
	}

	if e := C.Go_svn_repos_dir_delta2(c.baseRoot, c.root, editor, editBaton, pool); e != nil {
		return nil, makeError(e)
	}

	// if e := C.svn_repos_replay2(root, nil, -1, C.TRUE, editor, editBaton, nil, nil, pool); e != nil {
	// 	return nil, makeError(e)
	// }

	var tree *C.svn_repos_node_t
	tree = C.svn_repos_node_from_baton(editBaton)

	if err = c.collectChangedPaths(tree, ""); err != nil {
		return nil, err
	}

	rez.ChangedPaths = c.changes
	return rez, nil
}

// ChangedEntry is used for storing a changed entry
type ChangedEntry struct {
	Kind     int    // svn node kind
	Action   string // file operation A- added,R-changed,D-deleted
	Diff     string // file diff for changed text files
	IsBinary bool   // indicates if file is binary
}

type collector struct {
	baseRoot        *C.svn_fs_root_t
	root            *C.svn_fs_root_t
	r               *Repo
	changes         map[string]*ChangedEntry
	diffIgnoreSpace bool
}

// Collect changed paths
// similar with print_changed_tree() from svnlook.c
func (c *collector) collectChangedPaths(node *C.svn_repos_node_t, path string) error {
	if node == nil {
		return nil
	}

	countMe := true
	name := C.GoString(node.name)
	fullPath := filepath.Join(path, name)

	//log.Println(string(node.action), "Path", fullPath)

	switch node.action {
	case 'A':
	//node.copyfrom_path &&	node.copyfrom_rev
	case 'D':
	case 'R':
		if node.text_mod == C.FALSE && node.prop_mod == C.FALSE {
			countMe = false
		}
	default:
		countMe = false
	}

	var entry *ChangedEntry

	if countMe {
		entry = &ChangedEntry{Kind: int(node.kind), Action: string(node.action)}
		c.changes[fullPath] = entry
		//log.Printf("%s '%s'\n", string(node.action), fullPath)
	}

	if node.kind == C.svn_node_file {
		var (
			binaryA, binaryB bool
			rootA, rootB     *C.svn_fs_root_t
			rA, rB           int64
			err              error
		)

		// TODO: use can use here go routines, because it calculates, limit go routines to N
		header := fmt.Sprintf("Index: %s\n", fullPath)
		header += "===================================================================\n"
		cfullPath := C.CString(fullPath)
		defer C.free(unsafe.Pointer(cfullPath))

		if node.action == 'R' || node.action == 'D' {
			if binaryA, err = c.isBinary(c.baseRoot, cfullPath); err != nil {
				return err
			}
		}

		if node.action == 'R' || node.action == 'A' {
			if binaryB, err = c.isBinary(c.root, cfullPath); err != nil {
				return err
			}
		}

		var buf *C.svn_stringbuf_t
		cdiff := C.CString(entry.Diff)
		buf = C.svn_stringbuf_create(cdiff, c.r.pool)

		if binaryA || binaryB {
			entry.IsBinary = true
			header += "(Binary files differ)\n\n"
			C.svn_stringbuf_appendcstr(buf, C.CString(header))
		} else {
			// TODO check file size and do not dump files bigger than 10Mb for example
			rootA = c.baseRoot
			rootB = c.root

			if node.action == 'A' {
				rootA = nil
			}

			if node.action == 'D' {
				rootB = nil
			}

			f1Path, err := c.dumpFile(rootA, fullPath)
			if err != nil {
				return err
			}

			f2Path, err := c.dumpFile(rootB, fullPath)
			if err != nil {
				return err
			}

			//log.Println(f1Path, f2Path)
			defer os.Remove(f1Path)
			defer os.Remove(f2Path)

			cf1Path := C.CString(f1Path)
			defer C.free(unsafe.Pointer(cf1Path))
			cf2Path := C.CString(f2Path)
			defer C.free(unsafe.Pointer(cf2Path))

			var diff *C.svn_diff_t
			opts := C.svn_diff_file_options_create(c.r.pool) // svn_diff_file_options_t *

			if c.diffIgnoreSpace {
				opts.ignore_space = C.svn_diff_file_ignore_space_all
			}

			if err := C.svn_diff_file_diff_2(&diff, cf1Path, cf2Path, opts, c.r.pool); err != nil {
				return makeError(err)
			}

			if C.svn_diff_contains_diffs(diff) == C.TRUE {
				C.svn_stringbuf_appendcstr(buf, C.CString(header))

				//log.Println("Has diff")
				if rootA != nil {
					rA = int64(C.svn_fs_revision_root_revision(rootA))
				}
				if rootB != nil {
					rB = int64(C.svn_fs_revision_root_revision(rootB))
				}
				labelA := fmt.Sprintf("%s\t(revision %d)", fullPath, rA)
				labelB := fmt.Sprintf("%s\t(revision %d)", fullPath, rB)
				clabelA := C.CString(labelA)
				clabelB := C.CString(labelB)
				defer C.free(unsafe.Pointer(clabelA))
				defer C.free(unsafe.Pointer(clabelB))

				cstream := C.CreateWriterStream(unsafe.Pointer(buf), c.r.pool)

				if err := C.svn_diff_file_output_unified4(
					cstream,
					diff,
					cf1Path,
					cf2Path,
					clabelA,
					clabelB,
					C.defaultEncoding(),
					nil,
					C.FALSE,
					-1,
					nil,
					nil,
					c.r.pool,
				); err != nil {
					return makeError(err)
				}
			}
		} // end binary check

		entry.Diff = C.GoStringN(buf.data, (C.int)(buf.len))
	} // end file diff

	if node.prop_mod == C.TRUE {
		// TODO properties diff
	}

	tmp := (*C.svn_repos_node_t)(node.child)

	if tmp == nil {
		return nil
	}

	if err := c.collectChangedPaths(tmp, fullPath); err != nil {
		return err
	}

	for {
		tmp = (*C.svn_repos_node_t)(tmp.sibling)
		if tmp == nil {
			break
		}

		if err := c.collectChangedPaths(tmp, fullPath); err != nil {
			return err
		}
	}

	return nil
}

// Dumps repository file to temporary file
// Returns file path
// TODO check max file size to generate diff for
func (c *collector) dumpFile(fsRoot *C.svn_fs_root_t, path string) (string, error) {
	//log.Println("Dump", path)
	var err error
	tmp, err := ioutil.TempFile("", "svn-diff")

	if err != nil {
		return "", err
	}

	if fsRoot != nil {
		subPool := initSubPool(c.r.pool)
		svnStream, err := c.r.initSvnStream(fsRoot, subPool, path)
		if err != nil {
			return "", err
		}

		defer svnStream.Close()

		_, err = io.Copy(tmp, svnStream)

		if err != nil {
			return "", fmt.Errorf("Can not dump %s: %s", path, err.Error())
		}
	}

	if err = tmp.Close(); err != nil {
		return "", fmt.Errorf("Can not close temp file %s: %s", tmp.Name(), err.Error())
	}

	return tmp.Name(), nil
}

func (c *collector) isBinary(fsRoot *C.svn_fs_root_t, path *C.char) (bool, error) {
	var mimetype *C.svn_string_t

	if err := C.FileMimeType(&mimetype, fsRoot, path, c.r.pool); err != nil {
		return true, makeError(err)
	}

	if mimetype != nil && C.svn_mime_type_is_binary(mimetype.data) == C.TRUE {
		return true, nil
	}

	return false, nil
}
