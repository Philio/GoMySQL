include $(GOROOT)/src/Make.inc
 
TARG=mysql
GOFILES=mysql.go const.go error.go command.go result.go statement.go
 
include $(GOROOT)/src/Make.pkg 
