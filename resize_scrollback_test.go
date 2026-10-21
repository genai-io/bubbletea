package tea

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
)

// Measured in tmux: a 26-row frame of 84-column rows drawn at 90 columns, the
// cursor on frame row 23, then resized; a row pushed only in part is not
// counted.
func TestScrolledOffOnResize(t *testing.T) {
	drawn := make([]string, 23)
	for i := range drawn {
		drawn[i] = fmt.Sprintf("R%02d %s", i+1, strings.Repeat("x", 80))
	}
	for _, tc := range []struct {
		name   string
		width  int
		y      int
		rewrap bool
		want   int
	}{
		{"shorter", 90, 19, true, 4},
		{"narrower", 60, 20, true, 13},
		{"both", 60, 18, true, 14},
		{"row pushed in part", 40, 17, true, 17},
		{"taller", 90, 23, true, 0},
		{"truncating terminal", 60, 18, false, 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := scrolledOff(rewrappedHeights(drawn, tc.width, tc.rewrap), tc.y); got != tc.want {
				t.Fatalf("got %d rows, want %d", got, tc.want)
			}
		})
	}
}

// Padding the screen clears rather than writes does not rewrap; a styled space
// is written and does.
func TestDrawnText(t *testing.T) {
	line := uv.NewLine(10)
	for i, r := range "ab" {
		line[i] = uv.Cell{Content: string(r), Width: 1}
	}
	if got := drawnText(line); got != "ab" {
		t.Fatalf("got %q", got)
	}
	line[5].Style.Attrs = uv.AttrReverse
	if got := drawnText(line); got != "ab    " {
		t.Fatalf("got %q", got)
	}
}

func TestResizeProbeHoldsRedrawUntilReport(t *testing.T) {
	var out bytes.Buffer
	r := newCursedRenderer(&out, nil, 10, 4)
	r.resize(10, 4)
	r.render(View{Content: "a\nb\nc\nd"})
	if err := r.flush(false); err != nil {
		t.Fatal(err)
	}
	if _, probe := r.resize(10, 3); probe == 0 {
		t.Fatal("frame spanning the terminal not probed")
	}
	out.Reset()
	r.render(View{Content: "b\nc\nd"})
	if err := r.flush(false); err != nil || out.Len() != 0 {
		t.Fatalf("redraw not held: %q %v", out.String(), err)
	}
	msg, ok := r.cursorReport(CursorPositionMsg{Y: 2})
	got, _ := msg.(ResizeScrollbackMsg)
	if !ok || !got.Known || len(got.Lines) != 1 || got.Lines[0] != "a" {
		t.Fatalf("report = %#v, %v", msg, ok)
	}
	if err := r.flush(false); err != nil || out.Len() == 0 {
		t.Fatalf("redraw still held: %v", err)
	}
	if _, ok := r.cursorReport(CursorPositionMsg{}); ok {
		t.Fatal("unrequested report taken")
	}
}

func TestResizeProbeSkipsShortFrame(t *testing.T) {
	r := newCursedRenderer(&bytes.Buffer{}, nil, 10, 4)
	r.resize(10, 4)
	r.render(View{Content: "a\nb"})
	_ = r.flush(false)
	if _, probe := r.resize(10, 3); probe != 0 {
		t.Fatal("short frame probed")
	}
}

func TestResizeTakesBackUnshippedScroll(t *testing.T) {
	var out bytes.Buffer
	r := newCursedRenderer(&out, nil, 10, 4)
	r.resize(10, 4)
	r.render(View{Content: "a\nb\nc\nd"})
	if err := r.flush(false); err != nil {
		t.Fatal(err)
	}
	if !r.scrollUp(2, 10, 4) {
		t.Fatal("scroll refused")
	}
	out.Reset()
	if _, probe := r.resize(10, 3); probe == 0 {
		t.Fatal("frame not probed")
	}
	// The terminal kept the cursor on frame row 3 and pushed row "a".
	msg, _ := r.cursorReport(CursorPositionMsg{Y: 2})
	got := msg.(ResizeScrollbackMsg)
	if len(got.Unscrolled) != 1 || got.Unscrolled[0] != "b" || len(got.Lines) != 0 {
		t.Fatalf("report = %#v", got)
	}
	r.render(View{Content: "c\nd\ne"})
	if err := r.flush(false); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "\n\n") {
		t.Fatalf("taken-back line feeds reached the terminal: %q", out.String())
	}
}

func TestScrollUpRefusesStaleSize(t *testing.T) {
	r := newCursedRenderer(&bytes.Buffer{}, nil, 10, 4)
	r.resize(10, 4)
	r.render(View{Content: "a\nb\nc\nd"})
	_ = r.flush(false)
	if r.scrollUp(1, 10, 5) || r.scrollUp(1, 9, 4) {
		t.Fatal("scroll measured for another size accepted")
	}
}

func TestResizeReport(t *testing.T) {
	lines := []string{"a", "b", "c", "d"}
	for _, tc := range []struct {
		name              string
		pushed            int
		scrolled          []string
		lines, unscrolled string
	}{
		{"pushed past the scroll", 3, []string{"a"}, "b c", ""},
		{"pushed part of the scroll", 1, []string{"a", "b"}, "", "b"},
		{"no scroll", 2, nil, "a b", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := resizeReport(lines, tc.pushed, tc.scrolled)
			if strings.Join(got.Lines, " ") != tc.lines || strings.Join(got.Unscrolled, " ") != tc.unscrolled {
				t.Fatalf("got %#v", got)
			}
		})
	}
}
