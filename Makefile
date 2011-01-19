include $(GOROOT)/src/Make.inc
 
TARG=mysql
GOFILES=mysql.go const.go error.go reader.go writer.go result.go statement.go
 
include $(GOROOT)/src/Make.pkg 
