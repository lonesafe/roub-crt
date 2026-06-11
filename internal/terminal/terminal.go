package terminal

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"golang.org/x/term"
)

// MakeRawTerminal wrapper
func MakeRawTerminal() (interface{}, error) {
	return term.MakeRaw(int(os.Stdin.Fd()))
}

type Terminal struct {
	Width    int
	Height   int
	Screen   [][]rune
	mu       sync.RWMutex
	cursorX  int
	cursorY  int
	FG       lipgloss.Color
	BG       lipgloss.Color
	ScrollTop int
	ScrollBottom int
	Scrollback [][]rune
	ScrollbackSize int
	RawMode  bool
}

type TerminalOption func(*Terminal)

func NewTerminal(width, height int, options ...TerminalOption) *Terminal {
	t := &Terminal{
		Width:           width,
		Height:          height,
		Screen:          make([][]rune, height),
		cursorX:         0,
		cursorY:         0,
		FG:              lipgloss.Color("#D4D4D4"),
		BG:              lipgloss.Color("#1E1E1E"),
		ScrollTop:       0,
		ScrollBottom:    height - 1,
		ScrollbackSize:  10000,
		Scrollback:      make([][]rune, 0),
	}

	for i := range t.Screen {
		t.Screen[i] = make([]rune, width)
		for j := range t.Screen[i] {
			t.Screen[i][j] = ' '
		}
	}

	for _, opt := range options {
		opt(t)
	}

	return t
}

func WithScrollbackSize(size int) TerminalOption {
	return func(t *Terminal) {
		t.ScrollbackSize = size
		t.Scrollback = make([][]rune, 0, size)
	}
}

func (t *Terminal) Resize(width, height int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if width == t.Width && height == t.Height {
		return
	}

	newScreen := make([][]rune, height)
	for i := range newScreen {
		newScreen[i] = make([]rune, width)
		for j := range newScreen[i] {
			newScreen[i][j] = ' '
		}
	}

	minH := min(height, t.Height)
	minW := min(width, t.Width)
	for y := 0; y < minH; y++ {
		for x := 0; x < minW; x++ {
			newScreen[y][x] = t.Screen[y][x]
		}
	}

	t.Screen = newScreen
	t.Width = width
	t.Height = height

	if t.cursorY >= height {
		t.cursorY = height - 1
	}
	if t.cursorX >= width {
		t.cursorX = width - 1
	}
}

func (t *Terminal) WriteChar(ch rune) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.cursorX >= t.Width {
		t.cursorX = 0
		t.cursorY++
	}

	if t.cursorY >= t.Height {
		t.scrollUp()
		t.cursorY = t.Height - 1
	}

	t.Screen[t.cursorY][t.cursorX] = ch
	t.cursorX++
}

func (t *Terminal) WriteString(s string) {
	for _, ch := range s {
		t.WriteChar(ch)
	}
}

func (t *Terminal) scrollUp() {
	row := make([]rune, t.Width)
	for x := 0; x < t.Width; x++ {
		row[x] = t.Screen[0][x]
	}

	t.Scrollback = append(t.Scrollback, row)

	if len(t.Scrollback) > t.ScrollbackSize {
		t.Scrollback = t.Scrollback[1:]
	}

	for y := 0; y < t.Height-1; y++ {
		t.Screen[y] = t.Screen[y+1]
	}

	t.Screen[t.Height-1] = make([]rune, t.Width)
	for x := 0; x < t.Width; x++ {
		t.Screen[t.Height-1][x] = ' '
	}
}

func (t *Terminal) MoveCursor(x, y int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if x < 0 {
		x = 0
	}
	if x >= t.Width {
		x = t.Width - 1
	}
	if y < 0 {
		y = 0
	}
	if y >= t.Height {
		y = t.Height - 1
	}

	t.cursorX = x
	t.cursorY = y
}

