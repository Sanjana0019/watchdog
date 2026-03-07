package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gourish-mokashi/watchdog/daemon/internal/installers"
)

// ─────────────────────────────────────────────────────────────────────────────
// Color palette — blue-centric theme inspired by OpenClaw
// ─────────────────────────────────────────────────────────────────────────────

const (
	colorPrimary    = lipgloss.Color("#5EAEFF") // soft blue — accents
	colorSecondary  = lipgloss.Color("#3A7BD5") // mid-blue — borders, highlights
	colorTertiary   = lipgloss.Color("#1B3A5C") // dark navy — backgrounds
	colorAccent     = lipgloss.Color("#89CFF0") // light sky blue — active items
	colorDim        = lipgloss.Color("#4A5568") // muted grey — inactive text
	colorDimmer     = lipgloss.Color("#2D3748") // darker grey — decorative lines
	colorText       = lipgloss.Color("#CBD5E1") // off-white — body text
	colorSuccess    = lipgloss.Color("#5EEAD4") // teal-green — success
	colorError      = lipgloss.Color("#F87171") // soft red — errors
	colorWarn       = lipgloss.Color("#FBBF24") // amber — warnings / phase labels
	colorWhiteBold  = lipgloss.Color("#F1F5F9") // near-white for titles
	colorLogPhase   = lipgloss.Color("#818CF8") // indigo — phase badges
	colorLogToolBg  = lipgloss.Color("#1E293B") // dark slate — tool-name bg
)

// ─────────────────────────────────────────────────────────────────────────────
// Styles (lipgloss)
// ─────────────────────────────────────────────────────────────────────────────

var (
	// ── Branding ────────────────────────────────────────────────────
	logoStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	titleBarStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorWhiteBold).
			Background(colorTertiary).
			Padding(0, 2)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(colorDim).
			Italic(true)

	// ── Selection list ──────────────────────────────────────────────
	itemStyle = lipgloss.NewStyle().
			PaddingLeft(4).
			Foreground(colorText)

	selectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(4).
				Foreground(colorAccent).
				Bold(true)

	cursorStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	checkedStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	uncheckedStyle = lipgloss.NewStyle().
			Foreground(colorDimmer)

	toolNameStyle = lipgloss.NewStyle().
			Foreground(colorWhiteBold).
			Bold(true)

	toolDescStyle = lipgloss.NewStyle().
			Foreground(colorDim)

	selectedCountStyle = lipgloss.NewStyle().
				Foreground(colorAccent).
				PaddingLeft(4)

	dividerStyle = lipgloss.NewStyle().
			Foreground(colorDimmer)

	// ── Logs ────────────────────────────────────────────────────────
	logTimestampStyle = lipgloss.NewStyle().
				Foreground(colorDim)

	logPhaseBadgeStyle = lipgloss.NewStyle().
				Foreground(colorTertiary).
				Background(colorLogPhase).
				Bold(true).
				Padding(0, 1)

	logToolStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Background(colorLogToolBg).
			Bold(true).
			Padding(0, 1)

	logMsgStyle = lipgloss.NewStyle().
			Foreground(colorText)

	logSuccessIcon = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true)

	logErrorIcon = lipgloss.NewStyle().
			Foreground(colorError).
			Bold(true)

	logInfoIcon = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	// ── Progress ────────────────────────────────────────────────────
	progressLabelStyle = lipgloss.NewStyle().
				Foreground(colorText)

	// ── Containers ──────────────────────────────────────────────────
	outerBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorSecondary).
			Padding(1, 3).
			MarginTop(1)

	doneBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorSuccess).
			Padding(1, 3).
			MarginTop(1)

	errorBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorError).
			Padding(1, 3).
			MarginTop(1)

	doneHeaderSuccess = lipgloss.NewStyle().
				Foreground(colorSuccess).
				Bold(true)

	doneHeaderError = lipgloss.NewStyle().
			Foreground(colorError).
			Bold(true)

	hintStyle = lipgloss.NewStyle().
			Foreground(colorDim).
			Italic(true)
)

