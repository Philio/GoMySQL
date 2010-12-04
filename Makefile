include $(GOROOT)/src/Make.inc
 
TARG=mysql
GOFILES=mysql.go mysql_const.go mysql_error.go mysql_packet.go mysql_result.go mysql_statement.go
 
include $(GOROOT)/src/Make.pkg 
