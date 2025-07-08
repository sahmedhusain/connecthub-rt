/**
 * Sends a request to create a new conversation.
 * @param {number[]} participantIds - Array of user IDs (including the current user).
 * @returns {Promise<Object>} Result with success flag and data/error.
 */
export async function createConversation(participantIds) {
    if (!Array.isArray(participantIds) || participantIds.length < 2) {
        console.warn("[API] Invalid participant IDs for conversation:", participantIds);
        return { success: false, error: "At least two participants are required." };
    }

    try {
        console.debug("[API] Creating conversation with participants:", participantIds);

        const response = await fetch('/api/conversations', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ participants: participantIds }),
            credentials: 'include'
        });

        console.debug("[API] Conversation creation response status:", response.status);

        const data = await response.json();
        
        if (!response.ok) {
            const errorMsg = data.error || `Server error: ${response.status}`;
            console.error("[API] Conversation creation failed:", errorMsg);
            throw new Error(errorMsg);
        }

        const conversationId = data.conversation_id;

        if (!conversationId) {
            console.warn("[API] Conversation creation response missing conversation_id:", data);
            return { success: false, error: "Invalid server response format" };
        }

        console.info("[API] Conversation created successfully with ID:", conversationId);
        return {
            success: true,
            data: { conversation_id: conversationId }
        };

    } catch (error) {
        console.error('[API] Error creating conversation:', error.message || error);
        return { success: false, error: error.message || 'Unknown error occurred' };
    }
}

// --- Keep all other existing exported functions below ---
export async function login(identifier, password) {
    try {
        console.debug("[API] Attempting login with identifier:", identifier.substring(0, 2) + '***');
        
        const response = await fetch('/api/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ identifier, password }),
            credentials: 'include'
        });

        let data;
        try {
            data = await response.json();
        } catch (e) {
            console.error("[API] Failed to parse login response:", e.message || e);
            data = { success: false, error: 'Invalid response from server' };
        }

        if (!response.ok) {
            console.warn("[API] Login failed:", data.error || response.statusText);
            return {
                success: false,
                error: data.error || 'Login failed. Please check your credentials.'
            };
        }

        console.info("[API] Login successful for user:", data.username);
        localStorage.setItem('user', JSON.stringify({
            id: data.user_id || data.userID,
            userId: data.user_id || data.userID,
            username: data.username,
            email: data.email,
            firstName: data.firstName || '',
            lastName: data.lastName || '',
            gender: data.gender || '',
            dateOfBirth: data.dateOfBirth || '',
            avatar: data.avatar || '/static/assets/default-avatar.png'
        }));

        return { success: true, user: data };

    } catch (error) {
        console.error('[API] Login request error:', error.message || error);
        return { success: false, error: error.message || 'Unknown error occurred' };
    }
}

export async function signup(formData) {
    try {
        console.debug('[API] Sending signup request for user:', formData.username);

        const response = await fetch('/api/signup', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                firstName: formData.firstName,
                lastName: formData.lastName,
                username: formData.username,
                email: formData.email,
                gender: formData.gender,
                dateOfBirth: formData.dateOfBirth,
                password: formData.password
            }),
            credentials: 'include'
        });

        console.debug('[API] Signup response status:', response.status, response.statusText);

        let data;
        const contentType = response.headers.get('content-type');
        if (contentType && contentType.includes('application/json')) {
            data = await response.json();
        } else {
            const text = await response.text();
            console.warn('[API] Non-JSON signup response:', text);
            data = { error: text || 'Server returned an unexpected response format' };
        }

        if (!response.ok) {
            console.error('[API] Signup failed:', data.error || response.statusText);
            return {
                success: false,
                error: data.error || response.statusText
            };
        }

        if (data.success && data.user_id) {
            console.info('[API] User registered successfully:', formData.username);
            localStorage.setItem('user', JSON.stringify({
                id: data.user_id,
                userId: data.user_id,
                username: data.username,
                email: data.email,
                firstName: data.firstName || '',
                lastName: data.lastName || '',
                gender: data.gender || '',
                dateOfBirth: data.dateOfBirth || '',
                avatar: data.avatar || '/static/assets/default-avatar.png'
            }));
        }

        return {
            success: true,
            data
        };
    } catch (error) {
        console.error('[API] Signup request error:', error.message || error);
        return {
            success: false,
            error: error.message || 'Unknown error occurred'
        };
    }
}