// ─────────────────────────────────────────────────────────────────────────────
// ASCII logo
// ─────────────────────────────────────────────────────────────────────────────

func banner() string {
	art := `
 █ █ █▀█ ▀█▀ █▀▀ █ █ █▀▄ █▀█ █▀▀
 █▄█ █▀█  █  █   █▀█ █ █ █ █ █ █
 ▀ ▀ ▀ ▀  ▀  ▀▀▀ ▀ ▀ ▀▀  ▀▀▀ ▀▀▀`
	return logoStyle.Render(art)
}

// ─────────────────────────────────────────────────────────────────────────────
// Key bindings (bubbles/key)
// ─────────────────────────────────────────────────────────────────────────────

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
		key.WithHelp("enter", "deploy"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c", "q"),
		key.WithHelp("q", "quit"),
	),
}

// ─────────────────────────────────────────────────────────────────────────────
// UI States
// ─────────────────────────────────────────────────────────────────────────────

const (
	stateSelecting = iota
	stateInstalling
	stateDone
)

// ─────────────────────────────────────────────────────────────────────────────
// Messages
// ─────────────────────────────────────────────────────────────────────────────

// installLogMsg carries a single formatted log line from the install goroutine.
type installLogMsg struct {
	line string
}

// installDoneMsg is sent when the entire installation sequence finishes.
type installDoneMsg struct {
	err  error
	logs []string // all collected log lines
}

// ─────────────────────────────────────────────────────────────────────────────
// Model
// ─────────────────────────────────────────────────────────────────────────────

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

// ─────────────────────────────────────────────────────────────────────────────
// Constructor
// ─────────────────────────────────────────────────────────────────────────────

// InitialModel configures the starting state of the TUI.
//
// HOW TO ADD NEW TOOLS:
// ─────────────────────
//  1. Create a new file in internal/installers/ (e.g. mytool.go).
//  2. Define a struct that implements the installers.SecurityTools interface:
//     type MyTool struct{}
//     func (m *MyTool) Name() string        { return "MyTool" }
//     func (m *MyTool) Description() string  { return "What it does" }
//     func (m *MyTool) Install() error       { /* apt/dnf install logic */ }
//     func (m *MyTool) Configure() error     { /* write config files */ }
//     func (m *MyTool) Start() error         { /* systemctl enable --now */ }
//  3. Register it in cmd/daemon/main.go → RunInstallerUI() by appending to
//     the `tools` slice:
//     tools := []installers.SecurityTools{
//     &installers.FalcoTool{},
//     &installers.SuricataTool{},
//     &installers.MyTool{},          // ← add your new tool here
//     }
//     That's it — the TUI picks it up automatically.
func InitialModel(availableTools []installers.SecurityTools) model {
	// Spinner — use the MiniDot style for a cleaner look
	sp := spinner.New()
	sp.Spinner = spinner.MiniDot
	sp.Style = lipgloss.NewStyle().Foreground(colorPrimary)

	// Progress bar — blue gradient
	prog := progress.New(
		progress.WithScaledGradient("#3A7BD5", "#89CFF0"),
		progress.WithWidth(50),
		progress.WithoutPercentage(),
	)

	// Viewport for install logs — taller, wider
	vp := viewport.New(64, 14)
	vp.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorDimmer).
		Padding(0, 1)

	// Help — style it to match our palette
	h := help.New()
	h.ShowAll = false
	h.Styles.ShortKey = lipgloss.NewStyle().Foreground(colorPrimary).Bold(true)
	h.Styles.ShortDesc = lipgloss.NewStyle().Foreground(colorDim)
	h.Styles.ShortSeparator = lipgloss.NewStyle().Foreground(colorDimmer)

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

