package controller

import (
	"ChatApiServer/database"
	"ChatApiServer/models"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// CreateChat creates a new chat (group or one-on-one)
func CreateChat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Struct to safely decode incoming payload
	type ChatInput struct {
		Name        string              `json:"name"`
		Description string              `json:"description"`
		IsGroup     bool                `json:"is_group"`
		Members     []models.ChatMember `json:"members"`
	}

	var payload ChatInput
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, `{"error":"Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	// Get the creator's user ID from context
	userIDAny := r.Context().Value(userIDKey)
	userID, ok := userIDAny.(uint)
	if !ok {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Check for existing chat name (case-insensitive)
	var existingChat models.Chat
	if err := database.DB.Where("LOWER(name) = LOWER(?)", payload.Name).First(&existingChat).Error; err == nil {
		http.Error(w, `{"error":"Chat with this name already exists"}`, http.StatusConflict)
		return
	}

	// Create chat object (excluding members for now)
	chat := models.Chat{
		Name:        payload.Name,
		Description: payload.Description,
		IsGroup:     payload.IsGroup,
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
	}

	if err := database.DB.Create(&chat).Error; err != nil {
		http.Error(w, `{"error":"Failed to create chat"}`, http.StatusInternalServerError)
		return
	}

	// Deduplicate and prepare members
	uniqueMembers := make(map[uint]bool)
	var filteredMembers []models.ChatMember

	for _, m := range payload.Members {
		if !uniqueMembers[m.UserID] {
			filteredMembers = append(filteredMembers, models.ChatMember{
				ChatID:   chat.ID,
				UserID:   m.UserID,
				AddedBy:  &userID,
				JoinedAt: time.Now(),
				Role:     m.Role,
			})
			uniqueMembers[m.UserID] = true
		}
	}

	// Add creator if not already included
	if !uniqueMembers[userID] {
		filteredMembers = append(filteredMembers, models.ChatMember{
			ChatID:   chat.ID,
			UserID:   userID,
			AddedBy:  &userID,
			JoinedAt: time.Now(),
			Role:     "admin",
		})
	}

	// Save chat members
	if err := database.DB.Create(&filteredMembers).Error; err != nil {
		database.DB.Delete(&chat) // rollback chat
		http.Error(w, `{"error":"Failed to add chat members"}`, http.StatusInternalServerError)
		return
	}

	// Attach members to chat for response
	chat.Members = filteredMembers

	// Return success response
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Chat created successfully",
		"chat":    chat,
	})
}

// GetChat fetches chat details by ID
func GetChat(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error":"Invalid chat ID"}`, http.StatusBadRequest)
		return
	}

	// Fetch only necessary fields from Chat
	var chat models.Chat
	if err := database.DB.Select("id", "name", "description", "is_group", "last_message", "last_updated_at").
		First(&chat, id).Error; err != nil {
		http.Error(w, `{"error":"Chat not found"}`, http.StatusNotFound)
		return
	}

	// Respond with only required fields
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":              chat.ID,
		"name":            chat.Name,
		"description":     chat.Description,
		"is_group":        chat.IsGroup,
		"last_message":    chat.LastMessage,
		"last_updated_at": chat.LastUpdatedAt,
	})
}

// / DeleteChat permanently deletes chat and all related data (messages, members)
func DeleteChat(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid chat ID"}`, http.StatusBadRequest)
		return
	}

	var chat models.Chat
	// Include soft-deleted chats in the query
	if err := database.DB.Unscoped().First(&chat, id).Error; err != nil {
		http.Error(w, `{"error":"Chat not found"}`, http.StatusNotFound)
		return
	}

	// Use a transaction for safety
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		// Delete messages
		if err := tx.Where("chat_id = ?", chat.ID).Delete(&models.Message{}).Error; err != nil {
			return fmt.Errorf("failed to delete messages: %v", err)
		}

		// Delete members
		if err := tx.Where("chat_id = ?", chat.ID).Delete(&models.ChatMember{}).Error; err != nil {
			return fmt.Errorf("failed to remove chat members: %v", err)
		}

		// Permanently delete the chat
		if err := tx.Unscoped().Delete(&chat).Error; err != nil {
			return fmt.Errorf("failed to delete chat: %v", err)
		}

		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Chat and related data deleted permanently",
	})
}

