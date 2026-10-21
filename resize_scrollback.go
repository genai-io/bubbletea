package tea

import (
	"strings"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

// ResizeScrollbackMsg follows a resize, or a burst of them, and reports top
// rows of the inline frame the terminal itself pushed into its scrollback: a
// terminal that gets shorter keeps the cursor on its line and moves what no
// longer fits above it into history before the program redraws. Lines are
// those frame rows as drawn. Unscrolled are rows a [ScrollUp] had taken off the
// top that the resize kept from reaching the terminal; they are not in
// scrollback, and the program prints them again.
//
// Lines can fall short of what the terminal pushed, never past it: rows a
// narrowing rewraps are counted by the text drawn on them, broken greedily, the
// fewest rows any terminal breaks them into, and a row pushed only in part is
// left out.
//
// Known is false when the renderer cannot tell: the frame does not span the
// terminal, the alternate screen is active, input is disabled, something was
// written before the measurement, or the terminal did not report in time.
type ResizeScrollbackMsg struct {
	Known      bool
	Lines      []string
	Unscrolled []string
}

// resizeReportTimeout bounds how long a resize holds the redraw for the cursor
// report that measures it.
const resizeReportTimeout = 500 * time.Millisecond

type resizeReportTimeoutMsg struct{ probe int }

// resizeProbe is the frame a resize found on screen, kept until the terminal
// reports where that resize left the cursor.
type resizeProbe struct {
	lines     []string // frame rows as drawn
	drawn     []string // the text on each of lines, for rewrapping
	cursorRow int      // frame row the cursor was on
	pending   int      // cursor reports still to arrive
	id        int      // newest resize in this burst
	disturbed bool     // output moved the cursor before the report
	scrolled  []string // top rows of lines whose scroll the resize took back
}

// unshippedScroll is a scroll written to the renderer's buffer but not yet to
// the terminal: its line feeds would land on a terminal of another size.
type unshippedScroll struct {
	mark  int      // buffer length before the scroll
	x, y  int      // cursor before the scroll
	lines []string // rows it takes off the top, as drawn
}

// resizeReport is what a resize measured: the terminal pushed the first pushed
// lines into scrollback itself, and a scroll it took back is printed again
// where those did not cover it.
func resizeReport(lines []string, pushed int, scrolled []string) ResizeScrollbackMsg {
	if pushed < len(scrolled) {
		return ResizeScrollbackMsg{Known: true, Unscrolled: scrolled[pushed:]}
	}
	return ResizeScrollbackMsg{Known: true, Lines: lines[len(scrolled):pushed]}
}

// drawnText is the text a frame row puts on the terminal. Trailing cells the
// screen clears rather than writes (uv's canClearWith) hold nothing a terminal
// could rewrap.
func drawnText(line uv.Line) string {
	end := len(line)
	for end > 0 && clearedCell(&line[end-1]) {
		end--
	}
	var b strings.Builder
	for _, c := range line[:end] {
		b.WriteString(c.Content)
	}
	return b.String()
}

func clearedCell(c *uv.Cell) bool {
	return c.IsZero() || c.Content == " " && c.Width == 1 &&
		c.Style.Underline == uv.UnderlineNone &&
		c.Style.Attrs&^(uv.AttrBold|uv.AttrFaint|uv.AttrItalic|uv.AttrBlink|uv.AttrRapidBlink) == 0 &&
		c.Link.IsZero()
}

// rewrappedHeights returns how many terminal rows each of drawn takes once the
// terminal rewraps it at width, or one each when it does not rewrap. The text
// is broken greedily, the fewest rows any terminal breaks it into.
func rewrappedHeights(drawn []string, width int, rewrap bool) []int {
	heights := make([]int, len(drawn))
	for i, text := range drawn {
		heights[i] = 1
		if rewrap && width > 0 {
			heights[i] = strings.Count(ansi.Hardwrap(text, width, true), "\n") + 1
		}
	}
	return heights
}

// scrolledOff returns how many frame rows, of those with heights above the
// cursor, a terminal that left the cursor on screen row y has pushed whole into
// its scrollback; the terminal keeps the rows nearest the cursor.
func scrolledOff(heights []int, y int) int {
	pushed := -y
	for _, h := range heights {
		pushed += h
	}
	whole := 0
	for whole < len(heights) && heights[whole] <= pushed {
		pushed -= heights[whole]
		whole++
	}
	return whole
}
