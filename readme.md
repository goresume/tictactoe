# Multiplayer Tic Tac Toe

A simple multiplayer Tic Tac Toe game built with Go (backend) and HTML/JavaScript (frontend). Players can create games and share game IDs with others to play together in real-time.

## Quick Start

1. Clone the repository

2. Start the server:
```bash
go run main.go
```
The server will start on `http://localhost:8080`

3. Open the client:
- Either open `index.html` directly in your browser
- Or serve it using a simple HTTP server:
```bash
python -m http.server 8000
```
Then visit `http://localhost:8000`

## API Endpoints

| Endpoint | Method | Description |
|----------|---------|-------------|
| `/create` | POST | Create a new game |
| `/join` | POST | Join an existing game |
| `/move` | POST | Make a move in the game |
| `/state` | POST | Get current game state |

## Development

The project uses:
- Pure Go for the backend (no external dependencies)
- Vanilla JavaScript for the frontend (no frameworks)
- Simple CSS Grid for the game board layout

