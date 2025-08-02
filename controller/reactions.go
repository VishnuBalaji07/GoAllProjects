package controller

import (
	"ChatApiServer/database"
	"ChatApiServer/models"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// AddOrUpdateReaction handles adding or updating a reaction to a specific message
func AddOrUpdateReaction(w http.ResponseWriter, r *http.Request) {
	// Safely extract user ID from context
	userIDRaw := r.Context().Value(userIDKey)
	userID, ok := userIDRaw.(uint)
	if !ok {
		http.Error(w, `{"error":"Unauthorized or missing user ID"}`, http.StatusUnauthorized)
		return
	}

	// Get and validate message ID
	messageIDStr := mux.Vars(r)["message_id"]
	messageID, err := strconv.Atoi(messageIDStr)
	if err != nil || messageID <= 0 {
		http.Error(w, `{"error":"Invalid message ID"}`, http.StatusBadRequest)
		return
	}

	// Decode JSON body for emoji
	var payload struct {
		Emoji string `json:"emoji"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.Emoji == "" {
		http.Error(w, `{"error":"Emoji is required"}`, http.StatusBadRequest)
		return
	}

	db := database.DB
	var reaction models.Reaction

	// Check if the reaction already exists
	err = db.Where("message_id = ? AND user_id = ?", messageID, userID).
		First(&reaction).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new reaction
			reaction = models.Reaction{
				MessageID: uint(messageID),
				UserID:    userID,
				Emoji:     payload.Emoji,
			}
			if err := db.Create(&reaction).Error; err != nil {
				http.Error(w, `{"error":"Failed to save reaction"}`, http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusCreated)
		} else {
			http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
			return
		}
	} else {
		// Update existing reaction
		reaction.Emoji = payload.Emoji
		if err := db.Save(&reaction).Error; err != nil {
			http.Error(w, `{"error":"Failed to update reaction"}`, http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reaction)
}

// RemoveReaction deletes the current user's reaction to a specific message
func RemoveReaction(w http.ResponseWriter, r *http.Request) {
	// Safely extract user ID from context
	userIDRaw := r.Context().Value(userIDKey)
	userID, ok := userIDRaw.(uint)
	if !ok {
		http.Error(w, `{"error":"Unauthorized or missing user ID"}`, http.StatusUnauthorized)
		return
	}

	// Parse and validate message ID
	messageIDStr := mux.Vars(r)["message_id"]
	messageID, err := strconv.Atoi(messageIDStr)
	if err != nil || messageID <= 0 {
		http.Error(w, `{"error":"Invalid message ID"}`, http.StatusBadRequest)
		return
	}

	// Delete reaction for this user and message
	if err := database.DB.Where("message_id = ? AND user_id = ?", messageID, userID).
		Delete(&models.Reaction{}).Error; err != nil {
		http.Error(w, `{"error":"Failed to delete reaction"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetReactions returns all reactions for a message
func GetReactions(w http.ResponseWriter, r *http.Request) {
	messageID, err := strconv.Atoi(mux.Vars(r)["message_id"])
	if err != nil || messageID <= 0 {
		http.Error(w, `{"error":"Invalid message ID"}`, http.StatusBadRequest)
		return
	}

	var reactions []models.Reaction
	if err := database.DB.Where("message_id = ?", messageID).Find(&reactions).Error; err != nil {
		http.Error(w, `{"error":"Failed to fetch reactions"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reactions)
}