// ─────────────────────────────────────────────────────────────────────────────
// Bubble Tea interface
// ─────────────────────────────────────────────────────────────────────────────

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	// ── Installation progress ────────────────────────────────────────
	case installDoneMsg:
		m.logs = msg.logs
		if msg.err != nil {
			m.hadError = true
			m.logs = append(m.logs, fmtLogError(msg.err.Error()))
		} else {
			m.logs = append(m.logs, fmtLogSuccess("All modules deployed and active"))
		}
		m.viewport.SetContent(strings.Join(m.logs, "\n"))
		m.viewport.GotoBottom()
		m.state = stateDone
		return m, nil

	// ── Spinner / progress animation ────────────────────────────────
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		cmds = append(cmds, cmd)

	// ── Keyboard input ──────────────────────────────────────────────
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

				m.totalSteps = len(selectedTools) * 3
				m.completedSteps = 0

				m.logs = []string{fmtLogInfo("Initializing deployment pipeline...")}
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

// ─────────────────────────────────────────────────────────────────────────────
// View: Selection
// ─────────────────────────────────────────────────────────────────────────────

func (m model) viewSelection() string {
	var b strings.Builder

	// Banner
	b.WriteString(banner())
	b.WriteString("\n")
	b.WriteString(titleBarStyle.Render("  INIT  "))
	b.WriteString("  ")
	b.WriteString(subtitleStyle.Render("select security modules to deploy"))
	b.WriteString("\n\n")

	// Divider
	b.WriteString(dividerStyle.Render("  ─────────────────────────────────────────────") + "\n\n")

	// Tool list
	for i, tool := range m.tools {
		cursor := "   "
		if m.cursor == i {
			cursor = cursorStyle.Render(" ▸ ")
		}

		checkbox := uncheckedStyle.Render("○")
		if _, ok := m.selected[i]; ok {
			checkbox = checkedStyle.Render("◉")
		}

		name := toolNameStyle.Render(tool.Name())
		desc := toolDescStyle.Render("— " + tool.Description())

		row := fmt.Sprintf("%s %s  %s  %s", cursor, checkbox, name, desc)

		if m.cursor == i {
			b.WriteString(selectedItemStyle.Render(row) + "\n")
		} else {
			b.WriteString(itemStyle.Render(row) + "\n")
		}
	}

	// Selection count
	b.WriteString("\n")
	b.WriteString(dividerStyle.Render("  ─────────────────────────────────────────────") + "\n")
	selectedCount := len(m.selected)
	if selectedCount > 0 {
		b.WriteString(selectedCountStyle.Render(
			fmt.Sprintf("  ▪ %d module(s) selected — press enter to deploy", selectedCount)))
	} else {
		b.WriteString(selectedCountStyle.Render("  ▪ no modules selected"))
	}
	b.WriteString("\n\n")

	// Help footer
	b.WriteString("  " + m.help.View(m.keys))

	return outerBoxStyle.Render(b.String())
}

// ─────────────────────────────────────────────────────────────────────────────
// View: Installing
// ─────────────────────────────────────────────────────────────────────────────

func (m model) viewInstalling() string {
	var b strings.Builder

	b.WriteString(banner())
	b.WriteString("\n")
	b.WriteString(titleBarStyle.Render("  DEPLOY  "))
	b.WriteString("\n\n")

	// Spinner + status line
	pct := 0.0
	if m.totalSteps > 0 {
		pct = float64(m.completedSteps) / float64(m.totalSteps)
	}
	pctInt := int(pct * 100)

	b.WriteString(fmt.Sprintf("  %s  %s  %s\n\n",
		m.spinner.View(),
		progressLabelStyle.Render("Deploying modules..."),
		lipgloss.NewStyle().Foreground(colorPrimary).Bold(true).Render(fmt.Sprintf("%d%%", pctInt)),
	))

	// Progress bar
	b.WriteString("  " + m.progress.ViewAs(pct) + "\n\n")

	// Log viewport
	b.WriteString(m.viewport.View())
	b.WriteString("\n")

	return outerBoxStyle.Render(b.String())
}

// ─────────────────────────────────────────────────────────────────────────────
// View: Done
// ─────────────────────────────────────────────────────────────────────────────

