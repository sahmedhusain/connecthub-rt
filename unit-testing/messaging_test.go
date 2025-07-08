package unit_testing

import (
	"testing"

	"forum/repository"
	"forum/server/services"
)

func TestConversationCreation(t *testing.T) {
	testDB := TestSetup(t)

	// Setup test users
	userIDs, err := SetupTestUsers(testDB.DB)
	AssertNoError(t, err, "Failed to setup test users")

	messageRepo := repository.NewMessageRepository(testDB.DB)
	_ = services.NewMessageService(testDB.DB) // messageService not used in this test

	t.Run("CreateValidConversation", func(t *testing.T) {
		// Create conversation between two users
		participants := []int{userIDs[0], userIDs[1]}
		conversationID, err := messageRepo.CreateConversation(participants)
		AssertNoError(t, err, "Conversation creation should succeed")
		AssertTrue(t, conversationID > 0, "Conversation ID should be positive")

		// Verify participants were added
		conversationParticipants, err := messageRepo.GetConversationParticipants(conversationID)
		AssertNoError(t, err, "Should be able to get conversation participants")
		AssertEqual(t, len(conversationParticipants), 2, "Should have 2 participants")

		// Verify both users are participants
		participantIDs := make(map[int]bool)
		for _, participant := range conversationParticipants {
			participantIDs[participant.ID] = true
		}
		AssertTrue(t, participantIDs[userIDs[0]], "First user should be participant")
		AssertTrue(t, participantIDs[userIDs[1]], "Second user should be participant")
	})

	t.Run("CreateGroupConversation", func(t *testing.T) {
		// Create conversation with multiple users
		participants := []int{userIDs[0], userIDs[1], userIDs[2], userIDs[3]}
		conversationID, err := messageRepo.CreateConversation(participants)
		AssertNoError(t, err, "Group conversation creation should succeed")

		// Verify all participants were added
		conversationParticipants, err := messageRepo.GetConversationParticipants(conversationID)
		AssertNoError(t, err, "Should be able to get conversation participants")
		AssertEqual(t, len(conversationParticipants), 4, "Should have 4 participants")
	})

	t.Run("CreateConversationWithDuplicateParticipants", func(t *testing.T) {
		// Try to create conversation with duplicate participants
		participants := []int{userIDs[0], userIDs[1], userIDs[0]}
		conversationID, err := messageRepo.CreateConversation(participants)

		// This should either succeed with deduplicated participants or fail gracefully
		if err == nil {
			// If it succeeds, verify participants are deduplicated
			conversationParticipants, err := messageRepo.GetConversationParticipants(conversationID)
			AssertNoError(t, err, "Should be able to get conversation participants")
			AssertEqual(t, len(conversationParticipants), 2, "Should have 2 unique participants")
		}
		// If it fails, that's also acceptable behavior
	})

	t.Run("CreateConversationWithInvalidUser", func(t *testing.T) {
		// Try to create conversation with invalid user ID
		participants := []int{userIDs[0], 99999}
		_, err := messageRepo.CreateConversation(participants)
		AssertError(t, err, "Conversation creation should fail with invalid user")
	})

	t.Run("CreateConversationWithEmptyParticipants", func(t *testing.T) {
		// Try to create conversation with no participants
		participants := []int{}
		_, err := messageRepo.CreateConversation(participants)
		AssertError(t, err, "Conversation creation should fail with no participants")
	})
}

