package requestid

import (
	"fmt"
	"time"

	"github.com/jmcvetta/randutil"
)

type Config struct {
	Generator func() (string, error)
}

func (c Config) MakeID() (string, error) {
	if c.Generator != nil {
		id, err := c.Generator()
		return id, err
	}
	v, err := randutil.AlphaString(8)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%x", v, time.Now().Unix()), nil
}
