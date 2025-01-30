package logger

import (
	"bytes"
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
)

type ColorFormatter struct {
	logrus.TextFormatter
}

func (f *ColorFormatter) Format(entry *logrus.Entry) ([]byte, error) {

	var b bytes.Buffer

	// Colors for log components
	timestampColor := color.New(color.FgCyan)
	levelColor := color.New(color.FgWhite)
	messageColor := color.New(color.Reset) // Default terminal color

	// Determine level color
	switch entry.Level {
	case logrus.TraceLevel:
		levelColor = color.New(color.FgHiMagenta)
	case logrus.DebugLevel:
		levelColor = color.New(color.FgHiGreen)
	case logrus.InfoLevel:
		levelColor = color.New(color.FgHiBlue)
	case logrus.WarnLevel:
		levelColor = color.New(color.FgYellow)
	case logrus.ErrorLevel:
		levelColor = color.New(color.BgRed, color.FgWhite)
	case logrus.FatalLevel, logrus.PanicLevel:
		levelColor = color.New(color.BgRed, color.FgWhite)
	}

	// Format timestamp, level, and message
	timestamp := timestampColor.Sprint(entry.Time.Format(time.StampMilli))
	level := levelColor.Sprintf("[%s]", entry.Level.String())
	message := messageColor.Sprintf("%s", entry.Message)

	// Write main log line
	b.WriteString(fmt.Sprintf("%s %s %s", timestamp, level, message))

	// add a differet color for custom fields
	for key, value := range entry.Data {
		if key != "func" && key != "src" {
			fieldColor := color.New(color.FgHiYellow)
			fieldKey := fieldColor.Sprint(key)
			fieldValue := fmt.Sprintf("%v", value)
			b.WriteString(fmt.Sprintf("\t%s: %s", fieldKey, fieldValue))
		}
	}

	// ensure we add func and src fields at the end
	fieldColor := color.New(color.FgCyan)
	if funcVal, ok := entry.Data["func"]; ok {
		fieldKey := fieldColor.Sprint("func")
		fieldValue := fmt.Sprintf("%s", funcVal)
		b.WriteString(fmt.Sprintf("\t%s: %s", fieldKey, fieldValue))
	}
	if srcVal, ok := entry.Data["src"]; ok {
		fieldKey := fieldColor.Sprint("src")
		fieldValue := fmt.Sprintf("%s", srcVal)
		b.WriteString(fmt.Sprintf("\t%s: %s", fieldKey, fieldValue))
	}

	b.WriteString("\n")

	return b.Bytes(), nil
}
