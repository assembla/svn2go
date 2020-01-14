package svn

// #include <svn_version.h>
import "C"

// Version is used to keep SVN version details
type Version struct {
	Major int
	Minor int
	Patch int
	Tag   string
}

// GetVersion returns current version
func GetVersion() *Version {
	return &Version{
		Major: C.SVN_VER_MAJOR,
		Minor: C.SVN_VER_MINOR,
		Patch: C.SVN_VER_PATCH,
		Tag:   C.SVN_VER_TAG,
	}
}
