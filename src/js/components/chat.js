import {
  fetchUsers,
  fetchConversations,
} from "../utils/api.js";
import { addReceivedMessage, initNotificationSystem, handleTypingIndicator, cleanupTypingIndicators } from './chatView.js';
import { getUserFriendlyError, getErrorIcon, CHAT_ERRORS } from '../utils/errorMessages.js';
import { showEnhancedInlineError, showErrorToast } from './error.js';

const defaultAvatarPath = "/static/assets/default-avatar.png";

let currentUser = null;
let currentConversation = null;
let conversations = [];
let allUsers = [];
let socket = null;
let onlineUsers = new Set();
let isLoadingMoreMessages = false;
let typingUsers = new Map(); // Track typing users per conversation: conversationId -> {userId, userName, timestamp}
let unreadCounts = new Map(); // Track unread message counts per conversation: conversationId -> count

// ========================================
// UNIFIED CHAT STATE MANAGEMENT
// ========================================

/**
 * Centralized chat state management
 */
const ChatState = {
  current: {
    conversationId: null,
    recipientId: null,
    recipientName: null,
    recipientAvatar: null,
    isNewConversation: false,
    isRecipientOnline: false
  },

  /**
   * Update the current chat state
   * @param {Object} newState - New state properties
   */
  update(newState) {
    const oldState = { ...this.current };
    this.current = { ...this.current, ...newState };

    // Update legacy global references for backward compatibility
    window.currentChat = this.current;
    currentConversation = this.current.conversationId ? {
      id: this.current.conversationId,
      recipientId: this.current.recipientId,
      recipientName: this.current.recipientName
    } : null;

    console.debug('[ChatState] State updated:', {
      old: oldState,
      new: this.current
    });

    // Dispatch state change event
    window.dispatchEvent(new CustomEvent('chat-state-changed', {
      detail: { oldState, newState: this.current }
    }));

    return this.current;
  },

  /**
   * Clear the current chat state
   */
  clear() {
    return this.update({
      conversationId: null,
      recipientId: null,
      recipientName: null,
      recipientAvatar: null,
      isNewConversation: false,
      isRecipientOnline: false
    });
  },

  /**
   * Get the current state
   */
  get() {
    return { ...this.current };
  },

  /**
   * Check if a conversation is currently active
   * @param {number} conversationId - Conversation ID to check
   */
  isActiveConversation(conversationId) {
    return this.current.conversationId === conversationId;
  },

  /**
   * Check if a recipient is currently active
   * @param {number} recipientId - Recipient ID to check
   */
  isActiveRecipient(recipientId) {
    return this.current.recipientId === recipientId;
  }
};

// WebSocket Connection Manager - Singleton pattern to prevent multiple connections
class WebSocketConnectionManager {
  constructor() {
    this.socket = null;
    this.currentUser = null;
    this.connectionState = 'DISCONNECTED'; // DISCONNECTED, CONNECTING, CONNECTED, RECONNECTING
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
    this.reconnectDelay = 3000; // Reduced from 5000ms
    this.isInitialized = false;
    this.onlineUsers = new Set();

    // Add connection stability controls
    this.lastConnectionAttempt = 0;
    this.connectionCooldown = 2000; // Minimum time between connection attempts
    this.visibilityReconnectDelay = 5000; // Delay before reconnecting on visibility change
    this.focusReconnectDelay = 3000; // Delay before reconnecting on focus
    this.reconnectTimer = null;

    // Connection health monitoring
    this.lastPingTime = 0;
    this.pingInterval = null;
    this.pingTimeout = 30000; // 30 seconds ping timeout
    this.healthCheckInterval = 60000; // Check connection health every minute

    // Bind methods to preserve context
    this.connect = this.connect.bind(this);
    this.disconnect = this.disconnect.bind(this);
    this.handleBeforeUnload = this.handleBeforeUnload.bind(this);
    this.handleVisibilityChange = this.handleVisibilityChange.bind(this);

    // Set up page lifecycle event listeners
    this.setupPageLifecycleListeners();
  }

  setupPageLifecycleListeners() {
    // Clean up connection when page is about to unload
    window.addEventListener('beforeunload', this.handleBeforeUnload);
    window.addEventListener('unload', this.handleBeforeUnload);

    // Handle page visibility changes (tab switching, minimizing)
    document.addEventListener('visibilitychange', this.handleVisibilityChange);

    // Handle page focus/blur events with debouncing
    window.addEventListener('focus', () => {
      if (this.currentUser && this.connectionState === 'DISCONNECTED') {
        console.debug('[WSManager] Page focused, scheduling reconnection check');
        // Clear any existing timer
        if (this.reconnectTimer) {
          clearTimeout(this.reconnectTimer);
        }
        // Delay reconnection to avoid rapid cycling
        this.reconnectTimer = setTimeout(() => {
          if (this.currentUser && this.connectionState === 'DISCONNECTED') {
            console.info('[WSManager] Page focused, attempting to reconnect WebSocket');
            this.connect(this.currentUser);
          }
        }, this.focusReconnectDelay);
      }
    });
  }

  handleBeforeUnload() {
    console.info('[WSManager] Page unloading, closing WebSocket connection');

    // Clear all timers
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }

