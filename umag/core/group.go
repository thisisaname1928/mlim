package core

import "errors"

type Group struct {
	Name    string `json:"name"`
	GroupID uint   `json:"groupID"`
	User    uint   `json:"user"`   // user that hold this group
	IsFree  bool   `json:"isFree"` // free for using
}

func (db *UmagDB) IsFreeGroup(id uint) bool {
	if id >= uint(len(db.Groups)) {
		return false
	}

	return db.Groups[id].IsFree
}

func (db *UmagDB) AddGroup(u Group) {
	// search for free uid
	for i, v := range db.FreeGroupList {
		if db.IsFreeGroup(v) {
			// ok
			u.GroupID = v
			db.Groups[v] = u
			db.FreeGroupList = removeElement(db.FreeGroupList, i)
			db.Flush()
			return
		}
	}

	// just append to the list
	u.GroupID = uint(len(db.Groups))
	db.Groups = append(db.Groups, u)
	db.Flush()
}

func (db *UmagDB) RemoveGroup(u uint) error {
	if u == 0 {
		return errors.New("ATTEMP_DEL_ROOT_GROUP")
	}

	if !db.IsFreeGroup(u) {
		db.Groups[u].IsFree = true

		if uint(len(db.Groups)) > u {
			db.FreeGroupList = append(db.FreeGroupList, u)
		}

		db.Flush()

		return nil
	}

	return errors.New("GROUP_NOT_FOUND")
	// should remove their group
}
