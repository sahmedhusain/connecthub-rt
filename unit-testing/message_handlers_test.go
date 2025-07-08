package unit_testing

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"forum/database"
	"forum/server"
	"forum/websocket"
)

func TestSendMessage(t *testing.T) {
	testDB := TestSetup(t)
	defer testDB.Cleanup()

	// Setup test data
	userIDs, err := SetupTestUsers(testDB.DB)
	AssertNoError(t, err, "Failed to setup test users")

	conversationIDs, err := SetupTestConversations(testDB.DB, userIDs)
	AssertNoError(t, err, "Failed to setup test conversations")

	sessionToken := CreateTestSession(t, testDB, userIDs[0])

	// Setup WebSocket manager for testing
	wsManager := websocket.NewManager()
	server.SetWebSocketManager(wsManager)

	t.Run("ValidMessageSend", func(t *testing.T) {
		sendReq := server.SendMessageRequest{
			ConversationID: conversationIDs[0],
			Content:        "This is a test message",
		}

		body, _ := json.Marshal(sendReq)
		req := httptest.NewRequest("POST", "/api/send-message", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: sessionToken,
		})

		w := httptest.NewRecorder()
		server.SendMessageAPI(w, req)

		AssertEqual(t, w.Code, http.StatusOK, "Expected status OK")

		var response server.SendMessageResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		AssertNoError(t, err, "Failed to unmarshal response")

		AssertEqual(t, response.Success, true, "Message send should succeed")
		AssertNotEqual(t, response.Message, nil, "Message should be present in response")
	})

	t.Run("MessageSendWithoutSession", func(t *testing.T) {
		sendReq := server.SendMessageRequest{
			ConversationID: conversationIDs[0],
			Content:        "This is a test message",
		}

		body, _ := json.Marshal(sendReq)
		req := httptest.NewRequest("POST", "/api/send-message", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		server.SendMessageAPI(w, req)

		AssertEqual(t, w.Code, http.StatusUnauthorized, "Expected status Unauthorized")
	})

	t.Run("MessageSendWithEmptyContent", func(t *testing.T) {
		sendReq := server.SendMessageRequest{
			ConversationID: conversationIDs[0],
			Content:        "", // Empty content
		}

		body, _ := json.Marshal(sendReq)
		req := httptest.NewRequest("POST", "/api/send-message", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: sessionToken,
		})

		w := httptest.NewRecorder()
		server.SendMessageAPI(w, req)

		AssertEqual(t, w.Code, http.StatusBadRequest, "Expected status Bad Request")

		var response server.SendMessageResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		AssertNoError(t, err, "Failed to unmarshal response")

		AssertEqual(t, response.Success, false, "Message send should fail")
		AssertNotEqual(t, response.Error, "", "Error message should be present")
	})

	t.Run("MessageSendWithInvalidConversation", func(t *testing.T) {
		sendReq := server.SendMessageRequest{
			ConversationID: 99999, // Non-existent conversation
			Content:        "This is a test message",
		}

		body, _ := json.Marshal(sendReq)
		req := httptest.NewRequest("POST", "/api/send-message", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: sessionToken,
		})

		w := httptest.NewRecorder()
		server.SendMessageAPI(w, req)

		AssertEqual(t, w.Code, http.StatusBadRequest, "Expected status Bad Request")

		var response server.SendMessageResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		AssertNoError(t, err, "Failed to unmarshal response")

		AssertEqual(t, response.Success, false, "Message send should fail")
		AssertNotEqual(t, response.Error, "", "Error message should be present")
	})

	t.Run("MessageSendWithInvalidJSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/send-message", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: sessionToken,
		})

		w := httptest.NewRecorder()
		server.SendMessageAPI(w, req)

		AssertEqual(t, w.Code, http.StatusBadRequest, "Expected status Bad Request")
	})
}

