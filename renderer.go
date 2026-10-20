package tea

import (
	"fmt"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
)

const (
	// defaultFramerate specifies the maximum interval at which we should
	// update the view.
	defaultFPS = 60
	maxFPS     = 120
)

// renderer is the interface for Bubble Tea renderers.
type renderer interface {
	// start starts the renderer.
	start()

	// close closes the renderer and flushes any remaining data.
	close() error

	// render renders a frame to the output.
	render(View)

	// flush flushes the renderer's buffer to the output.
	flush(closing bool) error

	// reset resets the renderer's state to its initial state.
	reset()

	// insertAbove inserts unmanaged lines above the renderer.
	insertAbove(string) error

	// scrollUp scrolls the top rows of a frame spanning the terminal into its
	// scrollback, reporting whether it did.
	scrollUp(int) bool

	// setSyncdUpdates sets whether to use synchronized updates.
	setSyncdUpdates(bool)

	// setWidthMethod sets the method for calculating the width of the terminal.
	setWidthMethod(ansi.Method)

	// resize notify the renderer of a terminal resize.
	resize(int, int)

	// setColorProfile sets the color profile.
	setColorProfile(colorprofile.Profile)

	// clearScreen clears the screen.
	clearScreen()

	// writeString writes a string to the renderer's output.
	writeString(string) (int, error)

	// onMouse handles a mouse event.
	onMouse(MouseMsg) Cmd
}

type printLineMessage struct {
	messageBody string
}

// Println prints above the Program. This output is unmanaged by the program and
// will persist across renders by the Program.
//
// Unlike fmt.Println (but similar to log.Println) the message will be print on
// its own line.
//
// If the altscreen is active no output will be printed.
func Println(args ...any) Cmd {
	return func() Msg {
		return printLineMessage{
			messageBody: fmt.Sprint(args...),
		}
	}
}

type scrollUpMsg struct {
	rows int
}

// ScrollUpMsg reports the outcome of [ScrollUp]. When Scrolled is false
// nothing was written and the rows are still at the top of the frame.
type ScrollUpMsg struct {
	Rows     int
	Scrolled bool
}

// ScrollUp moves the top rows of an inline frame that spans the terminal into
// the terminal's scrollback, the way output scrolls in a plain terminal: the
// rows leave as they are already drawn, and the next frame paints only the rows
// that scrolled in. It is for rows the program is about to drop from the top of
// its view, so the view rendered when this runs must still show them there.
//
// The outcome arrives as a [ScrollUpMsg]. The scroll is refused when the frame
// does not span the terminal, when the alternate screen is active, or while a
// resize is being reconciled; the program then has to print the rows itself.
func ScrollUp(rows int) Cmd {
	return func() Msg {
		return scrollUpMsg{rows: rows}
	}
}

// Printf prints above the Program. It takes a format template followed by
// values similar to fmt.Printf. This output is unmanaged by the program and
// will persist across renders by the Program.
//
// Unlike fmt.Printf (but similar to log.Printf) the message will be print on
// its own line.
//
// If the altscreen is active no output will be printed.
func Printf(template string, args ...any) Cmd {
	return func() Msg {
		return printLineMessage{
			messageBody: fmt.Sprintf(template, args...),
		}
	}
}

// encodeCursorStyle returns the integer value for the given cursor style and
// blink state.
func encodeCursorStyle(style CursorShape, blink bool) int {
	// We're using the ANSI escape sequence values for cursor styles.
	// We need to map both [style] and [steady] to the correct value.
	style = (style * 2) + 1 //nolint:mnd
	if !blink {
		style++
	}
	return int(style)
}
