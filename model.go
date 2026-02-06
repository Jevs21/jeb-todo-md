package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Mode represents the current TUI mode.
type Mode int

const (
	ModeNormal Mode = iota
	ModeEditing
	ModeCreating
	ModeRearrange
)

type model struct {
	file          *TodoFile
	cursor        int
	mode          Mode
	textInput     textinput.Model
	pendingDelete bool
	windowWidth   int
	windowHeight  int
}

func initialModel(tf *TodoFile) model {
	ti := textinput.New()
	ti.CharLimit = 500
	ti.Width = 80

	return model{
		file:      tf,
		cursor:    0,
		mode:      ModeNormal,
		textInput: ti,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height
		return m, nil

	case tea.KeyMsg:
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

func (m model) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.pendingDelete {
		if msg.String() == "d" {
			m.file.DeleteTodo(m.cursor)
			_ = m.file.Save()
			if m.cursor >= m.file.TodoCount() && m.cursor > 0 {
				m.cursor--
			}
			m.pendingDelete = false
			return m, nil
		}
		m.pendingDelete = false
	}

	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "j", "down":
		if m.cursor < m.file.TodoCount()-1 {
			m.cursor++
		}
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case " ", "x":
		if m.file.TodoCount() > 0 {
			m.file.ToggleTodo(m.cursor)
			_ = m.file.Save()
		}
	case "e":
		if m.file.TodoCount() > 0 {
			m.mode = ModeEditing
			item := m.file.GetTodo(m.cursor)
			m.textInput.SetValue(item.Text)
			m.textInput.CursorEnd()
			m.textInput.Focus()
			return m, m.textInput.Focus()
		}
	case "c":
		m.mode = ModeCreating
		m.textInput.SetValue("")
		m.textInput.Focus()
		return m, m.textInput.Focus()
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

func (m model) updateEditing(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.file.SetTodoText(m.cursor, m.textInput.Value())
		_ = m.file.Save()
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
			_ = m.file.Save()
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
			_ = m.file.Save()
			m.cursor++
		}
	case "k", "up":
		if m.cursor > 0 {
			m.file.SwapTodos(m.cursor, m.cursor-1)
			_ = m.file.Save()
			m.cursor--
		}
	case "r", "esc":
		m.mode = ModeNormal
	}
	return m, nil
}

func (m model) View() string {
	var b strings.Builder

	// Title
	title := m.file.Title
	if title == "" {
		title = "Todo"
	}
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n")

	if m.file.TodoCount() == 0 && m.mode != ModeCreating {
		b.WriteString("\n  No todos. Press 'c' to create one.\n")
	}

	// Todo list
	for i := 0; i < m.file.TodoCount(); i++ {
		item := m.file.GetTodo(i)
		isCursor := i == m.cursor

		if isCursor && m.mode == ModeEditing {
			width := len(fmt.Sprintf("%d", m.file.TodoCount()))
			num := fmt.Sprintf("%*d  ", width, i+1)
			b.WriteString(cursorStyle.Render(" > ") + priorityStyle.Render(num) + m.textInput.View())
		} else {
			b.WriteString(m.renderTodoLine(i, item, isCursor))
		}
		b.WriteString("\n")

		// Show create input after cursor item
		if isCursor && m.mode == ModeCreating {
			width := len(fmt.Sprintf("%d", m.file.TodoCount()+1))
			num := fmt.Sprintf("%*d  ", width, m.cursor+2)
			b.WriteString(cursorStyle.Render(" > ") + priorityStyle.Render(num) + m.textInput.View())
			b.WriteString("\n")
		}
	}

	// If creating with no todos, show input
	if m.file.TodoCount() == 0 && m.mode == ModeCreating {
		b.WriteString(cursorStyle.Render(" > ") + priorityStyle.Render("1  ") + m.textInput.View())
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

	width := len(fmt.Sprintf("%d", m.file.TodoCount()))
	num := fmt.Sprintf("%*d  ", width, idx+1)
	numStr := priorityStyle.Render(num)

	if isCursor && m.pendingDelete {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
		return style.Render(cursor) + numStr + style.Render(item.Text)
	}
	if isCursor && m.mode == ModeRearrange {
		return rearrangeStyle.Render(cursor) + numStr + rearrangeStyle.Render(item.Text)
	}
	if isCursor {
		textStyle := cursorStyle
		if item.Checked {
			textStyle = textStyle.Strikethrough(true)
		}
		return textStyle.Render(cursor) + numStr + textStyle.Render(item.Text)
	}
	if item.Checked {
		return cursor + numStr + checkedStyle.Render(item.Text)
	}
	return cursor + numStr + item.Text
}

func (m model) renderHelp() string {
	switch m.mode {
	case ModeNormal:
		if m.pendingDelete {
			return helpStyle.Render("  press d again to delete  |  any other key to cancel")
		}
		return helpStyle.Render("  j/k: navigate  space/x: toggle  e: edit  c: create  r: rearrange  d: delete  q: quit")
	case ModeEditing:
		return helpStyle.Render("  enter: save  esc: cancel")
	case ModeCreating:
		return helpStyle.Render("  enter: create  esc: cancel")
	case ModeRearrange:
		return helpStyle.Render("  j/k: swap items  r/esc: done rearranging")
	}
	return ""
}
