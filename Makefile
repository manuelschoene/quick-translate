APP_NAME = quick-translate

INSTALL_DIR = $(HOME)/.local/bin
SYSTEMD_DIR = $(HOME)/.config/systemd/user
DESKTOP_DIR = $(HOME)/.local/share/applications
ICON_DIR = $(HOME)/.local/share/icons/hicolor/48x48/apps

SERVICE_FILE = quick-translate.service
DESKTOP_FILE = quick-translate.desktop
ICON_FILE = quick-translate_48x48.png

.PHONY: all build install-binary install-service install-desktop install clean uninstall

all: build

build:
	@printf "==> Compile Quick Translate...\n\n"
	wails build -tags webkit2_41 -upx

install-binary: build
	@printf "\n\n==> Install binary to $(INSTALL_DIR)...\n\n"
	mkdir -p $(INSTALL_DIR)
	cp build/bin/$(APP_NAME) $(INSTALL_DIR)/

install-service: install-binary
	@printf "\n\n==> Install and enable systemd service...\n\n"
	mkdir -p $(SYSTEMD_DIR)
	cp $(SERVICE_FILE) $(SYSTEMD_DIR)/
	systemctl --user daemon-reload
	systemctl --user enable --now $(SERVICE_FILE)

install-desktop: install-binary
	@printf "\n\n==> Create desktop entry...\n\n"

	mkdir -p $(ICON_DIR)
	cp art/$(ICON_FILE) $(ICON_DIR)/$(APP_NAME).png
	gtk-update-icon-cache -q -t -f ~/.local/share/icons/hicolor || true

	mkdir -p $(DESKTOP_DIR)
	@echo "[Desktop Entry]" > $(DESKTOP_DIR)/$(DESKTOP_FILE)
	@echo "Type=Application" >> $(DESKTOP_DIR)/$(DESKTOP_FILE)
	@echo "Name=Quick Translate" >> $(DESKTOP_DIR)/$(DESKTOP_FILE)
	@echo "Comment=Translate selected text anywhere" >> $(DESKTOP_DIR)/$(DESKTOP_FILE)
	@echo "Exec=$(INSTALL_DIR)/$(APP_NAME)" >> $(DESKTOP_DIR)/$(DESKTOP_FILE)
	@echo "Icon=$(ICON_DIR)/$(APP_NAME).png" >> $(DESKTOP_DIR)/$(DESKTOP_FILE)
	@echo "Terminal=false" >> $(DESKTOP_DIR)/$(DESKTOP_FILE)
	@echo "Categories=Utility;" >> $(DESKTOP_DIR)/$(DESKTOP_FILE)
	@echo "X-KDE-Shortcuts=Meta+T" >> $(DESKTOP_DIR)/$(DESKTOP_FILE)
	update-desktop-database $(DESKTOP_DIR) || true
	kbuildsycoca6 || kbuildsycoca5 || true

install: install-binary install-service install-desktop

clean:
	@printf "==> Cleaning build artifacts...\n\n"
	rm -rf build/bin/$(APP_NAME)

uninstall:
	@printf "==> Removing Quick Translate from system...\n\n"
	systemctl --user disable --now $(SERVICE_FILE) || true
	rm -f $(SYSTEMD_DIR)/$(SERVICE_FILE)
	systemctl --user daemon-reload

	rm -f $(INSTALL_DIR)/$(APP_NAME)

	rm -f $(DESKTOP_DIR)/$(DESKTOP_FILE)
	update-desktop-database $(DESKTOP_DIR) || true

	rm -f $(ICON_DIR)/$(APP_NAME).png
	gtk-update-icon-cache -q -t -f ~/.local/share/icons/hicolor || true
	