export async function logout() {
    try {
        console.debug("[API] Initiating logout request");

        const response = await fetch('/api/logout', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            credentials: 'include'
        });

        console.debug("[API] Clearing user data from local storage");
        localStorage.removeItem('user');

        let data = { success: false };
        try {
            if (response.headers.get('content-type')?.includes('application/json')) {
                data = await response.json();
            } else {
                const text = await response.text();
                console.debug("[API] Logout response not JSON:", text);
            }
        } catch (e) {
            console.error("[API] Failed to parse logout response:", e.message || e);
        }

        if (!response.ok) {
            console.error("[API] Logout failed:", data?.error || response.statusText);
            window.location.href = '/';
            return { success: false, error: data?.error || "Logout failed on server" };
        }

        console.info("[API] Logout successful, redirecting to login page");
        window.location.href = '/';

        return { success: true };

    } catch (error) {
        console.error("[API] Logout request error:", error.message || error);
        localStorage.removeItem('user');
        window.location.href = '/';
        return { success: false, error: error.message };
    }
}

export function getCurrentUser() {
    try {
        const userData = localStorage.getItem('user');
        if (!userData) {
            console.debug("[API] No user data found in local storage");
            return null;
        }

        const user = JSON.parse(userData);
        if (user && (user.id || user.userId) && user.username) {
            if (user.id && !user.userId) {
                user.userId = user.id;
            }
            return user;
        }
        
        console.warn("[API] Invalid user data found in local storage, removing");
        localStorage.removeItem('user');
        return null;
    } catch (e) {
        console.error("[API] Error parsing user data from localStorage:", e.message || e);
        localStorage.removeItem('user');
        return null;
    }
}

export function isLoggedIn() {
    const user = getCurrentUser();
    console.debug(`[API] User login status: ${user !== null}`);
    return user !== null;
}

export function clearUserData() {
    console.debug("[API] Manually clearing user data from local storage");
    localStorage.removeItem('user');
}

export async function fetchPosts(tab = 'posts', filter = 'all') {
    try {
        console.debug(`[API] Fetching posts with tab=${tab}, filter=${filter}`);

        let normalizedTab = tab.replace(/\s+/g, '+');

        if ((normalizedTab === 'your+posts' || normalizedTab === 'your+replies') && filter === 'all') {
            filter = 'newest';
            console.debug(`[API] Adjusted filter to 'newest' for personal tab: ${normalizedTab}`);
        }

        const response = await fetch(`/api/posts?tab=${encodeURIComponent(normalizedTab)}&filter=${encodeURIComponent(filter)}`, {
            credentials: 'include',
            headers: {
                'Accept': 'application/json'
            }
        });

        if (!response.ok) {
            let errorText = `Server error: ${response.status}`;
            let errorCode = response.status.toString();
            try {
                const errorData = await response.json();
                errorText = errorData.error || errorText;
                errorCode = errorData.code || errorCode;
            } catch (e) {
                console.debug("[API] Could not parse error response as JSON");
            }
            console.error(`[API] Failed to fetch posts:`, errorText);

            const error = new Error(errorText);
            error.status = response.status;
            error.code = errorCode;
            throw error;
        }

        const posts = await response.json();

        // Handle null response (empty database) as an empty array
        if (posts === null) {
            console.info("[API] No posts found (empty database), returning empty array");
            return { success: true, data: [] };
        }

        if (Array.isArray(posts)) {
            console.info(`[API] Successfully fetched ${posts.length} posts`);
            return { success: true, data: posts };
        } else {
            console.warn("[API] fetchPosts received non-array data:", posts);
            throw new Error("Invalid data format received from server.");
        }

    } catch (error) {
        console.error('[API] Error fetching posts:', error.message || error);
        return { success: false, error: error.message };
    }
}

