package ifelse

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"
)

const idxTimeFmt = "2006-01-02 15:04:05.99999"

type addresses struct {
	ID        string
	CreatedAt string
	UpdatedAt string
}

func (i addresses) switchParse(b []byte) (interface{}, error) {
	var d addresses
	err := json.Unmarshal(b, &d)
	if err != nil {
		return nil, err
	}
	id, err := strconv.ParseInt(d.ID, 10, 64)
	if err != nil {
		return nil, err
	}
	cr, err := time.Parse(idxTimeFmt, d.CreatedAt)
	if err != nil {
		return nil, err
	}
	ud, err := time.Parse(idxTimeFmt, d.UpdatedAt)
	if err != nil {
		return nil, err
	}
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
	if err := errors.New("test if"); err != nil {
		return false, err
	}
	return true, nil
}
