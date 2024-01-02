package bot

import (
	"exile-telegram-bot/pkg/db"
	"exile-telegram-bot/pkg/game"
	"fmt"
	"log"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const maxMessageLength = 4096

func StartBot(token string) error {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Printf("Error initializing bot: %v", err)
		return err
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Printf("Error getting updates channel: %v", err)
		return err
	}

	for update := range updates {
		if update.CallbackQuery != nil {
			handleCallbackQuery(bot, update.CallbackQuery)
			continue
		}

		if update.Message == nil {
			continue
		}

		telegramUserID := strconv.Itoa(update.Message.From.ID)

		if update.Message.IsCommand() && update.Message.Command() == "restart" {
			if err := db.RestartGame(telegramUserID); err != nil {
				log.Printf("Error restarting game: %v", err)
			}
			_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Game restarted."))
			continue
		}

		gameResponse, err := game.ProcessGameMessage(telegramUserID, update.Message.Text)
		if err != nil {
			log.Println("Error processing game message:", err)
			continue
		}

		text := gameResponse.Text + "\n\n"
		for i, choice := range gameResponse.Choices {
			text += fmt.Sprintf("%d. %s\n", i+1, choice)
		}
		if len(text) > maxMessageLength {
			text = text[:maxMessageLength] + "..."
		}

		keyboard := makeInlineKeyboard(gameResponse.Choices)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
		msg.ReplyMarkup = keyboard
		if _, err := bot.Send(msg); err != nil {
			log.Println("Error sending message:", err)
		}
	}

	return nil
}

func makeInlineKeyboard(choices []string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for i := range choices {
		callbackData := strconv.Itoa(i)
		button := tgbotapi.NewInlineKeyboardButtonData(strconv.Itoa(i+1), callbackData)
		row := tgbotapi.NewInlineKeyboardRow(button)
		rows = append(rows, row)
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func handleCallbackQuery(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery) {
	log.Printf("Callback Query Data: %s", callbackQuery.Data)

	choiceIndex, err := strconv.Atoi(callbackQuery.Data)
	if err != nil {
		log.Printf("Error converting callback data to int: %v", err)
		return
	}

	telegramUserID := strconv.Itoa(callbackQuery.From.ID)
	userChoice := fmt.Sprintf("Choice %d", choiceIndex)
	gameResponse, err := game.ProcessGameMessage(telegramUserID, userChoice)
	if err != nil {
		log.Println("Error processing game choice:", err)
		return
	}

	text := gameResponse.Text
	if len(text) > maxMessageLength {
		text = text[:maxMessageLength] + "..."
	}

	editMsgText := tgbotapi.NewEditMessageText(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID, text)
	if _, err := bot.Send(editMsgText); err != nil {
		log.Println("Error updating message text:", err)
	}

	if len(gameResponse.Choices) > 0 {
		keyboard := makeInlineKeyboard(gameResponse.Choices)
		editMarkup := tgbotapi.NewEditMessageReplyMarkup(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID, keyboard)
		if _, err := bot.Send(editMarkup); err != nil {
			log.Println("Error updating message reply markup:", err)
		}
	}
}