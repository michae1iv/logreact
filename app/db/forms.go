package db

import (
	"fmt"
	"slices"
)

// Automatically checks if given params sattisfy type of variable in db, also checks if password, username or group_id are blank
//
// password: string
//
// username: string
//
// fullname: string
//
// email:    string
//
// group:     uint
//
// is_active: bool
//
// cpofl:    bool
func (u *User) AddUserForm(params map[string]interface{}) error {
	pas := ""
	us := ""
	fu := ""
	em := ""
	var g_id uint = 0
	act := true
	cp := false

	if value, ok := params["password"].(string); !ok && params["password"] != nil {
		return fmt.Errorf("add user: password supposed to be of type %T", pas)
	} else if params["password"] != nil {
		pas = value
	}

	if value, ok := params["username"].(string); !ok && params["username"] != nil {
		return fmt.Errorf("add user: username supposed to be of type %T", us)
	} else if params["username"] != nil {
		us = value
	}

	if value, ok := params["fullname"].(string); !ok && params["fullname"] != nil {
		return fmt.Errorf("add user: fullname supposed to be of type %T", fu)
	} else if params["fullname"] != nil {
		fu = value
	}

	if value, ok := params["email"].(string); !ok && params["email"] != nil {
		return fmt.Errorf("add user: email supposed to be of type %T", em)
	} else if params["email"] != nil {
		em = value
	}

	if value, ok := params["group"].(float64); !ok && params["group"] != nil {
		return fmt.Errorf("add user: group_id supposed to be of type %T", g_id)
	} else if params["group"] != nil && value > 0 {
		g_id = uint(value)
	}

	if value, ok := params["is_active"].(bool); !ok && params["is_active"] != nil {
		return fmt.Errorf("add user: isActive supposed to be of type %T", act)
	} else if params["is_active"] != nil {
		act = value
	}

	if value, ok := params["cpofl"].(bool); !ok && params["cpofl"] != nil {
		return fmt.Errorf("add user: cpofl supposed to be of type %T", cp)
	} else if params["cpofl"] != nil {
		cp = value
	}

	if us == "" || pas == "" || g_id == 0 {
		return fmt.Errorf("add user: username, password and group_id cannot be blank")
	}

	res := DB.Where(&User{Username: us}).First(u)
	if res.RowsAffected != 0 {
		return fmt.Errorf("user with such username already exists")
	}
	group := Group{}
	res = DB.Where(&Group{ID: g_id}).First(&group)
	if res.RowsAffected == 0 {
		return fmt.Errorf("group doesn't exist")
	}

	u.Group = group
	u.Password = pas
	u.Username = us
	u.FullName = fu
	u.Email = em
	u.IsActive = act
	u.ChangePass = cp
	return nil
}

func (u *User) EditUserForm(params map[string]interface{}) error {

	for k, v := range params {
		switch k {
		case "password":
			if value, ok := v.(string); !ok {
				return fmt.Errorf("edit user: password supposed to be of type string")
			} else {
				if value == "" {
					return fmt.Errorf("edit user: password is blank")
				}
				u.Password = value
				err := u.HashPassword()
				if err != nil {
					return fmt.Errorf("edit user: bad password")
				}
			}
		case "username":
			if value, ok := v.(string); !ok {
				return fmt.Errorf("edit user: username supposed to be of type string")
			} else {
				if value == "" {
					return fmt.Errorf("edit user: username is blank")
				}
				res := DB.Where(&User{Username: value}).First(u)
				if res.RowsAffected != 0 {
					return fmt.Errorf("user with such username already exists")
				}
				u.Username = value
			}
		case "fullname":
			if value, ok := v.(string); !ok {
				return fmt.Errorf("edit user: fullname supposed to be of type string")
			} else {
				u.FullName = value
			}
		case "email":
			if value, ok := v.(string); !ok {
				return fmt.Errorf("edit user: email supposed to be of type string")
			} else {
				u.Email = value
			}
		case "group":
			if value, ok := v.(float64); !ok {
				return fmt.Errorf("edit user: group_id supposed to be of type uint")
			} else if value > 0 {
				group := Group{}
				res := DB.Where(&Group{ID: uint(value)}).First(&group)
				if res.RowsAffected == 0 {
					return fmt.Errorf("group doesn't exist")
				}
				u.Group = group
			}
		case "is_active":
			if value, ok := v.(bool); !ok {
				return fmt.Errorf("edit user: isActive supposed to be of type bool")
			} else {
				u.IsActive = value
			}
		case "cpofl":
			if value, ok := v.(bool); !ok {
				return fmt.Errorf("edit user: cpofl supposed to be of type bool")
			} else {
				u.ChangePass = value
			}
		}
	}
	return nil
}