    // Force disconnect without reconnection
    this.disconnect(true);
  }

  handleVisibilityChange() {
    if (document.hidden) {
      console.debug('[WSManager] Page hidden, WebSocket will remain connected');
      // Clear any pending reconnection timers when page is hidden
      if (this.reconnectTimer) {
        clearTimeout(this.reconnectTimer);
        this.reconnectTimer = null;
      }
    } else {
      console.debug('[WSManager] Page visible, checking WebSocket connection');
      if (this.currentUser && this.connectionState === 'DISCONNECTED') {
        // Clear any existing timer
        if (this.reconnectTimer) {
          clearTimeout(this.reconnectTimer);
        }
        // Delay reconnection to avoid rapid cycling when switching tabs
        this.reconnectTimer = setTimeout(() => {
          if (this.currentUser && this.connectionState === 'DISCONNECTED' && !document.hidden) {
            console.info('[WSManager] Page visible, attempting to reconnect WebSocket');
            this.connect(this.currentUser);
          }
        }, this.visibilityReconnectDelay);
      }
    }
  }

  isConnected() {
    return this.socket && this.socket.readyState === WebSocket.OPEN;
  }

  isConnecting() {
    return this.socket && this.socket.readyState === WebSocket.CONNECTING;
  }

  connect(user) {
    if (!user) {
      console.error('[WSManager] Cannot connect: No user provided');
      return false;
    }

    const userId = user.userId || user.id;
    if (!userId) {
      console.error('[WSManager] Cannot connect: No user ID found');
      return false;
    }

    // Implement connection cooldown to prevent rapid attempts
    const now = Date.now();
    if (now - this.lastConnectionAttempt < this.connectionCooldown) {
      console.debug(`[WSManager] Connection attempt too soon, waiting ${this.connectionCooldown - (now - this.lastConnectionAttempt)}ms`);
      return false;
    }
    this.lastConnectionAttempt = now;

    // Check if already connected or connecting for the same user
    if (this.isConnected() || this.isConnecting()) {
      if (this.currentUser && (this.currentUser.userId === userId || this.currentUser.id === userId)) {
        console.info('[WSManager] WebSocket already connected/connecting for current user');
        return true;
      } else {
        console.info('[WSManager] Different user detected, closing existing connection');
        this.disconnect(true);
      }
    }

    this.currentUser = user;
    this.connectionState = 'CONNECTING';

    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${wsProtocol}//${window.location.host}/ws?user_id=${userId}`;

    console.info(`[WSManager] Connecting to WebSocket: ${wsUrl}`);

    try {
      this.socket = new WebSocket(wsUrl);

      this.socket.onopen = () => {
        console.info('[WSManager] WebSocket connection established successfully');
        this.connectionState = 'CONNECTED';
        this.reconnectAttempts = 0; // Reset reconnection attempts on successful connection

        // Clear any pending reconnection timers
        if (this.reconnectTimer) {
          clearTimeout(this.reconnectTimer);
          this.reconnectTimer = null;
        }

        // Update global references for backward compatibility
        socket = this.socket;
        window.socket = this.socket;

        // Start connection health monitoring
        this.startHealthMonitoring();

        // Request current online users
        this.sendMessage({ type: 'get_online_users' });
      };

      this.socket.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          console.debug('[WSManager] Message received:', {
            type: data.type,
            hasContent: !!data.content,
            senderId: data.sender_id,
            recipientId: data.recipient_id,
            conversationId: data.conversation_id
          });
          handleWebSocketMessage(data);
        } catch (e) {
          console.error('[WSManager] Message parse error:', {
            raw: event.data,
            error: e.message
          });
        }
      };

      this.socket.onclose = (event) => {
        const reason = event.reason ? ` Reason: ${event.reason}` : '';
        console.info(`[WSManager] WebSocket closed (code: ${event.code})${reason}`);

        this.connectionState = 'DISCONNECTED';
        this.socket = null;
        socket = null;
        window.socket = null;

        // Clear online users
        this.onlineUsers.clear();
        onlineUsers.clear();

        // Stop health monitoring
        this.stopHealthMonitoring();

        // Clean up typing indicators
        cleanupTypingIndicators();

        // Update UI
        const sidebarContainer = document.getElementById('chat-sidebar-right-container');
        if (sidebarContainer) updateConversationsListUI(sidebarContainer);

        // Handle different close codes appropriately
        if (event.code === 1000 || event.code === 1001) {
          // Normal closure or going away - don't reconnect
          console.info('[WSManager] WebSocket closed normally, not attempting reconnection');
          return;
        }

        if (event.code === 1006) {
          // Abnormal closure - could be network issue or server problem
          console.warn('[WSManager] WebSocket closed abnormally (1006) - checking if reconnection is appropriate');

          // Add a longer delay for abnormal closures to avoid rapid cycling
          const abnormalDelay = Math.max(this.reconnectDelay * 2, 5000);

          if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.connectionState = 'RECONNECTING';
            this.reconnectAttempts++;

            console.info(`[WSManager] Attempting reconnection ${this.reconnectAttempts}/${this.maxReconnectAttempts} in ${abnormalDelay}ms (abnormal closure)`);

            // Clear any existing timer
            if (this.reconnectTimer) {
              clearTimeout(this.reconnectTimer);
            }

            this.reconnectTimer = setTimeout(() => {
              if (this.currentUser && this.connectionState === 'RECONNECTING') {
                this.connect(this.currentUser);
              }
            }, abnormalDelay);
          }
          return;
        }

        // For other unexpected closures, attempt reconnection with normal logic
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
          this.connectionState = 'RECONNECTING';
          this.reconnectAttempts++;

          // Use linear backoff instead of exponential to avoid very long delays
          const delay = Math.min(this.reconnectDelay * this.reconnectAttempts, 15000); // Cap at 15 seconds
          console.info(`[WSManager] Attempting reconnection ${this.reconnectAttempts}/${this.maxReconnectAttempts} in ${delay}ms`);

          // Clear any existing timer
          if (this.reconnectTimer) {
            clearTimeout(this.reconnectTimer);
          }

          this.reconnectTimer = setTimeout(() => {
            if (this.currentUser && this.connectionState === 'RECONNECTING') {
              this.connect(this.currentUser);
            }
          }, delay);
        } else if (this.reconnectAttempts >= this.maxReconnectAttempts) {
          console.error('[WSManager] Max reconnection attempts reached, giving up');
        }
      };

      this.socket.onerror = (error) => {
        console.error('[WSManager] WebSocket error:', error);

        // If we get an error during connection, it might be an authentication issue
        if (this.connectionState === 'CONNECTING') {
          console.warn('[WSManager] Error during connection attempt - possible authentication failure');
          this.connectionState = 'DISCONNECTED';

          // Don't attempt immediate reconnection for connection errors
          // as they might be authentication-related
          this.reconnectAttempts++;

          if (this.reconnectAttempts < this.maxReconnectAttempts) {
            const delay = Math.max(this.reconnectDelay * 3, 10000); // Longer delay for auth errors
            console.info(`[WSManager] Scheduling reconnection attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts} in ${delay}ms`);

            this.reconnectTimer = setTimeout(() => {
              if (this.currentUser && this.connectionState === 'DISCONNECTED') {
                this.connect(this.currentUser);
              }
            }, delay);
          }
        }
      };

      return true;
    } catch (error) {
      console.error('[WSManager] Failed to create WebSocket connection:', error);
      this.connectionState = 'DISCONNECTED';
      return false;
    }
  }

  disconnect(force = false) {
    if (this.socket) {
      console.info(`[WSManager] Disconnecting WebSocket (force: ${force})`);

      if (force) {
        // Prevent reconnection attempts
        this.reconnectAttempts = this.maxReconnectAttempts;
      }

      this.socket.close(1000, 'Client disconnect');
      this.socket = null;
      socket = null;
      window.socket = null;
    }

    this.connectionState = 'DISCONNECTED';
    this.onlineUsers.clear();
    onlineUsers.clear();

    // Clear any pending reconnection timers
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }

    // Stop health monitoring
    this.stopHealthMonitoring();

    // Clean up typing indicators
    cleanupTypingIndicators();
  }

  startHealthMonitoring() {
    // Clear any existing health monitoring
    this.stopHealthMonitoring();

    // Start ping interval for connection health
    this.pingInterval = setInterval(() => {
      if (this.isConnected()) {
        this.lastPingTime = Date.now();
        // Send a ping message to keep connection alive
        this.sendMessage({ type: 'ping' });
      }
    }, this.healthCheckInterval);

    console.debug('[WSManager] Started connection health monitoring');
  }

  stopHealthMonitoring() {
    if (this.pingInterval) {
      clearInterval(this.pingInterval);
      this.pingInterval = null;
      console.debug('[WSManager] Stopped connection health monitoring');
    }
  }

  isConnectionHealthy() {
    if (!this.isConnected()) {
      return false;
    }

    // Check if we've sent a ping recently and haven't timed out
    const timeSinceLastPing = Date.now() - this.lastPingTime;
    return timeSinceLastPing < this.pingTimeout;
  }

  sendMessage(message) {
    if (this.isConnected()) {
      try {
        // Always send user status updates with the correct message type
        if (message.type === 'status') {
          message.type = 'user_status';
        }
        this.socket.send(JSON.stringify(message));
        console.debug('[WSManager] Message sent:', message.type);
        return true;
      } catch (e) {
        console.error('[WSManager] Error sending message:', {
          message: message,
          error: e.message
        });
        return false;
      }
    } else {
      console.warn(`[WSManager] Cannot send message - connection state: ${this.connectionState}`, message);
      return false;
    }
  }

  getConnectionState() {
    return {
      state: this.connectionState,
      isConnected: this.isConnected(),
      isConnecting: this.isConnecting(),
      isHealthy: this.isConnectionHealthy(),
      reconnectAttempts: this.reconnectAttempts,
      onlineUsersCount: this.onlineUsers.size,
      currentUser: this.currentUser,
      lastPingTime: this.lastPingTime,
      timeSinceLastPing: this.lastPingTime ? Date.now() - this.lastPingTime : null
    };
  }
}

// Create singleton instance
const wsManager = new WebSocketConnectionManager();

// Export WebSocket manager for external access and debugging
export { wsManager };

// Global cleanup function
export function cleanupChatConnections() {
  console.info('[Chat] Cleaning up chat connections');
  wsManager.disconnect(true);
}

// Debug function to check WebSocket connection status
export function getWebSocketStatus() {
  const state = wsManager.getConnectionState();
  console.info('[Chat] WebSocket Connection Status:', state);
  return state;
}

// Debug function to force reconnection (for testing)
export function forceWebSocketReconnect() {
  if (wsManager.currentUser) {
    console.info('[Chat] Forcing WebSocket reconnection');
    wsManager.disconnect(true);
    setTimeout(() => wsManager.connect(wsManager.currentUser), 1000);
  } else {
    console.warn('[Chat] Cannot force reconnect: No current user');
  }
}

// Utility function for debounce (removed - unused)

// Utility function for throttle
function throttle(func, limit) {
  let inThrottle;
  return function (...args) {
    const context = this;
    if (!inThrottle) {
      func.apply(context, args);
      inThrottle = true;
      setTimeout(() => (inThrottle = false), limit);
    }
  };
}

export function renderChatSidebarHTML() {
  return `
        <div class="chat-sidebar-controls md3-enhanced">
            <h3 id="chats-heading" class="md3-enhanced">Chats</h3>
        </div>
        <div id="conversation-list" class="conversation-list" aria-live="polite">
            <div class="loading-indicator md3-loading">
                <div class="md3-skeleton-avatar md3-skeleton"></div>
                <div class="md3-skeleton-text md3-skeleton"></div>
                <div class="md3-skeleton-text md3-skeleton"></div>
                <p>Loading chats...</p>
            </div>
        </div>
    `;
}

