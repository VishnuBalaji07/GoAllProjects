package controller

import (
	"ChatApiServer/database"
	"ChatApiServer/models"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// ✅ Utility function to update chat metadata
func updateChatMetadata(chatID uint) {
	var lastMsg models.Message

	err := database.DB.
		Where("chat_id = ?", chatID).
		Order("created_at DESC").
		First(&lastMsg).Error

	if err == nil {
		database.DB.Model(&models.Chat{}).
			Where("id = ?", chatID).
			Updates(map[string]interface{}{
				"last_message":    lastMsg.Text,
				"last_updated_at": lastMsg.CreatedAt,
			})
	} else {
		// No messages found, clear metadata
		database.DB.Model(&models.Chat{}).
			Where("id = ?", chatID).
			Updates(map[string]interface{}{
				"last_message":    nil,
				"last_updated_at": nil,
			})
	}
}
func SendMessage(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ChatID uint   `json:"chat_id"`
		Text   string `json:"text"`
		Type   string `json:"type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Get user ID from context
	userIDAny := r.Context().Value(userIDKey)
	userID, ok := userIDAny.(uint)
	if !ok {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Fetch chat with members
	var chat models.Chat
	if err := database.DB.Preload("Members").First(&chat, input.ChatID).Error; err != nil {
		http.Error(w, "Chat not found", http.StatusNotFound)
		return
	}

	// Ensure sender is a member
	isMember := false
	for _, m := range chat.Members {
		if m.UserID == userID {
			isMember = true
			break
		}
	}
	if !isMember {
		http.Error(w, "Sender is not a member of this chat", http.StatusForbidden)
		return
	}

	// Create the message
	now := time.Now()
	msg := models.Message{
		ChatID:    input.ChatID,
		SenderID:  userID,
		Text:      input.Text,
		Type:      input.Type,
		CreatedAt: now,
	}

	// Save the message
	if err := database.DB.Create(&msg).Error; err != nil {
		http.Error(w, "Failed to save message", http.StatusInternalServerError)
		return
	}

	// Create status records for all other members
	var statuses []models.MessageStatus
	for _, m := range chat.Members {
		if m.UserID != userID {
			statuses = append(statuses, models.MessageStatus{
				MessageID:    msg.ID,
				UserID:       m.UserID,
				ChatMemberID: m.ID,
				Status:       "sent",
				SentAt:       &now,
			})
		}
	}
	if len(statuses) > 0 {
		if err := database.DB.Create(&statuses).Error; err != nil {
			http.Error(w, "Failed to insert message statuses", http.StatusInternalServerError)
			return
		}
	}

	// Call metadata update function
	updateChatMetadata(chat.ID)

	// Return enriched message
	var fullMsg models.Message
	if err := database.DB.
		Preload("Sender").
		Preload("StatusTrack").
		Preload("Reactions").
		First(&fullMsg, msg.ID).Error; err != nil {
		http.Error(w, "Failed to fetch message", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(fullMsg)
}

func GetMessage(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "Invalid message ID", http.StatusBadRequest)
		return
	}

	var msg models.Message
	if err := database.DB.
		Preload("ReplyTo"). // If you use ReplyTo
		Preload("Reactions").
		Preload("StatusTrack").
		First(&msg, id).Error; err != nil {
		http.Error(w, "Message not found", http.StatusNotFound)
		return
	}

	// Explicitly remove sender to omit it from JSON output

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(msg)
}

func GetMessagesBetweenUsers(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uint)
	receiverID, _ := strconv.Atoi(mux.Vars(r)["chat_id"])

	var chats []models.Chat
	err := database.DB.
		Joins("JOIN chat_members cm1 ON cm1.chat_id = chats.id").
		Joins("JOIN chat_members cm2 ON cm2.chat_id = chats.id").
		Where("cm1.user_id = ? AND cm2.user_id = ?", userID, receiverID).
		Where("is_group = ?", false).
		Group("chats.id").
		Find(&chats).Error

	if err != nil || len(chats) == 0 {
		http.Error(w, "No private chat found between users", http.StatusNotFound)
		return
	}

	chatID := chats[0].ID

	var messages []models.Message
	if err := database.DB.
		Preload("Sender").
		Where("chat_id = ?", chatID).
		Order("created_at ASC").
		Find(&messages).Error; err != nil {
		http.Error(w, "Failed to fetch messages", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}
func MarkDelivered(w http.ResponseWriter, r *http.Request) {
	messageID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid message ID"}`, http.StatusBadRequest)
		return
	}

	now := time.Now()
	result := database.DB.Model(&models.MessageStatus{}).
		Where("message_id = ?", messageID).
		Updates(map[string]interface{}{
			"delivered_at": now,
			"status":       "delivered",
		})

	if result.Error != nil {
		http.Error(w, `{"error":"Failed to update delivery status"}`, http.StatusInternalServerError)
		return
	}
	if result.RowsAffected == 0 {
		http.Error(w, `{"error":"No delivery status found for this message"}`, http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Marked as delivered",
	})
}

