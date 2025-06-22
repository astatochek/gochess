package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
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
)

type model struct {
	game      *chess.Game
	error     error
	width     int
	height    int
	textInput textinput.Model
}

func initialModel() model {
	ti := textinput.New()
	ti.Prompt = "Enter move: "
	ti.CharLimit = 4
	ti.Focus()
	return model{
		game:      chess.NewGame(),
		textInput: ti,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			err := m.game.MoveStr(m.textInput.Value())
			if err != nil {
				m.error = err
			} else {
				m.error = nil
				m.textInput.Reset() // Clear input after successful move
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
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

	// Board
	board := renderBoard(m.game, m.width)
	sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, board))
	sb.WriteString("\n\n")

	// Game status
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

		inputWidth := 16 // Fixed width for input area
		inputContainer := lipgloss.NewStyle().
			Width(inputWidth).
			Align(lipgloss.Left)

		// Build the input line
		inputLine := lipgloss.JoinHorizontal(
			lipgloss.Left,
			inputContainer.Render(m.textInput.View()),
		)

		// Center the entire line
		centeredInput := lipgloss.PlaceHorizontal(
			m.width,
			lipgloss.Center,
			inputLine,
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

func renderBoard(game *chess.Game, width int) string {
	board := game.Position().Board()
	var sb strings.Builder

	// The complete board line (including rank numbers) is exactly 26 characters:
	// 2 (left rank) + 24 (8 squares × 3 chars) + 2 (right rank) = 28 chars
	boardLineWidth := 28

	// Center the entire board block
	boardIndent := max((width-boardLineWidth)/2, 0)
	indentStr := strings.Repeat(" ", boardIndent)

	// File labels - perfectly aligned under squares
	files := strings.Join([]string{"", "a", "b", "c", "d", "e", "f", "g", "h", ""}, "  ")
	centeredFiles := lipgloss.PlaceHorizontal(width, lipgloss.Center, files)
	sb.WriteString(centeredFiles)
	sb.WriteString("\n")

	for rank := 7; rank >= 0; rank-- {
		sb.WriteString(indentStr)
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
	sb.WriteString(centeredFiles)
	return sb.String()
}

func main() {
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(), // add mouse support for good measure
	)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
	}
}
