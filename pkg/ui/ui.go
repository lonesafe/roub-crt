package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type ColorScheme struct {
	Background      lipgloss.Color
	Foreground      lipgloss.Color
	Cursor          lipgloss.Color
	Selection       lipgloss.Color
	Black           lipgloss.Color
	Red             lipgloss.Color
	Green           lipgloss.Color
	Yellow          lipgloss.Color
	Blue            lipgloss.Color
	Magenta         lipgloss.Color
	Cyan            lipgloss.Color
	White           lipgloss.Color
	BrightBlack     lipgloss.Color
	BrightRed       lipgloss.Color
	BrightGreen     lipgloss.Color
	BrightYellow    lipgloss.Color
	BrightBlue      lipgloss.Color
	BrightMagenta   lipgloss.Color
	BrightCyan      lipgloss.Color
	BrightWhite     lipgloss.Color
}

func GetDefaultColorScheme() ColorScheme {
	return ColorScheme{
		Background:      lipgloss.Color("#1E1E1E"),
		Foreground:      lipgloss.Color("#D4D4D4"),
		Cursor:          lipgloss.Color("#FFFFFF"),
		Selection:       lipgloss.Color("#264F78"),
		Black:           lipgloss.Color("#000000"),
		Red:             lipgloss.Color("#CD3131"),
		Green:           lipgloss.Color("#0DBC79"),
		Yellow:          lipgloss.Color("#E5E510"),
		Blue:            lipgloss.Color("#2472C8"),
		Magenta:         lipgloss.Color("#BC3FBC"),
		Cyan:            lipgloss.Color("#11A8CD"),
		White:           lipgloss.Color("#E5E5E5"),
		BrightBlack:     lipgloss.Color("#666666"),
		BrightRed:       lipgloss.Color("#F14C4C"),
		BrightGreen:     lipgloss.Color("#23D18B"),
		BrightYellow:    lipgloss.Color("#F5F543"),
		BrightBlue:      lipgloss.Color("#3B8EEA"),
		BrightMagenta:   lipgloss.Color("#D670D6"),
		BrightCyan:      lipgloss.Color("#29B8DB"),
		BrightWhite:     lipgloss.Color("#FFFFFF"),
	}
}

type TerminalUI struct {
	ColorScheme ColorScheme
	width      int
	height     int
}

func NewTerminalUI(width, height int) *TerminalUI {
	return &TerminalUI{
		ColorScheme: GetDefaultColorScheme(),
		width:       width,
		height:      height,
	}
}

func (ui *TerminalUI) SetColorScheme(scheme ColorScheme) {
	ui.ColorScheme = scheme
}

func (ui *TerminalUI) Resize(width, height int) {
	ui.width = width
	ui.height = height
}

func (ui *TerminalUI) GetWidth() int {
	return ui.width
}

func (ui *TerminalUI) GetHeight() int {
	return ui.height
}

type BoxStyle struct {
	BorderColor lipgloss.Color
	BGColor     lipgloss.Color
	FGColor     lipgloss.Color
}

func DefaultBoxStyle() BoxStyle {
	scheme := GetDefaultColorScheme()
	return BoxStyle{
		BorderColor: scheme.Blue,
		BGColor:     scheme.Background,
		FGColor:     scheme.Foreground,
	}
}

func (ui *TerminalUI) DrawBox(x, y, w, h int, title string, style BoxStyle) string {
	if w < 2 || h < 2 {
		return ""
	}

	top := "┌" + strings.Repeat("─", w-2) + "┐"
	bottom := "└" + strings.Repeat("─", w-2) + "┘"

	if title != "" {
		titleStr := fmt.Sprintf(" %s ", title)
		if len(titleStr) > w-2 {
			titleStr = titleStr[:w-2]
		}
		top = "┌" + titleStr + strings.Repeat(" ", w-2-len(titleStr)) + "┐"
	}

	result := top + "\n"

	for i := 0; i < h-2; i++ {
		result += "│" + strings.Repeat(" ", w-2) + "│\n"
	}

	result += bottom

	return lipgloss.NewStyle().
		Foreground(style.FGColor).
		Background(style.BGColor).
		Render(result)
}

