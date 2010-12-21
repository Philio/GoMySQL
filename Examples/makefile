include $(GOROOT)/src/Make.inc

EXAMPLES = query reconnect statement

all: $(EXAMPLES)

query:
	$(GC) -o $(@).$(O) $(@).go
	$(LD) -s -o $(@) $(@).$(O)

reconnect:
	$(GC) -o $(@).$(O) $(@).go
	$(LD) -s -o $(@) $(@).$(O)

statement:
	$(GC) -o $(@).$(O) $(@).go
	$(LD) -s -o $(@) $(@).$(O)

clean:
	rm -rf *.[568oa] $(EXAMPLES)



