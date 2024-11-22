package models

import (
	"time"
)

type Message struct {
	TimeSend     time.Time
	UserName     UserID
	FromUserName UserID
	Text         string
	Topic        string
	Command      string
}
