package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gourish-mokashi/watchdog/daemon/internal/installers"
)

// ──────────────────────────────────────────────────────────────────────────────
// Styles (lipgloss)
// ──────────────────────────────────────────────────────────────────────────────

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FF00")).
			Background(lipgloss.Color("#1a1a2e")).
			Padding(0, 2).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Italic(true).
			MarginBottom(1)

	itemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	selectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(lipgloss.Color("#00FF00")).
				Bold(true)

	checkedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Bold(true)

	uncheckedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#555555"))

	toolNameStyle = lipgloss.NewStyle().
			Bold(true)

	toolDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	logSuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00"))

	logErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	logInfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#5599FF"))

	borderBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#00FF00")).
			Padding(1, 2)

	doneBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("#00FF00")).
			Padding(1, 2).
			MarginTop(1)

	errorBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("#FF0000")).
			Padding(1, 2).
			MarginTop(1)
)

// ──────────────────────────────────────────────────────────────────────────────
// Key bindings (bubbles/key)
// ──────────────────────────────────────────────────────────────────────────────

type keyMap struct {
	Up      key.Binding
	Down    key.Binding
	Toggle  key.Binding
	Install key.Binding
	Quit    key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Toggle, k.Install, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down},
		{k.Toggle, k.Install},
		{k.Quit},
	}
}

var defaultKeys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Toggle: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "toggle"),
	),
	Install: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "install"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c", "q"),
		key.WithHelp("q", "quit"),
	),
}

// ──────────────────────────────────────────────────────────────────────────────
// UI States
// ──────────────────────────────────────────────────────────────────────────────

const (
	stateSelecting = iota
	stateInstalling
	stateDone
)

// ──────────────────────────────────────────────────────────────────────────────
// Messages
// ──────────────────────────────────────────────────────────────────────────────

// installStepMsg is sent after each tool completes a phase (install/configure/start).
type installStepMsg struct {
	toolName string
	phase    string // "install", "configure", "start"
}

// installDoneMsg is sent when the entire installation sequence finishes.
type installDoneMsg struct {
	err error
}

// ──────────────────────────────────────────────────────────────────────────────
// Model
// ──────────────────────────────────────────────────────────────────────────────

type model struct {
	// Tool data
	tools    []installers.SecurityTools
	cursor   int
	selected map[int]struct{}

	// UI state
	state int

	// Bubbles components
	spinner  spinner.Model  // spinner during installation
	progress progress.Model // overall progress bar
	viewport viewport.Model // scrollable log viewer
	help     help.Model     // keybinding help footer
	keys     keyMap

	// Installation tracking
	logs           []string
	totalSteps     int
	completedSteps int
	hadError       bool
}

// ──────────────────────────────────────────────────────────────────────────────
// Constructor
// ──────────────────────────────────────────────────────────────────────────────

// InitialModel configures the starting state of the TUI.
//
// HOW TO ADD NEW TOOLS:
// ─────────────────────
// 1. Create a new file in internal/installers/ (e.g. mytool.go).
// 2. Define a struct that implements the installers.SecurityTools interface:
//      type MyTool struct{}
//      func (m *MyTool) Name() string        { return "MyTool" }
//      func (m *MyTool) Description() string  { return "What it does" }
//      func (m *MyTool) Install() error       { /* apt/dnf install logic */ }
//      func (m *MyTool) Configure() error     { /* write config files */ }
//      func (m *MyTool) Start() error         { /* systemctl enable --now */ }
// 3. Register it in cmd/daemon/main.go → RunInstallerUI() by appending to
//    the `tools` slice:
//      tools := []installers.SecurityTools{
//          &installers.FalcoTool{},
//          &installers.SuricataTool{},
//          &installers.MyTool{},          // ← add your new tool here
//      }
//    That's it — the TUI picks it up automatically.
func InitialModel(availableTools []installers.SecurityTools) model {
	// Spinner
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))

	// Progress bar
	prog := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
	)

	// Viewport for install logs
	vp := viewport.New(60, 10)
	vp.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#444444")).
		Padding(0, 1)

	// Help
	h := help.New()
	h.ShowAll = false

	return model{
		tools:    availableTools,
		selected: make(map[int]struct{}),
		state:    stateSelecting,
		spinner:  sp,
		progress: prog,
		viewport: vp,
		help:     h,
		keys:     defaultKeys,
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Bubble Tea interface
// ──────────────────────────────────────────────────────────────────────────────

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	// ── Installation progress ────────────────────────────────────────────
	case installStepMsg:
		m.completedSteps++
		logLine := fmt.Sprintf("✓ [%s] %s complete", msg.toolName, msg.phase)
		m.logs = append(m.logs, logLine)
		m.viewport.SetContent(strings.Join(m.logs, "\n"))
		m.viewport.GotoBottom()
		return m, nil

	case installDoneMsg:
		if msg.err != nil {
			m.hadError = true
			m.logs = append(m.logs, fmt.Sprintf("✗ ERROR: %v", msg.err))
		} else {
			m.logs = append(m.logs, "✓ All modules deployed and running!")
		}
		m.viewport.SetContent(strings.Join(m.logs, "\n"))
		m.viewport.GotoBottom()
		m.state = stateDone
		return m, nil

	// ── Spinner / progress animation ────────────────────────────────────
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		cmds = append(cmds, cmd)

	// ── Keyboard input ──────────────────────────────────────────────────
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.Up):
			if m.state == stateSelecting && m.cursor > 0 {
				m.cursor--
			}

		case key.Matches(msg, m.keys.Down):
			if m.state == stateSelecting && m.cursor < len(m.tools)-1 {
				m.cursor++
			}

		case key.Matches(msg, m.keys.Toggle):
			if m.state == stateSelecting {
				if _, ok := m.selected[m.cursor]; ok {
					delete(m.selected, m.cursor)
				} else {
					m.selected[m.cursor] = struct{}{}
				}
			}

		case key.Matches(msg, m.keys.Install):
			if m.state == stateSelecting && len(m.selected) > 0 {
				m.state = stateInstalling

				var selectedTools []installers.SecurityTools
				for i := range m.selected {
					selectedTools = append(selectedTools, m.tools[i])
				}

				// 3 phases per tool: install, configure, start
				m.totalSteps = len(selectedTools) * 3
				m.completedSteps = 0

				m.logs = append(m.logs, "▶ Starting installation sequence...")
				m.viewport.SetContent(strings.Join(m.logs, "\n"))

				cmds = append(cmds, startInstallCmd(selectedTools))
			}
		}
	}

	// Keep spinner ticking during installation
	if m.state == stateInstalling {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	switch m.state {
	case stateSelecting:
		return m.viewSelection()
	case stateInstalling:
		return m.viewInstalling()
	case stateDone:
		return m.viewDone()
	default:
		return "Unknown state."
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// View: Selection
// ──────────────────────────────────────────────────────────────────────────────

func (m model) viewSelection() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(" WATCHDOG INIT "))
	b.WriteString("\n")
	b.WriteString(subtitleStyle.Render("Select security modules to deploy"))
	b.WriteString("\n\n")

	for i, tool := range m.tools {
		cursor := "  "
		if m.cursor == i {
			cursor = "▸ "
		}

		checkbox := uncheckedStyle.Render("○")
		if _, ok := m.selected[i]; ok {
			checkbox = checkedStyle.Render("●")
		}

		name := toolNameStyle.Render(tool.Name())
		desc := toolDescStyle.Render(tool.Description())

		row := fmt.Sprintf("%s%s %s  %s", cursor, checkbox, name, desc)

		if m.cursor == i {
			b.WriteString(selectedItemStyle.Render(row) + "\n")
		} else {
			b.WriteString(itemStyle.Render(row) + "\n")
		}
	}

	selectedCount := len(m.selected)
	if selectedCount > 0 {
		b.WriteString("\n")
		b.WriteString(logInfoStyle.Render(fmt.Sprintf("  %d module(s) selected", selectedCount)))
	}

	b.WriteString("\n\n")
	b.WriteString(m.help.View(m.keys))

	return borderBoxStyle.Render(b.String())
}

