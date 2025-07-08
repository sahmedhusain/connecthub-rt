import { getCurrentUser, markMessagesAsRead } from '../utils/api.js';
import { formatTime } from './chat.js';

/**
 * @param {string} timestamp - The timestamp to parse
 * @returns {Date} - The parsed date object
 */
function parseMessageDate(timestamp) {
  if (!timestamp) return new Date(0); // Return epoch for invalid dates

  try {

    let date = new Date(timestamp);

    // If the date is invalid, try alternative parsing
    if (isNaN(date.getTime())) {
      // Try parsing as ISO string with timezone adjustment
      const isoString = timestamp.replace(' ', 'T');
      date = new Date(isoString);

      if (isNaN(date.getTime())) {
        console.warn('[ChatView] Could not parse timestamp:', timestamp);
        return new Date(0); // Return epoch for unparseable dates
      }
    }

    return date;
  } catch (error) {
    console.error('[ChatView] Error parsing timestamp:', error, 'timestamp:', timestamp);
    return new Date(0); // Return epoch for error cases
  }
}

let allUsers = [];
let currentChatUserId = null;
let currentChatConversationId = null;
let isLoadingMessages = false;
let messages = [];
let messageOffset = 0; // Track pagination offset for loading more messages
let chatViewVisible = false;
let typingTimer = null;
let isTyping = false;

// CRITICAL FIX: Global message tracking to prevent duplicates
let processedMessageIds = new Set();
let messageProcessingTimeout = 30000; // 30 seconds timeout for processed message tracking

/**
 * Initialize the chat view functionality
 */
export function initChatView() {
    console.debug("[ChatView] Initializing chat view functionality");

    // Initialize global chat state
    if (!window.currentChat) {
        window.currentChat = {
            conversationId: null,
            recipientId: null,
            recipientName: null,
            isNewConversation: false,
            isRecipientOnline: false
        };
    }

    // ENHANCED: Make status indicator functions globally available
    window.updateMessageStatusIndicators = updateMessageStatusIndicators;
    window.updateMessageStatusToSent = updateMessageStatusToSent;
    window.updateMessageStatusToError = updateMessageStatusToError;
    
    // Create the chat view elements if they don't exist
    if (!document.getElementById('chat-view-container')) {
        createChatViewDOM();
    }

    // Get user list from window.allUsers set by chat.js
    allUsers = window.allUsers || [];
    // Listen for updates to window.allUsers
    const checkForUpdates = () => {
        if (window.allUsers && window.allUsers !== allUsers) {
            allUsers = window.allUsers;
        }
    };
    setInterval(checkForUpdates, 1000);
    
    // Listen for chat conversation clicks from the sidebar
    window.addEventListener('toggle-chat-view', handleToggleChatView);

    // Listen for chat state changes from the unified state manager
    window.addEventListener('chat-state-changed', handleChatStateChange);

    // CRITICAL FIX: Listen for read status updates from WebSocket
    window.addEventListener('read-status-update', (event) => {
        const { conversationId } = event.detail;
        console.debug('[ChatView] Received read-status-update event for conversation:', conversationId);

        if (conversationId === currentChatConversationId) {
            updateMessageStatusIndicators(conversationId);
            console.debug('[ChatView] Updated message status indicators for current conversation');
        }
    });
    
    // Add event listeners for the chat view
    const closeButton = document.getElementById('chat-view-close');
    if (closeButton) {
        closeButton.addEventListener('click', closeChatView);
    }

    // Add global ESC key handler for chat view
    document.addEventListener('keydown', handleGlobalKeyboardShortcuts);
    
    const messageForm = document.getElementById('chat-view-form');
    if (messageForm) {
        messageForm.addEventListener('submit', handleSendMessage);
    }

    // Add typing indicator functionality
    const messageInput = document.getElementById('chat-view-message');
    if (messageInput) {
        messageInput.addEventListener('input', handleTyping);
        messageInput.addEventListener('keydown', handleTyping);
        messageInput.addEventListener('keydown', handleKeyboardShortcuts);
        messageInput.addEventListener('focus', handleInputFocus);
        messageInput.addEventListener('blur', handleInputBlur);
    }

    // Add keyboard navigation for close button
    const chatCloseButton = document.getElementById('chat-view-close');
    if (chatCloseButton) {
        chatCloseButton.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                hideChatView();
            }
        });
    }
    
    // Add scroll event listener with throttling
    const messagesContainer = document.getElementById('chat-view-messages');
    if (messagesContainer) {
        messagesContainer.addEventListener('scroll', throttle(function() {
            console.debug("[ChatView] Scroll event:", {
                scrollTop: messagesContainer.scrollTop,
                scrollHeight: messagesContainer.scrollHeight,
                clientHeight: messagesContainer.clientHeight,
                messagesLength: messages.length,
                conversationId: currentChatConversationId,
                isLoading: isLoadingMessages
            });

            // CRITICAL FIX: Check if user scrolled to bottom and hide scroll indicator
            const isNearBottom = messagesContainer.scrollHeight - messagesContainer.scrollTop - messagesContainer.clientHeight < 100;
            if (isNearBottom && scrollIndicatorVisible) {
                hideScrollIndicator();
            }

            // PAGINATION FIX: Use exactly 100px trigger distance as per requirements
            if (messagesContainer.scrollTop < 100 && currentChatConversationId && !isLoadingMessages && messages.length >= 10) {
                console.debug("[ChatView] Near top, loading more messages...");
                loadMoreMessages(currentChatConversationId);
            }

            // PAGINATION FIX: Show loading indicator if user is scrolling up and there might be more messages
            if (messagesContainer.scrollTop < 100 && messages.length >= 10 && !isLoadingMessages) {
                showLoadMoreIndicator();
            } else {
                hideLoadMoreIndicator();
            }
        }, 300)); // Throttle to once every 300ms
    }
    
    console.info("[ChatView] Chat view initialized successfully");
}

/**
 * Create the DOM structure for the chat view
 */
function createChatViewDOM() {
    console.debug("[ChatView] Creating chat view DOM structure");
    
    // First check if the element already exists to prevent duplicates
    if (document.getElementById('chat-view-container')) {
        console.debug("[ChatView] Chat view container already exists, skipping creation");
        return;
    }
    
    const chatViewContainer = document.createElement('div');
    chatViewContainer.id = 'chat-view-container';
    chatViewContainer.className = 'chat-view-container';
    
    chatViewContainer.innerHTML = `
        <div id="chat-view" class="chat-view md3-enhanced">
            <div class="chat-view-header">
                <div class="chat-view-user-info">
                    <div class="user-avatar">
                        <img src="/static/assets/default-avatar.png" alt="User avatar" id="chat-view-avatar">
                    </div>
                    <div class="user-details">
                        <h3 id="chat-view-username">Select a conversation</h3>
                        <span class="user-status" id="chat-view-status-text" aria-live="polite">Offline</span>
                    </div>
                </div>
                <button id="chat-view-close" class="chat-view-close" aria-label="Close chat" tabindex="0">
                    <i class="fas fa-times" aria-hidden="true"></i>
                    <span class="sr-only">Close chat</span>
                </button>
            </div>
            <div id="chat-view-messages" class="chat-view-messages">
                <div class="empty-chat-prompt">
                    <i class="far fa-comment-dots"></i>
                    <p>No messages yet. Start the conversation!</p>
                </div>
            </div>
            <div id="chat-view-typing-indicator" class="chat-view-typing-indicator" style="display: none;"></div>
            <div class="chat-view-input">
                <form id="chat-view-form">
                    <label for="chat-view-message" class="sr-only">Type your message</label>
                    <textarea
                        id="chat-view-message"
                        placeholder="Type your message here..."
                        required
                        aria-label="Message input"
                        aria-describedby="chat-input-help"
                    ></textarea>
                    <div id="chat-input-help" class="sr-only">Press Enter to send, Shift+Enter for new line</div>
                    <button type="submit" class="btn btn-primary" aria-label="Send message" tabindex="0">
                        <i class="fas fa-paper-plane" aria-hidden="true"></i>
                        <span>Send</span>
                    </button>
                </form>
            </div>
        </div>
    `;
    
    document.body.appendChild(chatViewContainer);
    
    console.debug("[ChatView] Chat view DOM structure created");
}

/**
 * Handle chat state changes from the unified state manager
 * @param {CustomEvent} event - The state change event
 */
function handleChatStateChange(event) {
    const { newState, oldState } = event.detail;
    console.debug("[ChatView] Handling chat state change:", { oldState, newState });

    // Update local state variables
    if (newState.conversationId !== currentChatConversationId) {
        currentChatConversationId = newState.conversationId;
    }

    if (newState.recipientId !== currentChatUserId) {
        currentChatUserId = newState.recipientId;
    }

    // Update UI elements based on state changes
    updateChatViewState(newState);
}

/**
 * Update chat view UI based on current state
 * @param {Object} state - The current chat state
 */