export function renderChatContainerHTML() {
  return `
        <div class="chat-container">
            <div class="chat-header">
                <div class="chat-user-info">
                    <div class="user-avatar">
                        <img src="${defaultAvatarPath}" alt="User avatar" id="chat-recipient-avatar">
                        <span class="status-indicator" id="chat-recipient-status"></span>
                    </div>
                    <div class="user-details">
                        <h3 id="chat-recipient-name">Select a conversation</h3>
                    </div>
                </div>
                 <button class="close-chat-btn" id="close-chat-view-btn">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="chat-messages" id="chat-messages">
                <div class="empty-chat-prompt">
                    <i class="far fa-comment-dots"></i>
                    <p>Select a conversation to start chatting</p>
                </div>
            </div>
            <div class="chat-input-container" id="chat-input-container">
                <textarea id="chat-message-input" placeholder="Type a message..." disabled></textarea>
                <button id="chat-send-button" class="btn btn-primary" disabled>
                    <i class="fas fa-paper-plane"></i>
                </button>
            </div>
        </div>
    `;
}

export async function initChatSidebar(containerElement) {
  if (!containerElement) {
    console.error("Chat sidebar container not found");
    return;
  }

  // Check if already initialized for the current user
  const user = getCurrentUser();
  if (wsManager.currentUser && user &&
      (wsManager.currentUser.userId === user.userId || wsManager.currentUser.id === user.id) &&
      wsManager.isConnected()) {
    console.info("Chat sidebar already initialized for current user, updating UI only");
    updateConversationsListUI(containerElement);
    return;
  }

  console.log("Initializing chat sidebar inside:", containerElement);

  try {
    // Get current user using the new method
    currentUser = user;

    // Initialize global variables early
    window.onlineUsers = onlineUsers;
    window.socket = socket;

    // Check if user is logged in
    if (!currentUser) {
      // Show login required message
      containerElement.innerHTML = `
                <div class="chat-login-required">
                    <i class="fas fa-lock"></i>
                    <p>Please log in to use chat</p>
                </div>
            `;
      console.error("No user logged in - chat requires authentication");
      return;
    }

    // Normalize user ID (handle both userId and id formats)
    currentUser.userId = currentUser.userId || currentUser.id;

    // Proceed with regular initialization
    setupSidebarEventListeners(containerElement);

    // Load users and conversations only if not already loaded or if user changed
    if (!allUsers.length || !wsManager.currentUser ||
        (wsManager.currentUser.userId !== currentUser.userId && wsManager.currentUser.id !== currentUser.id)) {
      allUsers = await loadAllUsers();
      window.allUsers = allUsers; // Make users available globally
      conversations = await loadConversations();
      window.conversations = conversations; // Make conversations available globally for debugging
      console.debug('[Chat] Loaded conversations:', conversations);
    }

    // Update the UI after loading data
    updateConversationsListUI(containerElement);

    // Initialize notification system
    initNotificationSystem();

    // Set up event listeners for auto-read functionality
    setupAutoReadEventListeners();

    // Connect WebSocket using the connection manager
    connectWebSocket();
  } catch (error) {
    console.error("Error initializing chat sidebar:", error);

    // Use enhanced error display
    const userFriendlyMessage = getUserFriendlyError(
      'CHAT_INIT_FAILED',
      error.message || 'Failed to initialize chat',
      'chat'
    );

    showEnhancedInlineError(
      userFriendlyMessage,
      'CHAT_INIT_FAILED',
      'chat',
      containerElement,
      {
        actionHandlers: {
          'retry': 'location.reload()',
          'home': 'window.location.href = "/#/home"'
        },
        scrollToError: false
      }
    );
  }
}

export async function initChatContainer(containerElement) {
  if (!containerElement) {
    console.error("Chat container element not found");
    return;
  }

  containerElement.innerHTML = renderChatContainerHTML();

  const chatMessagesContainer =
    containerElement.querySelector("#chat-messages");
  const messageInput = containerElement.querySelector("#chat-message-input");
  const sendButton = containerElement.querySelector("#chat-send-button");

  // PAGINATION FIX: Use throttle instead of debounce for better responsiveness
  chatMessagesContainer.addEventListener(
    "scroll",
    throttle(function () {
      if (
        chatMessagesContainer.scrollTop < 100 &&
        !isLoadingMoreMessages &&
        currentConversation
      ) {
        loadMoreMessages();
      }
    }, 300)
  );

  sendButton.addEventListener("click", function () {
    sendMessage();
  });

  messageInput.addEventListener("keydown", function (event) {
    // CRITICAL FIX: Proper Enter key handling with Shift+Enter support
    if (event.key === "Enter") {
      if (event.shiftKey) {
        // Shift+Enter: Allow new line (default behavior)
        return;
      } else {
        // Enter alone: Send message
        event.preventDefault();
        event.stopPropagation(); // Prevent event bubbling
        sendMessage();
      }
    }
  });
}

async function loadAllUsers() {
    try {
        console.log('[Chat] Loading all users...');
        const response = await fetchUsers({ withStatus: true });
        
        // Handle the response format properly
        let users;
        let onlineIds = new Set();
        if (response && typeof response === 'object' && response.data) {
            users = response.data;
        } else if (Array.isArray(response)) {
            users = response;
        } else {
            throw new Error('Invalid user data format received');
        }
        
        if (!Array.isArray(users)) {
            throw new Error('Users data is not an array');
        }
        
        // Filter out the current user and collect online status
        const filteredUsers = users.filter(user => {
            if (user.id === currentUser?.userId) return false;
            
            // Check for online status in the response
            if (user.status === 'online' || user.is_online || onlineUsers.has(user.id)) {
                onlineIds.add(user.id);
            }
            return true;
        });
        
        // Update online users set
        onlineUsers = new Set(onlineIds);
        
        console.log('[Chat] Successfully loaded users:', {
            total: filteredUsers.length,
            online: onlineIds.length
        });
        return filteredUsers;
    } catch (error) {
        console.error('Error loading all users:', error);
        throw error;
    }
}

async function loadConversations() {
    try {
        console.log('[Chat] Loading conversations...');
        const response = await fetchConversations({ withStatus: true });

        // Handle the response format properly
        let conversations;
        if (response && typeof response === 'object' && response.data) {
            conversations = response.data;
        } else if (Array.isArray(response)) {
            conversations = response;
        } else {
            throw new Error('Invalid conversation data format received');
        }

        if (!Array.isArray(conversations)) {
            throw new Error('Conversations data is not an array');
        }

        conversations.forEach(conversation => {
            if (conversation && conversation.id) {
                // Initialize unread count if not already set
                if (!unreadCounts.has(conversation.id)) {
                    // Check if conversation has unread messages based on last message
                    let unreadCount = 0;
                    if (conversation.last_message &&
                        conversation.last_message.sender_id !== currentUser?.userId &&
                        !conversation.last_message.is_read) {
                        unreadCount = conversation.unread_count || 1;
                    }
                    unreadCounts.set(conversation.id, unreadCount);

                    if (unreadCount > 0) {
                        console.debug(`[Chat] Initialized unread count for conversation ${conversation.id}: ${unreadCount}`);
                    }
                }
            }
        });

        console.log('[Chat] Successfully loaded conversations:', conversations.length);
        return conversations;
    } catch (error) {
        console.error('Error loading conversations:', error);
        throw error;
    }
}

function getSortTimestamp(item) {
  const lastMessageTime = item.conversation?.last_message?.sent_at;
  const conversationCreationTime = item.conversation?.created_at;
  let timestamp = 0;
  if (lastMessageTime) {
    timestamp = new Date(lastMessageTime).getTime();
  } else if (conversationCreationTime) {
    timestamp = new Date(conversationCreationTime).getTime();
  }
  return isNaN(timestamp) ? 0 : timestamp;
}

function getUserDisplayName(user) {
  if (!user) return "Unknown User";

  // Format as "First Name Last Name\n@username"
  const fullName = `${user.first_name || ""} ${user.last_name || ""}`.trim();

  if (fullName && user.username) {
    return `${fullName}<br>@${user.username}`;
  } else if (user.username) {
    return `@${user.username}`;
  } else if (fullName) {
    return fullName;
  } else {
    return "User";
  }
}