func TestGetMessages(t *testing.T) {
	testDB := TestSetup(t)
	defer testDB.Cleanup()

	// Setup test data
	userIDs, err := SetupTestUsers(testDB.DB)
	AssertNoError(t, err, "Failed to setup test users")

	conversationIDs, err := SetupTestConversations(testDB.DB, userIDs)
	AssertNoError(t, err, "Failed to setup test conversations")

	_, err = SetupTestMessages(testDB.DB, conversationIDs, userIDs)
	AssertNoError(t, err, "Failed to setup test messages")

	sessionToken := CreateTestSession(t, testDB, userIDs[0])

	t.Run("ValidGetMessages", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/messages?conversation_id="+strconv.Itoa(conversationIDs[0]), nil)
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: sessionToken,
		})

		w := httptest.NewRecorder()
		server.GetMessages(w, req)

		AssertEqual(t, w.Code, http.StatusOK, "Expected status OK")

		var messages []database.Message
		err := json.Unmarshal(w.Body.Bytes(), &messages)
		AssertNoError(t, err, "Failed to unmarshal messages")

		AssertGreaterThan(t, len(messages), 0, "Should return messages")

		// Verify messages belong to the correct conversation
		for _, message := range messages {
			AssertEqual(t, message.ConversationID, conversationIDs[0], "Message should belong to correct conversation")
		}
	})

	t.Run("GetMessagesWithoutSession", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/messages?conversation_id="+strconv.Itoa(conversationIDs[0]), nil)

		w := httptest.NewRecorder()
		server.GetMessages(w, req)

		AssertEqual(t, w.Code, http.StatusUnauthorized, "Expected status Unauthorized")
	})

	t.Run("GetMessagesWithMissingConversationID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/messages", nil)
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: sessionToken,
		})

		w := httptest.NewRecorder()
		server.GetMessages(w, req)

		AssertEqual(t, w.Code, http.StatusBadRequest, "Expected status Bad Request")
	})

	t.Run("GetMessagesWithInvalidConversationID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/messages?conversation_id=invalid", nil)
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: sessionToken,
		})

		w := httptest.NewRecorder()
		server.GetMessages(w, req)

		AssertEqual(t, w.Code, http.StatusBadRequest, "Expected status Bad Request")
	})

	t.Run("GetMessagesWithNonExistentConversation", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/messages?conversation_id=99999", nil)
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: sessionToken,
		})

		w := httptest.NewRecorder()
		server.GetMessages(w, req)

		AssertEqual(t, w.Code, http.StatusOK, "Expected status OK")

		var messages []database.Message
		err := json.Unmarshal(w.Body.Bytes(), &messages)
		AssertNoError(t, err, "Failed to unmarshal messages")

		AssertEqual(t, len(messages), 0, "Should return empty array for non-existent conversation")
	})

	t.Run("GetMessagesWithPagination", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/messages?conversation_id="+strconv.Itoa(conversationIDs[0])+"&limit=5&offset=0", nil)
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: sessionToken,
		})

		w := httptest.NewRecorder()
		server.GetMessages(w, req)

		AssertEqual(t, w.Code, http.StatusOK, "Expected status OK")

		var messages []database.Message
		err := json.Unmarshal(w.Body.Bytes(), &messages)
		AssertNoError(t, err, "Failed to unmarshal messages")

		AssertLessThanOrEqual(t, len(messages), 5, "Should respect limit parameter")
	})
}

