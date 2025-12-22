package sse

// Defining the bare minimum events and their properties needed for the bot to function.

type GameOpened struct {
	GameId   string `json:"gameId"`
	PlayerId string `json:"playerId"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
}

type PlayerJoined struct {
	GameId      string `json:"gameId"`
	RedPlayerId string `json:"redPlayerId"`
}

type ChatAssigned struct {
	ChatId string `json:"chatId"`
}

type PlayerMoved struct {
	NextPlayerId string `json:"nextPlayerId"`
}

type GameAborted struct {
	GameId string `json:"gameId"`
}

type GameWon struct {
	GameId string `json:"gameId"`
}

type GameDrawn struct {
	GameId string `json:"gameId"`
}

type GameTimedOut struct {
	GameId string `json:"gameId"`
}

type GameResigned struct {
	GameId string `json:"gameId"`
}
