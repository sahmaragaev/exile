package models

type OpenAIRequest struct {
    Inputs []string `json:"inputs"`
}

type OpenAIResponse struct {
    Choices []struct {
        Message string `json:"message"`
    } `json:"choices"`
}