func (t *Terminal) MoveCursorUp(n int) {
	t.MoveCursor(t.cursorX, t.cursorY-n)
}

func (t *Terminal) MoveCursorDown(n int) {
	t.MoveCursor(t.cursorX, t.cursorY+n)
}

func (t *Terminal) MoveCursorLeft(n int) {
	t.MoveCursor(t.cursorX-n, t.cursorY)
}

func (t *Terminal) MoveCursorRight(n int) {
	t.MoveCursor(t.cursorX+n, t.cursorY)
}

func (t *Terminal) NewLine() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.cursorX = 0
	t.cursorY++

	if t.cursorY >= t.Height {
		t.scrollUp()
		t.cursorY = t.Height - 1
	}
}

func (t *Terminal) CarriageReturn() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.cursorX = 0
}

func (t *Terminal) Backspace() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.cursorX > 0 {
		t.cursorX--
		t.Screen[t.cursorY][t.cursorX] = ' '
	}
}

func (t *Terminal) DeleteLine() {
	t.mu.Lock()
	defer t.mu.Unlock()

	for x := 0; x < t.Width; x++ {
		t.Screen[t.cursorY][x] = ' '
	}
}

func (t *Terminal) DeleteChars(n int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for x := t.cursorX; x < t.Width && x < t.cursorX+n; x++ {
		t.Screen[t.cursorY][x] = ' '
	}
}

func (t *Terminal) InsertChars(n int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for x := t.Width - 1; x >= t.cursorX+n; x-- {
		t.Screen[t.cursorY][x] = t.Screen[t.cursorY][x-n]
	}
	for x := t.cursorX; x < t.cursorX+n && x < t.Width; x++ {
		t.Screen[t.cursorY][x] = ' '
	}
}

func (t *Terminal) ClearScreen() {
	t.mu.Lock()
	defer t.mu.Unlock()

	for y := 0; y < t.Height; y++ {
		for x := 0; x < t.Width; x++ {
			t.Screen[y][x] = ' '
		}
	}
	t.cursorX = 0
	t.cursorY = 0
}

func (t *Terminal) ClearEOS() {
	t.mu.Lock()
	defer t.mu.Unlock()

	for y := t.cursorY; y < t.Height; y++ {
		for x := 0; x < t.Width; x++ {
			if y == t.cursorY && x <= t.cursorX {
				continue
			}
			t.Screen[y][x] = ' '
		}
	}
}

func (t *Terminal) ClearBOS() {
	t.mu.Lock()
	defer t.mu.Unlock()

	for y := 0; y <= t.cursorY; y++ {
		for x := 0; x < t.Width; x++ {
			if y == t.cursorY && x >= t.cursorX {
				break
			}
			t.Screen[y][x] = ' '
		}
	}
}

func (t *Terminal) SetFg(color lipgloss.Color) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.FG = color
}

func (t *Terminal) SetBg(color lipgloss.Color) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.BG = color
}

func (t *Terminal) ResetAttributes() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.FG = lipgloss.Color("#D4D4D4")
	t.BG = lipgloss.Color("#1E1E1E")
}

func (t *Terminal) GetCell(x, y int) (rune, lipgloss.Color, lipgloss.Color) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if x < 0 || x >= t.Width || y < 0 || y >= t.Height {
		return ' ', t.FG, t.BG
	}

	return t.Screen[y][x], t.FG, t.BG
}

func (t *Terminal) Render() string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var result string

	for y := 0; y < t.Height; y++ {
		for x := 0; x < t.Width; x++ {
			ch := t.Screen[y][x]
			width := runewidth.RuneWidth(ch)
			if width == 0 {
				width = 1
			}
			result += string(ch)
			x += width - 1
		}
		if y < t.Height-1 {
			result += "\n"
		}
	}

	return lipgloss.NewStyle().
		Foreground(t.FG).
		Background(t.BG).
		Render(result)
}