// ──────────────────────────────────────────────────────────────────────────────
// View: Installing
// ──────────────────────────────────────────────────────────────────────────────

func (m model) viewInstalling() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(" DEPLOYING MODULES "))
	b.WriteString("\n\n")

	// Spinner + status
	pct := 0.0
	if m.totalSteps > 0 {
		pct = float64(m.completedSteps) / float64(m.totalSteps)
	}

	b.WriteString(fmt.Sprintf("  %s Installing... (%d/%d steps)\n\n",
		m.spinner.View(), m.completedSteps, m.totalSteps))

	// Progress bar
	b.WriteString("  " + m.progress.ViewAs(pct) + "\n\n")

	// Log viewport
	b.WriteString(m.viewport.View())

	return borderBoxStyle.Render(b.String())
}

// ──────────────────────────────────────────────────────────────────────────────
// View: Done
// ──────────────────────────────────────────────────────────────────────────────

func (m model) viewDone() string {
	var b strings.Builder

	if m.hadError {
		b.WriteString(logErrorStyle.Render("  INSTALLATION FAILED"))
	} else {
		b.WriteString(logSuccessStyle.Render("  INSTALLATION COMPLETE"))
	}
	b.WriteString("\n\n")

	for _, log := range m.logs {
		if strings.HasPrefix(log, "✗") {
			b.WriteString("  " + logErrorStyle.Render(log) + "\n")
		} else if strings.HasPrefix(log, "✓") {
			b.WriteString("  " + logSuccessStyle.Render(log) + "\n")
		} else {
			b.WriteString("  " + logInfoStyle.Render(log) + "\n")
		}
	}

	b.WriteString("\n  Press q to exit.")

	boxStyle := doneBoxStyle
	if m.hadError {
		boxStyle = errorBoxStyle
	}
	return boxStyle.Render(b.String())
}

// ──────────────────────────────────────────────────────────────────────────────
// Install command (runs in background, reports progress via messages)
// ──────────────────────────────────────────────────────────────────────────────

func startInstallCmd(tools []installers.SecurityTools) tea.Cmd {
	return func() tea.Msg {
		for _, tool := range tools {
			// Phase 1: Install
			if err := tool.Install(); err != nil {
				return installDoneMsg{err: fmt.Errorf("[%s] install failed: %v", tool.Name(), err)}
			}
			// NOTE: In a real async pipeline you'd send installStepMsg per phase
			// via p.Send(), but bubbletea Cmds return a single Msg.

			// Phase 2: Configure
			if err := tool.Configure(); err != nil {
				return installDoneMsg{err: fmt.Errorf("[%s] configure failed: %v", tool.Name(), err)}
			}

			// Phase 3: Start
			if err := tool.Start(); err != nil {
				return installDoneMsg{err: fmt.Errorf("[%s] start failed: %v", tool.Name(), err)}
			}
		}
		return installDoneMsg{err: nil}
	}
}
