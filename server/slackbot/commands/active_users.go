package commands

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"schej.it/server/db"
	"schej.it/server/logger"
	"schej.it/server/models"
	"schej.it/server/utils"
)

var activeUsers Command = Command{
	Name: "/active_users",
	Description: `Gets the number of active users in the database, based on last sign in date. 
  - if LIST is true, it will list the name/email of all users, otherwise, it will show a bar graph
  - DAYS is the amount of days since last sign in
  `,
	Usage: "/active_users [LIST=false] [DAYS=7]",
	Execute: func(args []string, webhookUrl string) {
		var err error

		// Parse args
		list := false
		days := 7
		if len(args) >= 1 {
			if args[0] == "true" {
				list = true
			} else if args[0] == "false" {
				list = false
			} else {
				SendRawMessage(&Response{ResponseType: "ephemeral", Text: fmt.Sprintf("LIST=%s is not a valid boolean!", args[0])}, webhookUrl)
				return
			}
		}
		if len(args) >= 2 {
			days, err = strconv.Atoi(args[1])
			if err != nil {
				SendRawMessage(&Response{ResponseType: "ephemeral", Text: fmt.Sprintf("DAYS=%s is not a valid number!", args[1])}, webhookUrl)
				return
			}
		}

		// Query for daily user logs starting from `days` days before the current date
		startDate := time.Now().AddDate(0, 0, -days)
		startDate = utils.GetDateAtTime(startDate, "00:00:00")
		var logs []models.DailyUserLog
		if err := db.ORM().Where("date >= ?", startDate).Order("date DESC").Find(&logs).Error; err != nil {
			logger.StdErr.Panicln(err)
		}
		if list {
			for i := range logs {
				if len(logs[i].UserIds) == 0 {
					logs[i].Users = []models.User{}
					continue
				}
				if err := db.ORM().Where("id IN ?", logs[i].UserIds).Find(&logs[i].Users).Error; err != nil {
					logger.StdErr.Panicln(err)
				}
			}
		}

		// Add empty days
		curDate := startDate
		for i := len(logs) - 1; i >= 0; i-- {
			// Add all dates up to the current log date
			for !logs[i].Date.Time().Equal(curDate) && curDate.Before(time.Now()) {
				// Insert curDate into logs, with an empty users array
				logs, err = utils.Insert(logs, i+1, models.DailyUserLog{
					Date:  models.NewDateTime(curDate),
					Users: make([]models.User, 0),
				})
				if err != nil {
					logger.StdErr.Panicln(err)
				}
				curDate = curDate.AddDate(0, 0, 1)
			}

			// Increase curDate by a day
			curDate = curDate.AddDate(0, 0, 1)
		}

		// Add all dates up to the current date
		for curDate.Before(time.Now()) {
			logs, err = utils.Insert(logs, 0, models.DailyUserLog{
				Date:  models.NewDateTime(curDate),
				Users: make([]models.User, 0),
			})
			if err != nil {
				logger.StdErr.Panicln(err)
			}
			curDate = curDate.AddDate(0, 0, 1)
		}

		// Define constants
		dayStrings := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}

		if list {
			// Display a list of all active users
			SendRawMessage(&Response{ResponseType: "in_channel", Text: "Active Users:\n"}, webhookUrl)
			message := ""
			for _, log := range logs {
				date := log.Date.Time()
				message += dayStrings[date.Weekday()] + " "
				message += utils.GetDateString(date) + " | "
				message += fmt.Sprintf("Count: %d\n", len(log.Users))

				for _, user := range log.Users {
					message += fmt.Sprintf("\t- %s %s (%s)\n", user.FirstName, user.LastName, user.Email)
				}
			}

			for _, msg := range splitLongMessage(message, "```") {
				SendRawMessage(&Response{ResponseType: "in_channel", Text: msg}, webhookUrl)
			}
		} else {
			// Display a bar graph of active users over time

			// Generate labels and data based on logs
			labels := make([]string, 0)
			data := make([]int, 0)
			for i := len(logs) - 1; i >= 0; i-- {
				labels = append(labels, utils.GetDateString(logs[i].Date.Time()))
				data = append(data, len(logs[i].UserIds))
			}

			// Generate chart using QuickChart API
			chart := map[string]interface{}{
				"type": "bar",
				"data": map[string]interface{}{
					"labels": labels,
					"datasets": []interface{}{map[string]interface{}{
						"label": "Active Users",
						"data":  data,
					}},
				},
				"options": map[string]interface{}{
					"scales": map[string]interface{}{
						"yAxes": []interface{}{map[string]interface{}{
							"ticks": map[string]interface{}{
								"stepSize": 1,
							},
						}},
					},
				},
			}
			jsonStr, _ := json.Marshal(chart)

			encodedChart := url.PathEscape(string(jsonStr))
			chartUrl := fmt.Sprintf(`https://quickchart.io/chart?c=%s&backgroundColor=white`, encodedChart)

			SendRawMessage(&Response{ResponseType: "in_channel", Blocks: []map[string]interface{}{
				{
					"type":      "image",
					"image_url": chartUrl,
					"alt_text":  "Active users chart",
				},
			}}, webhookUrl)
		}
	},
}
