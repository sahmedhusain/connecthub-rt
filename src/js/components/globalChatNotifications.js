/**
 * Global Chat Notification System - Material Design 3 Enhanced
 * Handles real-time chat message notifications throughout the entire SPA
 */

// Notification configuration
const NOTIFICATION_CONFIG = {
    maxNotifications: 3,
    autoHideDuration: 8000,
    animationDuration: 500,
    soundEnabled: true,
    desktopEnabled: true
};

// Active notifications tracking
let activeNotifications = new Map();
let notificationCounter = 0;

/**
 * Initialize the global chat notification system
 */
export function initGlobalChatNotifications() {
    console.debug('[GlobalChatNotifications] Initializing global chat notification system');
    
    // Create notifications container if it doesn't exist
    getOrCreateNotificationsContainer();
    
    // Request desktop notification permission
    requestDesktopNotificationPermission();
    
    // Make functions globally available
    window.showGlobalChatNotification = showGlobalChatNotification;
    window.hideGlobalChatNotification = hideGlobalChatNotification;
    window.hideAllGlobalChatNotifications = hideAllGlobalChatNotifications;
    
    console.debug('[GlobalChatNotifications] Global chat notification system initialized');
}

/**
 * Show a global chat message notification
 * @param {Object} message - The message object
 * @param {Object} options - Notification options
 */
export function showGlobalChatNotification(message, options = {}) {
    // Get user preferences from localStorage (same as chatView.js)
    const preferences = getNotificationPreferences();

    const {
        showDesktop = preferences.desktop,
        playSound = preferences.sound,
        autoHide = true,
        duration = NOTIFICATION_CONFIG.autoHideDuration
    } = options;

    // Don't show notifications for our own messages
    const currentUser = getCurrentUser();
    if (message.sender_id === (currentUser?.userId || currentUser?.id)) {
        console.debug('[GlobalChatNotifications] Skipping notification for own message');
        return;
    }

    console.debug('[GlobalChatNotifications] Showing notification for message from:', message.sender_name);

    // Remove oldest notification if we have too many
    if (activeNotifications.size >= NOTIFICATION_CONFIG.maxNotifications) {
        const oldestId = Array.from(activeNotifications.keys())[0];
        hideGlobalChatNotification(oldestId);
    }

    // Create visual notification (only if visual notifications are enabled)
    let notificationId = null;
    if (preferences.visual) {
        notificationId = createVisualNotification(message, { autoHide, duration });
    }

    // Show desktop notification
    if (showDesktop) {
        showDesktopNotification(message);
    }

    // Play sound with user preferences
    if (playSound) {
        playNotificationSound(preferences.volume || 0.3, preferences.soundType || 'default');
    }

    return notificationId;
}

/**
 * Create visual in-app notification
 * @param {Object} message - The message object
 * @param {Object} options - Options
 * @returns {string} Notification ID
 */
