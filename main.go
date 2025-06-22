package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/notnil/chess"
)

var (
	docStyle = lipgloss.NewStyle().Margin(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#779556")).
			Padding(0, 1)

	statusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#779556", Dark: "#779556"})

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000"))

	lightSquare = lipgloss.NewStyle().
			Background(lipgloss.Color("#EBECD0")).
			Width(3).
			Align(lipgloss.Center)

	darkSquare = lipgloss.NewStyle().
			Background(lipgloss.Color("#779556")).
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
)

type model struct {
	game      *chess.Game
	error     error
	moveInput string
	width     int
	height    int
}

func initialModel() model {
	return model{
		game: chess.NewGame(),
	}
}

func (m model) Init() tea.Cmd {
	return nil
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
			if m.game.Outcome() != chess.NoOutcome {
				return m, nil
			}
			err := m.game.MoveStr(m.moveInput)
			if err != nil {
				m.error = err
			} else {
				m.error = nil
			}
			m.moveInput = ""
			return m, nil
		case tea.KeyBackspace:
			if len(m.moveInput) > 0 {
				m.moveInput = m.moveInput[:len(m.moveInput)-1]
			}
			return m, nil
		case tea.KeyRunes:
			if m.game.Outcome() == chess.NoOutcome {
				m.moveInput += string(msg.Runes)
			}
			return m, nil
		}
	}
	return m, nil
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
		turn := "White"
		if m.game.Position().Turn() == chess.Black {
			turn = "Black"
		}

		turnStatus := statusMessageStyle.Render(fmt.Sprintf("%s to move", turn))
		sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, turnStatus))
		sb.WriteString("\n\n")

		// Move input
		inputPrompt := "Enter move (e.g. e2e4):\n" + m.moveInput + "_"
		sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, inputPrompt))

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
