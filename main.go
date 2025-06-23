package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/notnil/chess"
)

var (
	docStyle = lipgloss.NewStyle().Margin(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#BC7342")).
			Padding(0, 1)

	statusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#BC7342", Dark: "#BC7342"})

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000"))

	lightSquare = lipgloss.NewStyle().
			Background(lipgloss.Color("#DEBA90")).
			Width(3).
			Align(lipgloss.Center)

	darkSquare = lipgloss.NewStyle().
			Background(lipgloss.Color("#BC7342")).
			Width(3).
			Align(lipgloss.Center)

	whitePiece = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	blackPiece = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000"))

	// Piece notation (all uppercase)
	pieceNotation = map[chess.Piece]string{
		chess.WhiteKing:   "K",
		chess.WhiteQueen:  "Q",
		chess.WhiteRook:   "R",
		chess.WhiteBishop: "B",
		chess.WhiteKnight: "N",
		chess.WhitePawn:   "P",
		chess.BlackKing:   "K",
		chess.BlackQueen:  "Q",
		chess.BlackRook:   "R",
		chess.BlackBishop: "B",
		chess.BlackKnight: "N",
		chess.BlackPawn:   "P",
	}

	turnWhite = lipgloss.NewStyle().
			Background(lipgloss.Color("#BC7342")).
			Foreground(lipgloss.Color("#FFFFFF"))

	turnBlack = lipgloss.NewStyle().
			Background(lipgloss.Color("#BC7342")).
			Foreground(lipgloss.Color("#000000"))

	// Styles for history viewport
	historyStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#BC7342")).
			Padding(0, 1)

	inputBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#BC7342")).
				Padding(0, 1)
)

type model struct {
	game      *chess.Game
	error     error
	width     int
	height    int
	textInput textinput.Model
	viewport  viewport.Model // Viewport for game history
	history   []string       // Store game moves as strings
}

func initialModel() model {
	ti := textinput.New()
	ti.Prompt = "Enter move: "
	ti.CharLimit = 4
	ti.Focus()

	vp := viewport.New(0, 0) // Will be sized later
	vp.SetContent("Game History:\n")

	return model{
		game:      chess.NewGame(),
		textInput: ti,
		viewport:  vp,
		history:   []string{},
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Define fixed widths for board and history
		const boardRenderedWidth = 28  // 2 (rank) + 8*3 (squares) + 2 (rank)
		const historyDesiredWidth = 20 // Shorter width for history, plus padding/border later
		const spacingWidth = 4         // Space between board and history

		// Calculate available width for the content area (board + spacing + history)
		// contentAreaWidth := boardRenderedWidth + spacingWidth + historyDesiredWidth

		// Calculate available height for the main content (board + history)
		// Subtract space for title, turn/input, error, and docStyle margins
		topSectionHeight := lipgloss.Height(titleStyle.Render("Go Chess"))
		bottomSectionMinHeight := 6 // ~2 for turn, ~2 for input, ~2 for error/gap
		availableHeight := m.height - topSectionHeight - bottomSectionMinHeight - docStyle.GetVerticalFrameSize()

		// Board is square-ish, 8 ranks + 2 for file labels = 10 lines
		// boardDisplayHeight := 10

		// Set viewport dimensions based on calculations
		m.viewport.Width = historyDesiredWidth - historyStyle.GetHorizontalFrameSize()
		m.viewport.Height = availableHeight - historyStyle.GetVerticalFrameSize()

		if m.viewport.Width < 0 {
			m.viewport.Width = 0
		}
		if m.viewport.Height < 0 {
			m.viewport.Height = 0
		}

		// Update viewport content in case of resize
		m.updateHistoryViewport()
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			moveStr := m.textInput.Value()
			err := m.game.MoveStr(moveStr)
			if err != nil {
				m.error = err
			} else {
				m.error = nil
				m.textInput.Reset() // Clear input after successful move
				// Append move with proper numbering (e.g., "e4")
				m.history = append(m.history, fmt.Sprint(moveStr))
				m.updateHistoryViewport()
				// Scroll to bottom of history
				m.viewport.GotoBottom()
			}
			return m, nil
		// Pass key messages to viewport for scrolling
		case tea.KeyUp, tea.KeyDown, tea.KeyPgUp, tea.KeyPgDown:
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *model) updateHistoryViewport() {
	var historyContent strings.Builder
	historyContent.WriteString("Game History:\n\n") // Add an extra newline for spacing
	formattedHistory := make([]string, int(len(m.history)/2+1))

	// Group moves into pairs for display (e.g., "1. e4 e5")
	for i, move := range m.history {
		pos := i / 2
		if i%2 == 0 {
			formattedHistory[pos] = fmt.Sprintf("%d.", pos+1)
		}
		formattedHistory[pos] += " " + move
	}

	for _, line := range formattedHistory {
		historyContent.WriteString(line)
		historyContent.WriteString("\n")
	}
	m.viewport.SetContent(historyContent.String())
}

func (m model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	var sb strings.Builder

	// Title
	title := titleStyle.Render("Go Chess")
	sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, title))
	sb.WriteString("\n\n")

	// Board and History layout
	boardStr := renderBoard(m.game) // renderBoard no longer needs totalWidth

	// Ensure the history view is rendered with its styles
	historyView := historyStyle.Render(m.viewport.View())

	// Combine board and history side-by-side
	content := lipgloss.JoinHorizontal(
		lipgloss.Top,
		boardStr,
		lipgloss.NewStyle().Width(4).Render(""), // Spacer between board and history
		historyView,
	)

	// Center the combined board and history block within the terminal width
	sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, content))
	sb.WriteString("\n\n")

	// Game status and input
	if m.game.Outcome() != chess.NoOutcome {
		status := statusMessageStyle.Render(fmt.Sprintf("Game over! %s\n\nPress 'n' to start a new game or 'esc' to quit", outcomeString(m.game.Outcome())))
		sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, status))
	} else {
		// Current turn
		turnStyle := turnWhite
		turn := "White"
		if m.game.Position().Turn() == chess.Black {
			turnStyle = turnBlack
			turn = "Black"
		}

		turnStatus := turnStyle.Render(fmt.Sprint(turn)) + statusMessageStyle.Render(" to move")
		sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, turnStatus))
		sb.WriteString("\n")

		// Input area with border
		inputContent := lipgloss.JoinHorizontal(
			lipgloss.Left,
			m.textInput.View(),
		)

		// Apply border to the input content
		borderedInput := inputBorderStyle.Render(inputContent)

		// Center the bordered input within the terminal width
		centeredInput := lipgloss.PlaceHorizontal(
			m.width,
			lipgloss.Center,
			borderedInput,
		)
		sb.WriteString("\n" + centeredInput)
		// Error message
		if m.error != nil {
			sb.WriteString("\n\n")
			sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, errorStyle.Render(m.error.Error())))
		}
	}

	return docStyle.Render(sb.String())
}

