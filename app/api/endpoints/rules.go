package endpoints

import (
	"correlator/api/session"
	"correlator/db"
	"correlator/rule_manager"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

func ShowRules(w http.ResponseWriter, r *http.Request) {
	rules := []db.Rules{}
	res := db.DB.Find(&rules)
	if res.Error != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var data []map[string]interface{}

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
		for _, rule := range rules {
			if rule.UserID == claims["User"].(*db.User).ID {
				ru := &rule_manager.Rule{}
				if err := json.Unmarshal(rule.Rule, ru); err != nil {
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
				params := map[string]interface{}{
					"id":   rule.ID,
					"rule": ru.Rule_Name,
					"ukey": ru.UKeyValue,
				}
				data = append(data, params)
			}
		}
	default:
		if strings.Contains(rights, "r") {
			for _, rule := range rules {
				ru := &rule_manager.Rule{}
				if err := json.Unmarshal(rule.Rule, ru); err != nil {
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
				params := map[string]interface{}{
					"id":   rule.ID,
					"rule": ru.Rule_Name,
					"ukey": ru.UKeyValue,
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

func ShowSingleRule(w http.ResponseWriter, r *http.Request) {
	r_id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || r_id < 0 {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
	}

	rule := db.Rules{}
	res := db.DB.Where(db.Rules{ID: uint(r_id)}).First(&rule)
	if res.RowsAffected == 0 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
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
		if rule.UserID != claims["User"].(*db.User).ID {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
	default:
		if !strings.Contains(rights, "r") {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
	}

	var response map[string]interface{}
	json.Unmarshal(rule.Rule, &response)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func EditRule(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(session.UserCtxKey).(map[string]interface{})
	if !strings.Contains(claims["Rights"].(string), "w") {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	var response map[string]interface{}

	r_id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || r_id < 0 {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
	}

	rule := db.Rules{}
	if res := db.DB.Where(db.Rules{ID: uint(r_id)}).First(&rule); res.RowsAffected == 0 {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	NewRule := &rule_manager.Rule{}
	// Checking if rule is valid
	err = json.NewDecoder(r.Body).Decode(&NewRule)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	_, err = NewRule.ConvertMapToSteps(NewRule.Condition)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// marshaling to []byte
	jsonBytes, err := json.Marshal(NewRule)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
	rule.Rule = jsonBytes

	if tx := db.DB.Save(&rule); tx.Error != nil {
		http.Error(w, "Error occured while updating list", http.StatusInternalServerError)
		return
	}

	rule_manager.GlobalHandlerObj.RuleChan <- NewRule // First one to stop handler
	rule_manager.GlobalHandlerObj.RuleChan <- NewRule // Second one to start new handler

	response = map[string]interface{}{"operation": "update", "status": "success", "message": fmt.Sprintf("Rule %s updated", NewRule.Rule_Name)}
	json.NewEncoder(w).Encode(response)
}

func CreateRule(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(session.UserCtxKey).(map[string]interface{})
	if !strings.Contains(claims["Rights"].(string), "w") {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	var response map[string]interface{}
	rule := &rule_manager.Rule{}
	// Checking if rule is valid
	err := json.NewDecoder(r.Body).Decode(&rule)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	_, err = rule.ConvertMapToSteps(rule.Condition)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// marshaling to []byte
	jsonBytes, err := json.Marshal(rule)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	ru := db.Rules{Rule: jsonBytes, IsActive: true}

	ru.User = *claims["User"].(*db.User)

	if err := db.DB.Create(&ru).Error; err != nil {
		http.Error(w, "Error occured while creating rule", http.StatusInternalServerError)
		return
	}
	rule_manager.GlobalHandlerObj.RuleChan <- rule

	response = map[string]interface{}{"operation": "create", "status": "success", "message": fmt.Sprintf("Rule %s created", rule.Rule_Name)}
	json.NewEncoder(w).Encode(response)
}

func DeleteRule(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(session.UserCtxKey).(map[string]interface{})
	if !strings.Contains(claims["Rights"].(string), "w") {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	r_id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || r_id < 0 {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
	}

	// Finding rule and getting its name to stop it's handler
	rule := db.Rules{}
	res := db.DB.Where(db.Rules{ID: uint(r_id)}).First(&rule)
	if res.RowsAffected == 0 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	var data map[string]interface{}
	json.Unmarshal(rule.Rule, &data)
	if name, ok := data["rule"].(string); ok && name != "" {
		rule_manager.GlobalHandlerObj.RuleChan <- &rule_manager.Rule{Rule_Name: name}
	} else {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	res = db.DB.Delete(&db.Rules{}, r_id)
	if res.Error != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	response := map[string]interface{}{
		"operation": "delete",
		"object":    "rule",
		"status":    "success",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