func MarkRead(w http.ResponseWriter, r *http.Request) {
	messageID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid message ID"}`, http.StatusBadRequest)
		return
	}

	now := time.Now()
	result := database.DB.Model(&models.MessageStatus{}).
		Where("message_id = ?", messageID).
		Updates(map[string]interface{}{
			"read_at": now,
			"status":  "read",
		})

	if result.Error != nil {
		http.Error(w, `{"error":"Failed to update read status"}`, http.StatusInternalServerError)
		return
	}
	if result.RowsAffected == 0 {
		http.Error(w, `{"error":"No read status found for this message"}`, http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Marked as read",
	})
}

func DeleteMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error":"Invalid message ID"}`, http.StatusBadRequest)
		return
	}

	var msg models.Message
	if err := database.DB.First(&msg, id).Error; err != nil {
		http.Error(w, `{"error":"Message not found"}`, http.StatusNotFound)
		return
	}

	if err := database.DB.Delete(&msg).Error; err != nil {
		http.Error(w, `{"error":"Failed to delete message"}`, http.StatusInternalServerError)
		return
	}

	// Call your chat metadata update function after deletion
	updateChatMetadata(msg.ChatID)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Message deleted successfully",
	})
}

func UpdateMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get message ID from URL
	msgIDStr := mux.Vars(r)["id"]
	msgID, err := strconv.Atoi(msgIDStr)
	if err != nil {
		http.Error(w, `{"error":"Invalid message ID"}`, http.StatusBadRequest)
		return
	}

	// Parse new text from request body
	var input struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Text == "" {
		http.Error(w, `{"error":"Invalid or missing text field"}`, http.StatusBadRequest)
		return
	}

	// Find the message
	var msg models.Message
	if err := database.DB.First(&msg, msgID).Error; err != nil {
		http.Error(w, `{"error":"Message not found"}`, http.StatusNotFound)
		return
	}

	// Update and save
	msg.Text = input.Text
	if err := database.DB.Save(&msg).Error; err != nil {
		http.Error(w, `{"error":"Failed to update message"}`, http.StatusInternalServerError)
		return
	}

	// Respond with success
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"message":      "Message updated successfully",
		"message_id":   msg.ID,
		"updated_text": msg.Text,
	})
}

