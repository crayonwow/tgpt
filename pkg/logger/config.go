package logger

import (
	"io"
)

const (
	HandlerTypeText = "text"
	HandlerTypeJSON = "json"
)

type Config struct {
	Out         io.Writer
	Level       string
	HandlerType string
}
