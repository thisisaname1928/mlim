package core

import (
	"errors"
)

type User struct {
	Name         string `json:"name"`
	UserID       uint   `json:"userID"`
	NoLogin      bool   `json:"noLogin"`
	PasswordHash string `json:"passwordHash"`
	Group        uint   `json:"group"`  // group that this user own
	IsFree       bool   `json:"isFree"` // free for using
}

func (db *UmagDB) IsFreeUser(id uint) bool {
	if id >= uint(len(db.Users)) {
		return false
	}

	return db.Users[id].IsFree
}

func removeElement(slice []uint, index int) []uint {
	slice[index], slice[len(slice)-1] = slice[len(slice)-1], slice[index]
	return slice[:len(slice)-1]
}
func (db *UmagDB) AddUser(u User) {
	// search for free uid
	for i, v := range db.FreeUserList {
		if db.IsFreeUser(v) {
			// ok
			u.UserID = v
			db.Users[v] = u
			db.FreeUserList = removeElement(db.FreeUserList, i)
			db.Flush()
			return
		}
	}

	// just append to the list
	u.UserID = uint(len(db.Users))
	db.Users = append(db.Users, u)
	db.Flush()
}

func (db *UmagDB) RemoveUser(u uint) error {
	if u == 0 {
		return errors.New("ATTEMP_DEL_ROOT_USER")
	}

	if !db.IsFreeUser(u) {
		db.Users[u].IsFree = true

		if uint(len(db.Users)) > u {
			db.FreeUserList = append(db.FreeUserList, u)
		}

		db.Flush()

		return nil
	}

	return errors.New("USER_NOT_FOUND")
	// should remove their group
}
