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
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	statusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"})

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000"))

	lightSquare = lipgloss.NewStyle().
			Background(lipgloss.Color("#EBECD0")).
			Foreground(lipgloss.Color("#000000")).
			Padding(0, 1)

	darkSquare = lipgloss.NewStyle().
			Background(lipgloss.Color("#779556")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1)

	// Unicode piece symbols
	pieceSymbols = map[chess.Piece]string{
		chess.WhiteKing:   "♔",
		chess.WhiteQueen:  "♕",
		chess.WhiteRook:   "♖",
		chess.WhiteBishop: "♗",
		chess.WhiteKnight: "♘",
		chess.WhitePawn:   "♙",
		chess.BlackKing:   "♚",
		chess.BlackQueen:  "♛",
		chess.BlackRook:   "♜",
		chess.BlackBishop: "♝",
		chess.BlackKnight: "♞",
		chess.BlackPawn:   "♟",
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
	title := titleStyle.Render("Chess Terminal")
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

	// File labels
	files := "   a  b  c  d  e  f  g  h"
	sb.WriteString(lipgloss.PlaceHorizontal(width, lipgloss.Center, files))
	sb.WriteString("\n")

	for rank := 7; rank >= 0; rank-- {
		var rankLine strings.Builder
		rankLine.WriteString(fmt.Sprintf("%d ", rank+1))

		for file := range 8 {
			sq := chess.Square(file + rank*8)
			piece := board.Piece(sq)

			var style lipgloss.Style
			if (file+rank)%2 == 0 {
				style = darkSquare
			} else {
				style = lightSquare
			}

			if piece == chess.NoPiece {
				rankLine.WriteString(style.Render(" "))
			} else {
				symbol, exists := pieceSymbols[piece]
				if !exists {
					symbol = "?"
				}
				rankLine.WriteString(style.Render(symbol))
			}
		}

		rankLine.WriteString(fmt.Sprintf(" %d", rank+1))
		sb.WriteString(lipgloss.PlaceHorizontal(width, lipgloss.Center, rankLine.String()))
		sb.WriteString("\n")
	}

	// File labels
	sb.WriteString(lipgloss.PlaceHorizontal(width, lipgloss.Center, files))
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
