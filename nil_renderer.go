package tea

import (
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
)

// nilRenderer is a no-op renderer. It implements the Renderer interface but
// doesn't render anything to the terminal.
type nilRenderer struct{}

var _ renderer = nilRenderer{}

// start implements renderer.
func (n nilRenderer) start() {}

// clearScreen implements renderer.
func (n nilRenderer) clearScreen() {}

// insertAbove implements renderer.
func (n nilRenderer) insertAbove(string) error { return nil }

// scrollUp implements renderer.
func (n nilRenderer) scrollUp(int, int, int) bool { return false }

// resize implements renderer.
func (n nilRenderer) resize(int, int) (Msg, int) { return ResizeScrollbackMsg{}, 0 }

// cursorReport implements renderer.
func (n nilRenderer) cursorReport(CursorPositionMsg) (Msg, bool) { return nil, false }

// expireResizeProbe implements renderer.
func (n nilRenderer) expireResizeProbe(int) Msg { return nil }

// setColorProfile implements renderer.
func (n nilRenderer) setColorProfile(colorprofile.Profile) {}

// flush implements the Renderer interface.
func (nilRenderer) flush(bool) error { return nil }

// close implements the Renderer interface.
func (nilRenderer) close() error { return nil }

// render implements the Renderer interface.
func (nilRenderer) render(View) {}

// reset implements the Renderer interface.
func (nilRenderer) reset() {}

// writeString implements the Renderer interface.
func (nilRenderer) writeString(string) (int, error) { return 0, nil }

// setSyncdUpdates implements the Renderer interface.
func (n nilRenderer) setSyncdUpdates(bool) {}

// setWidthMethod implements the Renderer interface.
func (n nilRenderer) setWidthMethod(ansi.Method) {}

// onMouse implements the Renderer interface.
func (n nilRenderer) onMouse(MouseMsg) Cmd {
	return nil
}
