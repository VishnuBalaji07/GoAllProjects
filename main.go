package main

import (
	"ChatApiServer/controller"
	"ChatApiServer/database"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	// Initialize database
	database.InitDB()

	// Create router
	router := mux.NewRouter()

	// Public routes
	router.HandleFunc("/signup", controller.Signup).Methods("POST")
	router.HandleFunc("/login", controller.Login).Methods("POST")

	// Protected routes (require JWT auth)
	authRouter := router.PathPrefix("/api").Subrouter()
	authRouter.Use(controller.AuthMiddleware)

	// User-related
	authRouter.HandleFunc("/users", controller.CreateUser).Methods("POST")
	authRouter.HandleFunc("/user/chats", controller.GetUserChats).Methods("GET")

	// Chat-related
	authRouter.HandleFunc("/chats", controller.CreateChat).Methods("POST")
	authRouter.HandleFunc("/chats/{id}", controller.GetChat).Methods("GET")
	authRouter.HandleFunc("/chats/{id}", controller.UpdateChat).Methods("PUT")
	authRouter.HandleFunc("/chats/{id}", controller.DeleteChat).Methods("DELETE")
	authRouter.HandleFunc("/chats/{chat_id}/add-users", controller.AddUserToGroupChat).Methods("POST")
	authRouter.HandleFunc("/chats/{chat_id}/remove-users", controller.RemoveUserFromGroupChat).Methods("DELETE")

	// Message-related
	authRouter.HandleFunc("/messages", controller.SendMessage).Methods("POST")
	authRouter.HandleFunc("/messages/{id}", controller.GetMessage).Methods("GET")
	authRouter.HandleFunc("/messages/{id}", controller.UpdateMessage).Methods("PUT")
	authRouter.HandleFunc("/messages/{id}", controller.DeleteMessage).Methods("DELETE")
	authRouter.HandleFunc("/messages/{id}/delivered", controller.MarkDelivered).Methods("PUT")
	authRouter.HandleFunc("/messages/{id}/read", controller.MarkRead).Methods("PUT")
	authRouter.HandleFunc("/messages/private/{chat_id}", controller.GetMessagesBetweenUsers).Methods("GET")
	authRouter.HandleFunc("/chats/{chat_id}/messages", controller.GetMessagesInChat).Methods("GET")
	authRouter.HandleFunc("/chats/{chat_id}/messages/bulk", controller.SendMultipleMessages).Methods("POST")
	authRouter.HandleFunc("/chats/{chat_id}/messages/search", controller.SearchMessagesInChat).Methods("POST")

	// Reactions
	authRouter.HandleFunc("/messages/{message_id}/reactions", controller.AddOrUpdateReaction).Methods("POST")
	authRouter.HandleFunc("/messages/{message_id}/reactions", controller.RemoveReaction).Methods("DELETE")
	authRouter.HandleFunc("/messages/{message_id}/reactions", controller.GetReactions).Methods("GET")

	// Server start
	log.Println("✅ Server running at :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