function compareChatItems(a, b) {
  // First sort by timestamp (most recent first)
  const timeA = getSortTimestamp(a);
  const timeB = getSortTimestamp(b);
  if (timeA !== timeB) {
    return timeB - timeA;
  }
  
  // Then sort by username (alphabetically)
  const usernameA = (a.user.username || "").toLowerCase();
  const usernameB = (b.user.username || "").toLowerCase();
  if (usernameA < usernameB) return -1;
  if (usernameA > usernameB) return 1;
  
  // If usernames are the same, use name as fallback
  const nameA = (`${a.user.first_name || ""} ${a.user.last_name || ""}`).trim().toLowerCase();
  const nameB = (`${b.user.first_name || ""} ${b.user.last_name || ""}`).trim().toLowerCase();
  if (nameA < nameB) return -1;
  if (nameA > nameB) return 1;
  
  return 0;
}

function updateConversationsListUI(sidebarContainer) {
  const conversationList = sidebarContainer.querySelector("#conversation-list");
  if (!conversationList) {
    console.error("Conversation list element not found for UI update.");
    return;
  }
  
  // Clear loading indicator
  conversationList.innerHTML = '';
  
  if (!Array.isArray(allUsers)) {
    console.error(
      "Cannot update UI: `allUsers` data is not available or not an array.",
      allUsers
    );
    conversationList.innerHTML =
      '<div class="error" style="padding: 1rem;">Error displaying user list.</div>';
    return;
  }

  // Filter out current user and create combined items
  const filteredUsers = allUsers.filter(user => user.id !== currentUser?.userId);
  
  console.debug('[Chat] Processing conversations for UI:', {
    conversationsCount: conversations?.length || 0,
    conversations: conversations,
    currentUserId: currentUser?.userId,
    filteredUsersCount: filteredUsers.length
  });

  const combinedItems = filteredUsers
    .map((user) => {
      const conversation = conversations.find(
        (conv) =>
          conv &&
          Array.isArray(conv.participants) &&
          conv.participants.length === 2 &&
          conv.participants.some((p) => p && p.id === user.id) &&
          conv.participants.some((p) => p && p.id === currentUser?.userId)
      );

      console.debug(`[Chat] User ${user.id} (${user.username}):`, {
        hasConversation: !!conversation,
        conversationId: conversation?.id,
        participants: conversation?.participants?.map(p => ({ id: p.id, username: p.username }))
      });

      return { user: user, conversation: conversation || null };
    })
    .filter((item) => item.user);

  let filteredItems = combinedItems;

  // Custom sort: 
  // 1. Online users with conversations (by last message time desc)
  // 2. Online users without conversations (alphabetical)
  // 3. Offline users with conversations (by last message time desc)
  // 4. Offline users without conversations (alphabetical)
  const onlineWithConv = [];
  const onlineNoConv = [];
  const offlineWithConv = [];
  const offlineNoConv = [];

  filteredItems.forEach((item) => {
    const isOnline = onlineUsers.has(item.user.id);
    const hasConv = !!item.conversation;
    if (isOnline && hasConv) onlineWithConv.push(item);
    else if (isOnline && !hasConv) onlineNoConv.push(item);
    else if (!isOnline && hasConv) offlineWithConv.push(item);
    else offlineNoConv.push(item);
  });

  // Sort with conversations by last message/conversation time desc
  onlineWithConv.sort(compareChatItems);
  offlineWithConv.sort(compareChatItems);

  // Sort without conversations alphabetically
  function alphaSort(a, b) {
    const nameA = (a.user.username || "").toLowerCase();
    const nameB = (b.user.username || "").toLowerCase();
    return nameA.localeCompare(nameB);
  }
  onlineNoConv.sort(alphaSort);
  offlineNoConv.sort(alphaSort);

  const sortedItems = [
    ...onlineWithConv,
    ...onlineNoConv,
    ...offlineWithConv,
    ...offlineNoConv,
  ];

  if (sortedItems.length === 0) {
    const message = "No other users available";
    conversationList.innerHTML = `<div class="no-conversations" style="padding: 1rem; text-align: center;"><i class="fas fa-users-slash"></i><p>${message}</p></div>`;
    return;
  }

  conversationList.innerHTML = sortedItems
    .map((item) => {
      const user = item.user;
      const conversation = item.conversation;
      const isOnline = onlineUsers.has(user.id);
      const conversationId = conversation?.id || null;
      const isActive =
        currentConversation?.id === conversationId && conversationId !== null;
      
      // Improve last message display
      let lastMessageText = "Start a conversation";
      let lastMessageTime = "";
      let hasUnreadMessages = false;
      let unreadCount = 0;

      if (conversation) {
        lastMessageText = "No messages yet";
        lastMessageTime = formatTime(conversation.created_at);

        // Get unread count for this conversation
        unreadCount = unreadCounts.get(conversationId) || 0;
        hasUnreadMessages = unreadCount > 0;

        if (conversation.last_message) {
          const msg = conversation.last_message;
          const isSentByMe = msg.sender_id === currentUser?.userId;

        // Add prefix to show who sent the message
        const prefix = isSentByMe ? "You: " : "";

        // Handle both string content and object with nested content
        const messageContent = typeof msg.content === 'string' ?
          msg.content :
          (msg.content?.text || msg.content?.message || "");

        lastMessageText = prefix + messageContent;

        // Truncate long messages
        if (lastMessageText.length > 28) {
            lastMessageText = lastMessageText.substring(0, 25) + "...";
        }

          lastMessageTime = formatTime(msg.sent_at);
        }
      }

      const displayName = getUserDisplayName(user);
      const avatarSrc = user.avatar?.Valid
        ? user.avatar.String
        : defaultAvatarPath;
const avatarHTML = `
                <div class="user-avatar ${isOnline ? "online" : ""}" title="${isOnline ? "Online" : "Offline"}">
                    <img src="${avatarSrc}" alt="${displayName}'s Avatar" onerror="this.onerror=null; this.src='${defaultAvatarPath}';">
                </div>`;
      const dataAttrs = `data-user-id="${user.id}" ${
        conversationId ? `data-conversation-id="${conversationId}"` : ""
      }`;

      return `
            <div class="conversation-item md3-enhanced ${isActive ? "active" : ""} ${
        hasUnreadMessages ? "unread" : ""
      }"
                 ${dataAttrs} role="button" tabindex="0" aria-label="Chat with ${displayName}" data-tooltip="Chat with ${displayName}">
                ${avatarHTML}
                <div class="conversation-info md3-enhanced">
                    <div class="conversation-name md3-enhanced">${displayName}</div>
                    <div class="conversation-last-message md3-enhanced">${lastMessageText}</div>
                </div>
                <div class="conversation-meta md3-enhanced">
                    <div class="conversation-time md3-enhanced">${lastMessageTime}</div>
                    ${
                      hasUnreadMessages && unreadCount > 0
                        ? `<div class="unread-count-badge md3-enhanced" title="${unreadCount} unread message${unreadCount > 1 ? 's' : ''}">${unreadCount}</div>`
                        : ""
                    }
                </div>
                <div class="conversation-ripple"></div>
            </div>
        `;
    })
    .join("");

  conversationList.querySelectorAll(".conversation-item").forEach((item) => {
    item.addEventListener("click", (e) => {
      // Create ripple effect
      createConversationRipple(item, e);
      handleConversationItemClick.call(item, e);
    });

    item.addEventListener("keydown", (e) => {
      if (e.key === "Enter" || e.key === " ") {
        createConversationRipple(item, e);
        handleConversationItemClick.call(item, e);
      }
    });

    // Add MD3 hover effects
    item.addEventListener("mouseenter", () => {
      if (window.PerformanceOptimizer) {
        window.PerformanceOptimizer.forceGPUAcceleration(item);
      }
    });

    item.addEventListener("mouseleave", () => {
      if (window.PerformanceOptimizer) {
        setTimeout(() => {
          if (!item.matches(':hover, :focus, :active')) {
            window.PerformanceOptimizer.resetGPUAcceleration(item);
          }
        }, 300);
      }
    });
  });
  
  console.debug(`[Chat] Updated conversations list UI with ${sortedItems.length} items`);

  // Update global unread counter
  updateGlobalUnreadCounter();

  // Auto-scroll to selected or unread conversation
  if (conversationList) {
    const selected = conversationList.querySelector('.conversation-item.selected');
    const unread = conversationList.querySelector('.conversation-item.unread');
    const target = selected || unread;
    if (target) {
      target.scrollIntoView({ block: "nearest" });
    }
  }
}

