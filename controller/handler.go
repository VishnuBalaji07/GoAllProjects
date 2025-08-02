package controller

import (
	"ChatApiServer/database"
	"ChatApiServer/models"
	"encoding/json"
	"net/http"
)

func CreateUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	if err := database.DB.Create(&user).Error; err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
func GetUserChats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Use the same typed key as in AuthMiddleware
	userIDRaw := r.Context().Value(userIDKey)
	userID, ok := userIDRaw.(uint)
	if !ok {
		http.Error(w, `{"error":"Unauthorized or missing user ID"}`, http.StatusUnauthorized)
		return
	}

	var chatMembers []models.ChatMember
	if err := database.DB.Where("user_id = ?", userID).Find(&chatMembers).Error; err != nil {
		http.Error(w, `{"error":"Failed to fetch chat memberships"}`, http.StatusInternalServerError)
		return
	}

	var chatIDs []uint
	for _, member := range chatMembers {
		chatIDs = append(chatIDs, member.ChatID)
	}

	var chats []models.Chat
	if err := database.DB.Preload("Members").Preload("Messages").
		Where("id IN ?", chatIDs).
		Find(&chats).Error; err != nil {
		http.Error(w, `{"error":"Failed to fetch chats"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id": userID,
		"chats":   chats,
	})
}