func (t *Terminal) GetScrollback() [][]rune {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make([][]rune, len(t.Scrollback))
	copy(result, t.Scrollback)
	return result
}

func (t *Terminal) SetScrollRegion(top, bottom int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if top < 0 {
		top = 0
	}
	if bottom >= t.Height {
		bottom = t.Height - 1
	}
	if top > bottom {
		return
	}

	t.ScrollTop = top
	t.ScrollBottom = bottom
}

func (t *Terminal) ReverseIndex() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.cursorY > t.ScrollTop {
		t.cursorY--
	}
}

type PTY struct {
	Master  *os.File
	Slave   *os.File
	Terminal *Terminal
	width   int
	height  int
	done    chan struct{}
}

func NewPTY(width, height int) (*PTY, error) {
	pty, tty, err := openPTY()
	if err != nil {
		return nil, fmt.Errorf("failed to open PTY: %w", err)
	}

	terminal := NewTerminal(width, height)

	return &PTY{
		Master:    pty,
		Slave:     tty,
		Terminal:  terminal,
		width:     width,
		height:    height,
		done:      make(chan struct{}),
	}, nil
}

func (p *PTY) Read(buf []byte) (int, error) {
	return p.Master.Read(buf)
}

func (p *PTY) Write(buf []byte) (int, error) {
	return p.Master.Write(buf)
}

func (p *PTY) Resize(width, height int) error {
	p.width = width
	p.height = height
	return setSize(int(p.Master.Fd()), width, height)
}

func (p *PTY) Close() error {
	close(p.done)
	p.Master.Close()
	p.Slave.Close()
	return nil
}

func (p *PTY) HandleResize(signals chan os.Signal) {
	oldWidth, oldHeight := p.width, p.height

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-p.done:
			return
		case <-ticker.C:
			width, height, err := getSize(int(p.Master.Fd()))
			if err != nil {
				continue
			}
			if width != oldWidth || height != oldHeight {
				oldWidth, oldHeight = width, height
				p.Terminal.Resize(width, height)
				setSize(int(p.Master.Fd()), width, height)
			}
		case sig := <-signals:
			if isResizeSignal(sig) {
				width, height, _ := getSize(int(p.Master.Fd()))
				p.Terminal.Resize(width, height)
				setSize(int(p.Master.Fd()), width, height)
			}
		}
	}
}

func HandleTerminal(resizeWidth, resizeHeight int, input io.Reader, output io.Writer) error {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to make raw terminal: %w", err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	width, height, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		width, height = 80, 24
	}

	if resizeWidth > 0 {
		width = resizeWidth
	}
	if resizeHeight > 0 {
		height = resizeHeight
	}

	_ = NewTerminal(width, height)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, getResizeSignal()...)

	pty, err := NewPTY(width, height)
	if err != nil {
		return err
	}
	defer pty.Close()

	go pty.HandleResize(signals)

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := pty.Read(buf)
			if err != nil {
				return
			}
			output.Write(buf[:n])
		}
	}()

	io.Copy(pty.Master, input)

	return nil
}

func RestoreTerminal(state interface{}) error {
	if state == nil {
		return nil
	}
	return nil
}

func openPTY() (*os.File, *os.File, error) {
	pty, err := os.OpenFile("/dev/ptmx", os.O_RDWR, 0)
	if err != nil {
		return nil, nil, err
	}
	return pty, pty, nil
}

func setSize(fd, width, height int) error {
	return nil
}

func getSize(fd int) (int, int, error) {
	return 80, 24, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// getResizeSignal returns the OS signal(s) used to notify terminal resize events.
// SIGWINCH is only available on Unix-like systems; Windows has no equivalent.
func getResizeSignal() []os.Signal {
	return resizeSignals
}

// isResizeSignal reports whether the given signal indicates a terminal resize event.
func isResizeSignal(sig os.Signal) bool {
	for _, s := range resizeSignals {
		if sig == s {
			return true
		}
	}
	return false
}