func TestMessageSending(t *testing.T) {
	testDB := TestSetup(t)

	// Setup test users
	userIDs, err := SetupTestUsers(testDB.DB)
	AssertNoError(t, err, "Failed to setup test users")

	messageRepo := repository.NewMessageRepository(testDB.DB)
	messageService := services.NewMessageService(testDB.DB)

	// Create a conversation for testing
	participants := []int{userIDs[0], userIDs[1]}
	conversationID, err := messageRepo.CreateConversation(participants)
	AssertNoError(t, err, "Conversation creation should succeed")

	t.Run("SendValidMessage", func(t *testing.T) {
		// Send a message
		message, err := messageService.SendMessage(conversationID, userIDs[0], "Hello, this is a test message!")
		AssertNoError(t, err, "Message sending should succeed")
		AssertNotEqual(t, message, nil, "Message should not be nil")
		AssertTrue(t, message.ID > 0, "Message ID should be positive")
		AssertEqual(t, message.ConversationID, conversationID, "Message conversation ID should match")
		AssertEqual(t, message.SenderID, userIDs[0], "Message sender ID should match")
		AssertEqual(t, message.Content, "Hello, this is a test message!", "Message content should match")
		AssertFalse(t, message.IsRead, "Message should initially be unread")
	})

	t.Run("SendMessageWithEmptyContent", func(t *testing.T) {
		// Try to send message with empty content
		_, err := messageService.SendMessage(conversationID, userIDs[0], "")
		AssertError(t, err, "Message sending should fail with empty content")
	})

	t.Run("SendMessageWithWhitespaceOnly", func(t *testing.T) {
		// Try to send message with whitespace-only content
		_, err := messageService.SendMessage(conversationID, userIDs[0], "   ")
		AssertError(t, err, "Message sending should fail with whitespace-only content")
	})

	t.Run("SendMessageToInvalidConversation", func(t *testing.T) {
		// Try to send message to non-existent conversation
		_, err := messageService.SendMessage(99999, userIDs[0], "Message to invalid conversation")
		AssertError(t, err, "Message sending should fail to invalid conversation")
	})

	t.Run("SendMessageFromInvalidUser", func(t *testing.T) {
		// Try to send message from invalid user
		_, err := messageService.SendMessage(conversationID, 99999, "Message from invalid user")
		AssertError(t, err, "Message sending should fail from invalid user")
	})

	t.Run("SendMessageFromNonParticipant", func(t *testing.T) {
		// Try to send message from user who is not a participant
		_, err := messageService.SendMessage(conversationID, userIDs[2], "Message from non-participant")
		AssertError(t, err, "Message sending should fail from non-participant")
	})

	t.Run("SendMultipleMessages", func(t *testing.T) {
		// Send multiple messages
		messages := []string{
			"First message",
			"Second message",
			"Third message",
		}

		for i, content := range messages {
			message, err := messageService.SendMessage(conversationID, userIDs[i%2], content)
			AssertNoError(t, err, "Message sending should succeed")
			AssertEqual(t, message.Content, content, "Message content should match")
		}

		// Verify all messages were sent
		retrievedMessages, err := messageService.GetConversationMessages(conversationID, userIDs[0], 10, 0)
		AssertNoError(t, err, "Should be able to retrieve messages")
		AssertTrue(t, len(retrievedMessages) >= 3, "Should have at least 3 messages")
	})
}

func TestMessageRetrieval(t *testing.T) {
	testDB := TestSetup(t)

	// Setup test users
	userIDs, err := SetupTestUsers(testDB.DB)
	AssertNoError(t, err, "Failed to setup test users")

	messageRepo := repository.NewMessageRepository(testDB.DB)
	messageService := services.NewMessageService(testDB.DB)

	// Create conversation and send messages
	participants := []int{userIDs[0], userIDs[1]}
	conversationID, err := messageRepo.CreateConversation(participants)
	AssertNoError(t, err, "Conversation creation should succeed")

	// Send test messages
	testMessages := []string{
		"First message",
		"Second message",
		"Third message",
		"Fourth message",
		"Fifth message",
	}

	for i, content := range testMessages {
		_, err := messageService.SendMessage(conversationID, userIDs[i%2], content)
		AssertNoError(t, err, "Message sending should succeed")
	}

	t.Run("GetConversationMessages", func(t *testing.T) {
		// Get all messages
		messages, err := messageService.GetConversationMessages(conversationID, userIDs[0], 10, 0)
		AssertNoError(t, err, "Should be able to get conversation messages")
		AssertTrue(t, len(messages) >= 5, "Should have at least 5 messages")

		// Verify message data structure
		for _, message := range messages {
			AssertTrue(t, message.ID > 0, "Message ID should be positive")
			AssertEqual(t, message.ConversationID, conversationID, "Message conversation ID should match")
			AssertTrue(t, message.SenderID > 0, "Message sender ID should be positive")
			AssertNotEqual(t, message.Content, "", "Message content should not be empty")
		}
	})

	t.Run("GetMessagesWithPagination", func(t *testing.T) {
		// Get messages with pagination
		messages, err := messageService.GetConversationMessages(conversationID, userIDs[0], 3, 0)
		AssertNoError(t, err, "Should be able to get paginated messages")
		AssertTrue(t, len(messages) <= 3, "Should respect limit")

		// Get next page
		nextMessages, err := messageService.GetConversationMessages(conversationID, userIDs[0], 3, 3)
		AssertNoError(t, err, "Should be able to get next page")

		// Verify no overlap
		messageIDs := make(map[int]bool)
		for _, msg := range messages {
			messageIDs[msg.ID] = true
		}
		for _, msg := range nextMessages {
			AssertFalse(t, messageIDs[msg.ID], "Should not have overlapping messages")
		}
	})

	t.Run("GetMessagesFromNonParticipant", func(t *testing.T) {
		// Try to get messages from non-participant
		_, err := messageService.GetConversationMessages(conversationID, userIDs[2], 10, 0)
		AssertError(t, err, "Should fail to get messages from non-participant")
	})

	t.Run("GetMessagesFromInvalidConversation", func(t *testing.T) {
		// Try to get messages from invalid conversation
		_, err := messageService.GetConversationMessages(99999, userIDs[0], 10, 0)
		AssertError(t, err, "Should fail to get messages from invalid conversation")
	})

	t.Run("MessagesOrderedByTime", func(t *testing.T) {
		// Get messages and verify they're ordered by time
		messages, err := messageService.GetConversationMessages(conversationID, userIDs[0], 10, 0)
		AssertNoError(t, err, "Should be able to get messages")

		if len(messages) > 1 {
			for i := 0; i < len(messages)-1; i++ {
				// Messages should be ordered by sent time (oldest first or newest first, depending on implementation)
				// Adjust this assertion based on your actual ordering
				AssertTrue(t, messages[i].SentAt.Before(messages[i+1].SentAt) || messages[i].SentAt.Equal(messages[i+1].SentAt),
					"Messages should be ordered by time")
			}
		}
	})
}

