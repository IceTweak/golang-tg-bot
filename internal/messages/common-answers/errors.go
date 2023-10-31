package commonanswers

import "github.com/IceTweak/golang-tg-bot/internal/model"

func UnknownError() model.Message {
	return model.Message{
		Text: "Sorry something went wrong :(\nPlease try again later.",
	}
}
