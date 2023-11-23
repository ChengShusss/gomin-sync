TIME=$(shell date +"%Y-%m-%d %H:%M:%S")
LDFLAG= -X 'gomin-sync/internal/config.BuildTime=$(TIME)'

echo:
	@echo $(TIME)
	@echo $(LDFLAG)

.PHONY: build
build:
	go build --ldflags="$(LDFLAG)" -o build/gm-sync main.go

clean:
	rm build/gm-sync

install: build
	mkdir -p ~/.local/bin
	cp build/gm-sync ~/.local/bin