async function handleConversationItemClick(event) {
  const item = event.currentTarget;
  if (!item) return;

  try {
    // Parse IDs with better error handling
    const conversationIdStr = item.dataset.conversationId;
    const recipientIdStr = item.dataset.userId;
    
    const conversationId = conversationIdStr ? parseInt(conversationIdStr, 10) : null;
    const recipientId = recipientIdStr ? parseInt(recipientIdStr, 10) : null;

    console.debug("[Chat] Conversation click data:", {
      conversationIdStr: conversationIdStr,
      recipientIdStr: recipientIdStr,
      conversationId: conversationId,
      recipientId: recipientId,
      conversationIdValid: !isNaN(conversationId) && conversationId > 0,
      recipientIdValid: !isNaN(recipientId) && recipientId > 0
    });

    if ((!conversationId || isNaN(conversationId)) && (!recipientId || isNaN(recipientId))) {
      console.error("[Chat] Neither valid conversation ID nor user ID found in clicked item");
      return;
    }

    // Get recipient name from the DOM element first
    const recipientNameElement = item.querySelector('.conversation-name');
    const recipientName = recipientNameElement ? recipientNameElement.textContent : 'New Chat';

    // Update chat state using the unified state manager
    const recipient = allUsers.find(u => u.id === recipientId);
    const isRecipientOnline = onlineUsers.has(recipientId);

    ChatState.update({
        conversationId: conversationId,
        recipientId: recipientId,
        recipientName: recipientName,
        recipientAvatar: recipient?.avatar?.Valid ? recipient.avatar.String : '/static/assets/default-avatar.png',
        isNewConversation: !conversationId,
        isRecipientOnline: isRecipientOnline
    });

    // Update selected state in the sidebar with smooth animation
    const previousSelected = document.querySelector('.conversation-item.selected');
    if (previousSelected && previousSelected !== item) {
      previousSelected.classList.add('chat-deselecting');
      setTimeout(() => {
        previousSelected.classList.remove('selected', 'chat-deselecting');
      }, 300);
    }

    // Animate new selection
    item.classList.add('chat-selecting');
    setTimeout(() => {
      item.classList.add('selected');
      item.classList.remove('chat-selecting');
    }, 150);

    // Clear unread count for this conversation
    if (conversationId) {
      unreadCounts.set(conversationId, 0);
      item.classList.remove('unread');
      console.debug(`[Chat] Cleared unread count for conversation ${conversationId}`);

      // Update global counter
      updateGlobalUnreadCounter();
    }

    console.debug("[Chat] Dispatching toggle-chat-view event", {
      conversationId,
      recipientId,
      recipientName,
      chatState: ChatState.get()
    });

    // Update UI based on current state
    const currentState = ChatState.get();

    const messageInput = document.querySelector('#chat-message-input');
    const sendButton = document.querySelector('#chat-send-button');
    if (messageInput && sendButton) {
      messageInput.disabled = !currentState.isRecipientOnline;
      sendButton.disabled = !currentState.isRecipientOnline;
      messageInput.placeholder = currentState.isRecipientOnline ?
        "Type a message..." :
        "User is offline - messages can only be sent when online";
    }

    // Update status indicator in chat header
    const statusIndicator = document.querySelector('#chat-recipient-status');
    if (statusIndicator) {
      statusIndicator.className = `status-indicator ${currentState.isRecipientOnline ? 'online' : 'offline'}`;
      statusIndicator.title = currentState.isRecipientOnline ? "Online" : "Offline";
    }

    // Dispatch event for chat view with current state
    window.dispatchEvent(new CustomEvent('toggle-chat-view', {
      detail: ChatState.get()
    }));
  } catch (error) {
    console.error("[Chat] Error handling conversation click:", error);
  }
}

async function sendMessage() {
    const messageInput = document.querySelector('#chat-message-input');
    if (!messageInput || messageInput.disabled) return;

    const messageText = messageInput.value.trim();
    if (!messageText) return;

    const currentState = ChatState.get();
    const { conversationId, recipientId } = currentState;
    if (!recipientId) {
        console.error("[Chat] No recipient ID found for message");
        return;
    }

    // Verify recipient is online (shouldn't be possible to send if offline due to disabled input)
    if (!onlineUsers.has(recipientId)) {
        console.warn("[Chat] Attempted to send message to offline user");
        return;
    }

    try {
        if (recipientId && !conversationId) {
            // Create new conversation via WebSocket
            const message = {
                type: 'private',
                recipient_id: recipientId, // Fixed: Use snake_case to match backend JSON tag
                content: messageText,
                is_new_conversation: true // Fixed: Use snake_case to match backend JSON tag
            };
            console.debug('[Chat] Sending new conversation message:', message);
            sendWebSocketMessage(message);
        } else if (conversationId) {
            // Send message to existing conversation via WebSocket
            const message = {
                type: 'private',
                recipient_id: recipientId, // Fixed: Use snake_case to match backend JSON tag
                conversation_id: conversationId, // Fixed: Use snake_case to match backend JSON tag
                content: messageText
            };
            console.debug('[Chat] Sending existing conversation message:', message);
            sendWebSocketMessage(message);
        }

        // Clear input and focus
        messageInput.value = '';
        messageInput.focus();

        // CRITICAL FIX: Don't add optimistic message here - let the chatView handle it
        // The chatView system will handle message display to prevent conflicts
        console.debug('[Chat] Message sent via WebSocket, chatView will handle display');
    } catch (error) {
        console.error("Error sending message:", error);
    }
}

async function loadMoreMessages() {
  if (!currentConversation || isLoadingMoreMessages) return;

  isLoadingMoreMessages = true;
  const chatMessagesContainer = document.getElementById("chat-messages");

  // Remember current scroll height to maintain position
  const scrollHeight = chatMessagesContainer.scrollHeight;

  // Add loading indicator at the top
  const loadingIndicator = document.createElement("div");
  loadingIndicator.className = "loading-indicator loading-more";
  loadingIndicator.innerHTML = '<i class="fas fa-spinner fa-spin"></i>';
  chatMessagesContainer.prepend(loadingIndicator);

  try {
    // PAGINATION FIX: Use the actual number of messages we have as offset
    const currentMessageCount = chatMessagesContainer.children.length;
    const offset = currentMessageCount;

    const response = await fetch(
      `/api/messages?conversation_id=${currentConversation.id}&limit=10&offset=${offset}`
    );
    if (!response.ok) {
      throw new Error(`Failed to fetch more messages: ${response.status}`);
    }

    const messages = await response.json();
    console.debug(`[Chat] Loaded ${messages.length} older messages for conversation ${currentConversation.id}`);

    // Remove loading indicator
    loadingIndicator.remove();

    if (messages.length === 0) {
      // No more messages to load
      const noMoreMessages = document.createElement("div");
      noMoreMessages.className = "no-more-messages";
      noMoreMessages.textContent = "No more messages";
      chatMessagesContainer.prepend(noMoreMessages);
      setTimeout(() => noMoreMessages.remove(), 2000);
      return;
    }

    // PAGINATION FIX: Database returns messages in DESC order (newest first)
    // Reverse them to get chronological order for prepending
    const olderMessages = messages.reverse();

    // Prepend messages to the top of the chat container
    olderMessages.forEach(message => {
      const messageElement = document.createElement('div');
      messageElement.className = 'message';
      messageElement.innerHTML = `
        <div class="message-header">
          <span class="sender-name">${message.sender_name}</span>
          <span class="timestamp">${new Date(message.sent_at).toLocaleTimeString()}</span>
        </div>
        <div class="message-content">${message.content}</div>
      `;
      chatMessagesContainer.prepend(messageElement);
    });

    // Restore scroll position
    chatMessagesContainer.scrollTop =
      chatMessagesContainer.scrollHeight - scrollHeight;
  } catch (error) {
    console.error("[Chat] Error loading more messages:", error.message || error);
    loadingIndicator.innerHTML = "Failed to load more messages";
    setTimeout(() => loadingIndicator.remove(), 2000);
  } finally {
    isLoadingMoreMessages = false;
  }
}

function setupSidebarEventListeners() {
    // No-op, functionality moved
}

