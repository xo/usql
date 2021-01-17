package snowflake

import (
	"io/ioutil"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/snowflakedb/gosnowflake" // DRIVER: snowflake
	"github.com/xo/usql/drivers"
)

func init() {
	r := logrus.New()
	r.Out, r.Level = ioutil.Discard, logrus.PanicLevel
	var l gosnowflake.SFLogger = &logger{r}
	gosnowflake.SetLogger(&l)
	drivers.Register("snowflake", drivers.Driver{
		Err: func(err error) (string, string) {
			if e, ok := err.(*gosnowflake.SnowflakeError); ok {
				return strconv.Itoa(e.Number), e.Message
			}
			return "", err.Error()
		},
	})
}

// logger is an empty logger.
type logger struct {
	*logrus.Logger
}

func (*logger) SetLogLevel(string) error { return nil }