func GetMessagesInChat(w http.ResponseWriter, r *http.Request) {
	chatIDStr := mux.Vars(r)["chat_id"]
	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil || chatID <= 0 {
		http.Error(w, "Invalid chat ID", http.StatusBadRequest)
		return
	}

	pageStr := r.URL.Query().Get("page")
	page := 1
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}

	limit := 100
	offset := (page - 1) * limit

	var messages []models.Message

	if err := database.DB.
		Preload("Sender").
		Preload("StatusTrack").
		Preload("Reactions").
		Where("chat_id = ?", chatID).
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error; err != nil {
		http.Error(w, "Failed to load messages", http.StatusInternalServerError)
		return
	}

	var total int64
	database.DB.Model(&models.Message{}).Where("chat_id = ?", chatID).Count(&total)

	resp := map[string]interface{}{
		"page":           page,
		"limit":          limit,
		"total_messages": total,
		"messages":       messages,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func SendMultipleMessages(w http.ResponseWriter, r *http.Request) {
	chatIDStr := mux.Vars(r)["chat_id"]
	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil || chatID <= 0 {
		http.Error(w, "Invalid chat ID", http.StatusBadRequest)
		return
	}

	userIDAny := r.Context().Value(userIDKey)
	userID, ok := userIDAny.(uint)
	if !ok {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Parse incoming messages
	var inputMsgs []struct {
		Text string `json:"text"`
		Type string `json:"type"` // e.g., "text", "image"
	}
	if err := json.NewDecoder(r.Body).Decode(&inputMsgs); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	var chat models.Chat
	if err := database.DB.Preload("Members").First(&chat, chatID).Error; err != nil {
		http.Error(w, "Chat not found", http.StatusNotFound)
		return
	}

	// Ensure sender is a member
	isMember := false
	for _, member := range chat.Members {
		if member.UserID == userID {
			isMember = true
			break
		}
	}
	if !isMember {
		http.Error(w, "Sender is not a member of this chat", http.StatusForbidden)
		return
	}

	var messages []models.Message
	now := time.Now()
	for _, im := range inputMsgs {
		messages = append(messages, models.Message{
			ChatID:    uint(chatID),
			SenderID:  userID,
			Text:      im.Text,
			Type:      im.Type,
			CreatedAt: now,
		})
	}

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		// Save all messages
		if err := tx.Create(&messages).Error; err != nil {
			return err
		}

		// Prepare and insert status for each message & chat member (except sender)
		var statuses []models.MessageStatus
		for _, msg := range messages {
			for _, member := range chat.Members {
				if member.UserID != userID {
					statuses = append(statuses, models.MessageStatus{
						MessageID:    msg.ID,
						UserID:       member.UserID,
						ChatMemberID: member.ID,
						Status:       "sent",
						SentAt:       &now,
					})
				}
			}
		}

		if len(statuses) > 0 {
			if err := tx.Create(&statuses).Error; err != nil {
				return err
			}
		}

		// Update chat with latest message
		lastMsg := messages[len(messages)-1]
		if err := tx.Model(&models.Chat{}).Where("id = ?", chatID).
			Updates(map[string]interface{}{
				"last_message":    lastMsg.Text,
				"last_updated_at": lastMsg.CreatedAt,
			}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		http.Error(w, "Failed to send messages", http.StatusInternalServerError)
		return
	}

	// Return enriched message objects
	var fullMessages []models.Message
	messageIDs := make([]uint, len(messages))
	for i, m := range messages {
		messageIDs[i] = m.ID
	}

	if err := database.DB.
		Preload("Sender").
		Preload("StatusTrack").
		Preload("Reactions").
		Where("id IN ?", messageIDs).
		Find(&fullMessages).Error; err != nil {
		http.Error(w, "Failed to load messages", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(fullMessages)
}
func SearchMessagesInChat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	chatIDStr := mux.Vars(r)["chat_id"]
	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil || chatID <= 0 {
		http.Error(w, `{"error":"Invalid chat ID"}`, http.StatusBadRequest)
		return
	}

	var input struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Text == "" {
		http.Error(w, `{"error":"Invalid or missing 'text' in request body"}`, http.StatusBadRequest)
		return
	}

	// Get chat name
	var chat models.Chat
	if err := database.DB.Select("id", "name").First(&chat, chatID).Error; err != nil {
		http.Error(w, `{"error":"Chat not found"}`, http.StatusNotFound)
		return
	}

	// Search messages using LIKE (case-insensitive with LOWER)
	var texts []string
	if err := database.DB.
		Model(&models.Message{}).
		Where("chat_id = ? AND LOWER(text) LIKE ?", chatID, "%"+strings.ToLower(input.Text)+"%").
		Order("created_at DESC").
		Pluck("text", &texts).Error; err != nil {
		http.Error(w, `{"error":"Failed to search messages"}`, http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"chat_name": chat.Name,
		"messages":  texts,
	}
	json.NewEncoder(w).Encode(resp)
}