func outcomeString(outcome chess.Outcome) string {
	switch outcome {
	case chess.WhiteWon:
		return "White wins!"
	case chess.BlackWon:
		return "Black wins!"
	case chess.Draw:
		return "Draw"
	default:
		return "Unknown outcome"
	}
}

// renderBoard now only focuses on rendering the board string, without centering.
// Centering is handled by the View() method.
func renderBoard(game *chess.Game) string {
	board := game.Position().Board()
	var sb strings.Builder

	// File labels - two spaces between each letter for alignment with 3-char wide squares
	filesLine := "   a  b  c  d  e  f  g  h  "
	sb.WriteString(filesLine)
	sb.WriteString("\n")

	for rank := 7; rank >= 0; rank-- {
		sb.WriteString(fmt.Sprintf("%d ", rank+1))

		for file := range 8 {
			sq := chess.Square(file + rank*8)
			piece := board.Piece(sq)

			var squareStyle, pieceStyle lipgloss.Style
			if (file+rank)%2 == 0 {
				squareStyle = darkSquare
			} else {
				squareStyle = lightSquare
			}

			if piece != chess.NoPiece && piece.Color() == chess.White {
				pieceStyle = whitePiece
			} else {
				pieceStyle = blackPiece
			}

			if piece == chess.NoPiece {
				sb.WriteString(squareStyle.Render(" "))
			} else {
				notation := pieceNotation[piece]
				sb.WriteString(squareStyle.Render(pieceStyle.Render(notation)))
			}
		}

		sb.WriteString(fmt.Sprintf(" %d", rank+1))
		sb.WriteString("\n")
	}

	// File labels (same as top)
	sb.WriteString(filesLine)
	return sb.String()
}

func main() {
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(), // add mouse support for good measure
	)
	if _, err := p.Run(); err != nil {
		log.Fatalf("Alas, there's been an error: %v", err)
	}
}
