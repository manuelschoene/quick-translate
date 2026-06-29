package clipboard

import (
	"fmt"
	"os"
	"runtime"
	"sync"
)

type clipboard interface {
	Read() (string, error)
	Write(string) error
}

type ClipboardService struct {
	cp clipboard
	mx sync.Mutex
}

func NewService() (*ClipboardService, error) {
	var cp clipboard

	switch goos := runtime.GOOS; goos {
	case "linux":
		switch sessionType := os.Getenv("XDG_SESSION_TYPE"); sessionType {
		case "x11":
			cp = x11Clipboard{}
		case "wayland":
			cp = waylandClipboard{}
		default:
			return nil, fmt.Errorf("Error: Unsupported session type: %s", sessionType)
		}

	default:
		return nil, fmt.Errorf("Error: Unsupported operating system: %s", goos)
	}

	return &ClipboardService{cp: cp, mx: sync.Mutex{}}, nil
}

func (cs *ClipboardService) Read() (string, error) {
	cs.mx.Lock()
	defer cs.mx.Unlock()
	return cs.cp.Read()
}

func (cs *ClipboardService) Write(data string) error {
	cs.mx.Lock()
	defer cs.mx.Unlock()
	return cs.cp.Write(data)
}