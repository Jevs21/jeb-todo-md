package tui

import (
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	maxNavStackDepth   = 50
	textInputCharLimit = 500
	textInputWidth     = 80
)

var headerIcons = []string{
	"◆", "◇", "●", "○", "■", "□", "▲", "△",
	"★", "☆", "✦", "※", "›", "»", "→", "•", "‣", "⌘",
	"⌬", "⌭", "⏚", "⎈", "⌖", "⌑", "⏏", "⏍", "☊",
	"⚀", "⚁", "⚂", "⚃", "⚄", "⚅",
	"☽", "☿", "♃", "♄", "♅", "⚶", "⚷",
}

// Mode represents the current TUI mode.
type Mode int

const (
	// ModeNormal is the default browsing mode for navigating and triggering actions.
	ModeNormal Mode = iota
	// ModeEditing is active when inline-editing an existing todo's text.
	ModeEditing
	// ModeCreating is active when entering text for a new todo item.
	ModeCreating
	// ModeRearrange is active when reordering todos with j/k swaps.
	ModeRearrange
)

// navigationEntry stores position information for back-navigation.
type navigationEntry struct {
	FilePath       string
	CursorPosition int
}

type model struct {
	file          *TodoFile
	cursor        int
	mode          Mode
	textInput     textinput.Model
	pendingDelete bool
	headerIcon    string
	navStack      []navigationEntry
	statusMessage string
}

// switchFileMsg is returned by loadFileCmd after attempting to parse a file.
type switchFileMsg struct {
	newFile       *TodoFile
	restoreCursor int // -1 = start at 0 (forward nav), >= 0 = restore (back nav)
	err           error
}

// loadFileCmd returns a tea.Cmd that parses a file and sends a switchFileMsg.
func loadFileCmd(path string, restoreCursor int) tea.Cmd {
	return func() tea.Msg {
		todoFile, err := ParseFile(path)
		return switchFileMsg{newFile: todoFile, restoreCursor: restoreCursor, err: err}
	}
}

// fileBasenameWithoutExtension returns the filename without its directory or extension.
func fileBasenameWithoutExtension(filePath string) string {
	baseName := filepath.Base(filePath)
	return strings.TrimSuffix(baseName, filepath.Ext(baseName))
}

func initialModel(todoFile *TodoFile, navigationStack []navigationEntry) model {
	textInput := textinput.New()
	textInput.CharLimit = textInputCharLimit
	textInput.Width = textInputWidth

	return model{
		file:       todoFile,
		cursor:     0,
		mode:       ModeNormal,
		textInput:  textInput,
		headerIcon: headerIcons[rand.IntN(len(headerIcons))],
		navStack:   navigationStack,
	}
}

// Run parses the todo file at filePath and starts the TUI.
// returnStack provides file paths for back-navigation (each starts at cursor 0).
func Run(filePath string, returnStack []string) error {
	todoFile, err := ParseFile(filePath)
	if err != nil {
		return fmt.Errorf("loading file: %w", err)
	}

	var navigationStack []navigationEntry
	for _, returnPath := range returnStack {
		navigationStack = append(navigationStack, navigationEntry{
			FilePath:       returnPath,
			CursorPosition: 0,
		})
	}

	m := initialModel(todoFile, navigationStack)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case switchFileMsg:
		return m.handleSwitchFile(msg)
	case tea.KeyMsg:
		// Clear status message on any keypress
		if m.statusMessage != "" {
			m.statusMessage = ""
		}

		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		switch m.mode {
		case ModeNormal:
			return m.updateNormal(msg)
		case ModeEditing:
			return m.updateEditing(msg)
		case ModeCreating:
			return m.updateCreating(msg)
		case ModeRearrange:
			return m.updateRearrange(msg)
		}
	}
	return m, nil
}

// handleSwitchFile processes the result of a file load command.
func (m model) handleSwitchFile(msg switchFileMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.statusMessage = fmt.Sprintf("Error: %v", msg.err)
		// If this was a forward navigation, pop the entry we pushed
		if msg.restoreCursor == -1 && len(m.navStack) > 0 {
			m.navStack = m.navStack[:len(m.navStack)-1]
		}
		return m, nil
	}

	m.file = msg.newFile
	if msg.restoreCursor >= 0 {
		m.cursor = msg.restoreCursor
	} else {
		m.cursor = 0
	}

	// Clamp cursor to valid range
	if m.file.TodoCount() == 0 {
		m.cursor = 0
	} else if m.cursor >= m.file.TodoCount() {
		m.cursor = m.file.TodoCount() - 1
	}

	m.mode = ModeNormal
	m.pendingDelete = false
	m.statusMessage = ""
	return m, nil
}

// startTextInput sets up the text input with a value and focuses it.
func (m model) startTextInput(value string) (model, tea.Cmd) {
	m.textInput.SetValue(value)
	if value != "" {
		m.textInput.CursorEnd()
	}
	cmd := m.textInput.Focus()
	return m, cmd
}