func (ui *TerminalUI) DrawLine(y int, text string, style lipgloss.Style) string {
	return style.Render(text)
}

type ProgressBar struct {
	Width       int
	Height      int
	FillChar    string
	EmptyChar   string
	FillColor   lipgloss.Color
	BorderColor lipgloss.Color
}

func NewProgressBar(width int) *ProgressBar {
	return &ProgressBar{
		Width:       width,
		Height:      1,
		FillChar:    "█",
		EmptyChar:   "░",
		FillColor:   lipgloss.Color("#23D18B"),
		BorderColor: lipgloss.Color("#3B8EEA"),
	}
}

func (pb *ProgressBar) Render(progress float64) string {
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}

	filledWidth := int(float64(pb.Width-2) * progress)
	emptyWidth := pb.Width - 2 - filledWidth

	filled := strings.Repeat(pb.FillChar, filledWidth)
	empty := strings.Repeat(pb.EmptyChar, emptyWidth)

	bar := fmt.Sprintf("╡%s%s╞", filled, empty)

	return lipgloss.NewStyle().
		Foreground(pb.FillColor).
		Render(bar)
}

type TableColumn struct {
	Title    string
	Width    int
	Color    lipgloss.Color
}

type Table struct {
	Columns []TableColumn
	Rows    [][]string
	Style   BoxStyle
}

func NewTable(columns []TableColumn) *Table {
	return &Table{
		Columns: columns,
		Rows:    make([][]string, 0),
		Style:   DefaultBoxStyle(),
	}
}

func (t *Table) AddRow(row []string) {
	if len(row) == len(t.Columns) {
		t.Rows = append(t.Rows, row)
	}
}

func (t *Table) Render() string {
	var result strings.Builder

	header := ""
	for i, col := range t.Columns {
		title := truncate(col.Title, col.Width)
		headerStyle := lipgloss.NewStyle().
			Foreground(col.Color).
			Width(col.Width)
		header += headerStyle.Render(title)
		if i < len(t.Columns)-1 {
			header += " │ "
		}
	}
	result.WriteString(header + "\n")

	separator := ""
	for i, col := range t.Columns {
		separator += strings.Repeat("─", col.Width)
		if i < len(t.Columns)-1 {
			separator += "─┼─"
		}
	}
	result.WriteString(separator + "\n")

	for _, row := range t.Rows {
		line := ""
		for i, cell := range row {
			if i >= len(t.Columns) {
				break
			}
			col := t.Columns[i]
			cellStr := truncate(cell, col.Width)
			cellStyle := lipgloss.NewStyle().
				Foreground(t.Style.FGColor).
				Width(col.Width)
			line += cellStyle.Render(cellStr)
			if i < len(t.Columns)-1 {
				line += " │ "
			}
		}
		result.WriteString(line + "\n")
	}

	return lipgloss.NewStyle().
		Background(t.Style.BGColor).
		Render(result.String())
}

func truncate(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "…"
}

type ListItem struct {
	Label    string
	Value    string
	Selected bool
	Color    lipgloss.Color
}

type List struct {
	Items     []ListItem
	Style     BoxStyle
	ItemStyle lipgloss.Color
}

func NewList(items []ListItem) *List {
	scheme := GetDefaultColorScheme()
	_ = scheme
	return &List{
		Items: items,
		Style: DefaultBoxStyle(),
	}
}

