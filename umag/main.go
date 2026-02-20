package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
	"grd.umag/core"
)

const umagDB_PATH = "/home/ngqt/proj/mlim/rootfs/rootfs/etc/umag.json"

func main() {

	f, e := os.OpenFile(umagDB_PATH, os.O_RDWR, 0)
	if e != nil {
		fmt.Println(e)
		return
	}

	db, e := core.OpenUmagDB(f)
	if e != nil {
		fmt.Println(e)
		return
	}

	// e = db.RemoveUser(1)

	// fmt.Println(e)
	hash, e := bcrypt.GenerateFromPassword([]byte("root"), bcrypt.DefaultCost)

	db.AddUser(core.User{Name: "root", UserID: 1, NoLogin: false, PasswordHash: string(hash), Group: 123, IsFree: false})
	db.Flush()

}