function updateChatViewState(state) {
    // Update recipient name
    const usernameElement = document.getElementById('chat-view-username');
    if (usernameElement && state.recipientName) {
        usernameElement.textContent = state.recipientName;
    }

    // Update avatar
    const avatarElement = document.getElementById('chat-view-avatar');
    if (avatarElement && state.recipientAvatar) {
        avatarElement.src = state.recipientAvatar;
        avatarElement.onerror = () => {
            avatarElement.src = '/static/assets/default-avatar.png';
        };
    }

    // Update online status
    const statusTextElement = document.getElementById('chat-view-status-text');
    if (statusTextElement) {
        statusTextElement.textContent = state.isRecipientOnline ? 'Online' : 'Offline';
        statusTextElement.className = `user-status ${state.isRecipientOnline ? 'online' : 'offline'}`;
    }

    // Update input state
    const messageInput = document.getElementById('chat-view-message');
    const sendButton = document.querySelector('#chat-view-form button[type="submit"]');

    if (messageInput && sendButton) {
        messageInput.disabled = !state.isRecipientOnline;
        sendButton.disabled = !state.isRecipientOnline;
        messageInput.placeholder = state.isRecipientOnline ?
            'Type your message here...' :
            'User is offline, messages cannot be sent';
    }
}

/**
 * Handle toggle chat view event from sidebar
 * @param {CustomEvent} event - The custom event with chat details
 */
function handleToggleChatView(event) {
    console.debug("[ChatView] Handling toggle chat view event", event.detail);
    
    const { conversationId, recipientId, isNewConversation, recipientName, recipientAvatar, isRecipientOnline } = event.detail;
    
    // Add comprehensive debugging
    console.debug("[ChatView] Event detail values:", {
        conversationId: conversationId,
        recipientId: recipientId,
        conversationIdType: typeof conversationId,
        recipientIdType: typeof recipientId,
        isNewConversation: isNewConversation,
        recipientName: recipientName,
        isRecipientOnline: isRecipientOnline
    });
    
    if (conversationId === currentChatConversationId && recipientId === currentChatUserId && chatViewVisible) {
        // If clicking the same conversation and chat is visible, close it
        closeChatView();
        return;
    }

    // CRITICAL FIX: Clear messages immediately when switching conversations
    const messagesContainer = document.getElementById('chat-view-messages');
    if (messagesContainer) {
        messagesContainer.innerHTML = '';
        messages = []; // Clear the messages array immediately
        messageOffset = 0; // Reset pagination offset
    }

    // Set current chat information with proper type conversion
    currentChatConversationId = conversationId ? parseInt(conversationId, 10) : null;
    currentChatUserId = recipientId ? parseInt(recipientId, 10) : null;
    
    // IMPORTANT: Set the global chat state that other functions expect
    window.currentChat = {
        conversationId: currentChatConversationId,
        recipientId: currentChatUserId,
        recipientName: recipientName,
        isNewConversation: isNewConversation || !currentChatConversationId,
        isRecipientOnline: isRecipientOnline
    };
    
    // Debug the values after setting
    console.debug("[ChatView] Set chat variables:", {
        currentChatConversationId: currentChatConversationId,
        currentChatUserId: currentChatUserId,
        currentChatConversationIdType: typeof currentChatConversationId,
        currentChatUserIdType: typeof currentChatUserId,
        globalCurrentChat: window.currentChat
    });
    
    // Update chat view header information
    const usernameElement = document.getElementById('chat-view-username');
    if (usernameElement) {
        usernameElement.textContent = recipientName || 'New Chat';
    }

    // Update avatar and online status styling
    const avatarElement = document.getElementById('chat-view-avatar');
    const avatarContainer = avatarElement?.parentElement;
    if (avatarElement) {
        avatarElement.src = recipientAvatar || '/static/assets/default-avatar.png';
        avatarElement.onerror = () => {
            avatarElement.src = '/static/assets/default-avatar.png';
        };

        // Update avatar container online class
        if (avatarContainer) {
            avatarContainer.classList.toggle('online', isRecipientOnline);
            avatarContainer.title = isRecipientOnline ? 'Online' : 'Offline';
        }
    }
    
    // Update status indicators
    const statusTextElement = document.getElementById('chat-view-status-text');

    if (statusTextElement) {
        statusTextElement.textContent = isRecipientOnline ? 'Online' : 'Offline';
        statusTextElement.className = `user-status ${isRecipientOnline ? 'online' : 'offline'}`;
    }

    // Enable/disable chat input based on recipient's online status
    const messageInput = document.getElementById('chat-view-message');
    const sendButton = document.querySelector('#chat-view-form button[type="submit"]');
    
    if (messageInput && sendButton) {
        messageInput.disabled = !isRecipientOnline;
        sendButton.disabled = !isRecipientOnline;
        messageInput.placeholder = isRecipientOnline ? 'Type your message here...' : 'User is offline, messages cannot be sent';
    }
    
    // Show the chat view
    openChatView();
    
    // Load messages for this conversation
    if (conversationId) {
        loadMessages(conversationId);

        // CRITICAL FIX: Mark messages as read when conversation is actually opened
        console.debug("[ChatView] Marking messages as read for opened conversation", conversationId);
        markMessagesAsRead(conversationId).then(success => {
            if (success) {
                console.debug("[ChatView] Successfully marked messages as read on conversation open");
                // Clean up unread indicators since messages are now read
                setTimeout(() => cleanupUnreadIndicators(), 2000); // Delay to allow user to see unread messages first
                // Dispatch event to update sidebar unread counts
                window.dispatchEvent(new CustomEvent('messages-marked-read', {
                    detail: { conversationId: conversationId }
                }));
            }
        }).catch(error => {
            console.error("[ChatView] Error marking messages as read on conversation open:", error);
        });
    }
    
    console.debug("[ChatView] Toggled chat view", { 
        conversationId, 
        recipientId, 
        isNewConversation 
    });
}

/**
 * Open the chat view
 */
function openChatView() {
    console.debug("[ChatView] Opening chat view");

    const chatViewContainer = document.getElementById('chat-view-container');
    const chatView = document.getElementById('chat-view');

    if (chatViewContainer && chatView) {
        // Prevent body scrolling when chat view is open
        document.body.style.overflow = 'hidden';

        chatViewContainer.classList.add('active');
        chatView.classList.add('active');
        chatViewVisible = true;

        // Focus the message input
        setTimeout(() => {
            const messageInput = document.getElementById('chat-view-message');
            if (messageInput) {
                messageInput.focus();
            }
        }, 300);
    } else {
        console.error("[ChatView] Could not find chat view elements");
    }
}

/**
 * Close the chat view
 */
function closeChatView() {
    console.debug("[ChatView] Closing chat view");

    const chatViewContainer = document.getElementById('chat-view-container');
    const chatView = document.getElementById('chat-view');

    if (chatViewContainer && chatView) {
        // Restore body scrolling when chat view is closed
        document.body.style.overflow = '';

        chatViewContainer.classList.remove('active');
        chatView.classList.remove('active');
        chatViewVisible = false;

        hideScrollIndicator();

        clearChatState();
    }
}

/**
 * Hide the chat view (alias for closeChatView for backward compatibility)
 */
function hideChatView() {
    console.debug("[ChatView] Hiding chat view (via hideChatView)");
    closeChatView();
}

/**
 * Clear chat state and remove selected conversation styling
 */
function clearChatState() {
    console.debug("[ChatView] Clearing chat state");

    // Clear current chat variables
    currentChatConversationId = null;
    currentChatUserId = null;
    messages = [];
    messageOffset = 0;

    // CRITICAL FIX: Ensure chatViewVisible is properly set to false
    chatViewVisible = false;

    // Clear typing state and debounced function
    if (typingTimer) {
        clearTimeout(typingTimer);
        typingTimer = null;
    }
    isTyping = false;

    // Clear processed message tracking
    processedMessageIds.clear();

    // Clear global chat state
    if (window.currentChat) {
        window.currentChat = {
            conversationId: null,
            recipientId: null,
            recipientName: null,
            isNewConversation: false,
            isRecipientOnline: false
        };
    }

    // Remove selected class from all conversation items
    document.querySelectorAll('.conversation-item.selected').forEach(item => {
        item.classList.remove('selected');
        console.debug("[ChatView] Removed selected class from conversation item");
    });

    // Clear the chat state using the unified state manager if available
    if (window.ChatState && typeof window.ChatState.clear === 'function') {
        window.ChatState.clear();
        console.debug("[ChatView] Cleared unified chat state");
    }

    // Clear the messages container
    const messagesContainer = document.getElementById('chat-view-messages');
    if (messagesContainer) {
        messagesContainer.innerHTML = '';
    }

    // Dispatch event to notify other components that chat was closed
    window.dispatchEvent(new CustomEvent('chat-closed', {
        detail: { timestamp: Date.now() }
    }));
}

/**
 * Cleanup function to remove global event listeners
 */
function cleanup() {
    console.debug("[ChatView] Cleaning up global event listeners");
    document.removeEventListener('keydown', handleGlobalKeyboardShortcuts);
}

// Add cleanup on page unload to prevent memory leaks
window.addEventListener('beforeunload', cleanup);

/**
 * Load messages for a conversation
 * @param {number} conversationId - The conversation ID
 */