function setupAutoReadEventListeners() {
    // Listen for messages marked as read events from chatView
    window.addEventListener('messages-marked-read', (event) => {
        const { conversationId } = event.detail;
        console.debug('[Chat] Received messages-marked-read event for conversation:', conversationId);

        // Clear unread count for this conversation
        unreadCounts.delete(conversationId);

        // Remove unread class from conversation item
        const conversationItem = document.querySelector(
            `.conversation-item[data-conversation-id="${conversationId}"]`
        );
        if (conversationItem) {
            conversationItem.classList.remove('unread');
            // Remove unread count badge
            const unreadBadge = conversationItem.querySelector('.unread-count-badge');
            if (unreadBadge) {
                unreadBadge.remove();
            }
        }

        // Update global counter
        updateGlobalUnreadCounter();

        console.debug('[Chat] Updated unread status for conversation', conversationId);
    });

    window.addEventListener('chat-closed', () => {
        console.debug('[Chat] Received chat-closed event, clearing conversation selection');

        // Remove selected class from all conversation items
        document.querySelectorAll('.conversation-item.selected').forEach(item => {
            item.classList.remove('selected');
            console.debug('[Chat] Removed selected class from conversation item on chat close');
        });

        currentConversation = null;

        // Clear the unified chat state
        ChatState.clear();

        // Update conversations list UI if sidebar exists
        const sidebarContainer = document.getElementById('chat-sidebar-right-container');
        if (sidebarContainer) {
            updateConversationsListUI(sidebarContainer);
        }

        console.debug('[Chat] Chat state cleared after chat close, conversation list refreshed');
    });

    // Listen for open-conversation events from notifications
    window.addEventListener('open-conversation', (event) => {
        const { conversationId } = event.detail;
        console.debug('[Chat] Received open-conversation event for conversation:', conversationId);

        if (conversationId) {
            // Find the conversation in the global conversations array
            const conversation = conversations.find(conv => conv.id === conversationId);

            if (conversation) {
                // Trigger the same event that the sidebar uses to open conversations
                window.dispatchEvent(new CustomEvent('toggle-chat-view', {
                    detail: {
                        conversationId: conversation.id,
                        recipientId: conversation.participants?.[0]?.id || conversation.recipient_id,
                        recipientName: conversation.participants?.[0]?.name || conversation.recipient_name,
                        recipientAvatar: conversation.participants?.[0]?.avatar || conversation.recipient_avatar,
                        isRecipientOnline: conversation.participants?.[0]?.is_online || conversation.is_recipient_online,
                        isNewConversation: false
                    }
                }));
                console.debug(`[Chat] Opened conversation ${conversationId} from notification`);
            } else {
                console.warn(`[Chat] Conversation ${conversationId} not found in conversations array`);
            }
        }
    });
}


function connectWebSocket() {
  if (!currentUser) {
    console.error("[Chat] WebSocket connection failed: No user logged in");
    return;
  }

  // Use the WebSocket connection manager
  const success = wsManager.connect(currentUser);
  if (success) {
    console.info("[Chat] WebSocket connection initiated via connection manager");
  } else {
    console.error("[Chat] Failed to initiate WebSocket connection via connection manager");
  }
}

function sendWebSocketMessage(message) {
  const success = wsManager.sendMessage(message);
  if (!success) {
    const state = wsManager.getConnectionState();
    console.warn(
      `[Chat] Failed to send WebSocket message - connection state: ${state.state}`,
      message
    );
  }
}

function handleWebSocketMessage(data) {
  const sidebarContainer = document.getElementById(
    "chat-sidebar-right-container"
  );
  if (!sidebarContainer) {
    console.warn(
      "[Chat] Cannot handle WebSocket message: Sidebar container not found"
    );
    return;
  }
  
  if (!data || typeof data.type !== "string") {
    console.warn("[Chat] Invalid WebSocket message format:", data);
    return;
  }

  console.debug(`[Chat] Processing ${data.type} WebSocket message`);

  switch (data.type) {
    case "get_online_users":
      // This is just the request type, response will be handled by "online_users" case
      break;
    case "private":
    case "message":
      if (
        typeof data.sender_id === "number" &&
        typeof data.content === "string"
      ) {
        // Handle new conversation creation confirmation
        if (data.is_new_conversation && data.conversation_id) {
          ChatState.update({
            conversationId: data.conversation_id,
            isNewConversation: false
          });

          const selectedItem = document.querySelector('.conversation-item.selected');
          if (selectedItem) {
            selectedItem.dataset.conversationId = data.conversation_id;
            delete selectedItem.dataset.userId;
          }

          // CRITICAL FIX: Reload conversations for sender after creating new conversation
          console.debug('[Chat] New conversation created, reloading conversations for sender');
          loadConversations().then((newConversations) => {
            conversations = newConversations;
            window.conversations = conversations;
            updateConversationsListUI(sidebarContainer);
            console.debug(`[Chat] Sender conversations updated after new conversation creation: ${conversations.length} total`);
          }).catch(error => {
            console.error('[Chat] Failed to reload conversations for sender after new conversation:', error);
          });
        }

        handleNewMessage(data, sidebarContainer);
      } else {
        console.warn("[Chat] Malformed message data:", {
          conversationId: data.conversation_id,
          senderId: data.sender_id,
          hasContent: Boolean(data.content)
        });
      }
      break;
    case "user_status":
      if (data.content && typeof data.content.userId === "number" && typeof data.content.status === "string") {
        handleUserStatusUpdate({
          userId: data.content.userId,
          content: data.content.status
        }, sidebarContainer);
        
        // If user goes offline and we have an active chat with them, show notification
        const currentState = ChatState.get();
        if (data.content.status === "offline" &&
            currentState.recipientId === data.content.userId) {
          // Update state to reflect recipient is offline
          ChatState.update({ isRecipientOnline: false });

          const chatMessages = document.querySelector('#chat-messages');
          if (chatMessages) {
            const notification = document.createElement('div');
            notification.className = 'message-notification';
            notification.textContent = "User is now offline. Messages can only be sent when they are online.";
            chatMessages.appendChild(notification);
            setTimeout(() => notification.remove(), 5000);
          }
        } else if (data.content.status === "online" &&
                   currentState.recipientId === data.content.userId) {
          // Update state to reflect recipient is online
          ChatState.update({ isRecipientOnline: true });
        }
      } else {
        console.warn("[Chat] Malformed status update:", {
          userId: data.userId,
          status: data.content
        });
      }
      break;
    case "error":
      console.error("[Chat] WebSocket error received:", data.content, "Code:", data.code);

      // Get user-friendly error message
      const userFriendlyMessage = getUserFriendlyError(
        data.code || 'UNKNOWN_ERROR',
        data.content || 'An error occurred',
        'chat'
      );

      // Show error as toast notification for better UX
      showErrorToast(userFriendlyMessage, data.code || 'UNKNOWN_ERROR', 'chat', {
        duration: 5000
      });

      // Also handle specific error scenarios
      if (data.code === 'RECIPIENT_OFFLINE') {
        // Re-enable chat input if it was disabled
        const chatViewInput = document.querySelector('#chat-view-message');
        if (chatViewInput) {
          chatViewInput.disabled = false;
        }
      } else if (data.code === 'MESSAGE_SEND_FAILED') {
        // Re-enable chat input and focus it
        const chatViewInput = document.querySelector('#chat-view-message');
        if (chatViewInput) {
          chatViewInput.disabled = false;
          chatViewInput.focus();
        }
      }

      // For backward compatibility, also show in chat messages if available
      const chatMessages = document.querySelector('#chat-messages');
      if (chatMessages) {
        const errorElement = document.createElement('div');
        errorElement.className = 'message-error md3-enhanced';
        errorElement.innerHTML = `
          <i class="${getErrorIcon(data.code || 'UNKNOWN_ERROR', 'chat')}"></i>
          <span>${userFriendlyMessage}</span>
        `;
        chatMessages.appendChild(errorElement);
        setTimeout(() => errorElement.remove(), 5000);
      }
      break;
    case "online_users":
      if (data.content && Array.isArray(data.content.users)) {
        const validUserIds = data.content.users.filter((id) => typeof id === "number");
        onlineUsers = new Set(validUserIds);
        window.onlineUsers = onlineUsers; // Add this line
        
        console.info("[Chat] Updated online users list:", {
          count: onlineUsers.size,
          users: Array.from(onlineUsers)
        });
        updateConversationsListUI(sidebarContainer);
        
        // Update online status UI for current chat if exists
        const currentState = ChatState.get();
        if (currentState.recipientId) {
          const isRecipientOnline = onlineUsers.has(currentState.recipientId);

          // Update state with current online status
          ChatState.update({ isRecipientOnline });

          // Enable/disable input based on recipient's online status
          const messageInput = document.querySelector('#chat-message-input');
          const sendButton = document.querySelector('#chat-send-button');
          if (messageInput && sendButton) {
            messageInput.disabled = !isRecipientOnline;
            sendButton.disabled = !isRecipientOnline;

            // Update placeholder text
            messageInput.placeholder = isRecipientOnline ?
              "Type a message..." :
              "User is offline - messages can only be sent when online";
          }

          // Update status indicator
          const statusIndicator = document.querySelector('#chat-recipient-status');
          if (statusIndicator) {
            statusIndicator.className = `status-indicator ${isRecipientOnline ? 'online' : 'offline'}`;
          }
        }
      } else {
        console.warn("[Chat] Malformed online_users data:", data);
      }
      break;
    case "typing":
      if (data.sender_id && data.action) {
        handleTypingIndicator(data);
        handleSidebarTypingIndicator(data, sidebarContainer);
      } else {
        console.warn("[Chat] Malformed typing indicator data:", data);
      }
      break;
    case "new_conversation":
      if (data.conversation_id && data.sender_id) {
        handleNewConversationNotification(data, sidebarContainer);
      } else {
        console.warn("[Chat] Malformed new conversation data:", data);
      }
      break;
    case "ping":
      // Handle ping messages silently - server expects these for connection health
      console.debug(`[Chat] Processing ping WebSocket message`);
      break;
    case "pong":
      // Handle pong responses silently - part of connection health monitoring
      console.debug(`[Chat] Processing pong WebSocket message`);
      break;
    case "read_status":
      // CRITICAL FIX: Handle read status updates to update message indicators
      if (data.conversation_id && data.reader_id) {
        handleReadStatusUpdate(data);
      } else {
        console.warn("[Chat] Malformed read status data:", data);
      }
      break;
    default:
      console.warn(`[Chat] Unhandled WebSocket message type: ${data.type}`);
  }
}

