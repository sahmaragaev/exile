package models

type GameState struct {
    Health      int      `json:"health"`
    Inventory   []string `json:"inventory"`
    Environment string   `json:"environment"`
}