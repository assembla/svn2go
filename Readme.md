Subversion bindings for go.

#Install

It needs subversion >= 1.13.0

You will need libsvn-dev on Ubuntu/Debian or subversion installed with `brew` on OSX. apr-dev and apr-util-dev are also required.

    go get -u github.com/Assembla/svn2go

See svn_test.go for usage.

#TODO

* Make private svn callbacks
* Do not use pointer in slices, they already are
* Add Export for svnadmin export command