function handleNewMessage(messageData, sidebarContainer) {
  if (!Array.isArray(conversations)) {
    console.error(
      "[Chat] Cannot process new message: conversations not initialized properly"
    );
    return;
  }

  const conversationId = messageData.conversation_id;
  let conversation = conversations.find(
    (conv) => conv && conv.id === conversationId
  );
  let isNewConversation = false;

  const newMessage = {
    id: messageData.id || messageData.message_id || Date.now(),
    conversation_id: conversationId,
    sender_id: messageData.sender_id,
    sender_name: messageData.sender_name || "User",
    content: messageData.content || "",
    sent_at:
      messageData.sent_at || messageData.timestamp || new Date().toISOString(),
    is_read: messageData.sender_id === currentUser?.userId,
  };

  if (conversation) {
    conversation.last_message = newMessage;
    console.debug(`[Chat] Updated last message for conversation ${conversationId}`);
  } else {
    isNewConversation = true;
    console.info(
      `[Chat] Received message for unknown conversation ${conversationId}. Refreshing conversations.`
    );
  }

  const conversationItem = sidebarContainer.querySelector(
    `.conversation-item[data-conversation-id="${conversationId}"]`
  );
  
  // CRITICAL FIX: Only mark as unread if message is from someone else and conversation is not currently active
  if (newMessage.sender_id !== currentUser?.userId) {
    // Check if this conversation is currently active (being viewed)
    const isCurrentlyActive = currentConversation?.id === conversationId;

    console.debug(`[Chat] Checking if conversation ${conversationId} is active:`, {
      currentConversationId: currentConversation?.id,
      isCurrentlyActive: isCurrentlyActive,
      chatStateConversationId: ChatState.current.conversationId,
      chatStateActive: ChatState.isActiveConversation(conversationId)
    });

    if (!isCurrentlyActive) {
      // Add unread styling to conversation item if it exists
      if (conversationItem) {
        conversationItem.classList.add("unread");
      }

      // Update unread count
      const currentCount = unreadCounts.get(conversationId) || 0;
      unreadCounts.set(conversationId, currentCount + 1);

      console.debug(`[Chat] Marked conversation ${conversationId} as unread (count: ${currentCount + 1})`);

      // Update global counter
      updateGlobalUnreadCounter();
    } else {
      console.debug(`[Chat] Message received for active conversation ${conversationId}, not marking as unread`);
    }
  }

  if (isNewConversation) {
    // CRITICAL FIX: Reload conversations and update global array for new conversations
    loadConversations().then((newConversations) => {
      conversations = newConversations;
      window.conversations = conversations;
      updateConversationsListUI(sidebarContainer);
      console.debug(`[Chat] Updated conversations array after new conversation: ${conversations.length} total`);
    }).catch(error => {
      console.error('[Chat] Failed to reload conversations after new message:', error);
      // Fallback to just updating UI with existing data
      updateConversationsListUI(sidebarContainer);
    });
  } else {
    updateConversationsListUI(sidebarContainer);
  }

  if (currentConversation?.id === conversationId) {
    console.debug(`[Chat] Dispatching new-message-received event for active conversation ${conversationId}`);
    window.dispatchEvent(
      new CustomEvent("new-message-received", { detail: newMessage })
    );
  }

  // CRITICAL FIX: Only call addReceivedMessage for the current conversation
  // and distinguish between new messages and sender confirmations
  if (currentConversation?.id === conversationId) {
    console.debug(`[Chat] Processing message for active conversation ${conversationId}`);
    addReceivedMessage(messageData);
  } else {
    console.debug(`[Chat] Message for inactive conversation ${conversationId}, not updating chat view`);
  }

  // GLOBAL NOTIFICATION FIX: Show notification for ALL new messages (not just when chat is open)
  // This ensures notifications appear everywhere in the SPA
  if (window.showGlobalChatNotification && messageData.sender_id !== currentUser?.userId) {
    console.debug(`[Chat] Showing global notification for message from ${messageData.sender_name}`);
    window.showGlobalChatNotification(messageData, {
      showDesktop: true,
      playSound: true,
      autoHide: true,
      duration: 8000
    });
  }
}

function handleUserStatusUpdate(statusData, sidebarContainer) {
  console.debug("[Chat] Processing status update:", statusData);
  
  const userId = statusData.userId;
  const status = statusData.content;

  if (userId === currentUser?.userId) return;

  let statusChanged = false;
  if (status === "online") {
    if (!onlineUsers.has(userId)) {
      onlineUsers.add(userId);
      statusChanged = true;
      console.debug(`[Chat] User ${userId} is now online`);
    }
  } else if (status === "offline") {
    if (onlineUsers.has(userId)) {
      onlineUsers.delete(userId);
      statusChanged = true;
      console.debug(`[Chat] User ${userId} is now offline`);
    }
  }

  if (statusChanged) {
    console.info(`[Chat] User ${userId} status changed to ${status} (Online users: ${onlineUsers.size})`);
    window.onlineUsers = onlineUsers; // Add this line
    updateConversationsListUI(sidebarContainer);

    // Update UI if this is the current chat recipient
    const currentState = ChatState.get();
    if (currentState.recipientId === userId) {
      const isRecipientOnline = onlineUsers.has(userId);

      // Update state with new online status
      ChatState.update({ isRecipientOnline });

      // Update input state
      const messageInput = document.querySelector('#chat-message-input');
      const sendButton = document.querySelector('#chat-send-button');
      if (messageInput && sendButton) {
        messageInput.disabled = !isRecipientOnline;
        sendButton.disabled = !isRecipientOnline;
        messageInput.placeholder = isRecipientOnline ?
          "Type a message..." :
          "User is offline - messages can only be sent when online";
      }

      // Update status indicator and text
      const statusIndicator = document.querySelector('#chat-recipient-status');
      const statusText = document.querySelector('#chat-view-status-text');
      if (statusIndicator) {
        statusIndicator.className = `status-indicator ${isRecipientOnline ? 'online' : 'offline'}`;
        statusIndicator.title = isRecipientOnline ? "Online" : "Offline";
      }
      if (statusText) {
        statusText.textContent = isRecipientOnline ? 'Online' : 'Offline';
        statusText.className = `user-status ${isRecipientOnline ? 'online' : 'offline'}`;
      }
    }
  }
}

