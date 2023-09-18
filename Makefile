SHELL := /bin/bash

payloads/redirector:
	$(MAKE) -C src/go bin/redirector
	mkdir -p payloads
	cp src/go/bin/redirector payloads/redirector
