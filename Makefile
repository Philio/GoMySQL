include $(GOROOT)/src/Make.inc
 
TARG=mysql
GOFILES=mysql.go\
		types.go\
		const.go\
		error.go\
		password.go\
		reader.go\
		writer.go\
		packet.go\
		convert.go\
		handler.go\
		result.go\
		statement.go
 
include $(GOROOT)/src/Make.pkg 
