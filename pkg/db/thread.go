package db

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"exile-telegram-bot/pkg/config"
	"exile-telegram-bot/pkg/models"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"log"
	"net/http"
)

func GetOrCreateThread(userID primitive.ObjectID) (string, error) {
	collection := ThreadCollection
	filter := bson.M{"userId": userID.Hex()}
	var thread models.Thread

	err := collection.FindOne(context.Background(), filter).Decode(&thread)
	if errors.Is(err, mongo.ErrNoDocuments) {
		log.Println("No existing thread found for user, creating a new thread")
		threadID, err := createThread(userID.Hex())
		if err != nil {
			log.Printf("Error creating new thread: %v", err)
			return "", err
		}

		log.Printf("Created new thread with ID: %s", threadID)
		return threadID, nil
	} else if err != nil {
		log.Printf("Error finding thread for user in MongoDB: %v", err)
		return "", err
	}

	log.Printf("Found existing thread with ID: %s", thread.ThreadID)
	return thread.ThreadID, nil
}

func createThread(userId string) (string, error) {
	requestBody, _ := json.Marshal(map[string]interface{}{})

	request, err := http.NewRequest("POST", "https://api.openai.com/v1/threads", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Printf("Error creating request to OpenAI API: %v", err)
		return "", fmt.Errorf("error creating request to OpenAI API: %w", err)
	}

	setRequestHeaders(request)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Printf("Error sending request to OpenAI API: %v", err)
		return "", fmt.Errorf("error sending request to OpenAI API: %w", err)
	}
	defer response.Body.Close()

	var threadResponse struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(response.Body).Decode(&threadResponse); err != nil {
		log.Printf("Error decoding response from OpenAI API: %v", err)
		return "", fmt.Errorf("error decoding response from OpenAI API: %w", err)
	}

	threadCollection := ThreadCollection
	newThread := models.Thread{
		ID:          primitive.NewObjectID(),
		UserID:      userId,
		ThreadID:    threadResponse.ID,
		AssistantID: config.AppConfig.AssistantId,
	}

	_, err = threadCollection.InsertOne(context.Background(), newThread)
	if err != nil {
		log.Printf("Error inserting new thread into MongoDB: %v", err)
		return "", fmt.Errorf("error inserting new thread into MongoDB: %w", err)
	}

	return threadResponse.ID, nil
}

func AddMessageToThread(threadID, message string) error {
	requestBody, _ := json.Marshal(map[string]interface{}{
		"role":    "user",
		"content": message,
	})

	url := fmt.Sprintf("https://api.openai.com/v1/threads/%s/messages", threadID)
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		log.Printf("Error creating request to OpenAI API: %v", err)
		return fmt.Errorf("error creating request to OpenAI API: %w", err)
	}

	setRequestHeaders(request)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Printf("Error sending request to OpenAI API: %v", err)
		return fmt.Errorf("error sending request to OpenAI API: %w", err)
	}
	defer response.Body.Close()

	responseBody, _ := io.ReadAll(response.Body)
	log.Printf("Response from adding message: %s", string(responseBody))

	return nil
}

func RunThread(threadID string) error {
	log.Printf("Running thread with ID: %s", threadID)

	requestBody, err := json.Marshal(map[string]interface{}{
		"assistant_id": config.AppConfig.AssistantId,
	})
	if err != nil {
		log.Printf("Error marshaling request body: %v", err)
		return err
	}

	request, err := http.NewRequest("POST", fmt.Sprintf("https://api.openai.com/v1/threads/%s/runs", threadID), bytes.NewBuffer(requestBody))
	if err != nil {
		log.Printf("Error creating request to run thread: %v", err)
		return err
	}

	setRequestHeaders(request)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Printf("Error sending request to run thread: %v", err)
		return err
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		log.Printf("Error reading response body from run thread: %v", err)
		return err
	}

	log.Printf("Response from running thread: %s", string(responseBody))
	return nil
}

func GetGameResponse(threadID string) (*models.GameResponse, error) {
	url := fmt.Sprintf("https://api.openai.com/v1/threads/%s/messages", threadID)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Error creating request to get game response: %v", err)
		return nil, err
	}

	setRequestHeaders(request)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Printf("Error sending request to get game response: %v", err)
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Printf("Error reading response body from get game response: %v", err)
		return nil, err
	}

	log.Printf("Raw response body: %s", string(body))

	var messagesResponse models.MessagesResponse

	if err := json.Unmarshal(body, &messagesResponse); err != nil {
		log.Printf("Error unmarshaling game response: %v", err)
		return nil, fmt.Errorf("error unmarshaling game response: %w", err)
	}

	var latestAssistantMessage *models.ThreadMessage
	for _, message := range messagesResponse.Data {
		if message.Role == "assistant" {
			latestAssistantMessage = &message
			break
		}
	}

	if latestAssistantMessage == nil {
		log.Println("No assistant message found in the thread yet.")
		return nil, errors.New("assistant response not available yet")
	}

	if len(latestAssistantMessage.Content) == 0 || latestAssistantMessage.Content[0].Text.Value == "" {
		log.Println("Assistant message content is empty.")
		return nil, errors.New("assistant message content is empty")
	}

	var gameResponse models.GameResponse
	if err := json.Unmarshal([]byte(latestAssistantMessage.Content[0].Text.Value), &gameResponse); err != nil {
		log.Printf("Error unmarshaling game response content: %v", err)
		return nil, fmt.Errorf("error unmarshaling game response content: %w", err)
	}

	return &gameResponse, nil
}

func RestartGame(telegramUserId string) error {
	userId, err := EnsureUserExists(telegramUserId)
	if err != nil {
		log.Printf("Error ensuring user exists: %v", err)
		return fmt.Errorf("error ensuring user exists: %w", err)
	}

	threadCollection := ThreadCollection
	_, err = threadCollection.DeleteMany(context.Background(), bson.M{"userId": userId.Hex()})
	if err != nil {
		log.Printf("Error deleting user's thread: %v", err)
		return fmt.Errorf("error deleting user's thread: %w", err)
	}
	log.Println("Successfully restarted the game")
	return nil
}

func setRequestHeaders(request *http.Request) {
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.AppConfig.OpenAIKey))
	request.Header.Set("OpenAI-Beta", "assistants=v1")
}
