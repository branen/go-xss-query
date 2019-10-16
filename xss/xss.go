// Copyright 2019 Branen Salmon
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

// Package xss queries the X screen saver extension for information about
// screen saving and user input activity.
package xss

/*
#cgo LDFLAGS: -lX11 -lXss
#include <X11/Xlib.h>
#include <X11/extensions/saver.h>
#include <X11/extensions/scrnsaver.h>

Status QueryInfo (Display *disp, XScreenSaverInfo *info) {
	return XScreenSaverQueryInfo(disp, DefaultRootWindow(disp), info);
}
*/
import "C"
import (
	"fmt"
	"runtime"
	"sync"
	"time"
	"unsafe"
)

// Kind specifies one of three screen-saving strategies.
type Kind int

const (
	// The video PHY will be placed into a low-power mode when the screen
	// saver activates.
	Blanked Kind = iota
	// The X server will display an image when the screen saver activates.
	Internal
	// An X client will display an image when the screen saver activates.
	External
)

func (k Kind) String() string {
	switch k {
	case Blanked:
		return "blanked"
	case Internal:
		return "internal"
	case External:
		return "external"
	}
	return "unknown"
}

type Info struct {
	// Enabled specifies whether the screen saver is enabled (or disabled).
	Enabled bool
	// Active specifies whether the screen saver is active (or inactive).
	Active bool
	// Kind specifies which screen-saving strategy will be used when
	// the screen saver activates.
	Kind Kind
	// Countdown counts down the time until the screen saver activates.
	Countdown time.Duration
	// ActiveTime measures the time that the screen saver has been active.
	// (It resets to zero when the screen saver becomes inactive.)
	ActiveTime time.Duration
	// IdleTime measures the time since the last user input event.
	IdleTime time.Duration
}

type Client struct {
	disp  *C.Display
	info  *C.XScreenSaverInfo
	mutex sync.Mutex
}

// NewClient creates a persistent, thread-safe connection to the XSS extension.
func NewClient() (*Client, error) {
	disp := C.XOpenDisplay(nil)
	if disp == nil {
		return nil, fmt.Errorf("Could not open X display.")
	}
	var base, errbase C.int
	if C.XScreenSaverQueryExtension(
		disp,
		&base,
		&errbase,
	) == 0 {
		return nil, fmt.Errorf("XSS extension not active.")
	}
	cl := &Client{
		disp: disp,
		info: C.XScreenSaverAllocInfo(),
	}
	runtime.SetFinalizer(cl, func(cl *Client) {
		C.XFree(unsafe.Pointer(cl.info))
	})
	return cl, nil
}

// Query queries the XSS extension.
func (cl Client) Query() (i Info, err error) {
	if cl.disp == nil {
		panic("Client instances must be created with NewClient.")
	}
	var status C.Status
	(func() {
		cl.mutex.Lock()
		defer cl.mutex.Unlock()
		status = C.QueryInfo(cl.disp, cl.info)
	})()
	if status == 0 {
		err = fmt.Errorf("Error querying XSS.")
		return
	}
	i.IdleTime = time.Duration(cl.info.idle) * time.Millisecond
	switch cl.info.state {
	case C.ScreenSaverOn:
		i.Enabled = true
		i.Active = true
		i.ActiveTime =
			time.Duration(cl.info.til_or_since) * time.Millisecond
	case C.ScreenSaverOff:
		i.Enabled = true
		i.Active = false
		i.Countdown =
			time.Duration(cl.info.til_or_since) * time.Millisecond
	case C.ScreenSaverDisabled:
		i.Enabled = false
		i.Active = false
	}
	switch cl.info.kind {
	case C.ScreenSaverBlanked:
		i.Kind = Blanked
	case C.ScreenSaverInternal:
		i.Kind = Internal
	case C.ScreenSaverExternal:
		i.Kind = External
	}
	return
}