func (l *List) Render() string {
	var result strings.Builder

	for i, item := range l.Items {
		prefix := "  "
		if item.Selected {
			prefix = "▶ "
		}

		color := l.Style.FGColor
		if item.Color != "" {
			color = item.Color
		}

		line := lipgloss.NewStyle().
			Foreground(color).
			Render(fmt.Sprintf("%s%s", prefix, item.Label))

		if item.Value != "" {
			line += lipgloss.NewStyle().
				Foreground(l.Style.BGColor).
				Render(" → " + item.Value)
		}

		result.WriteString(line)

		if i < len(l.Items)-1 {
			result.WriteString("\n")
		}
	}

	return lipgloss.NewStyle().
		Background(l.Style.BGColor).
		Render(result.String())
}

func (ui *TerminalUI) DrawFileList(files []FileListItem, localPath, remotePath string, showHidden bool, selectedIndex int, isLocal bool) string {
	scheme := ui.ColorScheme

	headerStyle := lipgloss.NewStyle().
		Foreground(scheme.Blue).
		Background(scheme.Background)

	leftHeader := headerStyle.Render(fmt.Sprintf(" LOCAL (%s) ", localPath))
	rightHeader := headerStyle.Render(fmt.Sprintf(" REMOTE (%s) ", remotePath))

	lines := []string{
		fmt.Sprintf("┌%s┬%s┐",
			strings.Repeat("─", 40),
			strings.Repeat("─", 40)),
		fmt.Sprintf("│%s│%s│", leftHeader, rightHeader),
		fmt.Sprintf("├%s┼%s┤",
			strings.Repeat("─", 40),
			strings.Repeat("─", 40)),
	}

	maxLines := ui.height - 10
	if len(files) > maxLines {
		files = files[:maxLines]
	}

	for i, file := range files {
		prefix := "  "
		if i == selectedIndex {
			prefix = "▶ "
		}

		icon := "📄"
		if file.IsDir {
			icon = "📁"
		}
		if file.IsLink {
			icon = "🔗"
		}

		name := file.Name
		if !showHidden && strings.HasPrefix(name, ".") {
			name = ""
		}

		line := fmt.Sprintf("│ %s%s %-36s │ %-38s │",
			prefix, icon, truncate(name, 34), "")

		lineStyle := lipgloss.NewStyle().
			Foreground(scheme.Foreground).
			Background(scheme.Background)

		if i == selectedIndex {
			lineStyle = lipgloss.NewStyle().
				Foreground(scheme.Background).
				Background(scheme.Foreground)
		}

		lines = append(lines, lineStyle.Render(line))
	}

	lines = append(lines, fmt.Sprintf("├%s┴%s┤",
		strings.Repeat("─", 40),
		strings.Repeat("─", 40)))

	helpStyle := lipgloss.NewStyle().
		Foreground(scheme.BrightBlack).
		Background(scheme.Background)

	helpLine := "[F5] Upload │ [F6] Download │ [F10] Quit │ [Tab] Switch │ [↑↓] Navigate"
	lines = append(lines, helpStyle.Render(fmt.Sprintf("│ %-82s │", helpLine)))

	lines = append(lines, fmt.Sprintf("└%s┴%s┘",
		strings.Repeat("─", 40),
		strings.Repeat("─", 40)))

	return lipgloss.NewStyle().
		Background(scheme.Background).
		Render(strings.Join(lines, "\n"))
}

type FileListItem struct {
	Name    string
	Size    int64
	Mode    string
	ModTime string
	IsDir   bool
	IsLink  bool
}

func ClearScreen() {
	fmt.Print("\033[2J")
}

func MoveCursor(x, y int) {
	fmt.Printf("\033[%d;%dH", y, x)
}

func HideCursor() {
	fmt.Print("\033[?25l")
}

func ShowCursor() {
	fmt.Print("\033[?25h")
}

func ResetColors() {
	fmt.Print("\033[0m")
}

func ClearLine() {
	fmt.Print("\033[2K")
}

func GetTerminalSize() (int, int) {
	width, height, err := terminalSize()
	if err != nil {
		return 80, 24
	}
	return width, height
}

func terminalSize() (int, int, error) {
	fd := int(os.Stdin.Fd())
	return terminalSizeFd(fd)
}

