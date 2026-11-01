# Quick Translate @VERSION@ - linux/@ARCH@

The binary `@BINARY@` is built against GTK4 and WebKitGTK 6.0.

## Requirements

Install the requirements before you install the application. It does not start without WebKitGTK
6.0 and GTK 4.10 or newer, not even its CLI commands. Reading the selection needs no extra tool.
Under Wayland it needs a compositor with the data-control protocol (GNOME 49 or newer, KDE Plasma,
wlroots compositors); under older ones XWayland is used.

|           | Debian / Ubuntu      | Fedora            | Arch              |
|-----------|----------------------|-------------------|-------------------|
| WebKitGTK | `libwebkitgtk-6.0-4` | `webkitgtk6.0`    | `webkitgtk-6.0`   |

## Installation

The archive mirrors the directories it is installed into. For your user only:

```sh
install -D -m 0755 @BINARY@ ~/.local/bin/@APP_NAME@
cp -r share ~/.local/
```

Do not rename any of the files below `share/`: your desktop grants the global shortcut to the
application these names identify, and under different ones the shortcut stops working.

## Upgrading from a 0.1.0-alpha release

This version keeps the history and the language lists in a new database and no longer reads the
old files. Purge with the **old** binary before you install this one, then write your
configuration again:

```sh
@APP_NAME@ --purge
```

## Usage

An icon in the system tray shows that the application is running and offers general options. KDE
Plasma shows it out of the box. GNOME needs the AppIndicator extension. For other desktops, please
check yourself.

Start it once by hand. The first start writes a commented configuration to
`~/.config/quick-translate/config.yml` and exits, with a notification, until you have filled it in.

The shortcut is <kbd>Super</kbd>+<kbd>Shift</kbd>+<kbd>T</kbd> (some desktops call the key "Meta").
The application registers it with your desktop session on start, so there is nothing to bind by hand.
Under Wayland, most desktops ask you to grant the shortcut and allow you to change it.

## Help

The binary answers a few commands on the command line:

```sh
@APP_NAME@ version   # the version this binary was built as
@APP_NAME@ status    # what the application has put where on your system
@APP_NAME@ help      # every command the binary understands
```

## Uninstall

### Remove your files

```sh
@APP_NAME@ purge
```

Quit the application first. This removes everything it wrote for you: your configuration, the
database with your history, the autostart entry and more. **It cannot be undone.**

### Remove the application

This removes what you copied in during the installation. Run it after the purge, which needs the
binary:

```sh
rm ~/.local/bin/@APP_NAME@
rm ~/.local/share/applications/@APP_ID@.desktop
rm ~/.local/share/icons/hicolor/scalable/apps/@APP_ID@.svg
```
