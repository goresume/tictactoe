package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

type gameState struct {
	Board       [9]string         `json:"board"`
	Players     map[string]string `json:"players"` // playerID -> symbol
	CurrentTurn string            `json:"currentTurn"`
	Winner      string            `json:"winner"`
	mu          sync.Mutex
}

type server struct {
	games map[string]*gameState
	mu    sync.Mutex
}

func newServer() *server {
	return &server{
		games: make(map[string]*gameState),
	}
}

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func (s *server) createGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		PlayerID string `json:"playerID"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	gameID := "game-" + generateID()
	s.games[gameID] = &gameState{
		Players:     map[string]string{req.PlayerID: "X"},
		CurrentTurn: req.PlayerID,
	}
	s.mu.Unlock()

	json.NewEncoder(w).Encode(map[string]string{
		"gameID": gameID,
		"symbol": "X",
	})
}

func (s *server) joinGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		GameID   string `json:"gameID"`
		PlayerID string `json:"playerID"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	game, exists := s.games[req.GameID]
	if !exists {
		s.mu.Unlock()
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}
	s.mu.Unlock()

	game.mu.Lock()
	if len(game.Players) >= 2 {
		game.mu.Unlock()
		http.Error(w, "Game is full", http.StatusBadRequest)
		return
	}

	game.Players[req.PlayerID] = "O"
	game.mu.Unlock()

	json.NewEncoder(w).Encode(map[string]string{
		"symbol": "O",
	})
}

func (s *server) makeMove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		GameID   string `json:"gameID"`
		PlayerID string `json:"playerID"`
		Position int    `json:"position"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	game, exists := s.games[req.GameID]
	s.mu.Unlock()

	if !exists {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	game.mu.Lock()
	defer game.mu.Unlock()

	if game.CurrentTurn != req.PlayerID {
		http.Error(w, "Not your turn", http.StatusBadRequest)
		return
	}

	if req.Position < 0 || req.Position > 8 || game.Board[req.Position] != "" {
		http.Error(w, "Invalid move", http.StatusBadRequest)
		return
	}

	symbol := game.Players[req.PlayerID]
	game.Board[req.Position] = symbol

	// Switch turns
	for pid := range game.Players {
		if pid != req.PlayerID {
			game.CurrentTurn = pid
			break
		}
	}

	if winner := checkWinner(game.Board); winner != "" {
		game.Winner = winner
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"board":  game.Board,
		"winner": game.Winner,
	})
}

func (s *server) getGameState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		GameID string `json:"gameID"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	game, exists := s.games[req.GameID]
	s.mu.Unlock()

	if !exists {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	game.mu.Lock()
	defer game.mu.Unlock()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"board":       game.Board,
		"currentTurn": game.CurrentTurn,
		"winner":      game.Winner,
	})
}

func generateID() string {
	return "123" // Replace with proper ID generation
}

func checkWinner(board [9]string) string {
	lines := [][]int{
		{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, // Rows
		{0, 3, 6}, {1, 4, 7}, {2, 5, 8}, // Columns
		{0, 4, 8}, {2, 4, 6}, // Diagonals
	}

	for _, line := range lines {
		if board[line[0]] != "" &&
			board[line[0]] == board[line[1]] &&
			board[line[0]] == board[line[2]] {
			return board[line[0]]
		}
	}
	return ""
}

func main() {
	s := newServer()

	// Register routes with CORS middleware
	http.HandleFunc("/create", enableCORS(s.createGame))
	http.HandleFunc("/join", enableCORS(s.joinGame))
	http.HandleFunc("/move", enableCORS(s.makeMove))
	http.HandleFunc("/state", enableCORS(s.getGameState))

	log.Printf("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
