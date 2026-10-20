package tea

import (
	"bytes"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestCursedRenderer_scrollUpShipsWithTheNextFrame(t *testing.T) {
	var out bytes.Buffer
	r := newCursedRenderer(&out, []string{"TERM=xterm-256color"}, 10, 5)
	r.start()
	r.resize(10, 5) // what Program does on start
	r.render(NewView("r0\nr1\nr2\nr3\nr4"))
	if err := r.flush(false); err != nil {
		t.Fatal(err)
	}
	out.Reset()

	if !r.scrollUp(2) {
		t.Fatal("a frame spanning the terminal must scroll")
	}
	if out.Len() != 0 {
		t.Fatalf("scroll reached the terminal before the frame filling it: %q", out.String())
	}
	r.render(NewView("r2\nr3\nr4\nn0\nn1"))
	if err := r.flush(false); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	if !strings.Contains(got, "\n\n") || !strings.Contains(got, "n0") || !strings.Contains(got, "n1") {
		t.Fatalf("expected two line feeds and the new rows, got %q", got)
	}
	for _, kept := range []string{"r2", "r3", "r4"} {
		if strings.Contains(got, kept) {
			t.Fatalf("row %s already on screen was repainted: %q", kept, got)
		}
	}
	if strings.Contains(got, ansi.EraseScreenBelow) {
		t.Fatalf("scroll erased the frame: %q", got)
	}
}

func TestCursedRenderer_scrollUpRefusesShortFrameAndAltScreen(t *testing.T) {
	var out bytes.Buffer
	r := newCursedRenderer(&out, []string{"TERM=xterm-256color"}, 10, 5)
	r.start()
	r.resize(10, 5) // what Program does on start
	r.render(NewView("r0\nr1\nr2"))
	if err := r.flush(false); err != nil {
		t.Fatal(err)
	}
	if r.scrollUp(1) {
		t.Fatal("a frame shorter than the terminal cannot scroll its rows into history")
	}

	alt := NewView("r0\nr1\nr2\nr3\nr4")
	alt.AltScreen = true
	r.render(alt)
	if r.scrollUp(1) {
		t.Fatal("the alternate screen has no scrollback to scroll into")
	}
}

func TestCursedRenderer_inlineFramePaintsNoHardTabs(t *testing.T) {
	var out bytes.Buffer
	r := newCursedRenderer(&out, []string{"TERM=xterm-256color"}, 40, 5)
	r.setOptimizations(true, false, false)
	r.start()
	r.resize(40, 5)
	r.render(NewView("a" + strings.Repeat(" ", 23) + "b"))
	if err := r.flush(false); err != nil {
		t.Fatal(err)
	}
	r.render(NewView("x" + strings.Repeat(" ", 23) + "y"))
	if err := r.flush(false); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "\t") {
		t.Fatalf("inline frame moved the cursor with a hard tab: %q", out.String())
	}
}