async function loadMessages(conversationId) {
    if (!conversationId || isLoadingMessages) {
        console.debug("[ChatView] Skipping message load - invalid conversation ID or already loading");
        return;
    }

    console.debug(`[ChatView] Loading messages for conversation ${conversationId}`);
    isLoadingMessages = true;

    // CRITICAL FIX: Clear messages immediately and store the conversation ID we're loading for
    const loadingConversationId = conversationId;
    messages = [];
    messageOffset = 0;

    // CRITICAL FIX: Hide scroll indicator when switching conversations
    hideScrollIndicator();

    const messagesContainer = document.getElementById('chat-view-messages');
    if (!messagesContainer) {
        console.error("[ChatView] Messages container not found");
        isLoadingMessages = false;
        return;
    }

    // CRITICAL FIX: Clear the container immediately to prevent old messages from showing
    messagesContainer.innerHTML = `
        <div class="loading-indicator">
            <i class="fas fa-spinner fa-spin"></i>
            <p>Loading messages...</p>
        </div>
    `;

    try {
        // PAGINATION FIX: Load exactly 10 messages initially as per requirements
        const response = await fetch(`/api/messages?conversation_id=${conversationId}&limit=10&offset=0`, {
            credentials: 'include'
        });

        console.debug(`[ChatView] API response status: ${response.status}`);

        // CRITICAL FIX: Check if conversation changed while we were loading
        if (loadingConversationId !== currentChatConversationId) {
            console.debug(`[ChatView] Conversation changed during load (${loadingConversationId} -> ${currentChatConversationId}), discarding results`);
            isLoadingMessages = false;
            return;
        }

        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }

        const data = await response.json();
        console.debug(`[ChatView] Received ${Array.isArray(data) ? data.length : 0} messages from API:`, data);

        // CRITICAL FIX: Check again if conversation changed before updating UI
        if (loadingConversationId !== currentChatConversationId) {
            console.debug(`[ChatView] Conversation changed during API call, discarding results`);
            isLoadingMessages = false;
            return;
        }

        if (Array.isArray(data)) {
            // PAGINATION FIX: Database now returns messages in DESC order (newest first)
            // Reverse them to get chronological order (oldest first) for display
            messages = data.reverse();
            messageOffset = messages.length;
            console.debug(`[ChatView] Successfully loaded and sorted ${messages.length} messages for conversation ${conversationId}:`);
            console.debug(`[ChatView] First message:`, {id: messages[0]?.id, sent_at: messages[0]?.sent_at});
            console.debug(`[ChatView] Last message:`, {id: messages[messages.length-1]?.id, sent_at: messages[messages.length-1]?.sent_at});
            renderMessages(messagesContainer, messages);

            // Implement WhatsApp-like scroll behavior: scroll to first unread message
            setTimeout(() => {
                scrollToFirstUnreadMessage(messagesContainer, messages);
            }, 100);
        } else {
            throw new Error((data && data.error) || "Invalid response format - expected array");
        }
    } catch (error) {
        console.error("[ChatView] Error loading messages:", error);
        if (loadingConversationId === currentChatConversationId) {
            const { createInlineError } = await import('./error.js');
            messagesContainer.innerHTML = createInlineError(
                `Error loading messages: ${error.message}`,
                '',
                {
                    showRetry: true,
                    retryCallback: `loadMessages(${conversationId})`
                }
            );
        }
    } finally {
        isLoadingMessages = false;
    }
}

/**
 * Load more messages when scrolling to the top
 * @param {number} conversationId - The conversation ID
 */
async function loadMoreMessages(conversationId) {
  if (isLoadingMessages) {
    console.debug("[ChatView] Already loading messages, skipping");
    return;
  }

  console.debug(`[ChatView] Loading more messages for conversation ${conversationId}, current count: ${messages.length}`);
  isLoadingMessages = true;
  const messagesContainer = document.getElementById('chat-view-messages');

  // Remember current scroll position
  const scrollHeight = messagesContainer.scrollHeight;
  const scrollPosition = messagesContainer.scrollTop;

  // Add loading indicator at the top
  const loadingIndicator = document.createElement('div');
  loadingIndicator.className = 'loading-indicator loading-more';
  loadingIndicator.innerHTML = '<i class="fas fa-spinner fa-spin"></i><p>Loading more messages...</p>';
  messagesContainer.prepend(loadingIndicator);

  try {
    const response = await fetch(`/api/messages?conversation_id=${conversationId}&limit=10&offset=${messageOffset}`, {
      credentials: 'include'
    });

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }

    const data = await response.json();
    console.debug(`[ChatView] Received ${Array.isArray(data) ? data.length : 0} more messages from API`);

    if (Array.isArray(data) && data.length > 0) {
      const olderMessages = data.reverse();
      messages = [...olderMessages, ...messages];
      messageOffset += data.length;

      // Remove loading indicator before re-rendering
      loadingIndicator.remove();

      // Re-render messages
      renderMessages(messagesContainer, messages);

      // Restore scroll position to maintain user's place
      const newScrollHeight = messagesContainer.scrollHeight;
      messagesContainer.scrollTop = newScrollHeight - scrollHeight + scrollPosition;

      console.debug(`[ChatView] Loaded ${data.length} more messages, total: ${messages.length}, offset: ${messageOffset}`);
    } else {
      // No more messages to load
      loadingIndicator.remove();
      const noMoreElement = document.createElement('div');
      noMoreElement.className = 'no-more-messages';
      noMoreElement.innerHTML = '<p>No more messages</p>';
      messagesContainer.prepend(noMoreElement);

      // Remove the "no more" indicator after 2 seconds
      setTimeout(() => noMoreElement.remove(), 2000);
      console.debug("[ChatView] No more messages to load");
    }
  } catch (error) {
    console.error("[ChatView] Error loading more messages:", error);
    loadingIndicator.remove();

    const errorElement = document.createElement('div');
    errorElement.className = 'error-message';
    errorElement.innerHTML = `
      <p>Failed to load more messages: ${error.message}</p>
      <button onclick="loadMoreMessages(${conversationId})" class="retry-button">Retry</button>
    `;
    messagesContainer.prepend(errorElement);

    // Remove error after 5 seconds
    setTimeout(() => errorElement.remove(), 5000);
  } finally {
    isLoadingMessages = false;
  }
}

/**
 * Scroll to the first unread message (WhatsApp-like behavior)
 * @param {HTMLElement} container - The messages container
 * @param {Array} messagesList - The list of messages
 */
function scrollToFirstUnreadMessage(container, messagesList) {
    const currentUser = getCurrentUser();
    const currentUserId = currentUser?.userId || currentUser?.id;

    console.debug(`[ChatView] Implementing scroll-to-first-unread for ${messagesList.length} messages`);

    // Find the first unread message that was sent by someone else
    let firstUnreadIndex = -1;
    for (let i = 0; i < messagesList.length; i++) {
        const message = messagesList[i];
        // Check if message is unread and not sent by current user
        if (!message.is_read && message.sender_id !== currentUserId) {
            firstUnreadIndex = i;
            console.debug(`[ChatView] Found first unread message at index ${i}:`, message);
            break;
        }
    }

    if (firstUnreadIndex !== -1) {
        // Scroll to the first unread message
        const messageElements = container.querySelectorAll('.message');
        if (messageElements[firstUnreadIndex]) {
            const targetElement = messageElements[firstUnreadIndex];

            // Add visual indicator for unread messages
            addUnreadMessageIndicator(container, firstUnreadIndex, messageElements);

            // Scroll to the first unread message with some offset for better UX
            const containerHeight = container.clientHeight;
            const targetOffset = targetElement.offsetTop - (containerHeight * 0.2); // 20% from top

            container.scrollTo({
                top: Math.max(0, targetOffset),
                behavior: 'smooth'
            });

            console.debug(`[ChatView] Scrolled to first unread message at index ${firstUnreadIndex}`);

            // Add a subtle highlight to the first unread message
            targetElement.classList.add('first-unread-message');
            setTimeout(() => {
                targetElement.classList.remove('first-unread-message');
            }, 3000);
        }
    } else {
        // No unread messages, scroll to bottom (normal behavior)
        container.scrollTo({
            top: container.scrollHeight,
            behavior: 'smooth'
        });
        console.debug(`[ChatView] No unread messages found, scrolled to bottom`);
    }
}

/**
 * Add visual indicator for unread messages
 * @param {HTMLElement} container - The messages container
 * @param {number} firstUnreadIndex - Index of first unread message
 * @param {NodeList} messageElements - All message elements
 */
