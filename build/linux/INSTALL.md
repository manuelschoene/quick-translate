# Quick Translate @VERSION@ - linux/@ARCH@

Two binaries are included, one per Linux graphics stack:

- `@BINARY@` - GTK4 and WebKitGTK 6.0
- `@BINARY_GTK3@` - GTK3 and WebKit2GTK 4.1, for older distributions

## Requirements

Install the requirements before you install the application. Neither binary starts without its
WebKit library, not even CLI commands. Use the GTK4 binary where your distribution
offers WebKitGTK 6.0, and the GTK3 one only where it does not. Reading the selection needs
wl-clipboard in a Wayland session, and xclip or xsel in an X11 session.

|             | Debian / Ubuntu       | Fedora            | Arch              |
|-------------|-----------------------|-------------------|-------------------|
| GTK4 binary | `libwebkitgtk-6.0-4`  | `webkitgtk6.0`    | `webkitgtk-6.0`   |
| GTK3 binary | `libwebkit2gtk-4.1-0` | `webkit2gtk4.1`   | `webkit2gtk-4.1`  |
| Wayland     | `wl-clipboard`        | `wl-clipboard`    | `wl-clipboard`    |
| X11         | `xclip` or `xsel`     | `xclip` or `xsel` | `xclip` or `xsel` |

## Installation

The archive mirrors the directories it is installed into. For your user only, with the GTK4 binary:

```sh
install -D -m 0755 @BINARY@ ~/.local/bin/@APP_NAME@
cp -r share ~/.local/
```

For the GTK3 build, install `@BINARY_GTK3@` under the same name instead:

```sh
install -D -m 0755 @BINARY_GTK3@ ~/.local/bin/@APP_NAME@
cp -r share ~/.local/
```

Do not rename any of the files below `share/`: your desktop grants the global shortcut to the
application these names identify, and under different ones the shortcut stops working.

## Usage

An icon in the system tray shows that the application is running and offers general options. KDE
Plasma shows it out of the box. GNOME needs the AppIndicator extension. For other desktops, please
check yourself.

The shortcut is <kbd>Super</kbd>+<kbd>Shift</kbd>+<kbd>T</kbd> (Some desktops may call it "Meta" and not "Super"). 
The application registeres it with your desktop session on start, so there is nothing to bind by
hand. Under Wayland, most desktops will ask you to grant the shortcut and allow you to change it.

## Help

The binary answers a few commands on the command line:

```sh
@APP_NAME@ --version   # the version this binary was built as
@APP_NAME@ --status    # what the application has put where on your system
@APP_NAME@ --help      # every command the binary understands
```

## Uninstall

### Remove your files

```sh
@APP_NAME@ --purge
```

This removes everything the application wrote for you - your configuration, your history, the cache and more.
This cannot be undone afterwards! Quit the application before you
run it.

### Remove the application

This removes what you copied in during the installation. Run it after the purge, which needs the
binary:

```sh
rm ~/.local/bin/@APP_NAME@
rm ~/.local/share/applications/@APP_ID@.desktop
rm ~/.local/share/icons/hicolor/scalable/apps/@APP_ID@.svg
```