export async function fetchPostById(postId) {
    if (!postId) {
        console.warn("[API] fetchPostById called without a post ID");
        return { success: false, error: "Post ID is required." };
    }
    
    try {
        console.debug(`[API] Fetching post with ID: ${postId}`);
        
        const response = await fetch(`/api/post?id=${postId}`, {
            credentials: 'include'
        });

        if (!response.ok) {
            let errorText = `Failed to fetch post (${response.status})`;
            let errorCode = response.status.toString();
            try {
                const errorData = await response.json();
                errorText = errorData.error || errorText;
                errorCode = errorData.code || errorCode;
            } catch (e) {
                console.debug("[API] Could not parse error response as JSON");
            }
            console.error(`[API] Failed to fetch post ${postId}:`, errorText);

            const error = new Error(errorText);
            error.status = response.status;
            error.code = errorCode;
            throw error;
        }

        const data = await response.json();
        if (!data || !data.post || typeof data.post.PostID !== 'number') {
            console.warn(`[API] Invalid post data received for post ${postId}`);
            throw new Error("Invalid post data received from server.");
        }
        
        console.info(`[API] Successfully fetched post ${postId} with ${data.comments?.length || 0} comments`);
        return { success: true, data };

    } catch (error) {
        console.error(`[API] Error fetching post ${postId}:`, error.message || error);
        return { success: false, error: error.message };
    }
}

export async function createPost(formData) {
    if (!(formData instanceof FormData)) {
        console.warn("[API] createPost called with invalid data type");
        return { success: false, error: "Invalid data provided for post creation." };
    }

    try {
        console.debug("[API] Sending request to create new post");

        // Convert FormData to JSON for the API endpoint
        const postData = {
            title: formData.get('title'),
            content: formData.get('content'),
            categories: formData.getAll('categories')
        };

        const response = await fetch('/api/post/create', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(postData),
            credentials: 'include'
        });

        if (response.ok || response.redirected) {
            let resultData = { success: true };
            try {
                resultData = await response.json();
            } catch (e) { 
                console.debug("[API] Could not parse success response as JSON");
            }

            console.info("[API] Post created successfully");
            return {
                success: true,
                redirectUrl: response.redirected ? response.url : (resultData.redirect || null),
                data: resultData
            };
        } else {
            let errorMsg = `Failed to create post (${response.status})`;
            try {
                const errorData = await response.json();
                errorMsg = errorData.error || errorMsg;
            } catch (e) { 
                console.debug("[API] Could not parse error response as JSON");
            }
            console.error("[API] Post creation failed:", errorMsg);
            throw new Error(errorMsg);
        }

    } catch (error) {
        console.error('[API] Error creating post:', error.message || error);
        return { success: false, error: error.message };
    }
}

export async function addComment(postId, content) {
    const userData = getCurrentUser();
    const userId = userData?.userId;

    if (!userId) {
        console.warn("[API] addComment called without a logged-in user");
        return { success: false, error: "User not logged in." };
    }
    
    if (!postId || !content) {
        console.warn("[API] addComment called with missing data", { hasPostId: !!postId, contentLength: content?.length });
        return { success: false, error: "Post ID and comment content are required." };
    }

    try {
        console.debug(`[API] Adding comment to post ${postId}`);
        
        const response = await fetch('/addcomment', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded',
            },
            body: `post_id=${encodeURIComponent(postId)}&user_id=${encodeURIComponent(userId)}&content=${encodeURIComponent(content)}`,
            credentials: 'include'
        });

        if (response.ok || response.redirected) {
            console.info(`[API] Comment added successfully to post ${postId}`);
            return {
                success: true,
                redirectUrl: response.redirected ? response.url : null
            };
        } else {
            let errorMsg = `Failed to add comment (${response.status})`;
            try {
                const errorData = await response.json();
                errorMsg = errorData.error || errorMsg;
            } catch (e) { 
                console.debug("[API] Could not parse error response as JSON");
            }
            console.error(`[API] Failed to add comment to post ${postId}:`, errorMsg);
            throw new Error(errorMsg);
        }
    } catch (error) {
        console.error('[API] Error adding comment:', error.message || error);
        return { success: false, error: error.message };
    }
}

