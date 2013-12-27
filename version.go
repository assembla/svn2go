package svn

// #include <svn_version.h>
import "C"

type SvnVersion struct {
	Major int
	Minor int
	Patch int
	Tag   string
}

func Version() *SvnVersion {
	return &SvnVersion{
		Major: C.SVN_VER_MAJOR,
		Minor: C.SVN_VER_MINOR,
		Patch: C.SVN_VER_PATCH,
		Tag:   C.SVN_VER_TAG,
	}
}
