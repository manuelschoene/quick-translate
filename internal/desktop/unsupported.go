//go:build !linux

package desktop

// Installs the application into the desktop of the operating system. Every not currently supported operating system answers with the same error.
func install(environment string) error {
	return ErrUnsupported
}

// Removes the application from the desktop of the operating system. Answers with ErrUnsupported for the same reason as install.
func uninstall() error {
	return ErrUnsupported
}

// Reports where the application has installed itself. Answers with ErrUnsupported for the same reason as install.
func status() (*Report, error) {
	return nil, ErrUnsupported
}