function createVisualNotification(message, options = {}) {
    const notificationId = `global-chat-notification-${++notificationCounter}`;
    const senderName = message.sender_name || 'Someone';
    const content = message.content || '';
    const truncatedContent = content.length > 80 ? content.substring(0, 77) + '...' : content;
    const timestamp = formatNotificationTime(message.timestamp || new Date());

    // Get sender avatar from global users array
    const allUsers = window.allUsers || [];
    const senderUser = allUsers.find(u => u.id === message.sender_id);
    const senderAvatar = senderUser?.avatar?.Valid ? senderUser.avatar.String : '/static/assets/default-avatar.png';
    const initials = getInitials(senderName);

    // Create notification element
    const notification = document.createElement('div');
    notification.className = 'chat-notification md3-enhanced';
    notification.id = notificationId;
    notification.setAttribute('role', 'alert');
    notification.setAttribute('aria-live', 'polite');

    notification.innerHTML = `
        <div class="chat-notification-content">
            <div class="chat-notification-header">
                <div class="chat-notification-sender">
                    <div class="chat-notification-avatar-container">
                        <img src="${escapeHtml(senderAvatar)}"
                             alt="${escapeHtml(senderName)}'s Avatar"
                             class="chat-notification-avatar-img"
                             onerror="this.style.display='none'; this.nextElementSibling.style.display='flex';">
                        <div class="chat-notification-avatar-fallback" style="display: none;">${escapeHtml(initials)}</div>
                    </div>
                    <div class="chat-notification-sender-info">
                        <div class="chat-notification-sender-name">${escapeHtml(senderName)}</div>
                        <div class="chat-notification-time">${timestamp}</div>
                    </div>
                </div>
                <button class="chat-notification-close" aria-label="Close notification">
                    <i class="fas fa-times" aria-hidden="true"></i>
                </button>
            </div>
            <div class="chat-notification-message">${escapeHtml(truncatedContent)}</div>
            <div class="chat-notification-actions">
                <button class="chat-notification-action primary">
                    <i class="fas fa-eye" aria-hidden="true"></i>
                    View Chat
                </button>
            </div>
        </div>
    `;

    // Add event listeners
    setupNotificationEventListeners(notification, message, notificationId);

    // Add to container
    const container = getOrCreateNotificationsContainer();
    container.appendChild(notification);

    // Store notification reference
    activeNotifications.set(notificationId, {
        element: notification,
        message: message,
        createdAt: Date.now()
    });

    // Animate in
    requestAnimationFrame(() => {
        notification.classList.add('show');
    });

    // Auto-hide if enabled
    if (options.autoHide) {
        setTimeout(() => {
            hideGlobalChatNotification(notificationId);
        }, options.duration || NOTIFICATION_CONFIG.autoHideDuration);
    }

    console.debug(`[GlobalChatNotifications] Created notification: ${notificationId}`);
    return notificationId;
}

/**
 * Set up event listeners for notification
 * @param {HTMLElement} notification - Notification element
 * @param {Object} message - Message object
 * @param {string} notificationId - Notification ID
 */
function setupNotificationEventListeners(notification, message, notificationId) {
    // Close button
    const closeBtn = notification.querySelector('.chat-notification-close');
    closeBtn.addEventListener('click', (e) => {
        e.stopPropagation();
        hideGlobalChatNotification(notificationId);
    });

    // View action (only button now)
    const viewBtn = notification.querySelector('.chat-notification-action.primary');
    if (viewBtn) {
        viewBtn.addEventListener('click', (e) => {
            e.stopPropagation();
            openConversation(message.conversation_id, message);
            hideGlobalChatNotification(notificationId);
        });
    }

    // Click on notification to open conversation
    notification.addEventListener('click', () => {
        openConversation(message.conversation_id, message);
        hideGlobalChatNotification(notificationId);
    });

    // Keyboard accessibility
    notification.addEventListener('keydown', (e) => {
        if (e.key === 'Escape') {
            hideGlobalChatNotification(notificationId);
        } else if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault();
            openConversation(message.conversation_id, message);
            hideGlobalChatNotification(notificationId);
        }
    });
}

/**
 * Hide a specific notification
 * @param {string} notificationId - Notification ID
 */
export function hideGlobalChatNotification(notificationId) {
    const notificationData = activeNotifications.get(notificationId);
    if (!notificationData) return;

    console.debug(`[GlobalChatNotifications] Hiding notification: ${notificationId}`);

    const { element } = notificationData;
    
    // Add fade-out class
    element.classList.add('fade-out');

    // Remove from DOM after animation
    setTimeout(() => {
        if (element.parentNode) {
            element.parentNode.removeChild(element);
        }
        activeNotifications.delete(notificationId);
    }, NOTIFICATION_CONFIG.animationDuration);
}

/**
 * Hide all active notifications
 */
export function hideAllGlobalChatNotifications() {
    const notificationIds = Array.from(activeNotifications.keys());
    notificationIds.forEach(id => hideGlobalChatNotification(id));
}

/**
 * Get or create notifications container
 * @returns {HTMLElement} Notifications container
 */
function getOrCreateNotificationsContainer() {
    let container = document.querySelector('.chat-notifications-container');
    
    if (!container) {
        container = document.createElement('div');
        container.className = 'chat-notifications-container';
        container.setAttribute('aria-live', 'polite');
        container.setAttribute('aria-label', 'Chat notifications');
        document.body.appendChild(container);
    }
    
    return container;
}