// Automatically checks if given params sattisfy type of variable in db, also checks if groupname is blank
//
// groupname: string
//
// desc: string
//
// perm: map[string]interface{}
//
// is_active: bool
func (g *Group) AddGroupForm(params map[string]interface{}) error {
	var gn string = ""
	var des string = ""
	var act bool = true
	var perm map[string]interface{}
	perm = map[string]interface{}{"rule": "none", "list": "none", "admin": false}

	if value, ok := params["groupname"].(string); !ok && params["groupname"] != nil {
		return fmt.Errorf("add group: groupname supposed to be of type %T", gn)
	} else if params["groupname"] != nil {
		gn = value
	}

	if value, ok := params["desc"].(string); !ok && params["desc"] != nil {
		return fmt.Errorf("add group: description supposed to be of type %T", des)
	} else if params["desc"] != nil {
		des = value
	}

	if value, ok := params["is_active"].(bool); !ok && params["is_active"] != nil {
		return fmt.Errorf("add group: is_active supposed to be of type %T", act)
	} else if params["is_active"] != nil {
		act = value
	}

	if value, ok := params["perm"].(map[string]interface{}); !ok && params["perm"] != nil {
		return fmt.Errorf("add group: perm supposed to be of type %T", perm)
	} else if params["perm"] != nil {
		perm = value
	}

	if gn == "" {
		return fmt.Errorf("add group: groupname cannot be blank")
	}
	res := DB.Where(&Group{GroupName: gn}).First(g)
	if res.RowsAffected != 0 {
		return fmt.Errorf("Group %s already exists", gn)
	}

	g.GroupName = gn
	g.Description = des
	g.IsActive = act
	g.Permissions = perm

	return nil
}

func (g *Group) EditGroupForm(params map[string]interface{}) error {

	for k, v := range params {
		switch k {
		case "groupname":
			if value, ok := v.(string); !ok {
				return fmt.Errorf("edit group: groupname supposed to be of type string")
			} else if value != "" {
				g.GroupName = value
			}
		case "desc":
			if value, ok := v.(string); !ok {
				return fmt.Errorf("edit group: description supposed to be of type string")
			} else if value != "" {
				g.Description = value
			}
		case "perm":
			if value, ok := v.(map[string]interface{}); !ok {
				return fmt.Errorf("edit group: permission supposed to be of type map[string]interface{}")
			} else {
				g.Permissions = value
			}
		case "is_active":
			if value, ok := v.(bool); !ok {
				return fmt.Errorf("edit group: isActive supposed to be of type bool")
			} else {
				g.IsActive = value
			}
		}
	}
	return nil
}

// function to add list
//
// name: string
//
// desc: string
//
// list: []string
func (l *List) AddListForm(params map[string]interface{}) error {
	var name string = ""
	var desc string = ""
	var list []interface{}
	var items []string

	if value, ok := params["name"].(string); !ok && params["name"] != nil {
		return fmt.Errorf("create list: name supposed to be of type %T", name)
	} else if params["name"] != nil {
		name = value
	}

	if value, ok := params["desc"].(string); !ok {
		return fmt.Errorf("create list: desc supposed to be of type %T", desc)
	} else {
		desc = value
	}

	if value, ok := params["list"].([]interface{}); !ok && params["list"] != nil {
		fmt.Printf("list type of %T", params["list"])
		return fmt.Errorf("create list: list supposed to be of type []string")
	} else if params["list"] != nil {
		list = value
	}
	for _, el := range list {
		if v, ok := el.(string); ok {
			items = append(items, v)
		} else {
			return fmt.Errorf("create list: list supposed to be of type []string")
		}
	}

	if name == "" {
		return fmt.Errorf("add list: name cannot be blank")
	}
	res := DB.Where(&List{List_name: name}).First(l)
	if res.RowsAffected != 0 {
		return fmt.Errorf("List %s already exists", name)
	}

	l.List_name = name
	l.Description = desc
	l.Phrases = items

	return nil
}

// function to check fields, delete and add entries
//
// name: string
//
// desc: string
//
// add: []string
//
// del: []string
func (l *List) EditListForm(params map[string]interface{}) error {
	var to_add []string
	var to_del []string
	var to_save []string

	for k, v := range params {
		switch k {
		case "name":
			if value, ok := v.(string); !ok {
				return fmt.Errorf("edit list: name supposed to be of type %T", l.List_name)
			} else if value != "" {
				res := DB.Where(&List{List_name: value}).First(l)
				if res.RowsAffected != 0 {
					return fmt.Errorf("such list %s already exists", value)
				}
				l.List_name = value
			}
		case "desc":
			if value, ok := v.(string); !ok {
				return fmt.Errorf("edit list: description supposed to be of type %T", l.Description)
			} else if value != "" {
				l.Description = value
			}
		case "add":
			if value, ok := v.([]interface{}); !ok {
				return fmt.Errorf("edit list: add supposed to be of type %T", to_add)
			} else if value != nil {
				for _, el := range value {
					str, check := el.(string)
					if !check {
						continue
					}
					if slices.Contains(to_add, str) || slices.Contains(l.Phrases, str) {
						continue
					}
					to_add = append(to_add, str)
				}
				l.Phrases = append(l.Phrases, to_add...)
			}
		case "del":
			if value, ok := v.([]interface{}); !ok {
				return fmt.Errorf("edit list: del supposed to be of type %T", to_del)
			} else if value != nil {
				for _, el := range value {
					str, check := el.(string)
					if check {
						to_del = append(to_del, str)
					}
				}
				for _, el := range l.Phrases {
					if slices.Contains(to_del, el) {
						continue
					}
					to_save = append(to_save, el)
				}
				l.Phrases = to_save
			}
		}
	}

	return nil
}
