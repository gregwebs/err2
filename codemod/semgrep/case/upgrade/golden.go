package ifelse

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/gregwebs/try"
)

const idxTimeFmt = "2006-01-02 15:04:05.99999"

type addresses struct {
	ID        string
	CreatedAt string
	UpdatedAt string
}

func (i addresses) switchParse(b []byte) (interface{}, error) {
	defer try.Handle(&err, nil)
	var d addresses
	err := json.Unmarshal(b, &d)
	try.Check(err)
	id, err := strconv.ParseInt(d.ID, 10, 64)
	try.Check(err)
	cr, err := time.Parse(idxTimeFmt, d.CreatedAt)
	try.Check(err)
	ud, err := time.Parse(idxTimeFmt, d.UpdatedAt)
	try.Check(err)
	s := struct {
		id int64
		cr time.Time
		ud time.Time
	}{
		id: id,
		cr: cr,
		ud: ud,
	}
	return &s, nil
}

func ifErr() (bool, error) {
	defer try.Handle(&err, nil)
	{
		err := errors.New("test if")
		try.Check(err)
	}
	return true, nil
}
