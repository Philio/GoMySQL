include $(GOROOT)/src/Make.inc
 
TARG=mysql
GOFILES=mysql.go const.go error.go packet.go reader.go writer.go password.go result.go statement.go
 
include $(GOROOT)/src/Make.pkg 
