package endpoints

import (
	"correlator/api/session"
	"correlator/db"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

// Show all lists
func ShowLists(w http.ResponseWriter, r *http.Request) {
	lists := []db.List{}
	var data []map[string]interface{}
	res := db.DB.Select("ID", "List_name", "Description").Find(&lists)
	if res.Error != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	claims := r.Context().Value(session.UserCtxKey).(map[string]interface{})
	rights, ok := claims["Rights"].(string)
	if !ok {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	switch rights {
	case "none":
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	case "w":
		for _, list := range lists {
			if list.UserID == claims["User"].(*db.User).ID {
				params := map[string]interface{}{
					"id":   list.ID,
					"name": list.List_name,
					"desc": list.Description,
				}
				data = append(data, params)
			} else {
				continue
			}
		}
	default:
		if strings.Contains(rights, "r") {
			for _, list := range lists {
				params := map[string]interface{}{
					"id":   list.ID,
					"name": list.List_name,
					"desc": list.Description,
				}
				data = append(data, params)
			}
		}
	}
	w.Header().Set("Content-Type", "application/json")
	response, err := json.Marshal(data)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Write(response)

}

// Returns single list
func GetList(w http.ResponseWriter, r *http.Request) {
	l_id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || l_id < 0 {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
	}

	list := db.List{}
	res := db.DB.Where(db.List{ID: uint(l_id)}).First(&list)
	if res.RowsAffected == 0 {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	claims := r.Context().Value(session.UserCtxKey).(map[string]interface{})
	rights, ok := claims["Rights"].(string)
	if !ok {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	switch rights {
	case "none":
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	case "w":
		if list.UserID != claims["User"].(*db.User).ID {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
	default:
		if !strings.Contains(rights, "r") {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
	}

	response := map[string]interface{}{
		"id":   list.ID,
		"name": list.List_name,
		"desc": list.Description,
		"list": list.Phrases,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Creating new list
func CreateList(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(session.UserCtxKey).(map[string]interface{})
	if !strings.Contains(claims["Rights"].(string), "w") {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	var response map[string]interface{}
	var params map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		println(err.Error())
		return
	}

	list := db.List{}
	if err = list.AddListForm(params); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	list.User = *claims["User"].(*db.User)

	// Saving new list in database
	if err := db.DB.Create(&list).Error; err != nil {
		http.Error(w, "Error occured while creating list", http.StatusInternalServerError)
		return
	}

	response = map[string]interface{}{"operation": "create", "status": "success", "message": fmt.Sprintf("List %s created", list.List_name)}
	json.NewEncoder(w).Encode(response)
}

// Updating list with PATCH method
func UpdateList(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(session.UserCtxKey).(map[string]interface{})
	if !strings.Contains(claims["Rights"].(string), "w") {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	var response map[string]interface{}
	var params map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	l_id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || l_id < 0 {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
	}

	list := db.List{}
	res := db.DB.Where(db.List{ID: uint(l_id)}).First(&list)
	if res.RowsAffected == 0 {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	list.EditListForm(params)

	// Updating list
	if tx := db.DB.Save(&list); tx.Error != nil {
		http.Error(w, "Error occured while updating list", http.StatusInternalServerError)
		return
	}
	response = map[string]interface{}{"operation": "update", "status": "success", "message": fmt.Sprintf("List %s updated", list.List_name)}
	json.NewEncoder(w).Encode(response)
}

// Delete list by it's id
func DeleteList(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(session.UserCtxKey).(map[string]interface{})
	if !strings.Contains(claims["Rights"].(string), "w") {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	l_id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || l_id < 0 {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
	}

	res := db.DB.Delete(&db.List{}, l_id)
	if res.Error != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	response := map[string]interface{}{
		"operation": "delete",
		"object":    "list",
		"groupid":   l_id,
		"status":    "success",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
