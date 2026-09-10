APP_NAME = quick-translate

# Where the binary is installed to. Overridable so the application can also be installed somewhere else.
PREFIX ?= $(HOME)/.local
BIN_DIR = $(PREFIX)/bin
BINARY = build/bin/$(APP_NAME)
INSTALLED = $(BIN_DIR)/$(APP_NAME)

# Wails renders the frontend with webkit2gtk, which ships as 4.0 on older distributions and as 4.1 on
# current ones. Its 4.1 bindings are only compiled in when the build tag is set, so the tag is set for
# whichever version pkg-config finds installed.
WEBKIT_TAGS := $(shell pkg-config --exists webkit2gtk-4.1 2>/dev/null && echo webkit2_41)
TAGS = $(if $(WEBKIT_TAGS),-tags $(WEBKIT_TAGS),)

# The binary is compressed with UPX, which takes a noticeable while. Pass UPX=0 to skip it while testing
# locally; a UPX that is not installed is skipped either way.
UPX ?= 1
UPX_FOUND := $(shell command -v upx >/dev/null 2>&1 && echo yes)
COMPRESS = $(if $(filter-out 0,$(UPX)),$(if $(UPX_FOUND),-upx,),)

# The icon set the installation ships is rendered from the master artwork and checked in, so building the
# application needs no image tooling at all. Only 'make icons' does, and only after the artwork changed.
ICON_MASTER = art/quick-translate.png
ICON_DIR = internal/desktop/assets/linux/icons
ICON_SIZES = 16 22 24 32 48 64 128 256

# The desktop the application is installed for. The binary detects it from the session when this is not
# set; pass DESKTOP=kde for the Plasma installation or DESKTOP=generic to force the portable one.
DESKTOP ?=
DESKTOP_FLAG = $(if $(DESKTOP),--desktop=$(DESKTOP),)

# The install targets build on each other through the file system rather than through prerequisites, so
# they must never overlap: 'install-desktop' runs the binary that 'install-binary' has just put in place.
.NOTPARALLEL:

.PHONY: all help dev build icons install install-binary install-desktop status uninstall clean

all: build

help:
	@printf "Quick Translate\n\n"
	@printf "  make build            Compile the application to $(BINARY).\n"
	@printf "  make dev              Run the application with live reload.\n"
	@printf "  make install          Build, install the binary and register it with the desktop.\n"
	@printf "  make install-binary   Copy an already built binary to $(BIN_DIR).\n"
	@printf "  make install-desktop  Register the installed binary with the desktop.\n"
	@printf "  make status           Show where the application has installed itself.\n"
	@printf "  make uninstall        Remove the application from the desktop and delete the binary.\n"
	@printf "  make clean            Delete the built binary.\n"
	@printf "  make icons            Re-render the icon set from $(ICON_MASTER) (needs ImageMagick).\n\n"
	@printf "Variables: PREFIX=$(PREFIX)  DESKTOP=<kde|generic>  UPX=<1|0>\n"

dev:
	wails dev $(TAGS)

build:
	@printf "==> Compiling Quick Translate...\n"
	@$(if $(WEBKIT_TAGS),,printf "    webkit2gtk-4.1 was not found, building against webkit2gtk-4.0.\n")
	@$(if $(filter-out 0,$(UPX)),$(if $(UPX_FOUND),,printf "    upx is not installed, the binary is not compressed.\n"),)
	wails build $(TAGS) $(COMPRESS)

icons:
	@command -v magick >/dev/null || { printf "ImageMagick is required to render the icon set.\n"; exit 1; }
	@printf "==> Rendering the icon set from $(ICON_MASTER)...\n"
	@for size in $(ICON_SIZES); do \
		magick $(ICON_MASTER) -filter Lanczos -resize $${size}x$${size} -strip PNG32:$(ICON_DIR)/$${size}.png; \
	done
	@ls -1 $(ICON_DIR)

# Installing is deliberately kept apart from building, so a binary from a release can be installed with
# the same target without a toolchain being present.
install-binary:
	@test -f $(BINARY) || { printf "'$(BINARY)' does not exist. Run 'make build' first.\n"; exit 1; }
	@printf "\n==> Installing the binary to $(BIN_DIR)...\n"
	install -D -m 0755 $(BINARY) $(INSTALLED)

# The binary registers itself with the desktop, because what that means differs per operating system and
# is therefore implemented next to the code that knows about the operating system.
install-desktop:
	@printf "\n"
	$(INSTALLED) --install $(DESKTOP_FLAG)

install: build install-binary install-desktop

status:
	@test -x $(INSTALLED) || { printf "'$(INSTALLED)' is not installed.\n"; exit 1; }
	@$(INSTALLED) --status

uninstall:
	@test -x $(INSTALLED) && $(INSTALLED) --uninstall || printf "'$(INSTALLED)' is not installed.\n"
	@printf "\n==> Removing the binary from $(BIN_DIR)...\n"
	rm -f $(INSTALLED)

clean:
	@printf "==> Deleting build artifacts...\n"
	rm -f $(BINARY)