let cachedCategories = null;

export async function fetchCategories() {
    if (cachedCategories) {
        console.debug('[API] Using cached categories:', cachedCategories.length, 'items');
        return { success: true, data: cachedCategories };
    }

    try {
        console.debug('[API] Fetching categories from server');

        const response = await fetch('/api/categories', {
            method: 'GET',
            headers: {
                'Accept': 'application/json',
                'Cache-Control': 'no-cache'
            },
            credentials: 'include'
        });

        console.debug(`[API] Categories response status: ${response.status}, ok: ${response.ok}`);

        if (!response.ok) {
            const errorText = await response.text();
            console.error(`[API] Failed to load categories (${response.status}): ${errorText}`);
            throw new Error(`Failed to load categories (${response.status})`);
        }

        const categories = await response.json();
        if (!Array.isArray(categories)) {
            console.warn('[API] Invalid category data format received:', categories);
            throw new Error("Invalid category data format received.");
        }

        console.info('[API] Successfully loaded', categories.length, 'categories');
        cachedCategories = categories;
        return { success: true, data: categories };

    } catch (error) {
        console.error('[API] Error loading categories:', error.message || error);
        return { success: false, error: error.message };
    }
}

export function clearCategoriesCache() {
    cachedCategories = null;
    console.debug("[API] Categories cache cleared");
}

export async function fetchUsers() {
    console.log('[API] Fetching users');
    try {
        const response = await fetch('/api/users', {
            method: 'GET',
            credentials: 'include'
        });
        
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        const data = await response.json();
        console.log('[API] Users data received:', data);

        // Handle null response (empty database) as an empty array
        if (data === null) {
            console.info("[API] No users found (empty database), returning empty array");
            return [];
        }

        // Handle both direct array and {data: array, success: true} formats
        const users = data.data || data;

        // Validate data format
        if (!Array.isArray(users)) {
            console.error('[API] Invalid user data format received');
            throw new Error('Invalid user data format received.');
        }
        
        return users;
    } catch (error) {
        console.error('[API] Error loading users:', error.message);
        throw error;
    }
}

export async function fetchConversations() {
    console.log('[API] Fetching conversations');
    try {
        const response = await fetch('/api/conversations', {
            method: 'GET',
            credentials: 'include'
        });
        
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        const data = await response.json();
        console.log('[API] Conversations data received:', data);

        // Handle null response (empty database) as an empty array
        if (data === null) {
            console.info("[API] No conversations found (empty database), returning empty array");
            return [];
        }

        // Handle both direct array and {data: array, success: true} formats
        const conversations = data.data || data;

        // Validate data format
        if (!Array.isArray(conversations)) {
            console.error('[API] Invalid conversation data format received');
            throw new Error('Invalid conversation data format received.');
        }
        
        return conversations;
    } catch (error) {
        console.error('[API] Error fetching conversations:', error.message);
        throw error;
    }
}

