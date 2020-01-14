#include <svn_types.h>
#include <svn_props.h>
#include <svn_repos.h>
#include <svn_hash.h>

#include "_cgo_export.h"

svn_error_t *
Go_svn_repos_get_logs5(svn_repos_t *repos,
                    const apr_array_header_t *paths,
                    svn_revnum_t start,
                    svn_revnum_t end,
                    int limit,
                    svn_boolean_t strict_node_history,
                    svn_boolean_t include_merged_revisions,
                    const apr_array_header_t *revprops,
                    /*svn_repos_authz_func_t authz_read_func,
                    void *authz_read_baton,
                    svn_log_entry_receiver_t receiver,*/
                    void *receiver_baton,
                    apr_pool_t *pool) {
	return svn_repos_get_logs5(repos,
                    paths,
                    start,
                    end,
                    limit,
                    strict_node_history,
                    include_merged_revisions,
                    revprops,
                    NULL,
                    NULL,
                    NULL,
                    NULL,
                    (svn_repos_log_entry_receiver_t)&LogEntryReceiverCallback,
                    receiver_baton,
                    pool);
}

svn_error_t *
Go_svn_repos_history2(svn_fs_t *fs,
                   const char *path,
                   void *history_baton,
                   svn_revnum_t start,
                   svn_revnum_t end,
                   svn_boolean_t cross_copies,
                   apr_pool_t *pool) {
    return svn_repos_history2(fs,
                   path,
                   (svn_repos_history_func_t)&HistoryReceiverCallback,
                   history_baton,
                   NULL,
                   NULL,
                   start,
                   end,
                   cross_copies,
                   pool);
}

svn_error_t * Go_svn_repos_dir_delta2(svn_fs_root_t *src_root,
                     svn_fs_root_t *tgt_root,
                     const svn_delta_editor_t *editor,
                     void *edit_baton,
                     apr_pool_t *pool) {
  return svn_repos_dir_delta2(src_root,
                     "",
                     "",
                     tgt_root,
                     "",
                     editor,
                     edit_baton,
                     NULL, // authz func
                     NULL, // authz baton
                     TRUE,
                     svn_depth_infinity,
                     FALSE,
                     FALSE,
                     pool);
}

void init_fs_config(apr_hash_t *fsConfig) {
	//svn_hash_sets(fsConfig, SVN_FS_CONFIG_PRE_1_5_COMPATIBLE, "1");
	svn_hash_sets(fsConfig, SVN_FS_CONFIG_PRE_1_8_COMPATIBLE, "1");
	svn_hash_sets(fsConfig, SVN_FS_CONFIG_FS_TYPE, SVN_FS_TYPE_FSFS);
}

apr_array_header_t * GoCreateAprArrayForPath(const char *path, apr_pool_t *pool) {
	apr_array_header_t *paths = apr_array_make(pool, 1, sizeof(const char *));
	APR_ARRAY_IDX(paths, 0, const char *) = path;
	return paths;
}

apr_array_header_t * GoDefaultLogProps(apr_pool_t *pool) {
      apr_array_header_t *revprops = apr_array_make(pool, 3, sizeof(char *));
      APR_ARRAY_PUSH(revprops, const char *) = SVN_PROP_REVISION_AUTHOR;
      APR_ARRAY_PUSH(revprops, const char *) = SVN_PROP_REVISION_DATE;
      APR_ARRAY_PUSH(revprops, const char *) = SVN_PROP_REVISION_LOG;
      return revprops;
}

char * GoPropAuthor() {
	return SVN_PROP_REVISION_AUTHOR;
}

// Go can not get defined properties that use other defined properties
char * PropAuthor(apr_hash_t *hash) {
	return (char*)svn_hash_gets(hash, SVN_PROP_REVISION_AUTHOR);
}

svn_stream_t * CreateWriterStream(void *baton, apr_pool_t *pool) {
  svn_stream_t * rez = svn_stream_create(baton, pool);
  svn_stream_set_write(rez, &StreamWrite);
  return rez;
}

svn_error_t * FileMimeType(svn_string_t **mimetype, svn_fs_root_t *root, const char *path, apr_pool_t *pool) {
  SVN_ERR(svn_fs_node_prop(mimetype, root, path, SVN_PROP_MIME_TYPE, pool));
  return SVN_NO_ERROR;
}

char *
defaultEncoding() {
  return "UTF-8";
}
