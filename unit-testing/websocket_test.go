package unit_testing

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestWebSocketConnections(t *testing.T) {
	testDB := TestSetup(t)

	// Setup test users
	userIDs, err := SetupTestUsers(testDB.DB)
	AssertNoError(t, err, "Failed to setup test users")

	// Create WebSocket test helper
	wsHelper := NewWebSocketTestHelper()
	defer wsHelper.Close()

	t.Run("ConnectUser", func(t *testing.T) {
		// Create session for user
		sessionToken := CreateTestSession(t, testDB, userIDs[0])

		// Connect user to WebSocket
		conn, err := wsHelper.ConnectUser(userIDs[0], sessionToken)
		AssertNoError(t, err, "User should be able to connect to WebSocket")
		AssertNotEqual(t, conn, nil, "Connection should not be nil")
		AssertTrue(t, conn.IsConnected(), "Connection should be active")
		AssertEqual(t, conn.UserID, userIDs[0], "Connection user ID should match")
	})

	t.Run("ConnectMultipleUsers", func(t *testing.T) {
		// Connect multiple users
		for i := 0; i < 3; i++ {
			sessionToken := CreateTestSession(t, testDB, userIDs[i])
			conn, err := wsHelper.ConnectUser(userIDs[i], sessionToken)
			AssertNoError(t, err, "User should be able to connect")
			AssertTrue(t, conn.IsConnected(), "Connection should be active")
		}

		// Verify all connections exist
		for i := 0; i < 3; i++ {
			conn := wsHelper.GetConnection(userIDs[i])
			AssertNotEqual(t, conn, nil, "Connection should exist")
			AssertTrue(t, conn.IsConnected(), "Connection should be active")
		}
	})

	t.Run("ConnectWithInvalidSession", func(t *testing.T) {
		// Try to connect with invalid session
		_, err := wsHelper.ConnectUser(userIDs[0], "invalid_session")
		AssertError(t, err, "Connection should fail with invalid session")
	})

	t.Run("DisconnectUser", func(t *testing.T) {
		// Connect user
		sessionToken := CreateTestSession(t, testDB, userIDs[0])
		conn, err := wsHelper.ConnectUser(userIDs[0], sessionToken)
		AssertNoError(t, err, "User should be able to connect")
		AssertTrue(t, conn.IsConnected(), "Connection should be active")

		// Disconnect user
		wsHelper.DisconnectUser(userIDs[0])

		// Verify disconnection
		conn = wsHelper.GetConnection(userIDs[0])
		if conn != nil {
			AssertFalse(t, conn.IsConnected(), "Connection should be inactive after disconnect")
		}
	})
}

func TestWebSocketMessaging(t *testing.T) {
	testDB := TestSetup(t)

	// Setup test users
	userIDs, err := SetupTestUsers(testDB.DB)
	AssertNoError(t, err, "Failed to setup test users")

	// Create WebSocket test helper
	wsHelper := NewWebSocketTestHelper()
	defer wsHelper.Close()

	// Connect users
	sessionToken1 := CreateTestSession(t, testDB, userIDs[0])
	sessionToken2 := CreateTestSession(t, testDB, userIDs[1])

	conn1, err := wsHelper.ConnectUser(userIDs[0], sessionToken1)
	AssertNoError(t, err, "First user should connect")

	conn2, err := wsHelper.ConnectUser(userIDs[1], sessionToken2)
	AssertNoError(t, err, "Second user should connect")

	t.Run("SendDirectMessage", func(t *testing.T) {
		// Create test message
		message := map[string]interface{}{
			"type":            "message",
			"conversation_id": 1,
			"content":         "Hello from user 1",
			"sender_id":       userIDs[0],
		}

		// Send message from user 1
		err := conn1.SendMessage(message)
		AssertNoError(t, err, "Should be able to send message")

		// Wait for message to be received by user 2
		receivedMessage, err := conn2.WaitForMessage(time.Second * 2)
		AssertNoError(t, err, "User 2 should receive message")
		AssertEqual(t, receivedMessage.Type, "message", "Message type should match")

		// Parse message data
		var messageData map[string]interface{}
		err = json.Unmarshal(receivedMessage.Data, &messageData)
		AssertNoError(t, err, "Should be able to parse message data")
		AssertEqual(t, messageData["content"], "Hello from user 1", "Message content should match")
	})

	t.Run("BroadcastMessage", func(t *testing.T) {
		// Connect third user
		sessionToken3 := CreateTestSession(t, testDB, userIDs[2])
		conn3, err := wsHelper.ConnectUser(userIDs[2], sessionToken3)
		AssertNoError(t, err, "Third user should connect")

		// Clear previous messages
		conn1.ClearMessages()
		conn2.ClearMessages()
		conn3.ClearMessages()

		// Broadcast message to all users
		broadcastMessage := map[string]interface{}{
			"type":    "broadcast",
			"content": "Broadcast message to all users",
		}

		err = wsHelper.BroadcastMessage(broadcastMessage)
		AssertNoError(t, err, "Should be able to broadcast message")

		// Verify all users received the message
		for i, conn := range []*TestWebSocketConnection{conn1, conn2, conn3} {
			receivedMessage, err := conn.WaitForMessage(time.Second * 2)
			AssertNoError(t, err, fmt.Sprintf("User %d should receive broadcast message", i+1))
			AssertEqual(t, receivedMessage.Type, "broadcast", "Message type should be broadcast")
		}
	})

	t.Run("MessageToSpecificUser", func(t *testing.T) {
		// Clear messages
		conn1.ClearMessages()
		conn2.ClearMessages()

		// Send message to specific user
		specificMessage := map[string]interface{}{
			"type":      "private_message",
			"content":   "Private message",
			"recipient": userIDs[1],
		}

		err := wsHelper.SendMessage(userIDs[0], userIDs[1], specificMessage)
		AssertNoError(t, err, "Should be able to send message to specific user")

		// User 2 should receive the message
		receivedMessage, err := conn2.WaitForMessage(time.Second * 2)
		AssertNoError(t, err, "User 2 should receive private message")
		AssertEqual(t, receivedMessage.Type, "private_message", "Message type should match")

		// User 1 should not receive the message (since it was sent to user 2)
		_, err = conn1.WaitForMessage(time.Millisecond * 500)
		AssertError(t, err, "User 1 should not receive the private message")
	})
}

