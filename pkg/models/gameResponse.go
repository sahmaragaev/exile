package models

type GameResponse struct {
    Text         string    `json:"text"`
    Choices      []string  `json:"choices"`
    NumChoices   int       `json:"num_choices"`
    FreeResponse string    `json:"free_response"`
    GameState    GameState `json:"game_state"`
}