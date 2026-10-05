package chess

type Board [64]Piece

// Render renders a board into string for printing. Used for debugging.
func (board Board) Render() string {
	out := ""
	for i := range 64 {
		out += " " + board[i].String()
		if i&7 == 7 {
			out += "\n"
		}
	}
	return out
}
