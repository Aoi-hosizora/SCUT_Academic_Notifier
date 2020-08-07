package logger

import (
	"github.com/Aoi-hosizora/ahlib-web/xserverchan"
	"github.com/Aoi-hosizora/ahlib-web/xtelebot"
	"github.com/Aoi-hosizora/ahlib/xlogger"
	"github.com/Aoi-hosizora/scut-academic-notifier/src/config"
	"github.com/sirupsen/logrus"
	"time"
)

var (
	Logger     *logrus.Logger
	Telebot    *xtelebot.TelebotLogrus
	Serverchan *xserverchan.ServerchanLogrus
)

func Setup() error {
	Logger = logrus.New()
	logLevel := logrus.WarnLevel
	if config.Configs.Meta.RunMode == "debug" {
		logLevel = logrus.DebugLevel
	}

	Logger.SetLevel(logLevel)
	Logger.SetReportCaller(false)
	Logger.AddHook(xlogger.NewRotateLogHook(&xlogger.RotateLogConfig{
		MaxAge:       15 * 24 * time.Hour,
		RotationTime: 24 * time.Hour,
		Filepath:     config.Configs.Meta.LogPath,
		Filename:     config.Configs.Meta.LogName,
		Level:        logLevel,
		Formatter:    &logrus.JSONFormatter{TimestampFormat: time.RFC3339},
	}))
	Logger.SetFormatter(&xlogger.CustomFormatter{
		ForceColor: true,
	})

	Telebot = xtelebot.NewTelebotLogrus(Logger, true)
	Serverchan = xserverchan.NewServerchanLogrus(Logger, true)

	return nil
}