func TestTypingIndicators(t *testing.T) {
	testDB := TestSetup(t)

	// Setup test users
	userIDs, err := SetupTestUsers(testDB.DB)
	AssertNoError(t, err, "Failed to setup test users")

	// Create WebSocket test helper
	wsHelper := NewWebSocketTestHelper()
	defer wsHelper.Close()

	// Connect users
	sessionToken1 := CreateTestSession(t, testDB, userIDs[0])
	sessionToken2 := CreateTestSession(t, testDB, userIDs[1])

	conn1, err := wsHelper.ConnectUser(userIDs[0], sessionToken1)
	AssertNoError(t, err, "First user should connect")

	conn2, err := wsHelper.ConnectUser(userIDs[1], sessionToken2)
	AssertNoError(t, err, "Second user should connect")

	t.Run("StartTypingIndicator", func(t *testing.T) {
		// Clear messages
		conn1.ClearMessages()
		conn2.ClearMessages()

		// Simulate typing start
		err := wsHelper.SimulateTypingIndicator(userIDs[0], 1, true)
		AssertNoError(t, err, "Should be able to simulate typing start")

		// User 2 should receive typing indicator
		receivedMessage, err := conn2.WaitForMessage(time.Second * 2)
		AssertNoError(t, err, "User 2 should receive typing indicator")
		AssertEqual(t, receivedMessage.Type, "typing_indicator", "Message type should be typing_indicator")

		// Parse typing indicator data
		var typingData map[string]interface{}
		err = json.Unmarshal(receivedMessage.Data, &typingData)
		AssertNoError(t, err, "Should be able to parse typing data")
		AssertEqual(t, int(typingData["user_id"].(float64)), userIDs[0], "User ID should match")
		AssertEqual(t, typingData["is_typing"], true, "Should indicate typing started")
	})

	t.Run("StopTypingIndicator", func(t *testing.T) {
		// Clear messages
		conn1.ClearMessages()
		conn2.ClearMessages()

		// Simulate typing stop
		err := wsHelper.SimulateTypingIndicator(userIDs[0], 1, false)
		AssertNoError(t, err, "Should be able to simulate typing stop")

		// User 2 should receive typing indicator
		receivedMessage, err := conn2.WaitForMessage(time.Second * 2)
		AssertNoError(t, err, "User 2 should receive typing indicator")
		AssertEqual(t, receivedMessage.Type, "typing_indicator", "Message type should be typing_indicator")

		// Parse typing indicator data
		var typingData map[string]interface{}
		err = json.Unmarshal(receivedMessage.Data, &typingData)
		AssertNoError(t, err, "Should be able to parse typing data")
		AssertEqual(t, typingData["is_typing"], false, "Should indicate typing stopped")
	})
}