/**
 * Open conversation by triggering the proper chat system events
 * @param {number} conversationId - Conversation ID
 * @param {Object} message - Optional message object for fallback
 */
function openConversation(conversationId, message = null) {
    if (!conversationId) return;

    console.debug(`[GlobalChatNotifications] Opening conversation: ${conversationId}`);

    // Find the conversation in the global conversations array
    const conversations = window.conversations || [];
    const conversation = conversations.find(conv => conv.id === conversationId);

    if (conversation) {
        // Extract proper recipient information
        const currentUser = getCurrentUser();
        const currentUserId = currentUser?.userId || currentUser?.id;

        // Find the other participant (not the current user)
        let recipientInfo = null;

        if (conversation.participants && Array.isArray(conversation.participants)) {
            recipientInfo = conversation.participants.find(p => p.id !== currentUserId);
        }

        // Fallback to direct properties if participants array is not available
        if (!recipientInfo) {
            recipientInfo = {
                id: conversation.recipient_id || conversation.other_user_id,
                name: conversation.recipient_name || conversation.other_user_name,
                avatar: conversation.recipient_avatar || conversation.other_user_avatar,
                is_online: conversation.is_recipient_online || conversation.other_user_online
            };
        }

        // Get user info from global users array if available (same as sidebar)
        const allUsers = window.allUsers || [];
        const onlineUsers = window.onlineUsers || new Set();
        const userDetails = allUsers.find(u => u.id === recipientInfo?.id);

        if (userDetails) {
            recipientInfo = {
                id: userDetails.id,
                name: `${userDetails.firstName || ''} ${userDetails.lastName || ''}`.trim() || userDetails.username,
                avatar: userDetails.avatar?.Valid ? userDetails.avatar.String : '/static/assets/default-avatar.png',
                is_online: onlineUsers.has(userDetails.id) // Use actual online status like sidebar
            };
        }

        // Update ChatState exactly like the sidebar does
        if (window.ChatState) {
            window.ChatState.update({
                conversationId: conversation.id,
                recipientId: recipientInfo?.id,
                recipientName: recipientInfo?.name || 'Unknown User',
                recipientAvatar: recipientInfo?.avatar || '/static/assets/default-avatar.png',
                isNewConversation: false,
                isRecipientOnline: recipientInfo?.is_online || false
            });
        }

        // Update sidebar selection state exactly like sidebar click does
        const conversationItems = document.querySelectorAll('.conversation-item');
        conversationItems.forEach(item => {
            const itemConversationId = parseInt(item.dataset.conversationId, 10);
            const itemUserId = parseInt(item.dataset.userId, 10);

            if ((itemConversationId === conversation.id) || (itemUserId === recipientInfo?.id)) {
                // Remove previous selection with animation
                const previousSelected = document.querySelector('.conversation-item.selected');
                if (previousSelected && previousSelected !== item) {
                    previousSelected.classList.add('chat-deselecting');
                    setTimeout(() => {
                        previousSelected.classList.remove('selected', 'chat-deselecting');
                    }, 300);
                }

                // Add selection to current item with animation
                item.classList.add('chat-selecting');
                setTimeout(() => {
                    item.classList.add('selected');
                    item.classList.remove('chat-selecting');
                }, 150);
            }
        });

        // Trigger the same event that the sidebar uses to open conversations
        window.dispatchEvent(new CustomEvent('toggle-chat-view', {
            detail: {
                conversationId: conversation.id,
                recipientId: recipientInfo?.id,
                recipientName: recipientInfo?.name || 'Unknown User',
                recipientAvatar: recipientInfo?.avatar || '/static/assets/default-avatar.png',
                isRecipientOnline: recipientInfo?.is_online || false,
                isNewConversation: false
            }
        }));
        console.debug(`[GlobalChatNotifications] Successfully opened conversation ${conversationId} with complete state management like sidebar`);
    } else {
        console.warn(`[GlobalChatNotifications] Conversation ${conversationId} not found, using enhanced fallback`);

        // Enhanced fallback: try to create conversation with message sender
        if (message && message.sender_id) {
            const allUsers = window.allUsers || [];
            const onlineUsers = window.onlineUsers || new Set();
            const sender = allUsers.find(u => u.id === message.sender_id);

            if (sender) {
                const senderName = `${sender.firstName || ''} ${sender.lastName || ''}`.trim() || sender.username;
                const senderAvatar = sender.avatar?.Valid ? sender.avatar.String : '/static/assets/default-avatar.png';
                const isSenderOnline = onlineUsers.has(message.sender_id);

                // Update ChatState for new conversation
                if (window.ChatState) {
                    window.ChatState.update({
                        conversationId: null,
                        recipientId: message.sender_id,
                        recipientName: senderName,
                        recipientAvatar: senderAvatar,
                        isNewConversation: true,
                        isRecipientOnline: isSenderOnline
                    });
                }

                // Update sidebar selection for the sender
                const conversationItems = document.querySelectorAll('.conversation-item');
                conversationItems.forEach(item => {
                    const itemUserId = parseInt(item.dataset.userId, 10);
                    if (itemUserId === message.sender_id) {
                        // Remove previous selection
                        const previousSelected = document.querySelector('.conversation-item.selected');
                        if (previousSelected && previousSelected !== item) {
                            previousSelected.classList.remove('selected');
                        }
                        // Add selection to sender's item
                        item.classList.add('selected');
                    }
                });

                // Create a new conversation with the sender
                window.dispatchEvent(new CustomEvent('toggle-chat-view', {
                    detail: {
                        conversationId: null,
                        recipientId: message.sender_id,
                        recipientName: senderName,
                        recipientAvatar: senderAvatar,
                        isRecipientOnline: isSenderOnline,
                        isNewConversation: true
                    }
                }));
                console.debug(`[GlobalChatNotifications] Created new conversation with sender ${message.sender_id} using complete state management`);
                return;
            }
        }

        // Final fallback: dispatch the open-conversation event
        window.dispatchEvent(new CustomEvent('open-conversation', {
            detail: { conversationId: conversationId }
        }));
        console.debug(`[GlobalChatNotifications] Used final fallback for conversation ${conversationId}`);
    }
}

