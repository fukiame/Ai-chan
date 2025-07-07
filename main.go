package main

import (
	"embed"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/fukiame/Ai-chan/ai/config"
	"github.com/fukiame/Ai-chan/ai/i18n"
	"github.com/fukiame/Ai-chan/ai/utils/helpers"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"github.com/fukiame/Ai-chan/ai"
)

//go:embed locales
var Locales embed.FS

func main() {
	// logs if bot is running in debug mode or not
	if config.Debug {
		log.Info("Running in DEBUG Mode...")
	} else {
		log.Info("Running in RELEASE Mode...")
	}

	// Load Locales
	i18n.LoadLocaleFiles(&Locales, "locales")

	// create a new bot with default HTTP client (BotOpts doesn't support custom client in this version)
	b, err := gotgbot.NewBot(config.BotToken, nil)
	if err != nil {
		panic("failed to create new bot: " + err.Error())
	}

	// some initial checks before running bot
	ai.InitialChecks(b)

	// Create updater and dispatcher with limited max routines
	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
		// If an error is returned by a handler, log it and continue going.
		Error: func(_ *gotgbot.Bot, _ *ext.Context, err error) ext.DispatcherAction {
			log.Println("an error occurred while handling update:", err.Error())
			return ext.DispatcherActionNoop
		},
		MaxRoutines: 100, // Limit concurrent goroutines to prevent explosion
	})
	updater := ext.NewUpdater(dispatcher, nil) // create updater with dispatcher

	if _, err = b.DeleteWebhook(nil); err != nil {
		log.Fatalf("[Polling] Failed to remove webhook: %v", err)
	}
	log.Info("[Polling] Removed Webhook!")

	// start the bot in polling mode
	err = updater.StartPolling(b,
		&ext.PollingOpts{
			DropPendingUpdates: config.DropPendingUpdates,
			GetUpdatesOpts: &gotgbot.GetUpdatesOpts{
				AllowedUpdates: config.AllowedUpdates,
			},
		},
	)
	if err != nil {
		log.Fatalf("[Polling] Failed to start polling: %v", err)
	}
	log.Info("[Polling] Started Polling...!")

	// Log the message that bot started
	log.Infof("[Bot] %s has been started...", b.Username)

	// Set Commands of Bot
	log.Info("Setting Custom Commands for PM...!")
	_, err = b.SetMyCommands(
		[]gotgbot.BotCommand{
			{Command: "start", Description: "Starts the Bot"},
			{Command: "help", Description: "Check Help section of bot"},
		},
		&gotgbot.SetMyCommandsOpts{
			Scope:        gotgbot.BotCommandScopeAllPrivateChats{},
			LanguageCode: "en",
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	// Loading Modules
	ai.LoadModules(dispatcher)

	// list modules from modules dir
	log.Infof(
		fmt.Sprintf(
			"[Modules] Loaded modules: %s", ai.ListModules(),
		),
	)

	// send message to log group
	_, err = b.SendMessage(config.MessageDump,
		fmt.Sprintf("<b>Started Bot!</b>\n<b>Mode:</b> %s\n<b>Loaded Modules:</b>\n%s", config.WorkingMode, ai.ListModules()),
		&gotgbot.SendMessageOpts{
			ParseMode: helpers.HTML,
		},
	)
	if err != nil {
		log.Errorf("[Bot] Failed to send message to log group: %v", err)
		log.Fatal(err)
	}

	// Idle, to keep updates coming in, and avoid bot stopping.
	updater.Idle()
}