export async function fetchMessages(conversationId, limit = 100, offset = 0) {
    if (!conversationId) {
        console.warn("[API] fetchMessages called without conversation ID");
        return { success: false, error: "Conversation ID is required." };
    }
    
    try {
        console.debug(`[API] Fetching messages for conversation ${conversationId} (limit: ${limit}, offset: ${offset})`);
        
        const response = await fetch(`/api/messages?conversation_id=${conversationId}&limit=${limit}&offset=${offset}`, {
            credentials: 'include'
        });

        if (!response.ok) {
            let errorMsg = `Failed to fetch messages (${response.status})`;
            try {
                const errorData = await response.json();
                errorMsg = errorData.error || errorMsg;
            } catch (e) { 
                console.debug("[API] Could not parse error response as JSON");
            }
            console.error(`[API] Failed to fetch messages for conversation ${conversationId}:`, errorMsg);
            throw new Error(errorMsg);
        }

        const messages = await response.json();

        // Handle null response (empty database) as an empty array
        if (messages === null) {
            console.info(`[API] No messages found for conversation ${conversationId} (empty database), returning empty array`);
            return { success: true, data: [] };
        }

        if (!Array.isArray(messages)) {
            console.warn('[API] Invalid messages data format received');
            throw new Error("Invalid messages data format received.");
        }
        
        console.info(`[API] Successfully loaded ${messages.length} messages for conversation ${conversationId}`);
        return { success: true, data: messages };

    } catch (error) {
        console.error(`[API] Error fetching messages for conversation ${conversationId}:`, error.message || error);
        return { success: false, error: error.message };
    }
}

export async function getConversations() {
    try {
        console.debug('[API] Fetching conversations (getConversations)');
        
        const response = await fetch('/api/conversations', {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
            },
            credentials: 'include'
        });

        const data = await response.json();

        if (!response.ok) {
            console.error('[API] Failed to fetch conversations:', data.error || response.statusText);
            throw new Error(data.error || `Server error: ${response.status}`);
        }

        console.info(`[API] Successfully loaded conversations (getConversations)`);
        return {
            success: true,
            data: data
        };
    } catch (error) {
        console.error('[API] Error fetching conversations:', error.message || error);
        return {
            success: false,
            error: error.message || 'Failed to load conversations'
        };
    }
}

export async function apiGetConversations(conversationId) {
    try {
        console.debug(`[API] Fetching specific conversation: ${conversationId}`);
        
        const response = await fetch(`/api/conversations/${conversationId}`, {
            credentials: 'include'
        });

        if (!response.ok) {
            let errorMsg = `Failed to fetch conversation (${response.status})`;
            try {
                const errorData = await response.json();
                errorMsg = errorData.error || errorMsg;
            } catch (e) { 
                console.debug("[API] Could not parse error response as JSON");
            }
            console.error(`[API] Failed to fetch conversation ${conversationId}:`, errorMsg);
            throw new Error(errorMsg);
        }

        const conversation = await response.json();
        console.info(`[API] Successfully loaded conversation ${conversationId}`);
        return { success: true, data: conversation };

    } catch (error) {
        console.error(`[API] Error fetching conversation ${conversationId}:`, error.message || error);
        return { success: false, error: error.message };
    }
}

// Unused function removed

/**
 * Mark messages in a conversation as read
 * @param {number} conversationId - The conversation ID
 * @returns {Promise<boolean>} - Success status
 */
export async function markMessagesAsRead(conversationId) {
    if (!conversationId) {
        console.warn("[API] markMessagesAsRead called without conversation ID");
        return false;
    }

    try {
        console.debug(`[API] Marking messages as read for conversation ${conversationId}`);

        const response = await fetch(`/api/messages/read`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            credentials: 'include',
            body: JSON.stringify({
                conversation_id: conversationId
            })
        });

        if (!response.ok) {
            let errorMsg = `Failed to mark messages as read (${response.status})`;
            try {
                const errorData = await response.json();
                errorMsg = errorData.error || errorMsg;
            } catch (e) {
                console.debug("[API] Could not parse error response as JSON");
            }
            console.error(`[API] Failed to mark messages as read for conversation ${conversationId}:`, errorMsg);
            return false;
        }

        console.info(`[API] Successfully marked messages as read for conversation ${conversationId}`);
        return true;

    } catch (error) {
        console.error(`[API] Error marking messages as read for conversation ${conversationId}:`, error.message || error);
        return false;
    }
}

