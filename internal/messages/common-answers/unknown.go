package commonanswers

import "github.com/IceTweak/golang-tg-bot/internal/model"

func UnknownCommand() model.Message {
	return model.Message{
		Text: "Sorry, I don't know such a command :(",
	}
}

func UnknownMessage() model.Message {
	return model.Message{
		Text: "Sorry, I didn't understand this message :(",
	}
}