func TestOnlineStatus(t *testing.T) {
	testDB := TestSetup(t)

	// Setup test users
	userIDs, err := SetupTestUsers(testDB.DB)
	AssertNoError(t, err, "Failed to setup test users")

	// Create WebSocket test helper
	wsHelper := NewWebSocketTestHelper()
	defer wsHelper.Close()

	t.Run("UserGoesOnline", func(t *testing.T) {
		// Connect first user to monitor status changes
		sessionToken1 := CreateTestSession(t, testDB, userIDs[0])
		conn1, err := wsHelper.ConnectUser(userIDs[0], sessionToken1)
		AssertNoError(t, err, "First user should connect")

		// Clear messages
		conn1.ClearMessages()

		// Simulate user going online
		err = wsHelper.SimulateOnlineStatusChange(userIDs[1], "online")
		AssertNoError(t, err, "Should be able to simulate online status change")

		// User 1 should receive online status notification
		receivedMessage, err := conn1.WaitForMessage(time.Second * 2)
		AssertNoError(t, err, "User 1 should receive online status notification")
		AssertEqual(t, receivedMessage.Type, "online_status", "Message type should be online_status")

		// Parse status data
		var statusData map[string]interface{}
		err = json.Unmarshal(receivedMessage.Data, &statusData)
		AssertNoError(t, err, "Should be able to parse status data")
		AssertEqual(t, int(statusData["user_id"].(float64)), userIDs[1], "User ID should match")
		AssertEqual(t, statusData["status"], "online", "Status should be online")
	})

	t.Run("UserGoesOffline", func(t *testing.T) {
		// Connect user
		sessionToken2 := CreateTestSession(t, testDB, userIDs[1])
		_, err := wsHelper.ConnectUser(userIDs[1], sessionToken2)
		AssertNoError(t, err, "Second user should connect")

		// Disconnect user (simulating going offline)
		wsHelper.DisconnectUser(userIDs[1])

		// Simulate offline status change
		err = wsHelper.SimulateOnlineStatusChange(userIDs[1], "offline")
		AssertNoError(t, err, "Should be able to simulate offline status change")

		// Note: In a real implementation, other connected users would receive this notification
		// For this test, we're just verifying the simulation works without errors
	})
}

func TestWebSocketMessageTypes(t *testing.T) {
	t.Run("CreateWebSocketMessage", func(t *testing.T) {
		// Test creating different message types
		messageTypes := []string{
			"message",
			"typing_indicator",
			"online_status",
			"conversation_update",
			"user_joined",
			"user_left",
		}

		for _, msgType := range messageTypes {
			testData := map[string]interface{}{
				"test": "data",
				"type": msgType,
			}

			message := CreateTestWebSocketMessage(msgType, testData)
			AssertEqual(t, message.Type, msgType, "Message type should match")
			AssertNotEqual(t, message.Data, nil, "Message data should not be nil")

			// Verify data can be unmarshaled
			var parsedData map[string]interface{}
			err := json.Unmarshal(message.Data, &parsedData)
			AssertNoError(t, err, "Should be able to unmarshal message data")
			AssertEqual(t, parsedData["test"], "data", "Data should be preserved")
		}
	})

	t.Run("AssertWebSocketMessage", func(t *testing.T) {
		// Create test message
		testData := map[string]interface{}{
			"content": "test content",
			"user_id": 123,
		}

		message := CreateTestWebSocketMessage("test_message", testData)

		// Test assertion (should not fail)
		AssertWebSocketMessage(t, message, "test_message", testData)

		// Test with wrong type (would fail in real test)
		// AssertWebSocketMessage(t, message, "wrong_type", testData) // This would fail
	})
}

func TestWebSocketConnectionManagement(t *testing.T) {
	testDB := TestSetup(t)

	// Setup test users
	userIDs, err := SetupTestUsers(testDB.DB)
	AssertNoError(t, err, "Failed to setup test users")

	// Create WebSocket test helper
	wsHelper := NewWebSocketTestHelper()
	defer wsHelper.Close()

	t.Run("ConnectionLifecycle", func(t *testing.T) {
		// Connect user
		sessionToken := CreateTestSession(t, testDB, userIDs[0])
		conn, err := wsHelper.ConnectUser(userIDs[0], sessionToken)
		AssertNoError(t, err, "User should be able to connect")
		AssertTrue(t, conn.IsConnected(), "Connection should be active")

		// Send message to verify connection works
		testMessage := map[string]interface{}{
			"type":    "test",
			"content": "Connection test",
		}

		err = conn.SendMessage(testMessage)
		AssertNoError(t, err, "Should be able to send message")

		// Verify message was sent
		sentMessages := conn.GetSentMessages()
		AssertEqual(t, len(sentMessages), 1, "Should have one sent message")

		// Close connection
		conn.Close()
		AssertFalse(t, conn.IsConnected(), "Connection should be inactive after close")

		// Try to send message after close (should fail)
		err = conn.SendMessage(testMessage)
		AssertError(t, err, "Should not be able to send message after close")
	})

	t.Run("MessageHistory", func(t *testing.T) {
		// Connect user
		sessionToken := CreateTestSession(t, testDB, userIDs[1])
		conn, err := wsHelper.ConnectUser(userIDs[1], sessionToken)
		AssertNoError(t, err, "User should be able to connect")

		// Send multiple messages
		messages := []string{"Message 1", "Message 2", "Message 3"}
		for _, content := range messages {
			testMessage := map[string]interface{}{
				"type":    "test",
				"content": content,
			}
			err = conn.SendMessage(testMessage)
			AssertNoError(t, err, "Should be able to send message")
		}

		// Verify message history
		sentMessages := conn.GetSentMessages()
		AssertEqual(t, len(sentMessages), 3, "Should have three sent messages")

		// Clear messages
		conn.ClearMessages()
		sentMessages = conn.GetSentMessages()
		AssertEqual(t, len(sentMessages), 0, "Should have no messages after clear")
	})
}