function addUnreadMessageIndicator(container, firstUnreadIndex, messageElements) {
    // Remove any existing unread indicators
    const existingIndicators = container.querySelectorAll('.unread-messages-indicator');
    existingIndicators.forEach(indicator => indicator.remove());

    const currentUser = getCurrentUser();
    const currentUserId = currentUser?.userId || currentUser?.id;

    // Count unread messages from the messages array (more reliable than DOM)
    let unreadCount = 0;
    for (let i = firstUnreadIndex; i < messages.length; i++) {
        const message = messages[i];
        if (!message.is_read && message.sender_id !== currentUserId) {
            unreadCount++;
            // Add unread styling to the corresponding DOM element
            if (messageElements[i]) {
                messageElements[i].classList.add('unread-message');
            }
        }
    }

    if (unreadCount > 0) {
        // Create unread messages indicator
        const indicator = document.createElement('div');
        indicator.className = 'unread-messages-indicator';
        indicator.innerHTML = `
            <div class="unread-indicator-content">
                <i class="fas fa-arrow-down"></i>
                <span>${unreadCount} unread message${unreadCount > 1 ? 's' : ''}</span>
            </div>
        `;

        // Position the indicator above the first unread message
        const firstUnreadElement = messageElements[firstUnreadIndex];
        container.insertBefore(indicator, firstUnreadElement);

        console.debug(`[ChatView] Added unread indicator for ${unreadCount} messages`);
    }
}

/**
 * Clean up unread message indicators and styling
 */
function cleanupUnreadIndicators() {
    const messagesContainer = document.getElementById('chat-view-messages');
    if (!messagesContainer) return;

    // Remove unread indicators
    const indicators = messagesContainer.querySelectorAll('.unread-messages-indicator');
    indicators.forEach(indicator => indicator.remove());

    // Remove unread styling from messages
    const unreadMessages = messagesContainer.querySelectorAll('.message.unread-message');
    unreadMessages.forEach(message => message.classList.remove('unread-message'));

    // Remove first unread highlighting
    const firstUnreadMessages = messagesContainer.querySelectorAll('.message.first-unread-message');
    firstUnreadMessages.forEach(message => message.classList.remove('first-unread-message'));

    console.debug('[ChatView] Cleaned up unread indicators and styling');
}

/**
 * Render messages in the container
 * @param {HTMLElement} container - The container element
 * @param {Array} messagesList - The list of messages
 */
function renderMessages(container, messagesList) {
  if (!messagesList || messagesList.length === 0) {
    container.innerHTML = `
      <div class="empty-chat-prompt">
        <i class="far fa-comment-dots"></i>
        <p>No messages yet. Start the conversation!</p>
      </div>
    `;
    return;
  }

  const currentUser = getCurrentUser();
  const userId = currentUser?.userId || currentUser?.id;

  console.debug(`[ChatView] Rendering ${messagesList.length} messages for user ${userId}`);

  const messagesHTML = messagesList.map(message => {
    return createMessageHTML(message, userId);
  }).join('');

  container.innerHTML = messagesHTML;
  console.debug(`[ChatView] Successfully rendered ${messagesList.length} messages`);
}

/**
 * Create HTML for a single message
 * @param {Object} message - The message object
 * @param {number} userId - The current user ID
 * @returns {string} - The HTML string for the message
 */
function createMessageHTML(message, userId) {
  const isSent = message.sender_id === userId;
  const messageClass = `message ${isSent ? 'sent' : 'received'} ${message.pending ? 'pending' : ''}`;
  const formattedTime = formatTime(message.sent_at);

  // Escape HTML content to prevent XSS
  const escapedContent = escapeHtml(message.content || '');
  const escapedSenderName = escapeHtml(message.sender_name || 'User');

  // ENHANCED: Generate status indicator HTML for sent messages
  let statusIndicatorHTML = '';
  if (isSent) {
    statusIndicatorHTML = generateStatusIndicatorHTML(message);
  }

  return `
    <div class="${messageClass}" data-message-id="${message.id}" data-is-read="${message.is_read}" data-sender-id="${message.sender_id}">
      <div class="message-content">
        ${!isSent ? `<div class="sender-name">${escapedSenderName}</div>` : ''}
        <div class="message-text">${escapedContent}</div>
      </div>
      <div class="message-meta">
        <div class="message-timestamp">${formattedTime}</div>
        ${statusIndicatorHTML}
      </div>
    </div>
  `;
}

/**
 * ENHANCED: Generate status indicator HTML based on message state
 * @param {Object} message - The message object
 * @returns {string} - The HTML string for the status indicator
 */
function generateStatusIndicatorHTML(message) {
  if (message.pending) {
    return `<div class="message-status" data-message-id="${message.id}">
      <i class="fas fa-clock pending" title="Sending..."></i>
    </div>`;
  }

  if (message.error) {
    return `<div class="message-status" data-message-id="${message.id}">
      <i class="fas fa-exclamation-triangle error" title="Failed to send"></i>
    </div>`;
  }

  if (message.is_read) {
    return `<div class="message-status" data-message-id="${message.id}">
      <i class="fas fa-check-double read" title="Read"></i>
    </div>`;
  }

  // Default: sent but not read
  return `<div class="message-status" data-message-id="${message.id}">
    <i class="fas fa-check unread" title="Sent"></i>
  </div>`;
}

/**
 * CRITICAL FIX: Append a single message to the container (for real-time messages)
 * @param {HTMLElement} container - The container element
 * @param {Object} message - The message object
 */
function appendSingleMessage(container, message) {
  const currentUser = getCurrentUser();
  const userId = currentUser?.userId || currentUser?.id;

  // Remove empty chat prompt if it exists
  const emptyPrompt = container.querySelector('.empty-chat-prompt');
  if (emptyPrompt) {
    emptyPrompt.remove();
  }

  // Create message element
  const messageElement = document.createElement('div');
  messageElement.innerHTML = createMessageHTML(message, userId);

  // CRITICAL FIX: Append to the END of the container (bottom) for chronological order
  container.appendChild(messageElement.firstElementChild);

  console.debug(`[ChatView] Appended single message to bottom of chat`);
}

/**
 * Escape HTML to prevent XSS attacks
 * @param {string} text - The text to escape
 * @returns {string} - The escaped text
 */