/**
 * Show desktop notification
 * @param {Object} message - Message object
 */
async function showDesktopNotification(message) {
    if (!('Notification' in window)) {
        console.debug('[GlobalChatNotifications] Desktop notifications not supported');
        return;
    }

    const hasPermission = await requestDesktopNotificationPermission();
    if (!hasPermission) {
        console.debug('[GlobalChatNotifications] Desktop notification permission denied');
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
            silent: true
        });

        // Auto-close after 6 seconds
        setTimeout(() => notification.close(), 6000);

        // Handle click to focus window and open conversation
        notification.onclick = () => {
            window.focus();
            openConversation(message.conversation_id, message);
            notification.close();
        };

    } catch (error) {
        console.debug('[GlobalChatNotifications] Error showing desktop notification:', error);
    }
}

/**
 * Request desktop notification permission
 * @returns {Promise<boolean>} Permission granted
 */
async function requestDesktopNotificationPermission() {
    if (!('Notification' in window)) {
        return false;
    }

    if (Notification.permission === 'granted') {
        return true;
    }

    if (Notification.permission === 'denied') {
        return false;
    }

    try {
        const permission = await Notification.requestPermission();
        return permission === 'granted';
    } catch (error) {
        console.debug('[GlobalChatNotifications] Error requesting notification permission:', error);
        return false;
    }
}

/**
 * Play notification sound with enhanced sound type support
 * @param {number} volume - Volume level (0-1)
 * @param {string} soundType - Type of sound to play
 */
