package share

import (
	"fmt"
	"time"
)

type MsgKind string

type String []byte

const (
	LogOk   MsgKind = "SUCCESS"
	LogInfo MsgKind = "INFO"
	LogWarn MsgKind = "WARN"
	LogErr  MsgKind = "ERR"

	DGray  string = "\033[90m"
	LGray  string = "\033[37m"
	Blue   string = "\033[34m"
	Yellow string = "\033[33m"
	Green  string = "\033[32m"
	Red    string = "\033[31m"
	Bold   string = "\033[1m"
	Reset  string = "\033[0m"
)

// WriteLog padroniza o log do server e do client
// =================================================>
func WriteLog(kind MsgKind, msg string, origin string) {
	if len(origin) == 0 {
		origin = "internal"
	}
	color := Green
	switch kind {
	case LogOk:
		color = Green
		break
	case LogInfo:
		color = Blue
		break
	case LogWarn:
		color = Yellow
		break
	case LogErr:
		color = Red
		break
	default:
		color = Red
		break
	}
	fmt.Printf("%s|%s|%s - %s%s%s%s :: %s - %s%s%s\n", color, kind, Reset, Bold, LGray, origin, Reset, msg, Bold, time.Now().Format("02-Jan-2006 03:04 PM"), Reset)
}