func (m model) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.pendingDelete {
		if msg.String() == "d" {
			m.file.DeleteTodo(m.cursor)
			if err := m.file.Save(); err != nil {
				m.statusMessage = fmt.Sprintf("Error saving: %v", err)
			}
			if m.cursor >= m.file.TodoCount() && m.cursor > 0 {
				m.cursor--
			}
			m.pendingDelete = false
			return m, nil
		}
		m.pendingDelete = false
	}

	switch msg.String() {
	case "q", "esc":
		if len(m.navStack) > 0 {
			entry := m.navStack[len(m.navStack)-1]
			m.navStack = m.navStack[:len(m.navStack)-1]
			return m, loadFileCmd(entry.FilePath, entry.CursorPosition)
		}
		return m, tea.Quit
	case "j", "down":
		if m.cursor < m.file.TodoCount()-1 {
			m.cursor++
		}
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case " ", "enter":
		if m.file.TodoCount() > 0 {
			item := m.file.GetTodo(m.cursor)
			linkedPath := item.LinkedPath()
			if linkedPath != "" {
				return m.navigateToLinkedFile(linkedPath)
			}
			m.file.ToggleTodo(m.cursor)
			if err := m.file.Save(); err != nil {
				m.statusMessage = fmt.Sprintf("Error saving: %v", err)
			}
		}
	case "x":
		if m.file.TodoCount() > 0 {
			m.file.ToggleTodo(m.cursor)
			if err := m.file.Save(); err != nil {
				m.statusMessage = fmt.Sprintf("Error saving: %v", err)
			}
		}
	case "e":
		if m.file.TodoCount() > 0 {
			m.mode = ModeEditing
			item := m.file.GetTodo(m.cursor)
			return m.startTextInput(item.Text)
		}
	case "c":
		m.mode = ModeCreating
		return m.startTextInput("")
	case "r":
		if m.file.TodoCount() > 0 {
			m.mode = ModeRearrange
		}
	case "d":
		if m.file.TodoCount() > 0 {
			m.pendingDelete = true
		}
	}
	return m, nil
}

// navigateToLinkedFile resolves a linked path and initiates navigation.
func (m model) navigateToLinkedFile(linkedPath string) (tea.Model, tea.Cmd) {
	resolvedPath := ResolveLinkedPath(m.file.Path, linkedPath)

	// Check for self-reference
	currentAbsolutePath, err := filepath.Abs(m.file.Path)
	if err == nil {
		resolvedAbsolutePath, err := filepath.Abs(resolvedPath)
		if err == nil && currentAbsolutePath == resolvedAbsolutePath {
			m.statusMessage = "Error: cannot link to current file"
			return m, nil
		}
	}

	// Check file exists
	if _, err := os.Stat(resolvedPath); os.IsNotExist(err) {
		m.statusMessage = fmt.Sprintf("Error: file not found: %s", resolvedPath)
		return m, nil
	}

	// Check max stack depth
	if len(m.navStack) >= maxNavStackDepth {
		m.statusMessage = fmt.Sprintf("Error: maximum navigation depth (%d) reached", maxNavStackDepth)
		return m, nil
	}

	// Push current file + cursor onto the stack
	m.navStack = append(m.navStack, navigationEntry{
		FilePath:       m.file.Path,
		CursorPosition: m.cursor,
	})

	return m, loadFileCmd(resolvedPath, -1)
}

func (m model) updateEditing(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.file.SetTodoText(m.cursor, m.textInput.Value())
		if err := m.file.Save(); err != nil {
			m.statusMessage = fmt.Sprintf("Error saving: %v", err)
		}
		m.textInput.Blur()
		m.mode = ModeNormal
		return m, nil
	case "esc":
		m.textInput.Blur()
		m.mode = ModeNormal
		return m, nil
	default:
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}
}

func (m model) updateCreating(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		text := strings.TrimSpace(m.textInput.Value())
		if text != "" {
			newItem := TodoItem{Text: text, Checked: false}
			m.file.InsertTodo(m.cursor, newItem)
			if err := m.file.Save(); err != nil {
				m.statusMessage = fmt.Sprintf("Error saving: %v", err)
			}
			m.cursor++
		}
		m.textInput.Blur()
		m.mode = ModeNormal
		return m, nil
	case "esc":
		m.textInput.Blur()
		m.mode = ModeNormal
		return m, nil
	default:
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}
}

func (m model) updateRearrange(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if m.cursor < m.file.TodoCount()-1 {
			m.file.SwapTodos(m.cursor, m.cursor+1)
			if err := m.file.Save(); err != nil {
				m.statusMessage = fmt.Sprintf("Error saving: %v", err)
			}
			m.cursor++
		}
	case "k", "up":
		if m.cursor > 0 {
			m.file.SwapTodos(m.cursor, m.cursor-1)
			if err := m.file.Save(); err != nil {
				m.statusMessage = fmt.Sprintf("Error saving: %v", err)
			}
			m.cursor--
		}
	case "r", "esc":
		m.mode = ModeNormal
	}
	return m, nil
}

