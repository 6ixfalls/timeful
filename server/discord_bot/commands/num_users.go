package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"schej.it/server/db"
	"schej.it/server/logger"
	"schej.it/server/models"
)

var numUsers Command = Command{
	Name:        "!num_users",
	Description: "Returns the number of signed up users",
	Execute: func(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
		var n int64
		if err := db.ORM().Model(&models.User{}).Count(&n).Error; err != nil {
			logger.StdErr.Panicln(err)
		}

		sendMessage(s, m, fmt.Sprintf("Number of currently signed up users: %v", n))
	},
}
