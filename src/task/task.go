package task

import (
	"fmt"
	"github.com/Aoi-hosizora/scut-academic-notifier/src/bot/server"
	"github.com/Aoi-hosizora/scut-academic-notifier/src/bot/serverchan"
	"github.com/Aoi-hosizora/scut-academic-notifier/src/config"
	"github.com/Aoi-hosizora/scut-academic-notifier/src/database"
	"github.com/Aoi-hosizora/scut-academic-notifier/src/logger"
	"github.com/Aoi-hosizora/scut-academic-notifier/src/model"
	"github.com/Aoi-hosizora/scut-academic-notifier/src/service"
	"github.com/Aoi-hosizora/scut-academic-notifier/src/static"
	"github.com/robfig/cron/v3"
	"gopkg.in/tucnak/telebot.v2"
	"math"
	"strings"
	"sync"
)

var Cron *cron.Cron

func Setup() error {
	Cron = cron.New(cron.WithSeconds())

	_, err := Cron.AddFunc(config.Configs.Task.Cron, task)
	if err != nil {
		return err
	}

	return nil
}

func foreachUsers(users []*model.User, fn func(user *model.User)) {
	wg := sync.WaitGroup{}
	for _, user := range users {
		wg.Add(1)
		go func(user *model.User) {
			fn(user)
			wg.Done()
		}(user)
	}
	wg.Wait()
}

func task() {
	defer func() { recover() }()

	users := database.GetUsers()
	if len(users) == 0 {
		return
	}

	foreachUsers(users, func(user *model.User) {
		// get new items
		jwItems, err := service.GetJwItems()
		if err != nil {
			return
		}
		seItems, err := service.GetSeItems()
		if err != nil {
			return
		}

		// filter new items
		newItems := make([]*model.PostItem, 0)
		for _, jw := range jwItems {
			if service.CheckTime(jw.Date, config.Configs.Send.Range) {
				newItems = append(newItems, jw)
			}
		}
		for _, se := range seItems {
			if service.CheckTime(se.Date, config.Configs.Send.Range) {
				newItems = append(newItems, se)
			}
		}
		if len(newItems) == 0 {
			return
		}

		// get old items and get diff
		oldItems, ok := database.GetOldItems(user.ChatID)
		if !ok {
			return
		}
		logger.Logger.Infof("Get old data: #%d | %d", len(oldItems), user.ChatID)
		sendItems := model.ItemSliceDiff(newItems, oldItems)
		logger.Logger.Infof("Get diff data: #%d | %d", len(sendItems), user.ChatID)
		if len(sendItems) == 0 {
			return
		}

		// update old items
		ok = database.SetOldItems(user.ChatID, newItems)
		logger.Logger.Infof("Set new data: #%d | %d", len(newItems), user.ChatID)
		if !ok {
			return
		}

		// ===============
		// render and send
		// ===============

		moreStr := fmt.Sprintf("更多信息，请查阅 [华工教务通知](%s) 以及 [软院公务通知](%s)。", static.JwHomepage, static.SeHomepage)

		// send to telebot
		sb := strings.Builder{}
		sb.WriteString("*学校相关通知*\n=====\n")
		for idx, item := range sendItems {
			sb.WriteString(fmt.Sprintf("%d. %s\n", idx+1, item.String()))
		}
		sb.WriteString("=====\n")
		sb.WriteString(moreStr)
		msg := sb.String()
		_ = server.Bot.SendToChat(user.ChatID, msg, telebot.ModeMarkdown)

		// send to wechat
		maxCnt := int(config.Configs.Send.MaxCount)
		sendTimes := int(math.Ceil(float64(len(sendItems)) / float64(maxCnt)))
		for i := 0; i < sendTimes; i++ {
			from := i * maxCnt
			to := (i + 1) * maxCnt
			if l := len(sendItems); to > l {
				to = l
			}
			sb := strings.Builder{}
			for j := from; j < to; j++ {
				sb.WriteString(fmt.Sprintf("%d. %s\n", j+1, sendItems[j].String()))
			}
			if i == sendTimes-1 {
				sb.WriteString("\n--- \n")
				sb.WriteString(moreStr)
			}
			msg := sb.String()
			title := fmt.Sprintf("学校相关通知 (第 %d 条，共 %d 条)", i+1, sendTimes)
			_ = serverchan.SendToChat(user.Sckey, title, msg)
		}
	})
}
