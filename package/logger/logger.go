package logger

import (
	"os"

	"github.com/rs/zerolog"
)

func SetLogger() error {
	file, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	multi := zerolog.MultiLevelWriter(os.Stdout, file)
	zerolog.New(multi).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	return nil
}
