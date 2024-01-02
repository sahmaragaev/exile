package game

import (
	"errors"
	"exile-telegram-bot/pkg/db"
	"exile-telegram-bot/pkg/models"
	"fmt"
	"log"
	"time"
)

func ProcessGameMessage(telegramUserID, userMessage string) (*models.GameResponse, error) {
	userID, err := db.EnsureUserExists(telegramUserID)
	if err != nil {
		log.Printf("Error ensuring user exists: %v", err)
		return nil, err
	}

	threadID, err := db.GetOrCreateThread(userID)
	if err != nil {
		log.Printf("Error getting or creating thread: %v", err)
		return nil, err
	}

	if err := db.AddMessageToThread(threadID, userMessage); err != nil {
		log.Printf("Error adding message to thread: %v", err)
		return nil, fmt.Errorf("error adding message to thread: %w", err)
	}

	if err := db.RunThread(threadID); err != nil {
		log.Printf("Error running thread: %v", err)
		return nil, fmt.Errorf("error running thread: %w", err)
	}

	immediateResponse, err := db.GetGameResponse(threadID)
	if err == nil && immediateResponse != nil {
		return immediateResponse, nil
	}

	responseChan := make(chan *models.GameResponse)
	errChan := make(chan error)

	go func() {
		time.Sleep(5 * time.Second)
		for i := 0; i < 5; i++ {
			response, err := db.GetGameResponse(threadID)
			if err == nil && response != nil {
				responseChan <- response
				return
			}
			time.Sleep(5 * time.Second)
		}
		errChan <- errors.New("timeout waiting for game response")
	}()

	select {
	case response := <-responseChan:
		return response, nil
	case err := <-errChan:
		log.Printf("Error waiting for game response: %v", err)
		return nil, err
	}
}