func TestConversationManagement(t *testing.T) {
	testDB := TestSetup(t)

	// Setup test users
	userIDs, err := SetupTestUsers(testDB.DB)
	AssertNoError(t, err, "Failed to setup test users")

	messageRepo := repository.NewMessageRepository(testDB.DB)
	messageService := services.NewMessageService(testDB.DB)

	t.Run("GetUserConversations", func(t *testing.T) {
		// Create multiple conversations for a user
		conv1Participants := []int{userIDs[0], userIDs[1]}
		conv1ID, err := messageRepo.CreateConversation(conv1Participants)
		AssertNoError(t, err, "First conversation creation should succeed")

		conv2Participants := []int{userIDs[0], userIDs[2]}
		conv2ID, err := messageRepo.CreateConversation(conv2Participants)
		AssertNoError(t, err, "Second conversation creation should succeed")

		// Send messages to both conversations
		_, err = messageService.SendMessage(conv1ID, userIDs[0], "Message in conversation 1")
		AssertNoError(t, err, "Message sending should succeed")

		_, err = messageService.SendMessage(conv2ID, userIDs[0], "Message in conversation 2")
		AssertNoError(t, err, "Message sending should succeed")

		// Get user's conversations
		conversations, err := messageRepo.GetUserConversations(userIDs[0])
		AssertNoError(t, err, "Should be able to get user conversations")
		AssertTrue(t, len(conversations) >= 2, "User should have at least 2 conversations")

		// Verify conversation IDs
		conversationIDs := make(map[int]bool)
		for _, conv := range conversations {
			conversationIDs[conv.ID] = true
		}
		AssertTrue(t, conversationIDs[conv1ID], "Should include first conversation")
		AssertTrue(t, conversationIDs[conv2ID], "Should include second conversation")
	})

	t.Run("IsUserParticipant", func(t *testing.T) {
		// Create conversation
		participants := []int{userIDs[0], userIDs[1]}
		conversationID, err := messageRepo.CreateConversation(participants)
		AssertNoError(t, err, "Conversation creation should succeed")

		// Check participants
		isParticipant, err := messageRepo.IsUserParticipant(conversationID, userIDs[0])
		AssertNoError(t, err, "Should be able to check participant status")
		AssertTrue(t, isParticipant, "User should be participant")

		isParticipant, err = messageRepo.IsUserParticipant(conversationID, userIDs[1])
		AssertNoError(t, err, "Should be able to check participant status")
		AssertTrue(t, isParticipant, "User should be participant")

		// Check non-participant
		isParticipant, err = messageRepo.IsUserParticipant(conversationID, userIDs[2])
		AssertNoError(t, err, "Should be able to check participant status")
		AssertFalse(t, isParticipant, "User should not be participant")
	})

	t.Run("GetConversationParticipants", func(t *testing.T) {
		// Create conversation
		participants := []int{userIDs[0], userIDs[1], userIDs[2]}
		conversationID, err := messageRepo.CreateConversation(participants)
		AssertNoError(t, err, "Conversation creation should succeed")

		// Get participants
		conversationParticipants, err := messageRepo.GetConversationParticipants(conversationID)
		AssertNoError(t, err, "Should be able to get conversation participants")
		AssertEqual(t, len(conversationParticipants), 3, "Should have 3 participants")

		// Verify participant data
		for _, participant := range conversationParticipants {
			AssertTrue(t, participant.ID > 0, "Participant ID should be positive")
			AssertNotEqual(t, participant.Username, "", "Participant should have username")
			AssertNotEqual(t, participant.FirstName, "", "Participant should have first name")
			AssertNotEqual(t, participant.LastName, "", "Participant should have last name")
		}
	})
}
