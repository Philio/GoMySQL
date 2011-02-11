include $(GOROOT)/src/Make.inc
 
TARG=mysql
GOFILES=mysql.go\
        const.go\
        error.go\
        password.go\
        reader.go\
        writer.go\
        packet.go\
	convert.go\
        result.go\
        statement.go
 
include $(GOROOT)/src/Make.pkg 
