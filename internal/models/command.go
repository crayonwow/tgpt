package models

type Command string

func (t Command) String() string {
	return string(t)
}

const (
	CommandSummarize = Command("summarize")
)
