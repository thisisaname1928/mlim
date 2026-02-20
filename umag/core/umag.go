package core

import (
	"encoding/json"
	"errors"
	"io"
	"os"
)

var CUR_UMAG_VER = 1

type UmagDB struct {
	Revision      int     `json:"revision"`
	Users         []User  `json:"user"`
	Groups        []Group `json:"group"`
	FreeUserList  []uint  `json:"freeUserList"`
	FreeGroupList []uint  `json:"freeGroupList"`
	file          *os.File
}

const UMAGDB_PATH = "/etc/umag.json"

func OpenUmagDB(f *os.File) (UmagDB, error) {
	var res UmagDB

	f.Seek(0, io.SeekStart)
	var r io.Reader = f
	b, e := io.ReadAll(r)
	if e != nil {
		return res, e
	}

	e = json.Unmarshal(b, &res)
	if e != nil {
		return res, e
	}

	res.file = f

	return res, nil
}

func (db *UmagDB) Flush() error {
	b, e := json.Marshal(db)

	if e != nil {
		return errors.New("CANT_FLUSH")
	}

	db.file.Seek(0, io.SeekStart)
	db.file.Truncate(0)
	n, e := db.file.Write(b)

	if n < len(b) || e != nil {
		return errors.New("CANT_FLUSH")
	}

	db.file.Sync()

	return nil
}