function updateGlobalUnreadCounter() {
  const chatsHeading = document.getElementById('chats-heading');
  if (!chatsHeading) return;

  // Count conversations with unread messages
  let unreadConversationsCount = 0;
  for (const [, count] of unreadCounts.entries()) {
    if (count > 0) {
      unreadConversationsCount++;
    }
  }

  // Update heading text
  if (unreadConversationsCount > 0) {
    chatsHeading.textContent = `Chats (${unreadConversationsCount})`;
  } else {
    chatsHeading.textContent = 'Chats';
  }

  console.debug(`[Chat] Updated global unread counter: ${unreadConversationsCount} conversations with unread messages`);
}

function handleNewConversationNotification(data, sidebarContainer) {
  console.info(`[Chat] New conversation notification received: conversation ${data.conversation_id} from user ${data.sender_id}`);

  // CRITICAL FIX: Reload conversations and update global array
  loadConversations().then((newConversations) => {
    console.debug('[Chat] Conversations reloaded after new conversation notification');

    // Update the global conversations array
    conversations = newConversations;
    window.conversations = conversations;

    // Update the UI with the new conversations
    updateConversationsListUI(sidebarContainer);

    console.debug(`[Chat] Updated conversations array with ${conversations.length} conversations`);
  }).catch(error => {
    console.error('[Chat] Failed to reload conversations after new conversation notification:', error);
  });

  // Show a notification to the user
  if (window.showNotification) {
    const senderName = data.sender_name || `User ${data.sender_id}`;
    window.showNotification(`New conversation started by ${senderName}`, 'info');
  }
}

// CRITICAL FIX: Handle read status updates from WebSocket
function handleReadStatusUpdate(data) {
  console.debug('[Chat] Read status update received:', {
    conversationId: data.conversation_id,
    readerId: data.reader_id,
    readerName: data.reader_name
  });

  // CRITICAL FIX: Update message status indicators in the chat view if it's the active conversation
  if (window.updateMessageStatusIndicators) {
    window.updateMessageStatusIndicators(data.conversation_id);
    console.debug('[Chat] Called updateMessageStatusIndicators for conversation:', data.conversation_id);
  } else {
    console.warn('[Chat] updateMessageStatusIndicators function not available');
  }

  // Also try to import and call the function directly
  try {
    if (window.ChatView && typeof window.ChatView.updateMessageStatusIndicators === 'function') {
      window.ChatView.updateMessageStatusIndicators(data.conversation_id);
      console.debug('[Chat] Called ChatView.updateMessageStatusIndicators');
    }
  } catch (error) {
    console.debug('[Chat] Error calling ChatView.updateMessageStatusIndicators:', error);
  }

  // CRITICAL FIX: Also dispatch a custom event for any other listeners
  window.dispatchEvent(new CustomEvent('read-status-update', {
    detail: {
      conversationId: data.conversation_id,
      readerId: data.reader_id,
      readerName: data.reader_name
    }
  }));
}

function handleSidebarTypingIndicator(data, sidebarContainer) {
  if (!sidebarContainer) return;

  const senderId = data.sender_id || data.user_id;
  const senderName = data.sender_name || `User ${senderId}`;
  const action = data.action;
  const conversationId = data.conversation_id;

  console.debug(`[Chat] Sidebar typing indicator: ${senderName} ${action} in conversation ${conversationId}`);

  // Find the conversation item in the sidebar
  let conversationItem = null;
  if (conversationId) {
    conversationItem = sidebarContainer.querySelector(
      `.conversation-item[data-conversation-id="${conversationId}"]`
    );
  } else {
    // If no conversation ID, find by user ID (for new conversations)
    conversationItem = sidebarContainer.querySelector(
      `.conversation-item[data-user-id="${senderId}"]`
    );
  }

  if (!conversationItem) {
    console.debug(`[Chat] No conversation item found for typing indicator from user ${senderId}`);
    return;
  }

  const lastMessageElement = conversationItem.querySelector('.conversation-last-message');
  if (!lastMessageElement) return;

  if (action === 'start') {
    // Store original message for restoration
    if (!conversationItem.dataset.originalMessage) {
      conversationItem.dataset.originalMessage = lastMessageElement.textContent;
    }

    // Add typing class and update message
    conversationItem.classList.add('typing');
    lastMessageElement.innerHTML = `
      <span class="typing-text">${senderName} is typing</span>
      <span class="typing-indicator">
        <span></span>
        <span></span>
        <span></span>
      </span>
    `;

    // Track typing user
    typingUsers.set(conversationId || `user_${senderId}`, {
      userId: senderId,
      userName: senderName,
      timestamp: Date.now()
    });

  } else if (action === 'stop') {
    // Remove typing class and restore original message
    conversationItem.classList.remove('typing');

    if (conversationItem.dataset.originalMessage) {
      lastMessageElement.textContent = conversationItem.dataset.originalMessage;
      delete conversationItem.dataset.originalMessage;
    }

    // Remove from typing users
    typingUsers.delete(conversationId || `user_${senderId}`);
  }
}

export function formatTime(dateString) {
  if (!dateString) return "";
  try {
    const date = new Date(dateString);
    if (isNaN(date.getTime())) {
      console.warn("Invalid date string received for formatting:", dateString);
      return "";
    }
    const now = new Date();
    const diffSeconds = Math.floor((now - date) / 1000);
    const diffDays = Math.floor(diffSeconds / (60 * 60 * 24));
    const isToday =
      date.getDate() === now.getDate() &&
      date.getMonth() === now.getMonth() &&
      date.getFullYear() === now.getFullYear();
    const yesterday = new Date(now);
    yesterday.setDate(now.getDate() - 1);
    const isYesterday =
      date.getDate() === yesterday.getDate() &&
      date.getMonth() === yesterday.getMonth() &&
      date.getFullYear() === yesterday.getFullYear();

    if (isToday) {
      return date.toLocaleTimeString("en-US", {
        hour: "numeric",
        minute: "2-digit",
        hour12: true,
      });
    } else if (isYesterday) {
      return `Yesterday ${date.toLocaleTimeString("en-US", {
        hour: "numeric",
        minute: "2-digit",
        hour12: true,
      })}`;
    } else if (diffDays < 7) {
      return `${date.toLocaleDateString("en-US", { weekday: "short" })} ${date.toLocaleTimeString("en-US", {
        hour: "numeric",
        minute: "2-digit",
        hour12: true,
      })}`;
    } else {
      return `${date.toLocaleDateString("en-US", {
        month: "short",
        day: "numeric",
        year: "numeric",
      })} ${date.toLocaleTimeString("en-US", {
        hour: "numeric",
        minute: "2-digit",
        hour12: true,
      })}`;
    }
  } catch (e) {
    console.error("Error formatting time:", dateString, e);
    return "";
  }
}

// Add a more reliable method to check the current user
function getCurrentUser() {
  // First check for user data in sessionStorage (more secure for authentication)
  const userData = sessionStorage.getItem("user");

  // If not in sessionStorage, try localStorage as fallback
  if (!userData && localStorage.getItem("user")) {
    return JSON.parse(localStorage.getItem("user"));
  }

  // If we have userData in sessionStorage, parse and return it
  if (userData) {
    return JSON.parse(userData);
  }

  // As a last resort, check for auth token
  const token =
    sessionStorage.getItem("authToken") || localStorage.getItem("authToken");
  if (token) {
    // If we have a token but no user data, we should fetch the user data from the server
    // For now, return a minimal user object to prevent the error
    return { id: "authenticated", token: token };
  }

  return null;
}

// All unused functions removed for cleanup

/**
 * Create MD3 ripple effect for conversation items
 * @param {HTMLElement} element - The conversation item element
 * @param {Event} event - The click event
 */
function createConversationRipple(element, event) {
  const rippleContainer = element.querySelector('.conversation-ripple');
  if (!rippleContainer) return;

  const rect = element.getBoundingClientRect();
  const size = Math.max(rect.width, rect.height);
  const x = event.clientX - rect.left - size / 2;
  const y = event.clientY - rect.top - size / 2;

  const ripple = document.createElement('div');
  ripple.className = 'ripple-effect';
  ripple.style.cssText = `
    position: absolute;
    width: ${size}px;
    height: ${size}px;
    left: ${x}px;
    top: ${y}px;
    background: currentColor;
    border-radius: 50%;
    opacity: 0.2;
    transform: scale(0);
    animation: md3Ripple 0.6s ease-out;
    pointer-events: none;
  `;

  rippleContainer.appendChild(ripple);

  // Remove ripple after animation
  setTimeout(() => {
    if (ripple.parentNode) {
      ripple.parentNode.removeChild(ripple);
    }
  }, 600);
}