func TestGetConversations(t *testing.T) {
	testDB := TestSetup(t)
	defer testDB.Cleanup()

	// Setup test data
	userIDs, err := SetupTestUsers(testDB.DB)
	AssertNoError(t, err, "Failed to setup test users")

	_, err = SetupTestConversations(testDB.DB, userIDs)
	AssertNoError(t, err, "Failed to setup test conversations")

	sessionToken := CreateTestSession(t, testDB, userIDs[0])

	t.Run("ValidGetConversations", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/conversations", nil)
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: sessionToken,
		})

		w := httptest.NewRecorder()
		server.GetConversations(w, req)

		AssertEqual(t, w.Code, http.StatusOK, "Expected status OK")

		var conversations []database.Conversation
		err := json.Unmarshal(w.Body.Bytes(), &conversations)
		AssertNoError(t, err, "Failed to unmarshal conversations")

		AssertGreaterThan(t, len(conversations), 0, "Should return conversations")

		// Verify conversations contain the user
		for _, conversation := range conversations {
			AssertNotEqual(t, conversation.ID, 0, "Conversation ID should be set")
			AssertGreaterThan(t, len(conversation.Participants), 0, "Conversation should have participants")

			// User should be one of the participants
			userInConversation := false
			for _, participant := range conversation.Participants {
				if participant.ID == userIDs[0] {
					userInConversation = true
					break
				}
			}
			AssertEqual(t, userInConversation, true, "User should be part of the conversation")
		}
	})

	t.Run("GetConversationsWithoutSession", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/conversations", nil)

		w := httptest.NewRecorder()
		server.GetConversations(w, req)

		AssertEqual(t, w.Code, http.StatusUnauthorized, "Expected status Unauthorized")
	})

	t.Run("GetConversationsWithInvalidSession", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/conversations", nil)
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: "invalid_session_token",
		})

		w := httptest.NewRecorder()
		server.GetConversations(w, req)

		AssertEqual(t, w.Code, http.StatusUnauthorized, "Expected status Unauthorized")
	})
}

func TestCreateConversation(t *testing.T) {
	testDB := TestSetup(t)
	defer testDB.Cleanup()

	// Setup test data
	userIDs, err := SetupTestUsers(testDB.DB)
	AssertNoError(t, err, "Failed to setup test users")

	sessionToken := CreateTestSession(t, testDB, userIDs[0])

	t.Run("ValidConversationCreation", func(t *testing.T) {
		createReq := server.CreateConversationRequest{
			Participants: []int{userIDs[1]}, // Create conversation with second user
		}

		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest("POST", "/api/create-conversation", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: sessionToken,
		})

		w := httptest.NewRecorder()
		server.CreateConversationAPI(w, req)

		AssertEqual(t, w.Code, http.StatusOK, "Expected status OK")

		var response server.CreateConversationResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		AssertNoError(t, err, "Failed to unmarshal response")

		AssertEqual(t, response.Success, true, "Conversation creation should succeed")
		AssertNotEqual(t, response.ConversationID, 0, "Conversation ID should be set")
	})

	t.Run("ConversationCreationWithoutSession", func(t *testing.T) {
		createReq := server.CreateConversationRequest{
			Participants: []int{userIDs[1]},
		}

		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest("POST", "/api/create-conversation", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		server.CreateConversationAPI(w, req)

		AssertEqual(t, w.Code, http.StatusUnauthorized, "Expected status Unauthorized")
	})

	t.Run("ConversationCreationWithSameUser", func(t *testing.T) {
		createReq := server.CreateConversationRequest{
			Participants: []int{userIDs[0]}, // Same user as session
		}

		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest("POST", "/api/create-conversation", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: sessionToken,
		})

		w := httptest.NewRecorder()
		server.CreateConversationAPI(w, req)

		AssertEqual(t, w.Code, http.StatusBadRequest, "Expected status Bad Request")

		var response server.CreateConversationResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		AssertNoError(t, err, "Failed to unmarshal response")

		AssertEqual(t, response.Success, false, "Conversation creation should fail")
		AssertNotEqual(t, response.Error, "", "Error message should be present")
	})

	t.Run("ConversationCreationWithNonExistentUser", func(t *testing.T) {
		createReq := server.CreateConversationRequest{
			Participants: []int{99999}, // Non-existent user
		}

		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest("POST", "/api/create-conversation", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: sessionToken,
		})

		w := httptest.NewRecorder()
		server.CreateConversationAPI(w, req)

		AssertEqual(t, w.Code, http.StatusBadRequest, "Expected status Bad Request")

		var response server.CreateConversationResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		AssertNoError(t, err, "Failed to unmarshal response")

		AssertEqual(t, response.Success, false, "Conversation creation should fail")
		AssertNotEqual(t, response.Error, "", "Error message should be present")
	})
}