func (m model) viewDone() string {
	var b strings.Builder

	b.WriteString(banner())
	b.WriteString("\n\n")

	if m.hadError {
		b.WriteString(doneHeaderError.Render("  ✗  DEPLOYMENT FAILED"))
	} else {
		b.WriteString(doneHeaderSuccess.Render("  ✓  DEPLOYMENT COMPLETE"))
	}
	b.WriteString("\n")
	b.WriteString(dividerStyle.Render("  ─────────────────────────────────────────────") + "\n\n")

	// Render all collected logs
	for _, line := range m.logs {
		b.WriteString("  " + line + "\n")
	}

	b.WriteString("\n")
	b.WriteString(dividerStyle.Render("  ─────────────────────────────────────────────") + "\n")
	b.WriteString(hintStyle.Render("  press q to exit"))

	boxStyle := doneBoxStyle
	if m.hadError {
		boxStyle = errorBoxStyle
	}
	return boxStyle.Render(b.String())
}

// ─────────────────────────────────────────────────────────────────────────────
// Log formatting helpers
// ─────────────────────────────────────────────────────────────────────────────

func timestamp() string {
	return logTimestampStyle.Render(time.Now().Format("15:04:05"))
}

func fmtLogPhase(tool, phase, detail string) string {
	ts := timestamp()
	badge := logPhaseBadgeStyle.Render(strings.ToUpper(phase))
	name := logToolStyle.Render(tool)
	msg := logMsgStyle.Render(detail)
	return fmt.Sprintf("%s  %s  %s  %s", ts, badge, name, msg)
}

func fmtLogSuccess(msg string) string {
	icon := logSuccessIcon.Render("✓")
	return fmt.Sprintf("%s  %s  %s", timestamp(), icon, logSuccessIcon.Render(msg))
}

func fmtLogError(msg string) string {
	icon := logErrorIcon.Render("✗")
	return fmt.Sprintf("%s  %s  %s", timestamp(), icon, logErrorIcon.Render(msg))
}

func fmtLogInfo(msg string) string {
	icon := logInfoIcon.Render("›")
	return fmt.Sprintf("%s  %s  %s", timestamp(), icon, logMsgStyle.Render(msg))
}

// ─────────────────────────────────────────────────────────────────────────────
// Install command (runs in background, collects formatted logs)
// ─────────────────────────────────────────────────────────────────────────────

func startInstallCmd(tools []installers.SecurityTools) tea.Cmd {
	return func() tea.Msg {
		var logs []string

		for _, tool := range tools {
			name := tool.Name()

			// Phase 1: Install
			logs = append(logs, fmtLogPhase(name, "install", "downloading packages..."))
			if err := tool.Install(); err != nil {
				return installDoneMsg{
					err:  fmt.Errorf("[%s] install failed: %v", name, err),
					logs: logs,
				}
			}
			logs = append(logs, fmtLogPhase(name, "install", "packages installed"))

			// Phase 2: Configure
			logs = append(logs, fmtLogPhase(name, "config", "writing configuration..."))
			if err := tool.Configure(); err != nil {
				return installDoneMsg{
					err:  fmt.Errorf("[%s] configure failed: %v", name, err),
					logs: logs,
				}
			}
			logs = append(logs, fmtLogPhase(name, "config", "configuration applied"))

			// Phase 3: Start
			logs = append(logs, fmtLogPhase(name, "start", "enabling systemd service..."))
			if err := tool.Start(); err != nil {
				return installDoneMsg{
					err:  fmt.Errorf("[%s] start failed: %v", name, err),
					logs: logs,
				}
			}
			logs = append(logs, fmtLogPhase(name, "start", "service active ✓"))

			// Summary line per tool
			logs = append(logs, fmtLogSuccess(fmt.Sprintf("%s — fully deployed", name)))
			logs = append(logs, "") // blank line separator between tools
		}

		return installDoneMsg{err: nil, logs: logs}
	}
}
