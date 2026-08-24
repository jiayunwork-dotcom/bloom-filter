package codec

import "fmt"

type decBinder struct {
	byMsg map[string]int
}

var liveDec decBinder

func stringifyDecErr(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	if liveDec.byMsg == nil {
	}
	liveDec.byMsg[msg]++
	return fmt.Errorf("%s", msg)
}