// UpdateChat updates chat info like name or description
func UpdateChat(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid chat ID"}`, http.StatusBadRequest)
		return
	}

	var chat models.Chat
	if err := database.DB.First(&chat, id).Error; err != nil {
		http.Error(w, `{"error":"Chat not found"}`, http.StatusNotFound)
		return
	}

	var updatedData struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updatedData); err != nil {
		http.Error(w, `{"error":"Invalid JSON body"}`, http.StatusBadRequest)
		return
	}

	// Input validation
	if updatedData.Name == "" && updatedData.Description == "" {
		http.Error(w, `{"error":"Name or Description must be provided"}`, http.StatusBadRequest)
		return
	}

	// Update only if data is provided
	if updatedData.Name != "" {
		chat.Name = updatedData.Name
	}
	if updatedData.Description != "" {
		chat.Description = updatedData.Description
	}

	// Save the updated chat
	if err := database.DB.Save(&chat).Error; err != nil {
		http.Error(w, `{"error":"Failed to update chat"}`, http.StatusInternalServerError)
		return
	}

	// Send success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Chat updated successfully",
		"chat":    chat,
	})
}

// AddUserToGroupChat adds members to a group chat
func AddUserToGroupChat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	chatIDStr := mux.Vars(r)["chat_id"]
	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil {
		http.Error(w, `{"error":"Invalid chat ID"}`, http.StatusBadRequest)
		return
	}

	// Input includes user_ids, role, added_by
	var input struct {
		UserIDs []uint `json:"user_ids"`
		Role    string `json:"role"`     // optional, default to "member"
		AddedBy uint   `json:"added_by"` // who adds these users
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error":"Invalid input data"}`, http.StatusBadRequest)
		return
	}

	if len(input.UserIDs) == 0 {
		http.Error(w, `{"error":"No user IDs provided"}`, http.StatusBadRequest)
		return
	}

	if input.Role == "" {
		input.Role = "member"
	}

	// Check chat exists and is group
	var chat models.Chat
	if err := database.DB.First(&chat, chatID).Error; err != nil {
		http.Error(w, `{"error":"Chat not found"}`, http.StatusNotFound)
		return
	}
	if !chat.IsGroup {
		http.Error(w, `{"error":"Cannot add users to a private chat"}`, http.StatusBadRequest)
		return
	}

	// Add members
	for _, userID := range input.UserIDs {
		var existing models.ChatMember
		err := database.DB.Where("chat_id = ? AND user_id = ?", chatID, userID).First(&existing).Error
		if err == nil {
			// Already member, skip
			continue
		}

		member := models.ChatMember{
			ChatID:  uint(chatID),
			UserID:  userID,
			AddedBy: &input.AddedBy,
			Role:    input.Role,
		}
		if err := database.DB.Create(&member).Error; err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"Failed to add user %d: %v"}`, userID, err), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Users added successfully",
	})
}

func RemoveUserFromGroupChat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse chat ID from the URL
	chatID, err := strconv.Atoi(mux.Vars(r)["chat_id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid chat ID"}`, http.StatusBadRequest)
		return
	}

	// Find the chat
	var chat models.Chat
	if err := database.DB.First(&chat, chatID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, `{"error":"Chat not found"}`, http.StatusNotFound)
		} else {
			http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		}
		return
	}

	// Ensure it's a group chat
	if !chat.IsGroup {
		http.Error(w, `{"error":"Cannot remove users from a private chat"}`, http.StatusForbidden)
		return
	}

	// Parse user IDs to remove
	var payload struct {
		UserIDs []uint `json:"user_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || len(payload.UserIDs) == 0 {
		http.Error(w, `{"error":"Invalid or empty user_ids list"}`, http.StatusBadRequest)
		return
	}

	// Remove each user
	for _, userID := range payload.UserIDs {
		if err := database.DB.
			Where("chat_id = ? AND user_id = ?", chat.ID, userID).
			Delete(&models.ChatMember{}).Error; err != nil {
			http.Error(w, `{"error":"Failed to remove some users"}`, http.StatusInternalServerError)
			return
		}
	}

	// Success
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Users removed from group chat",
	})
}