// renderHeader builds the header string with repeated depth icons, file basename, and date.
func (m model) renderHeader() string {
	navigationDepth := len(m.navStack)
	repeatedIcons := strings.Repeat(m.headerIcon, navigationDepth+1)
	currentFileBasename := fileBasenameWithoutExtension(m.file.Path)
	currentDateFormatted := time.Now().Format("Jan 2, 2006")
	headerText := fmt.Sprintf("%s %s [%s]", repeatedIcons, currentFileBasename, currentDateFormatted)
	return titleStyle.Render(headerText)
}

func (m model) View() string {
	var b strings.Builder

	// Header
	b.WriteString(m.renderHeader())
	b.WriteString("\n")

	// Status message (e.g., link errors)
	if m.statusMessage != "" {
		b.WriteString(errorStyle.Render("  " + m.statusMessage))
		b.WriteString("\n")
	}

	if m.file.TodoCount() == 0 && m.mode != ModeCreating {
		b.WriteString("\n  No todos. Press 'c' to create one.\n")
	}

	// Todo list
	for i := 0; i < m.file.TodoCount(); i++ {
		item := m.file.GetTodo(i)
		isCursor := i == m.cursor

		if isCursor && m.mode == ModeEditing {
			b.WriteString(m.renderInputLine(i+1, m.file.TodoCount()))
		} else {
			b.WriteString(m.renderTodoLine(i, item, isCursor))
		}
		b.WriteString("\n")

		if isCursor && m.mode == ModeCreating {
			b.WriteString(m.renderInputLine(m.cursor+2, m.file.TodoCount()+1))
			b.WriteString("\n")
		}
	}

	if m.file.TodoCount() == 0 && m.mode == ModeCreating {
		b.WriteString(m.renderInputLine(1, 1))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(m.renderHelp())

	return b.String()
}

func (m model) renderTodoLine(idx int, item TodoItem, isCursor bool) string {
	cursor := "   "
	if isCursor {
		cursor = " > "
	}

	numStr := priorityStyle.Render(m.fmtLineNum(idx+1, m.file.TodoCount()))

	if isCursor && m.pendingDelete {
		return deleteStyle.Render(cursor) + numStr + deleteStyle.Render(item.Text)
	}
	if isCursor && m.mode == ModeRearrange {
		return rearrangeStyle.Render(cursor) + numStr + rearrangeStyle.Render(item.Text)
	}
	if isCursor {
		textStyle := cursorStyle
		if item.Checked {
			textStyle = textStyle.Strikethrough(true)
		}
		if item.IsLinkedTodo() {
			textStyle = textStyle.Underline(true)
		}
		return textStyle.Render(cursor) + numStr + textStyle.Render(item.Text)
	}
	if item.IsLinkedTodo() {
		if item.Checked {
			return cursor + numStr + linkStyle.Strikethrough(true).Render(item.Text)
		}
		return cursor + numStr + linkStyle.Render(item.Text)
	}
	if item.Checked {
		return cursor + numStr + checkedStyle.Render(item.Text)
	}
	return cursor + numStr + item.Text
}

// fmtLineNum formats a 1-based line number right-aligned to the width
// needed for totalItems, followed by two spaces.
func (m model) fmtLineNum(oneBasedIdx, totalItems int) string {
	width := len(fmt.Sprintf("%d", totalItems))
	return fmt.Sprintf("%*d  ", width, oneBasedIdx)
}

// renderInputLine renders the text input row with cursor marker and line number.
func (m model) renderInputLine(oneBasedIdx, totalItems int) string {
	num := m.fmtLineNum(oneBasedIdx, totalItems)
	return cursorStyle.Render(" > ") + priorityStyle.Render(num) + m.textInput.View()
}

func (m model) renderHelp() string {
	switch m.mode {
	case ModeNormal:
		if m.pendingDelete {
			return helpStyle.Render("  press d again to delete  |  any other key to cancel")
		}
		quitOrBackLabel := "esc/q: quit"
		if len(m.navStack) > 0 {
			quitOrBackLabel = "esc/q: back"
		}
		return helpStyle.Render("  j/k: navigate  space/enter: toggle/open  x: toggle  e: edit  c: create  r: rearrange  d: delete  " + quitOrBackLabel)
	case ModeEditing:
		return helpStyle.Render("  enter: save  esc: cancel")
	case ModeCreating:
		return helpStyle.Render("  enter: create  esc: cancel")
	case ModeRearrange:
		return helpStyle.Render("  j/k: swap items  r/esc: done rearranging")
	}
	return ""
}