function playNotificationSound(volume = 0.3, soundType = 'default') {
    const preferences = getNotificationPreferences();
    const actualVolume = Math.max(0, Math.min(1, volume));

    console.debug(`[GlobalChatNotifications] Playing notification sound: ${soundType} at volume ${actualVolume}`);

    try {
        let audioSource = '/static/assets/notification.mp3';


        const audio = new Audio(audioSource);
        audio.volume = actualVolume;

        // Handle both loading errors and playback errors gracefully
        audio.onerror = () => {
            console.debug('[GlobalChatNotifications] Notification sound file not found, using fallback');
            playFallbackNotificationSound(actualVolume);
        };

        audio.play().catch(err => {
            console.debug('[GlobalChatNotifications] Could not play notification sound:', err);
            playFallbackNotificationSound(actualVolume);
        });
    } catch (error) {
        console.debug('[GlobalChatNotifications] Error playing notification sound:', error);
        playFallbackNotificationSound(actualVolume);
    }
}

/**
 * Play fallback notification sound using Web Audio API
 * @param {number} volume - Volume level (0-1)
 */
function playFallbackNotificationSound(volume = 0.3) {
    try {
        // Try Web Audio API beep
        if (window.AudioContext || window.webkitAudioContext) {
            const audioContext = new (window.AudioContext || window.webkitAudioContext)();
            const oscillator = audioContext.createOscillator();
            const gainNode = audioContext.createGain();

            oscillator.connect(gainNode);
            gainNode.connect(audioContext.destination);

            oscillator.frequency.setValueAtTime(800, audioContext.currentTime);
            gainNode.gain.setValueAtTime(volume * 0.3, audioContext.currentTime);
            gainNode.gain.exponentialRampToValueAtTime(0.01, audioContext.currentTime + 0.5);

            oscillator.start(audioContext.currentTime);
            oscillator.stop(audioContext.currentTime + 0.5);
        }
    } catch (error) {
        console.debug('[GlobalChatNotifications] Fallback sound not available:', error);
    }
}

/**
 * Get notification preferences from localStorage (synchronized with header dropdown)
 * @returns {Object} Notification preferences
 */
function getNotificationPreferences() {
    const prefs = localStorage.getItem('chatNotificationPreferences');
    const defaultPrefs = {
        sound: true,
        desktop: true,
        visual: true,
        volume: 0.3,
        dndSchedule: 'off',
        priorityOnly: false,
        soundType: 'default'
    };

    if (prefs) {
        try {
            const parsed = JSON.parse(prefs);
            // Merge with defaults to ensure all properties exist
            return { ...defaultPrefs, ...parsed };
        } catch (error) {
            console.warn('[GlobalChatNotifications] Error parsing preferences, using defaults:', error);
            return defaultPrefs;
        }
    }

    return defaultPrefs;
}

/**
 * Get current user from global state
 * @returns {Object|null} Current user
 */
function getCurrentUser() {
    return window.currentUser || null;
}

/**
 * Get initials from name
 * @param {string} name - Full name
 * @returns {string} Initials
 */
function getInitials(name) {
    if (!name) return '?';
    return name.split(' ')
        .map(word => word.charAt(0).toUpperCase())
        .slice(0, 2)
        .join('');
}

/**
 * Format notification timestamp
 * @param {string|Date} timestamp - Timestamp
 * @returns {string} Formatted time
 */
function formatNotificationTime(timestamp) {
    const date = new Date(timestamp);
    const now = new Date();
    const diffMs = now - date;
    const diffMins = Math.floor(diffMs / 60000);

    if (diffMins < 1) return 'now';
    if (diffMins < 60) return `${diffMins}m ago`;

    const diffHours = Math.floor(diffMins / 60);
    if (diffHours < 24) return `${diffHours}h ago`;

    return date.toLocaleDateString();
}

/**
 * Escape HTML to prevent XSS
 * @param {string} text - Text to escape
 * @returns {string} Escaped text
 */
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

/**
 * Clean up notification system
 */
export function cleanupGlobalChatNotifications() {
    hideAllGlobalChatNotifications();

    // Remove container
    const container = document.querySelector('.chat-notifications-container');
    if (container) {
        container.remove();
    }

    // Remove global functions
    delete window.showGlobalChatNotification;
    delete window.hideGlobalChatNotification;
    delete window.hideAllGlobalChatNotifications;

    console.debug('[GlobalChatNotifications] Global chat notification system cleaned up');
}