function escapeHtml(text) {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

/**
 * Handle sending a message
 * @param {Event} event - The form submit event
 */
async function handleSendMessage(event) {
    event.preventDefault();
    
    const messageInput = document.querySelector('#chat-view-message');
    const content = messageInput?.value?.trim();
    
    if (!content) {
        console.warn('[ChatView] Cannot send empty message');
        return;
    }
    
    // Add comprehensive debugging
    console.debug('[ChatView] Message send attempt:', {
        currentChatUserId: currentChatUserId,
        currentChatConversationId: currentChatConversationId,
        currentChatUserIdType: typeof currentChatUserId,
        globalCurrentChatRecipientId: window.currentChat?.recipientId,
        globalCurrentChatRecipientIdType: typeof window.currentChat?.recipientId,
        onlineUsersSize: window.onlineUsers?.size,
        onlineUsersHasRecipient: window.onlineUsers?.has(currentChatUserId),
        socketExists: !!window.socket,
        socketState: window.socket?.readyState
    });
    
    // Use currentChatConversationId and currentChatUserId for context
    if (!currentChatConversationId && !currentChatUserId) {
        console.error('[ChatView] No active conversation or recipient');
        showMessageError('No conversation selected');
        return;
    }

    // Additional validation for recipient ID
    if (!currentChatUserId || currentChatUserId === 0) {
        console.error('[ChatView] Invalid recipient ID:', currentChatUserId);
        showMessageError('Invalid recipient selected');
        return;
    }

    // Double-check with the global state
    if (!window.currentChat?.recipientId || window.currentChat.recipientId === 0) {
        console.error('[ChatView] Global chat state has invalid recipient ID:', window.currentChat);
        showMessageError('Chat state error - please try selecting the conversation again');
        return;
    }

    // Check if recipient is online
    if (!window.onlineUsers?.has(currentChatUserId)) {
        showMessageError('Cannot send message: Recipient must be online');
        return;
    }

    // Disable input while sending
    if (messageInput) {
        messageInput.disabled = true;
    }

    try {
        // Send message via WebSocket
        if (!window.socket || window.socket.readyState !== WebSocket.OPEN) {
            throw new Error('WebSocket connection not available');
        }

        const message = {
            type: 'private',
            recipient_id: parseInt(currentChatUserId, 10), // Fixed: Use snake_case to match backend JSON tag
            content: content,
            conversation_id: currentChatConversationId, // Fixed: Use snake_case to match backend JSON tag
            is_new_conversation: !currentChatConversationId // Fixed: Use snake_case to match backend JSON tag
        };

        // Add debug logging to see what we're sending
        console.debug('[ChatView] Sending WebSocket message:', {
            type: message.type,
            recipient_id: message.recipient_id,
            recipientIdType: typeof message.recipient_id,
            conversation_id: message.conversation_id,
            contentLength: message.content.length,
            is_new_conversation: message.is_new_conversation
        });

        window.socket.send(JSON.stringify(message));

        // Clear input
        if (messageInput) {
            messageInput.value = '';
        }

        // Stop typing indicator when message is sent
        if (isTyping) {
            isTyping = false;
            sendTypingIndicator('stop');
            // Note: debounced function will handle its own cleanup
        }

        // Clear any existing error messages
        const existingError = document.querySelector('.message-error');
        if (existingError) {
            existingError.remove();
        }

        // CRITICAL FIX: Add message to UI optimistically with proper ordering
        const currentUser = getCurrentUser();
        const tempId = `temp_${Date.now()}`;
        const optimisticMessage = {
            id: tempId,
            sender_id: currentUser?.userId || currentUser?.id,
            sender_name: currentUser?.username || 'You',
            conversation_id: currentChatConversationId,
            content: content,
            sent_at: new Date().toISOString(),
            pending: true
        };

        // Add to messages array (messages are in chronological order - oldest first)
        messages.push(optimisticMessage);
        messageOffset++;

        const messagesContainer = document.getElementById('chat-view-messages');
        if (messagesContainer) {
            // CRITICAL FIX: Use appendMessage for new messages instead of full re-render
            appendSingleMessage(messagesContainer, optimisticMessage);

            // Scroll to bottom to show the new message
            setTimeout(() => {
                messagesContainer.scrollTop = messagesContainer.scrollHeight;
            }, 50);
        }

        // ENHANCED: Set a timeout to mark message as error if no confirmation received
        setTimeout(() => {
            const messageIndex = messages.findIndex(m => m.id === tempId && m.pending);
            if (messageIndex !== -1) {
                console.warn('[ChatView] Message confirmation timeout, marking as error:', tempId);
                messages[messageIndex].pending = false;
                messages[messageIndex].error = true;
                updateMessageStatusToError(tempId);
            }
        }, 10000); // 10 second timeout

        console.debug("[ChatView] Added optimistic message to UI");

    } catch (error) {
        console.error('[ChatView] Error sending message:', error);
        showMessageError(error.message || 'Failed to send message');
    } finally {
        // Re-enable input
        if (messageInput) {
            messageInput.disabled = false;
            messageInput.focus();
        }
    }
}

function showMessageError(message) {
    // Look for the correct message input selector
    const messageInput = document.querySelector('#chat-view-message'); // Fixed selector
    if (!messageInput) {
        console.warn('[ChatView] Message input not found for error display');
        return;
    }
    
    // Remove existing error
    const existingError = document.querySelector('.message-error');
    if (existingError) {
        existingError.remove();
    }
    
    // Create error element
    const errorElement = document.createElement('div');
    errorElement.className = 'message-error';
    errorElement.style.cssText = `
        color: #e74c3c; 
        font-size: 0.9em; 
        margin-top: 0.5rem; 
        padding: 0.5rem; 
        background: #fdf2f2; 
        border-radius: 4px; 
        border-left: 3px solid #e74c3c;
        animation: fadeIn 0.3s ease-in;
    `;
    errorElement.textContent = message;
    
    // Insert after message input container
    const inputContainer = messageInput.closest('.chat-view-input') || messageInput.parentNode;
    inputContainer.appendChild(errorElement);
    
    // Auto-hide after 5 seconds
    setTimeout(() => {
        if (errorElement.parentNode) {
            errorElement.remove();
        }
    }, 5000);
}

/**
 * Add a message received from WebSocket to the current chat
 * @param {Object} message - The message object
 */
export function addReceivedMessage(message) {
    if (!message) {
        console.warn("[ChatView] Received null/undefined message");
        return;
    }

    // CRITICAL FIX: Global duplicate prevention using message ID and content hash
    const messageKey = `${message.id || 'no-id'}_${message.sender_id}_${message.conversation_id}_${message.content?.substring(0, 50)}`;

    if (processedMessageIds.has(messageKey)) {
        console.debug("[ChatView] Message already processed, ignoring duplicate:", messageKey);
        return;
    }

    // Add to processed messages set
    processedMessageIds.add(messageKey);

    // Clean up old processed message IDs after timeout
    setTimeout(() => {
        processedMessageIds.delete(messageKey);
    }, messageProcessingTimeout);

    console.debug("[ChatView] Processing received message:", {
        messageId: message.id,
        senderId: message.sender_id,
        conversationId: message.conversation_id,
        currentConversationId: currentChatConversationId,
        content: message.content?.substring(0, 50) + '...',
        messageKey: messageKey
    });

    const currentUser = getCurrentUser();
    const currentUserId = currentUser?.userId || currentUser?.id;

    // CRITICAL FIX: Strict conversation ID checking to prevent cross-contamination
    if (message.conversation_id !== currentChatConversationId) {
        console.debug("[ChatView] Message is for different conversation, ignoring", {
            messageConversationId: message.conversation_id,
            currentConversationId: currentChatConversationId
        });

        // Still show notification for messages from other conversations
        if (message.sender_id !== currentUserId) {
            showMessageNotification(message);
        }
        return; // Don't update UI for messages from other conversations
    }

    // CRITICAL FIX: If this is a message from the current user, it's likely a confirmation
    // We should only process it if we have a pending message to update
    if (message.sender_id === currentUserId) {
        console.debug("[ChatView] Received confirmation message from current user");

        // Check if we have any pending messages that match this confirmation
        const hasPendingMessage = messages.some(m =>
            m.pending &&
            m.content === message.content &&
            m.sender_id === currentUserId
        );

        if (!hasPendingMessage) {
            console.debug("[ChatView] No pending message found for confirmation, ignoring to prevent duplicate");
            return;
        }
    }

    // CRITICAL FIX: Double-check that we're still in the same conversation
    if (!currentChatConversationId || !chatViewVisible) {
        console.debug("[ChatView] No active conversation or chat view not visible, ignoring message");
        return;
    }

    const messagesContainer = document.getElementById('chat-view-messages');
    let shouldAppendMessage = false;

    // If this is a confirmation of our sent message, update the pending message
    if (message.sender_id === currentUserId) {
        // CRITICAL FIX: Find pending message by content and timestamp proximity (within 5 seconds)
        const messageTime = new Date(message.sent_at || message.timestamp).getTime();
        const pendingIndex = messages.findIndex(m => {
            if (!m.pending || m.content !== message.content) return false;

            // Check if the message was sent within the last 5 seconds
            const pendingTime = new Date(m.sent_at).getTime();
            const timeDiff = Math.abs(messageTime - pendingTime);
            return timeDiff <= 5000; // 5 second tolerance
        });

        if (pendingIndex !== -1) {
            console.debug("[ChatView] Updating pending message to confirmed", {
                pendingId: messages[pendingIndex].id,
                confirmedId: message.id,
                content: message.content
            });

            const oldTempId = messages[pendingIndex].id;

            // Update the message in the array with confirmed data
            messages[pendingIndex] = {
                ...message,
                pending: false,
                is_read: false // Initially unread until recipient reads it
            };

            // CRITICAL FIX: Update the existing pending message in the DOM without creating duplicates
            if (messagesContainer && chatViewVisible) {
                const pendingElement = messagesContainer.querySelector(`[data-message-id="${oldTempId}"]`);
                if (pendingElement) {
                    // Update the data-message-id to the real ID
                    pendingElement.setAttribute('data-message-id', messages[pendingIndex].id);

                    // Remove pending class
                    pendingElement.classList.remove('pending');

                    // CRITICAL FIX: Update status indicator from pending to sent
                    const statusElement = pendingElement.querySelector('.message-status');
                    if (statusElement) {
                        statusElement.setAttribute('data-message-id', messages[pendingIndex].id);
                        const statusIcon = statusElement.querySelector('i');
                        if (statusIcon) {
                            statusIcon.className = 'fas fa-check unread';
                            statusIcon.title = 'Sent';
                        }
                    }

                    console.debug("[ChatView] Successfully updated pending message DOM element with smooth transition");

                    // CRITICAL FIX: Set shouldAppendMessage to false to prevent duplicate
                    shouldAppendMessage = false;
                } else {
                    console.warn("[ChatView] Could not find pending message element in DOM", oldTempId);
                    // Only append if we can't find the pending element
                    shouldAppendMessage = true;
                }
            } else {
                // If chat view is not visible, don't append
                shouldAppendMessage = false;
            }
        } else {
            console.debug("[ChatView] No matching pending message found, checking if this is a duplicate");

            // CRITICAL FIX: Check if this message already exists to prevent duplicates
            const existingMessage = messages.find(m =>
                m.id === message.id ||
                (m.content === message.content &&
                 m.sender_id === message.sender_id &&
                 Math.abs(new Date(m.sent_at).getTime() - new Date(message.sent_at).getTime()) < 5000)
            );

            if (!existingMessage) {
                messages.push(message);
                messageOffset++;
                shouldAppendMessage = true;
                console.debug("[ChatView] Added new confirmed message");
            } else {
                console.debug("[ChatView] Message already exists, skipping to prevent duplicate");
                shouldAppendMessage = false;
            }
        }
    } else {
        // Add received message to current chat only if it's for the active conversation
        console.debug("[ChatView] Adding received message from other user");

        // CRITICAL FIX: Check for duplicates before adding received messages
        const existingMessage = messages.find(m =>
            m.id === message.id ||
            (m.content === message.content &&
             m.sender_id === message.sender_id &&
             Math.abs(new Date(m.sent_at).getTime() - new Date(message.sent_at).getTime()) < 5000)
        );

        console.debug("[ChatView] Duplicate check for received message:", {
            messageId: message.id,
            content: message.content?.substring(0, 30),
            senderId: message.sender_id,
            existingMessage: existingMessage ? {
                id: existingMessage.id,
                content: existingMessage.content?.substring(0, 30),
                pending: existingMessage.pending
            } : null,
            messagesCount: messages.length
        });

        if (!existingMessage) {
            // CRITICAL FIX: Insert message in correct chronological position
            const messageDate = parseMessageDate(message.sent_at);
            let insertIndex = messages.length; // Default to end

            // Find the correct position to insert the message (maintain chronological order)
            for (let i = messages.length - 1; i >= 0; i--) {
                const existingMessageDate = parseMessageDate(messages[i].sent_at);
                if (messageDate >= existingMessageDate) {
                    insertIndex = i + 1;
                    break;
                }
            }

            // Insert at the correct position
            messages.splice(insertIndex, 0, message);
            shouldAppendMessage = true;
            console.debug("[ChatView] Added new received message at index", insertIndex);
        } else {
            console.debug("[ChatView] Received message already exists, skipping to prevent duplicate");
            shouldAppendMessage = false;
        }

        // CRITICAL FIX: Only auto-mark messages as read when conversation is actively open AND visible
        // AND the user is actually viewing the conversation (not just when it's in the background)
        if (currentChatConversationId &&
            chatViewVisible &&
            message.conversation_id === currentChatConversationId &&
            !document.hidden && // Page is visible
            document.hasFocus()) { // Window has focus

            console.debug("[ChatView] Auto-marking messages as read for actively viewed conversation", currentChatConversationId);
            markMessagesAsRead(currentChatConversationId).then(success => {
                if (success) {
                    console.debug("[ChatView] Successfully auto-marked messages as read");
                    // Update the message status in the UI
                    message.is_read = true;
                    // Clean up unread indicators since messages are now read
                    cleanupUnreadIndicators();
                    // Dispatch event to update sidebar unread counts
                    window.dispatchEvent(new CustomEvent('messages-marked-read', {
                        detail: { conversationId: currentChatConversationId }
                    }));
                } else {
                    console.warn("[ChatView] Failed to auto-mark messages as read");
                }
            }).catch(error => {
                console.error("[ChatView] Error auto-marking messages as read:", error);
            });
        } else {
            console.debug("[ChatView] Not auto-marking as read - conversation not actively viewed", {
                currentConversationId: currentChatConversationId,
                chatViewVisible: chatViewVisible,
                messageConversationId: message.conversation_id,
                documentHidden: document.hidden,
                documentHasFocus: document.hasFocus()
            });
        }

        // Note: Global notifications are now handled in chat.js for all messages
        // This ensures notifications work everywhere in the SPA, not just when chat is open
        // Still update page title with unread count
        updatePageTitleWithUnreadCount();

        // CRITICAL FIX: Increment unread count and show scroll indicator if user is not at bottom
        const messagesContainer = document.getElementById('chat-view-messages');
        if (messagesContainer) {
            const isNearBottom = messagesContainer.scrollHeight - messagesContainer.scrollTop - messagesContainer.clientHeight < 100;
            if (!isNearBottom) {
                incrementUnreadMessageCount();
                showScrollIndicatorIfNeeded();
            }
        }
    }

    // CRITICAL FIX: Re-render messages to maintain proper chronological order
    if (messagesContainer && chatViewVisible && shouldAppendMessage) {
        // Check if the new message should be at the end (most common case)
        const isNewestMessage = messages.length === 0 ||
            parseMessageDate(message.sent_at) >= parseMessageDate(messages[messages.length - 1].sent_at);

        if (isNewestMessage) {
            // Optimize: Just append if it's the newest message
            appendSingleMessage(messagesContainer, message);
        } else {
            // Re-render all messages to maintain chronological order
            renderMessages(messagesContainer, messages);
        }

        // Scroll to the bottom if we're already near the bottom or if it's our own message
        const isNearBottom = messagesContainer.scrollHeight - messagesContainer.scrollTop - messagesContainer.clientHeight < 100;
        const isOwnMessage = message.sender_id === currentUserId;

        if (isNearBottom || isOwnMessage) {
            setTimeout(() => {
                messagesContainer.scrollTop = messagesContainer.scrollHeight;
            }, 50);
        }

        console.debug("[ChatView] Updated chat view with new message in chronological order");
    }
}


/**
 * Throttle utility function
 * @param {Function} func - The function to throttle
 * @param {number} limit - The throttle limit in milliseconds
 * @returns {Function} - The throttled function
 */
function throttle(func, limit) {
  let inThrottle;
  return function(...args) {
    if (!inThrottle) {
      func.apply(this, args);
      inThrottle = true;
      setTimeout(() => inThrottle = false, limit);
    }
  };
}

/**
 * Debounce utility function
 * @param {Function} func - The function to debounce
 * @param {number} wait - The debounce wait time in milliseconds
 * @returns {Function} - The debounced function
 */
function debounce(func, wait) {
  let timeout;
  return function(...args) {
    const context = this;
    clearTimeout(timeout);
    timeout = setTimeout(() => func.apply(context, args), wait);
  };
}

// ========================================
// NOTIFICATION SYSTEM
// ========================================

/**
 * Notification preferences (stored in localStorage)
 */
const NotificationPreferences = {
  get: () => {
    const prefs = localStorage.getItem('chatNotificationPreferences');
    return prefs ? JSON.parse(prefs) : {
      sound: true,
      desktop: true,
      visual: true,
      volume: 0.3
    };
  },

  set: (preferences) => {
    localStorage.setItem('chatNotificationPreferences', JSON.stringify(preferences));
  }
};

/**
 * Request notification permission if not already granted
 */
async function requestNotificationPermission() {
  if ('Notification' in window && Notification.permission === 'default') {
    try {
      const permission = await Notification.requestPermission();
      console.debug('[ChatView] Notification permission:', permission);
      return permission === 'granted';
    } catch (error) {
      console.debug('[ChatView] Error requesting notification permission:', error);
      return false;
    }
  }
  return Notification.permission === 'granted';
}

/**
 * Show comprehensive notification for new messages
 * NOTE: This function is deprecated - notifications are now handled globally in globalChatNotifications.js
 * @param {Object} message - The message object
 */
async function showMessageNotification(message) {
  // This function is now handled globally by the chat notification system
  // See globalChatNotifications.js and the integration in chat.js
  console.debug('[ChatView] Notification handling moved to global system');

  // Still update page title with unread count
  updatePageTitleWithUnreadCount();
}

/**
 * Play notification sound
 * @param {number} volume - Volume level (0-1)
 */
function playNotificationSound(volume = 0.3) {
  try {
    const audio = new Audio('/static/assets/notification.mp3');
    audio.volume = Math.max(0, Math.min(1, volume));

    // Handle both loading errors and playback errors gracefully
    audio.onerror = () => {
      console.debug('[ChatView] Notification sound file not found, using fallback');
      playFallbackNotificationSound();
    };

    audio.play().catch(err => {
      console.debug('[ChatView] Could not play notification sound:', err);
      playFallbackNotificationSound();
    });
  } catch (err) {
    console.debug('[ChatView] Notification sound not available:', err);
    playFallbackNotificationSound();
  }
}

/**
 * Play fallback notification sound when audio file is not available
 */
function playFallbackNotificationSound() {
  try {
    // Try Web Audio API beep
    if (window.AudioContext || window.webkitAudioContext) {
      const audioContext = new (window.AudioContext || window.webkitAudioContext)();
      const oscillator = audioContext.createOscillator();
      const gainNode = audioContext.createGain();

      oscillator.connect(gainNode);
      gainNode.connect(audioContext.destination);

      oscillator.frequency.setValueAtTime(800, audioContext.currentTime);
      gainNode.gain.setValueAtTime(0.1, audioContext.currentTime);
      gainNode.gain.exponentialRampToValueAtTime(0.01, audioContext.currentTime + 0.1);

      oscillator.start(audioContext.currentTime);
      oscillator.stop(audioContext.currentTime + 0.1);
    } else if (window.speechSynthesis) {
      // Fallback to speech synthesis (silent utterance for system beep)
      const utterance = new SpeechSynthesisUtterance('');
      utterance.volume = 0;
      window.speechSynthesis.speak(utterance);
    }
  } catch (err) {
    console.debug('[ChatView] Fallback notification sound also failed:', err);
  }
}

// ========================================
// SCROLL INDICATOR SYSTEM
// ========================================

let unreadMessageCount = 0;
let scrollIndicatorVisible = false;

/**
 * Show scroll indicator if user is not at the bottom of chat
 */
function showScrollIndicatorIfNeeded() {
  const chatContainer = document.getElementById('chat-view-messages');
  if (!chatContainer || !chatViewVisible) return;

  // Check if user is near the bottom (within 100px)
  const isNearBottom = chatContainer.scrollHeight - chatContainer.scrollTop - chatContainer.clientHeight < 100;

  if (!isNearBottom && !scrollIndicatorVisible) {
    showScrollIndicator();
  } else if (isNearBottom && scrollIndicatorVisible) {
    hideScrollIndicator();
  }
}

/**
 * Show the scroll indicator button
 */
function showScrollIndicator() {
  // Remove existing indicator if present
  hideScrollIndicator();

  const chatContainer = document.getElementById('chat-view-messages');
  if (!chatContainer) return;

  scrollIndicatorVisible = true;

  // Create scroll indicator
  const scrollIndicator = document.createElement('div');
  scrollIndicator.id = 'scroll-to-bottom-indicator';
  scrollIndicator.className = 'scroll-indicator';

  const messageText = unreadMessageCount > 0 ?
    `${unreadMessageCount} new message${unreadMessageCount > 1 ? 's' : ''}` :
    'New message';

  scrollIndicator.innerHTML = `
    <button type="button" class="scroll-to-bottom-btn" title="Scroll to bottom">
      <i class="fas fa-chevron-down"></i>
      <span class="new-message-text">${messageText}</span>
    </button>
  `;

  // Add click handler
  const button = scrollIndicator.querySelector('.scroll-to-bottom-btn');
  button.addEventListener('click', scrollToBottom);

  // Insert indicator into chat view container (relative to messages container)
  const chatViewContainer = chatContainer.parentElement;
  if (chatViewContainer) {
    chatViewContainer.appendChild(scrollIndicator);
  }

  console.debug('[ChatView] Scroll indicator shown with message:', messageText);
}

/**
 * Hide the scroll indicator
 */
function hideScrollIndicator() {
  const indicator = document.getElementById('scroll-to-bottom-indicator');
  if (indicator) {
    indicator.remove();
    scrollIndicatorVisible = false;
    unreadMessageCount = 0; // Reset count when hiding
    console.debug('[ChatView] Scroll indicator hidden');
  }
}

/**
 * Scroll to the bottom of the chat
 */
function scrollToBottom() {
  const chatContainer = document.getElementById('chat-view-messages');
  if (chatContainer) {
    chatContainer.scrollTop = chatContainer.scrollHeight;
    hideScrollIndicator();
    console.debug('[ChatView] Scrolled to bottom via indicator');
  }
}

/**
 * Show load more indicator when user is near the top
 */
function showLoadMoreIndicator() {
  if (document.getElementById('load-more-indicator')) return; // Already showing

  const chatContainer = document.getElementById('chat-view-messages');
  if (!chatContainer) return;

  const indicator = document.createElement('div');
  indicator.id = 'load-more-indicator';
  indicator.className = 'load-more-hint';
  indicator.innerHTML = `
    <div class="load-more-content">
      <i class="fas fa-arrow-up"></i>
      <span>Scroll up to load older messages</span>
    </div>
  `;

  chatContainer.appendChild(indicator);
  console.debug('[ChatView] Load more indicator shown');
}

/**
 * Hide load more indicator
 */
function hideLoadMoreIndicator() {
  const indicator = document.getElementById('load-more-indicator');
  if (indicator) {
    indicator.remove();
    console.debug('[ChatView] Load more indicator hidden');
  }
}

/**
 * Increment unread message count for scroll indicator
 */
function incrementUnreadMessageCount() {
  unreadMessageCount++;

  // Update indicator text if it's visible
  if (scrollIndicatorVisible) {
    const indicator = document.getElementById('scroll-to-bottom-indicator');
    if (indicator) {
      const textSpan = indicator.querySelector('.new-message-text');
      if (textSpan) {
        const messageText = unreadMessageCount > 0 ?
          `${unreadMessageCount} new message${unreadMessageCount > 1 ? 's' : ''}` :
          'New message';
        textSpan.textContent = messageText;
      }
    }
  }
}

/**
 * ENHANCED: Update message status indicators in real-time with animations
 * @param {number} conversationId - The conversation ID
 * @param {Array} messageIds - Array of message IDs that were marked as read
 */
export function updateMessageStatusIndicators(conversationId, messageIds = []) {
  if (conversationId !== currentChatConversationId) {
    return; // Only update if it's the current conversation
  }

  console.debug('[ChatView] Updating message status indicators for conversation:', conversationId, 'messageIds:', messageIds);

  // Update all sent messages in the current conversation to read status
  const messageStatusElements = document.querySelectorAll('.message.sent .message-status');

  messageStatusElements.forEach(statusElement => {
    const messageId = statusElement.dataset.messageId;

    // If specific message IDs provided, only update those; otherwise update all
    if (messageIds.length === 0 || messageIds.includes(parseInt(messageId))) {
      updateSingleMessageStatus(statusElement, 'read');
    }
  });

  // Also update the messages array
  messages.forEach(message => {
    if (message.sender_id === (getCurrentUser()?.userId || getCurrentUser()?.id)) {
      if (messageIds.length === 0 || messageIds.includes(message.id)) {
        message.is_read = true;
      }
    }
  });
}

/**
 * ENHANCED: Update a single message status indicator with animation
 * @param {HTMLElement} statusElement - The status element to update
 * @param {string} newStatus - The new status ('pending', 'sent', 'read', 'error')
 */
function updateSingleMessageStatus(statusElement, newStatus) {
  const statusIcon = statusElement.querySelector('i');
  if (!statusIcon) return;

  const messageId = statusElement.dataset.messageId;

  // Add animation class for smooth transition
  statusIcon.classList.add('status-changed');

  // Remove animation class after animation completes
  setTimeout(() => {
    statusIcon.classList.remove('status-changed');
  }, 400);

  // Update icon based on new status
  switch (newStatus) {
    case 'pending':
      statusIcon.className = 'fas fa-clock pending';
      statusIcon.title = 'Sending...';
      break;
    case 'sent':
      statusIcon.className = 'fas fa-check unread';
      statusIcon.title = 'Sent';
      break;
    case 'read':
      statusIcon.className = 'fas fa-check-double read';
      statusIcon.title = 'Read';
      break;
    case 'error':
      statusIcon.className = 'fas fa-exclamation-triangle error';
      statusIcon.title = 'Failed to send';
      break;
    default:
      console.warn('[ChatView] Unknown status:', newStatus);
      return;
  }

  console.debug('[ChatView] Updated message status indicator for message:', messageId, 'to:', newStatus);
}

/**
 * ENHANCED: Update message status from pending to sent
 * @param {string} messageId - The message ID
 */
export function updateMessageStatusToSent(messageId) {
  console.debug('[ChatView] Updating message status to sent for ID:', messageId);

  // Try multiple selectors to find the status element
  let statusElement = document.querySelector(`.message-status[data-message-id="${messageId}"]`);

  if (!statusElement) {
    // Try finding by message element and then the status within it
    const messageElement = document.querySelector(`[data-message-id="${messageId}"]`);
    if (messageElement) {
      statusElement = messageElement.querySelector('.message-status');
    }
  }

  if (!statusElement) {
    // Try finding pending indicators that might not have the data attribute yet
    const pendingIndicators = document.querySelectorAll('.message-status .pending');
    for (const indicator of pendingIndicators) {
      const messageEl = indicator.closest('[data-message-id]');
      if (messageEl && messageEl.getAttribute('data-message-id') === messageId) {
        statusElement = indicator.closest('.message-status');
        break;
      }
    }
  }

  if (statusElement) {
    updateSingleMessageStatus(statusElement, 'sent');
    console.debug('[ChatView] Successfully updated message status to sent');
  } else {
    console.warn('[ChatView] Could not find status element for message ID:', messageId);
  }
}

/**
 * ENHANCED: Update message status to error state
 * @param {string} messageId - The message ID
 */
export function updateMessageStatusToError(messageId) {
  console.debug('[ChatView] Updating message status to error for ID:', messageId);

  // Try multiple selectors to find the status element
  let statusElement = document.querySelector(`.message-status[data-message-id="${messageId}"]`);

  if (!statusElement) {
    // Try finding by message element and then the status within it
    const messageElement = document.querySelector(`[data-message-id="${messageId}"]`);
    if (messageElement) {
      statusElement = messageElement.querySelector('.message-status');
    }
  }

  if (statusElement) {
    updateSingleMessageStatus(statusElement, 'error');
    console.debug('[ChatView] Successfully updated message status to error');
  } else {
    console.warn('[ChatView] Could not find status element for message ID:', messageId);
  }
}

/**
 * Show visual in-app notification
 * @param {Object} message - The message object
 */
function showVisualNotification(message) {
  const senderName = message.sender_name || 'Someone';
  const content = message.content || '';
  const truncatedContent = content.length > 50 ? content.substring(0, 47) + '...' : content;

  // Create notification element
  const notification = document.createElement('div');
  notification.className = 'chat-notification';
  notification.innerHTML = `
    <div class="chat-notification-content">
      <div class="chat-notification-header">
        <strong>${escapeHtml(senderName)}</strong>
        <button class="chat-notification-close" aria-label="Close notification">×</button>
      </div>
      <div class="chat-notification-message">${escapeHtml(truncatedContent)}</div>
    </div>
  `;

  // Add click handler to open conversation
  notification.addEventListener('click', () => {
    if (message.conversation_id) {
      // Dispatch event to open the conversation
      window.dispatchEvent(new CustomEvent('open-conversation', {
        detail: { conversationId: message.conversation_id }
      }));
    }
    notification.remove();
  });

  // Add close button handler
  const closeBtn = notification.querySelector('.chat-notification-close');
  closeBtn.addEventListener('click', (e) => {
    e.stopPropagation();
    notification.remove();
  });

  // Add to page
  let container = document.querySelector('.chat-notifications-container');
  if (!container) {
    container = document.createElement('div');
    container.className = 'chat-notifications-container';
    document.body.appendChild(container);
  }

  container.appendChild(notification);

  // Auto-remove after 5 seconds
  setTimeout(() => {
    if (notification.parentNode) {
      notification.remove();
    }
  }, 5000);

  // Animate in
  setTimeout(() => notification.classList.add('show'), 100);
}

/**
 * Show desktop notification
 * @param {Object} message - The message object
 */
async function showDesktopNotification(message) {
  if (!('Notification' in window)) {
    console.debug('[ChatView] Desktop notifications not supported');
    return;
  }

  const hasPermission = await requestNotificationPermission();
  if (!hasPermission) {
    console.debug('[ChatView] Desktop notification permission denied');
    return;
  }

  const senderName = message.sender_name || 'Someone';
  const content = message.content || '';
  const truncatedContent = content.length > 100 ? content.substring(0, 97) + '...' : content;

  try {
    const notification = new Notification(`New message from ${senderName}`, {
      body: truncatedContent,
      icon: '/static/assets/logo.png',
      badge: '/static/assets/logo.png',
      tag: `chat-message-${message.conversation_id}`,
      requireInteraction: false,
      silent: true // We handle sound separately
    });

    // Auto-close after 5 seconds
    setTimeout(() => notification.close(), 5000);

    // Handle click to focus window and open conversation
    notification.onclick = () => {
      window.focus();
      if (message.conversation_id) {
        window.dispatchEvent(new CustomEvent('open-conversation', {
          detail: { conversationId: message.conversation_id }
        }));
      }
      notification.close();
    };

  } catch (error) {
    console.debug('[ChatView] Error showing desktop notification:', error);
  }
}

/**
 * Update page title with unread message count
 */
function updatePageTitleWithUnreadCount() {
  // Count unread conversations
  const unreadCount = document.querySelectorAll('.conversation-item.unread').length;
  const baseTitle = 'Real-Time Forum';

  if (unreadCount > 0) {
    document.title = `(${unreadCount}) ${baseTitle}`;
  } else {
    document.title = baseTitle;
  }
}

// ========================================
// TYPING INDICATORS
// ========================================

/**
 * Handle input focus events
 */
function handleInputFocus() {
    console.debug('[ChatView] Input focused');
    // Focus doesn't start typing indicator, only actual typing does
}

/**
 * Handle input blur events - stop typing indicator when user leaves input
 */
function handleInputBlur() {
    console.debug('[ChatView] Input blurred, stopping typing indicator');
    if (isTyping) {
        isTyping = false;
        sendTypingIndicator('stop');
        if (typingTimer) {
            clearTimeout(typingTimer);
            typingTimer = null;
        }
    }
}

// Debounced function to stop typing indicator
const debouncedStopTyping = debounce(() => {
    if (isTyping) {
        isTyping = false;
        sendTypingIndicator('stop');
        console.debug('[ChatView] Stopped typing indicator (debounced)');
    }
}, 2000);

/**
 * Handle typing events for typing indicators
 * @param {Event} event - The input or keydown event
 */
function handleTyping(event) {
    if (!currentChatUserId || !window.socket || window.socket.readyState !== WebSocket.OPEN) {
        return;
    }

    const messageInput = event.target;
    const hasContent = messageInput.value.trim().length > 0;

    // Send typing start indicator
    if (hasContent && !isTyping) {
        isTyping = true;
        sendTypingIndicator('start');
        console.debug('[ChatView] Started typing indicator');
    }

    // Use debounced function to stop typing after inactivity
    if (hasContent) {
        debouncedStopTyping();
    }

    // Stop typing immediately if input is empty
    if (!hasContent && isTyping) {
        isTyping = false;
        sendTypingIndicator('stop');
        console.debug('[ChatView] Stopped typing indicator (input empty)');
    }
}

/**
 * Send typing indicator via WebSocket
 * @param {string} action - 'start' or 'stop'
 */
function sendTypingIndicator(action) {
    if (!window.socket || window.socket.readyState !== WebSocket.OPEN) {
        return;
    }

    const message = {
        type: 'typing',
        recipient_id: currentChatUserId,
        conversation_id: currentChatConversationId,
        action: action
    };

    try {
        window.socket.send(JSON.stringify(message));
    } catch (error) {
        console.debug('[ChatView] Error sending typing indicator:', error);
    }
}

/**
 * Handle received typing indicators
 * @param {Object} data - The typing indicator data
 */
export function handleTypingIndicator(data) {
    if (!data.sender_id || data.sender_id === (getCurrentUser()?.userId || getCurrentUser()?.id)) {
        return; // Ignore our own typing indicators
    }

    // Only show typing indicator for current conversation
    if (data.conversation_id !== currentChatConversationId) {
        return;
    }

    const typingContainer = document.getElementById('chat-view-typing-indicator');
    if (!typingContainer) {
        return;
    }

    if (data.action === 'start') {
        showTypingIndicator(data.sender_name || 'Someone');
    } else if (data.action === 'stop') {
        hideTypingIndicator();
    }
}

/**
 * Show typing indicator in chat
 * @param {string} senderName - Name of the person typing
 */
function showTypingIndicator(senderName) {
    const typingContainer = document.getElementById('chat-view-typing-indicator');
    if (!typingContainer) {
        return;
    }

    typingContainer.innerHTML = `
        <div class="typing-indicator-content">
            <span class="typing-text">${escapeHtml(senderName)} is typing</span>
            <div class="typing-dots">
                <span></span>
                <span></span>
                <span></span>
            </div>
        </div>
    `;
    typingContainer.style.display = 'block';

    // Auto-hide after 5 seconds as fallback
    setTimeout(() => {
        hideTypingIndicator();
    }, 5000);
}

/**
 * Hide typing indicator
 */
function hideTypingIndicator() {
    const typingContainer = document.getElementById('chat-view-typing-indicator');
    if (typingContainer) {
        typingContainer.style.display = 'none';
        typingContainer.innerHTML = '';
    }
}

/**
 * Clean up typing indicators and timers (called on connection loss, page unload, etc.)
 */
export function cleanupTypingIndicators() {
    console.debug('[ChatView] Cleaning up typing indicators');

    // Stop our own typing indicator
    if (isTyping) {
        isTyping = false;
        // Don't try to send stop indicator if connection is lost
    }

    // Clear typing timer (debounced function handles its own cleanup)
    if (typingTimer) {
        clearTimeout(typingTimer);
        typingTimer = null;
    }

    // Hide any visible typing indicators
    hideTypingIndicator();
}

/**
 * Handle global keyboard shortcuts (works when chat view is open)
 * @param {KeyboardEvent} event - The keyboard event
 */
function handleGlobalKeyboardShortcuts(event) {
    // Only handle shortcuts when chat view is visible
    if (!chatViewVisible) {
        return;
    }

    // Escape to close chat (works globally when chat is open)
    if (event.key === 'Escape') {
        event.preventDefault();
        event.stopPropagation();
        console.debug('[ChatView] ESC key pressed - closing chat view');
        closeChatView();
        return;
    }

    // Ctrl/Cmd + K to focus message input (global shortcut)
    if ((event.ctrlKey || event.metaKey) && event.key === 'k') {
        event.preventDefault();
        event.stopPropagation();
        const messageInput = document.getElementById('chat-view-message');
        if (messageInput && !messageInput.disabled) {
            messageInput.focus();
            console.debug('[ChatView] Ctrl/Cmd+K pressed - focused message input');
        }
        return;
    }
}

/**
 * Handle keyboard shortcuts in chat input field
 * @param {KeyboardEvent} event - The keyboard event
 */
function handleKeyboardShortcuts(event) {
    // CRITICAL FIX: Proper Enter key handling with event bubbling prevention
    if (event.key === 'Enter' && !event.shiftKey) {
        event.preventDefault();
        event.stopPropagation(); // Prevent event bubbling that causes navigation issues
        const form = document.getElementById('chat-view-form');
        if (form) {
            form.dispatchEvent(new Event('submit'));
        }
    }
    // Shift+Enter: Allow new line (default behavior - no preventDefault needed)

    // Note: ESC and Ctrl/Cmd+K are now handled globally by handleGlobalKeyboardShortcuts
    // This prevents conflicts when input is not focused
}

// Flag to track if notification system has been initialized
let notificationSystemInitialized = false;

/**
 * Initialize notification system
 */
export function initNotificationSystem() {
  // Request permission on first load (safe to call multiple times)
  requestNotificationPermission();

  // Note: Notification settings are now handled in the header dropdown
  // No longer adding to chat sidebar

  // Only add the visibility change listener once
  if (!notificationSystemInitialized) {
    // Listen for page visibility changes to update title
    document.addEventListener('visibilitychange', () => {
      if (!document.hidden) {
        // Reset title when page becomes visible
        setTimeout(updatePageTitleWithUnreadCount, 100);
      }
    });

    notificationSystemInitialized = true;
    console.debug('[ChatView] Notification system initialized for the first time');
  } else {
    console.debug('[ChatView] Notification system already initialized, skipping duplicate setup');
  }
}

// Notification settings functions removed - now handled in header dropdown
