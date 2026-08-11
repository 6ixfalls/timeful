package commands

import (
	"fmt"

	"schej.it/server/db"
	"schej.it/server/logger"
	"schej.it/server/models"
)

var numUsers Command = Command{
	Name:        "/num_users",
	Description: "Returns the number of signed up users",
	Execute: func(args []string, webhookUrl string) {
		var n int64
		if err := db.ORM().Model(&models.User{}).Count(&n).Error; err != nil {
			logger.StdErr.Panicln(err)
		}

		response := Response{
			ResponseType: "in_channel",
			Text:         fmt.Sprintf("Number of currently signed up users: %v", n),
		}
		SendRawMessage(&response, webhookUrl)
	},
}
