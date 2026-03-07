package ui

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gourish-mokashi/watchdog/daemon/internal/dispatcher"
	"github.com/gourish-mokashi/watchdog/daemon/internal/installers"
)

// ─────────────────────────────────────────────────────────────────────────────
// Color palette — blue-centric theme inspired by OpenClaw
// ─────────────────────────────────────────────────────────────────────────────

const (
	colorPrimary   = lipgloss.Color("#5EAEFF") // soft blue — accents
	colorSecondary = lipgloss.Color("#3A7BD5") // mid-blue — borders, highlights
	colorTertiary  = lipgloss.Color("#1B3A5C") // dark navy — backgrounds
	colorAccent    = lipgloss.Color("#89CFF0") // light sky blue — active items
	colorDim       = lipgloss.Color("#4A5568") // muted grey — inactive text
	colorDimmer    = lipgloss.Color("#2D3748") // darker grey — decorative lines
	colorText      = lipgloss.Color("#CBD5E1") // off-white — body text
	colorSuccess   = lipgloss.Color("#5EEAD4") // teal-green — success
	colorError     = lipgloss.Color("#F87171") // soft red — errors
	colorWarn      = lipgloss.Color("#FBBF24") // amber — warnings / phase labels
	colorWhiteBold = lipgloss.Color("#F1F5F9") // near-white for titles
	colorLogPhase  = lipgloss.Color("#818CF8") // indigo — phase badges
	colorLogToolBg = lipgloss.Color("#1E293B") // dark slate — tool-name bg
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

	logCommandStyle = lipgloss.NewStyle().
			Foreground(colorDim)

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
██╗    ██╗ █████╗ ████████╗ ██████╗██╗  ██╗██████╗  ██████╗  ██████╗ 
██║    ██║██╔══██╗╚══██╔══╝██╔════╝██║  ██║██╔══██╗██╔═══██╗██╔════╝ 
██║ █╗ ██║███████║   ██║   ██║     ███████║██║  ██║██║   ██║██║  ███╗
██║███╗██║██╔══██║   ██║   ██║     ██╔══██║██║  ██║██║   ██║██║   ██║
╚███╔███╔╝██║  ██║   ██║   ╚██████╗██║  ██║██████╔╝╚██████╔╝╚██████╔╝
 ╚══╝╚══╝ ╚═╝  ╚═╝   ╚═╝    ╚═════╝╚═╝  ╚═╝╚═════╝  ╚═════╝  ╚═════╝ 
                                                                     `
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
	stateProjectPath
	stateAskAnalyze
	stateAskWriteRules
	stateGeneratingSummary
	stateAskRestart
	stateRestarting
	stateDone
)

// ─────────────────────────────────────────────────────────────────────────────
// Messages
// ─────────────────────────────────────────────────────────────────────────────

// installLogMsg carries a single formatted log line from the install goroutine.
type installLogMsg struct {
	line    string
	advance int
}

// installDoneMsg is sent when the entire installation sequence finishes.
type installDoneMsg struct {
	err error
}

type summaryDoneMsg struct {
	text string
	err  error
}

type restartDoneMsg struct {
	err error
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
	installEvents  <-chan tea.Msg
	selectedTools  []installers.SecurityTools
	backendURL     string
	projectPath    string
	runAnalysis    bool
	writeRules     bool
	yesSelected    bool
	textInput      textinput.Model
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
func InitialModel(availableTools []installers.SecurityTools, backendURL string) model {
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

	input := textinput.New()
	input.Placeholder = "/path/to/project"
	input.Prompt = "> "
	input.Focus()
	input.CharLimit = 512
	input.Width = 64

	return model{
		tools:       availableTools,
		selected:    make(map[int]struct{}),
		state:       stateSelecting,
		spinner:     sp,
		progress:    prog,
		viewport:    vp,
		help:        h,
		keys:        defaultKeys,
		backendURL:  backendURL,
		yesSelected: true,
		textInput:   input,
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
	case installLogMsg:
		if msg.line != "" {
			m.logs = append(m.logs, msg.line)
			m.viewport.SetContent(strings.Join(m.logs, "\n"))
			m.viewport.GotoBottom()
		}
		if msg.advance > 0 {
			m.completedSteps += msg.advance
			if m.completedSteps > m.totalSteps {
				m.completedSteps = m.totalSteps
			}
		}
		if m.state == stateInstalling && m.installEvents != nil {
			cmds = append(cmds, waitForInstallMsg(m.installEvents))
		}

	case installDoneMsg:
		m.completedSteps = m.totalSteps
		if msg.err != nil {
			m.hadError = true
			m.logs = append(m.logs, fmtLogError(msg.err.Error()))
			m.state = stateDone
		} else {
			m.logs = append(m.logs, fmtLogSuccess("All modules deployed and active"))
			m.state = stateProjectPath
			m.textInput.SetValue("")
			m.textInput.Focus()
		}
		m.viewport.SetContent(strings.Join(m.logs, "\n"))
		m.viewport.GotoBottom()
		m.installEvents = nil
		return m, nil

	case summaryDoneMsg:
		if msg.err != nil {
			m.logs = append(m.logs, fmtLogError(msg.err.Error()))
		} else {
			if msg.text != "" {
				m.logs = append(m.logs, fmtLogInfo("Backend summary generated successfully"))
				m.logs = append(m.logs, fmtLogCommand(truncateLine(msg.text, 140)))
			}
			m.logs = append(m.logs, fmtLogInfo("Falco rules will be written to /etc/falco/rules.d/watchdog-rules.yaml"))
			m.logs = append(m.logs, fmtLogInfo("Falco config will be written to /etc/falco/config.d/watchdog.yaml"))
		}
		m.viewport.SetContent(strings.Join(m.logs, "\n"))
		m.viewport.GotoBottom()
		m.yesSelected = true
		m.state = stateAskRestart

	case restartDoneMsg:
		if msg.err != nil {
			m.logs = append(m.logs, fmtLogError(msg.err.Error()))
		} else {
			m.logs = append(m.logs, fmtLogSuccess("Selected services restarted"))
		}
		m.viewport.SetContent(strings.Join(m.logs, "\n"))
		m.viewport.GotoBottom()
		m.state = stateDone

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

		case m.state == stateProjectPath:
			if msg.Type == tea.KeyEnter {
				path := strings.TrimSpace(m.textInput.Value())
				if path == "" {
					m.logs = append(m.logs, fmtLogError("Project path cannot be empty"))
					m.viewport.SetContent(strings.Join(m.logs, "\n"))
					m.viewport.GotoBottom()
					return m, nil
				}

				m.projectPath = path
				m.logs = append(m.logs, fmtLogInfo("Project path set to "+path))
				m.viewport.SetContent(strings.Join(m.logs, "\n"))
				m.viewport.GotoBottom()
				m.yesSelected = true
				m.state = stateAskAnalyze
				return m, nil
			}

			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd

		case m.state == stateAskAnalyze:
			if toggled, handled := handleYesNoInput(msg, m.yesSelected); handled {
				m.yesSelected = toggled
				return m, nil
			}
			if msg.Type == tea.KeyEnter {
				m.runAnalysis = m.yesSelected
				if m.runAnalysis {
					m.logs = append(m.logs, fmtLogInfo("Project analysis enabled"))
				} else {
					m.logs = append(m.logs, fmtLogInfo("Project analysis skipped"))
				}
				m.viewport.SetContent(strings.Join(m.logs, "\n"))
				m.viewport.GotoBottom()
				m.yesSelected = true
				m.state = stateAskWriteRules
				return m, nil
			}

		case m.state == stateAskWriteRules:
			if toggled, handled := handleYesNoInput(msg, m.yesSelected); handled {
				m.yesSelected = toggled
				return m, nil
			}
			if msg.Type == tea.KeyEnter {
				m.writeRules = m.yesSelected
				if !m.writeRules {
					m.logs = append(m.logs, fmtLogInfo("Rule generation skipped"))
					m.viewport.SetContent(strings.Join(m.logs, "\n"))
					m.viewport.GotoBottom()
					m.yesSelected = true
					m.state = stateAskRestart
					return m, nil
				}

				m.logs = append(m.logs, fmtLogInfo("Requesting project summary from backend..."))
				m.viewport.SetContent(strings.Join(m.logs, "\n"))
				m.viewport.GotoBottom()
				m.state = stateGeneratingSummary
				return m, generateSummaryCmd(m.projectPath, m.backendURL)
			}

		case m.state == stateAskRestart:
			if toggled, handled := handleYesNoInput(msg, m.yesSelected); handled {
				m.yesSelected = toggled
				return m, nil
			}
			if msg.Type == tea.KeyEnter {
				if !m.yesSelected {
					m.logs = append(m.logs, fmtLogInfo("Service restart skipped"))
					m.viewport.SetContent(strings.Join(m.logs, "\n"))
					m.viewport.GotoBottom()
					m.state = stateDone
					return m, nil
				}

				m.logs = append(m.logs, fmtLogInfo("Restarting installed services..."))
				m.viewport.SetContent(strings.Join(m.logs, "\n"))
				m.viewport.GotoBottom()
				m.state = stateRestarting
				return m, restartServicesCmd(m.selectedTools)
			}

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

				m.selectedTools = selectedTools
				m.totalSteps = len(selectedTools) * 3
				m.completedSteps = 0

				m.logs = []string{
					fmtLogInfo("Initializing deployment pipeline..."),
					fmtLogInfo("Streaming installer output..."),
				}
				m.viewport.SetContent(strings.Join(m.logs, "\n"))
				m.viewport.GotoBottom()

				eventCh := make(chan tea.Msg, 256)
				m.installEvents = eventCh
				go startInstall(selectedTools, eventCh)
				cmds = append(cmds, waitForInstallMsg(eventCh))
			}
		}
	}

	// Keep spinner ticking during long-running states
	if m.state == stateInstalling || m.state == stateGeneratingSummary || m.state == stateRestarting {
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
	case stateProjectPath:
		return m.viewProjectPath()
	case stateAskAnalyze:
		return m.viewYesNoPrompt("PROJECT ANALYSIS", "Do you want to run an agent to analyze the project so it can write better rules?")
	case stateAskWriteRules:
		return m.viewYesNoPrompt("RULE GENERATION", "Do you want the agent to write rules for the installed security tools?")
	case stateGeneratingSummary:
		return m.viewWorking("Generating project summary from backend...")
	case stateAskRestart:
		return m.viewYesNoPrompt("RESTART SERVICES", "Do you want to restart the installed services?")
	case stateRestarting:
		return m.viewWorking("Restarting installed services...")
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

	b.WriteString(fmt.Sprintf("  %s  %s\n\n",
		m.spinner.View(),
		progressLabelStyle.Render("Deploying modules..."),
	))

	pct := 0.0
	if m.totalSteps > 0 {
		pct = float64(m.completedSteps) / float64(m.totalSteps)
	}
	pctInt := int(pct * 100)

	b.WriteString(fmt.Sprintf("  %s\n\n",
		lipgloss.NewStyle().Foreground(colorPrimary).Bold(true).Render(fmt.Sprintf("%d%% complete", pctInt)),
	))

	b.WriteString("  " + m.progress.ViewAs(pct) + "\n\n")

	b.WriteString(m.viewport.View())
	b.WriteString("\n")

	return outerBoxStyle.Render(b.String())
}

func (m model) viewProjectPath() string {
	var b strings.Builder

	b.WriteString(banner())
	b.WriteString("\n")
	b.WriteString(titleBarStyle.Render("  PROJECT PATH  "))
	b.WriteString("\n\n")
	b.WriteString("  Enter your project path.\n\n")
	b.WriteString("  " + m.textInput.View() + "\n\n")
	b.WriteString(m.viewport.View())
	b.WriteString("\n")

	return outerBoxStyle.Render(b.String())
}

func (m model) viewYesNoPrompt(title, question string) string {
	var b strings.Builder

	b.WriteString(banner())
	b.WriteString("\n")
	b.WriteString(titleBarStyle.Render("  " + title + "  "))
	b.WriteString("\n\n")
	b.WriteString("  " + question + "\n\n")
	b.WriteString("  " + renderYesNoChoice(m.yesSelected) + "\n\n")
	b.WriteString(m.viewport.View())
	b.WriteString("\n")

	return outerBoxStyle.Render(b.String())
}

func (m model) viewWorking(message string) string {
	var b strings.Builder

	b.WriteString(banner())
	b.WriteString("\n")
	b.WriteString(titleBarStyle.Render("  POST INSTALL  "))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("  %s  %s\n\n", m.spinner.View(), message))
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

func fmtLogCommand(msg string) string {
	return fmt.Sprintf("%s      %s", timestamp(), logCommandStyle.Render(msg))
}

func handleYesNoInput(msg tea.KeyMsg, current bool) (bool, bool) {
	switch msg.String() {
	case "left", "h", "y", "Y":
		return true, true
	case "right", "l", "n", "N":
		return false, true
	default:
		return current, false
	}
}

func renderYesNoChoice(yesSelected bool) string {
	selected := lipgloss.NewStyle().Foreground(colorTertiary).Background(colorSuccess).Bold(true).Padding(0, 1)
	unselected := lipgloss.NewStyle().Foreground(colorDim)

	if yesSelected {
		return selected.Render("YES") + "   " + unselected.Render("NO")
	}

	return unselected.Render("YES") + "   " + selected.Render("NO")
}

// ─────────────────────────────────────────────────────────────────────────────
// Install command (runs in background, streams logs through Bubble Tea messages)
// ─────────────────────────────────────────────────────────────────────────────

func startInstall(tools []installers.SecurityTools, ch chan<- tea.Msg) {
	defer close(ch)

	emitter := &installLogWriter{ch: ch}
	installers.SetCommandOutput(emitter)
	defer installers.SetCommandOutput(nil)

	for _, tool := range tools {
		name := tool.Name()

		sendInstallLog(ch, fmtLogPhase(name, "install", "downloading packages..."), 0)
		if err := tool.Install(); err != nil {
			emitter.Flush()
			ch <- installDoneMsg{err: fmt.Errorf("[%s] install failed: %v", name, err)}
			return
		}
		emitter.Flush()
		sendInstallLog(ch, fmtLogPhase(name, "install", "packages installed"), 1)

		sendInstallLog(ch, fmtLogPhase(name, "config", "writing configuration..."), 0)
		if err := tool.Configure(); err != nil {
			emitter.Flush()
			ch <- installDoneMsg{err: fmt.Errorf("[%s] configure failed: %v", name, err)}
			return
		}
		emitter.Flush()
		sendInstallLog(ch, fmtLogPhase(name, "config", "configuration applied"), 1)

		sendInstallLog(ch, fmtLogPhase(name, "start", "enabling systemd service..."), 0)
		if err := tool.Start(); err != nil {
			emitter.Flush()
			ch <- installDoneMsg{err: fmt.Errorf("[%s] start failed: %v", name, err)}
			return
		}
		emitter.Flush()
		sendInstallLog(ch, fmtLogPhase(name, "start", "service active ✓"), 1)
		sendInstallLog(ch, fmtLogSuccess(fmt.Sprintf("%s - fully deployed", name)), 0)
		sendInstallLog(ch, fmtLogCommand("------------------------------------------------------------"), 0)
		sendInstallLog(ch, "", 0)
	}

	ch <- installDoneMsg{err: nil}
}

func generateSummaryCmd(projectPath, backendURL string) tea.Cmd {
	return func() tea.Msg {
		if strings.TrimSpace(backendURL) == "" {
			return summaryDoneMsg{err: fmt.Errorf("WATCHDOG_BACKEND_URL is not configured")}
		}

		text, err := dispatcher.GenerateSummary(projectPath, backendURL)
		return summaryDoneMsg{text: text, err: err}
	}
}

func restartServicesCmd(tools []installers.SecurityTools) tea.Cmd {
	return func() tea.Msg {
		for _, tool := range tools {
			if err := tool.Start(); err != nil {
				return restartDoneMsg{err: fmt.Errorf("[%s] restart failed: %v", tool.Name(), err)}
			}
		}
		return restartDoneMsg{err: nil}
	}
}

func waitForInstallMsg(ch <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return msg
	}
}

func sendInstallLog(ch chan<- tea.Msg, line string, advance int) {
	ch <- installLogMsg{line: line, advance: advance}
}

type installLogWriter struct {
	ch      chan<- tea.Msg
	mu      sync.Mutex
	pending strings.Builder
}

func (w *installLogWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.pending.WriteString(strings.ReplaceAll(string(p), "\r", "\n"))
	w.flushLocked(false)
	return len(p), nil
}

func (w *installLogWriter) Flush() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.flushLocked(true)
}

func (w *installLogWriter) flushLocked(force bool) {
	for {
		current := w.pending.String()
		idx := strings.IndexByte(current, '\n')
		if idx == -1 {
			if force && strings.TrimSpace(current) != "" {
				sendInstallLog(w.ch, fmtLogCommand(truncateLine(strings.TrimSpace(current), 140)), 0)
				w.pending.Reset()
			}
			return
		}

		line := strings.TrimSpace(current[:idx])
		rest := current[idx+1:]
		w.pending.Reset()
		w.pending.WriteString(rest)

		if line == "" {
			continue
		}

		sendInstallLog(w.ch, fmtLogCommand(truncateLine(line, 140)), 0)
	}
}

func truncateLine(line string, limit int) string {
	if len(line) <= limit {
		return line
	}
	return line[:limit-3] + "..."
}