func terminalSizeFd(fd int) (int, int, error) {
	return 80, 24, nil
}

type StatusBar struct {
	Text      string
	Color     lipgloss.Color
	BGColor   lipgloss.Color
}

func (ui *TerminalUI) RenderStatusBar(text string) string {
	scheme := ui.ColorScheme

	bar := lipgloss.NewStyle().
		Foreground(scheme.Foreground).
		Background(scheme.Black).
		Padding(0, 1).
		Width(ui.width).
		Render(text)

	return bar
}

type MenuItem struct {
	Label       string
	Description string
	Key         string
	Disabled    bool
}

type Menu struct {
	Title   string
	Items   []MenuItem
	Style   BoxStyle
}

func NewMenu(title string, items []MenuItem) *Menu {
	return &Menu{
		Title: title,
		Items: items,
		Style: DefaultBoxStyle(),
	}
}

func (m *Menu) Render(selected int) string {
	var result strings.Builder

	titleStyle := lipgloss.NewStyle().
		Foreground(m.Style.BorderColor).
		Background(m.Style.BGColor).
		Bold(true)

	result.WriteString(titleStyle.Render(fmt.Sprintf(" %s ", m.Title)) + "\n")

	for i, item := range m.Items {
		prefix := "  "
		if i == selected {
			prefix = " ▶ "
		}

		labelStyle := lipgloss.NewStyle().
			Foreground(m.Style.FGColor).
			Background(m.Style.BGColor)

		if item.Disabled {
			labelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#666666")).
				Background(m.Style.BGColor)
		}

		if i == selected {
			labelStyle = lipgloss.NewStyle().
				Foreground(m.Style.BGColor).
				Background(m.Style.FGColor)
		}

		line := labelStyle.Render(fmt.Sprintf("%s%s", prefix, item.Label))

		if item.Key != "" {
			keyStyle := lipgloss.NewStyle().
				Foreground(m.Style.BorderColor).
				Background(m.Style.BGColor)
			if i == selected {
				keyStyle = lipgloss.NewStyle().
					Foreground(m.Style.FGColor).
					Background(m.Style.BGColor)
			}
			line += keyStyle.Render(fmt.Sprintf(" [%s]", item.Key))
		}

		result.WriteString(line + "\n")

		if item.Description != "" {
			descStyle := lipgloss.NewStyle().
				Foreground(m.Style.BGColor).
				Background(m.Style.BGColor)
			result.WriteString(descStyle.Render(fmt.Sprintf("    %s", item.Description)) + "\n")
		}
	}

	return lipgloss.NewStyle().
		Background(m.Style.BGColor).
		Render(result.String())
}

func (ui *TerminalUI) DrawConnectionStatus(host string, port int, protocol string, connected bool) string {
	scheme := ui.ColorScheme

	status := "● Disconnected"
	statusColor := scheme.Red
	if connected {
		status = "● Connected"
		statusColor = scheme.Green
	}

	_ = statusColor

	hostStyle := lipgloss.NewStyle().
		Foreground(scheme.Foreground).
		Background(scheme.Background)

	hostText := fmt.Sprintf(" %s:%d (%s) ", host, port, protocol)
	hostText += hostStyle.Render(status)

	return hostText
}

func RenderTransferProgress(filename string, progress float64, bytesTransferred, totalBytes int64) string {
	scheme := GetDefaultColorScheme()

	pb := NewProgressBar(50)
	pb.FillColor = scheme.Green

	bar := pb.Render(progress)

	percent := int(progress * 100)

	sizeStr := formatSize(bytesTransferred) + " / " + formatSize(totalBytes)

	info := lipgloss.NewStyle().
		Foreground(scheme.Foreground).
		Render(fmt.Sprintf("%s %s %d%%", truncate(filename, 20), sizeStr, percent))

	return fmt.Sprintf("%s\n%s", info, bar)
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
