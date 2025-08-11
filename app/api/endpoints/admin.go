package endpoints

import (
	"correlator/db"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// function which handles admin page
func AdminStats(w http.ResponseWriter, r *http.Request) {

	// Counting all records in db
	var total_u int64 = 0
	db.DB.Model(&db.User{}).Count(&total_u)

	var total_g int64 = 0
	db.DB.Model(&db.Group{}).Count(&total_g)

	var total_l int64 = 0
	db.DB.Model(&db.List{}).Count(&total_l)

	var total_r int64 = 0
	db.DB.Model(&db.Rules{}).Count(&total_r)

	var total_a int64 = 0
	db.DB.Model(&db.Alert{}).Count(&total_a)

	response := map[string]interface{}{
		"users":  total_u,
		"groups": total_g,
		"lists":  total_l,
		"rules":  total_r,
		"alerts": total_a,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Get's all users from database
func GetAllUsers(w http.ResponseWriter, r *http.Request) {

	// Selecting all users from database
	var users []db.User
	res := db.DB.Find(&users)
	if res.Error != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	var response []map[string]interface{}

	for _, user := range users {
		res := db.DB.Preload("Group").Where("id = ?", user.ID).First(&user)
		if res.Error != nil {
			http.Error(w, http.StatusText(http.StatusFailedDependency), http.StatusFailedDependency)
			return
		}

		data := map[string]interface{}{
			"id":        user.ID,
			"username":  user.Username,
			"group":     user.Group.GroupName,
			"is_active": user.IsActive,
		}
		response = append(response, data)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Returns info about user by id
func GetUser(w http.ResponseWriter, r *http.Request) {
	u_id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || u_id < 0 {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
	}
	var user db.User
	res := db.DB.Where(&db.User{ID: uint(u_id)}).First(&user)
	if res.RowsAffected != 1 {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"id":        user.ID,
		"username":  user.Username,
		"fullname":  user.FullName,
		"group":     user.GroupID,
		"email":     user.Email,
		"is_active": user.IsActive,
		"cpofl":     user.ChangePass,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Creating user, password and selecting group
func CreateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var response map[string]interface{}
	var params map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	u := db.User{}
	err = u.AddUserForm(params)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		response = map[string]interface{}{"operation": "create", "status": "error", "message": err.Error()}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Generating salt and hashing user password
	err = u.HashPassword()
	if err != nil {
		http.Error(w, "Error occured while creating user", http.StatusInternalServerError)
		return
	}

	// Saving new user in database
	if err := db.DB.Create(&u).Error; err != nil {
		http.Error(w, "Error occured while creating user", http.StatusInternalServerError)
		return
	}

	response = map[string]interface{}{"operation": "create", "status": "success", "message": fmt.Sprintf("User %s created", u.Username)}
	json.NewEncoder(w).Encode(response)
}

// Updating user with method PATCH, which allows send very fields that needs to be updated
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var response map[string]interface{}
	var params map[string]interface{}

	u_id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || u_id < 0 {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
	}

	err = json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	u := db.User{}
	res := db.DB.Where(&db.User{ID: uint(u_id)}).First(&u)
	if res.RowsAffected != 1 {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	err = u.EditUserForm(params)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		response = map[string]interface{}{"operation": "update", "status": "error", "message": err.Error()}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Updating user
	if tx := db.DB.Save(&u); tx.Error != nil {
		http.Error(w, "Error occured while updating user", http.StatusInternalServerError)
		return
	}
	response = map[string]interface{}{"operation": "update", "status": "success", "message": fmt.Sprintf("User %s updated", u.Username)}
	json.NewEncoder(w).Encode(response)
}

// Deleting User
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	u_id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || u_id < 0 {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
	}

	res := db.DB.Delete(&db.User{}, uint(u_id))
	if res.Error != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	response := map[string]interface{}{
		"operation": "delete",
		"userid":    fmt.Sprintf("User id: %v deleted", u_id),
		"status":    "success",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

/*--------------------------------------------------------GROUPS API------------------------------------------------------------------------------------------------------*/

// Get's all users from database
func GetAllGroups(w http.ResponseWriter, r *http.Request) {

	// Selecting all users from database
	var groups []db.Group
	res := db.DB.Find(&groups)
	if res.Error != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	var response []map[string]interface{}

	for _, group := range groups {
		data := map[string]interface{}{
			"id":        group.ID,
			"groupname": group.GroupName,
			"desc":      group.Description,
		}
		response = append(response, data)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Returns info about group by id
func GetGroup(w http.ResponseWriter, r *http.Request) {
	g_id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || g_id < 0 {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
	}
	group := db.Group{}
	res := db.DB.Where(&db.Group{ID: uint(g_id)}).First(&group)
	if res.RowsAffected != 1 {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"id":        group.ID,
		"groupname": group.GroupName,
		"desc":      group.Description,
		"is_active": group.IsActive,
		"perm": map[string]interface{}{
			"rule":  group.Permissions["rule"],
			"admin": group.Permissions["admin"],
			"list":  group.Permissions["list"],
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Creating user, password and selecting group
func CreateGroup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var response map[string]interface{}

	var params map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	group := db.Group{}
	err = group.AddGroupForm(params)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		response = map[string]interface{}{"operation": "create", "status": "error", "message": err.Error()}
		json.NewEncoder(w).Encode(response)
		return
	}

	if err := db.DB.Create(&group).Error; err != nil {
		http.Error(w, "Error occured while creating group", http.StatusInternalServerError)
		return
	}

	response = map[string]interface{}{
		"operation": "create",
		"status":    "success",
		"message":   fmt.Sprintf("Group %s created", group.GroupName),
	}
	json.NewEncoder(w).Encode(response)
}

// Updating group with method PATCH, which allows send very fields that needs to be updated
func UpdateGroup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var response map[string]interface{}

	g_id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || g_id < 0 {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
	}
	var params map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	g := db.Group{}
	res := db.DB.Where(&db.Group{ID: uint(g_id)}).First(&g)
	if res.RowsAffected != 1 {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	err = g.EditGroupForm(params)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		response = map[string]interface{}{"operation": "update", "status": "error", "message": err.Error()}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Updating user
	if tx := db.DB.Save(&g); tx.Error != nil {
		http.Error(w, "Error occured while creating user", http.StatusInternalServerError)
		return
	}

	response = map[string]interface{}{
		"operation": "update",
		"message":   fmt.Sprintf("Group %s updated", g.GroupName),
		"status":    "success",
	}
	json.NewEncoder(w).Encode(response)
}

// Deleting User
func DeleteGroup(w http.ResponseWriter, r *http.Request) {
	g_id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || g_id < 0 {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
	}
	res := db.DB.Delete(&db.Group{}, uint(g_id))
	if res.Error != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	response := map[string]interface{}{
		"operation": "delete",
		"message":   fmt.Sprintf("Group id: %v deleted", g_id),
		"status":    "success",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
