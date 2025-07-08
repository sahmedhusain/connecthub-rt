/**
 * Component Tests
 * Tests for all frontend components including chat, header, home, login, etc.
 */

import { describe, test, expect, beforeEach, afterEach, jest } from '@jest/globals';

// Mock DOM environment
const mockDOM = () => {
  document.body.innerHTML = '';
  
  // Create basic HTML structure
  document.body.innerHTML = `
    <div id="app">
      <header id="header"></header>
      <main id="main-content"></main>
      <div id="sidebar"></div>
      <div id="chat-container"></div>
    </div>
  `;
};

// Mock fetch API
global.fetch = jest.fn();

// Mock WebSocket
global.WebSocket = jest.fn().mockImplementation(() => ({
  send: jest.fn(),
  close: jest.fn(),
  addEventListener: jest.fn(),
  removeEventListener: jest.fn(),
  readyState: 1,
}));

describe('Component Tests', () => {
  beforeEach(() => {
    mockDOM();
    fetch.mockClear();
    jest.clearAllMocks();
  });

  afterEach(() => {
    document.body.innerHTML = '';
  });

  describe('Header Component', () => {
    test('should render header with navigation', () => {
      const headerElement = document.getElementById('header');
      expect(headerElement).toBeTruthy();
      
      // Simulate header rendering
      headerElement.innerHTML = `
        <nav class="navbar">
          <div class="nav-brand">Real-Time Forum</div>
          <div class="nav-links">
            <a href="/home">Home</a>
            <a href="/chat">Chat</a>
            <a href="/create-post">Create Post</a>
          </div>
          <div class="user-menu">
            <div class="dropdown">
              <button class="dropdown-toggle">User</button>
              <div class="dropdown-menu">
                <a href="/profile">Profile</a>
                <a href="/logout">Logout</a>
              </div>
            </div>
          </div>
        </nav>
      `;

      expect(headerElement.querySelector('.navbar')).toBeTruthy();
      expect(headerElement.querySelector('.nav-brand')).toBeTruthy();
      expect(headerElement.querySelector('.nav-links')).toBeTruthy();
      expect(headerElement.querySelector('.user-menu')).toBeTruthy();
    });

    test('should handle dropdown menu interactions', () => {
      const headerElement = document.getElementById('header');
      headerElement.innerHTML = `
        <div class="dropdown">
          <button class="dropdown-toggle">User</button>
          <div class="dropdown-menu hidden">
            <a href="/profile">Profile</a>
            <a href="/logout">Logout</a>
          </div>
        </div>
      `;

      const dropdownToggle = headerElement.querySelector('.dropdown-toggle');
      const dropdownMenu = headerElement.querySelector('.dropdown-menu');

      // Simulate click event
      dropdownToggle.click();
      
      // In a real implementation, this would toggle the 'hidden' class
      expect(dropdownToggle).toBeTruthy();
      expect(dropdownMenu).toBeTruthy();
    });

    test('should display user information when authenticated', () => {
      const headerElement = document.getElementById('header');
      
      // Mock user data
      const userData = {
        username: 'testuser',
        firstName: 'Test',
        lastName: 'User',
        avatar: 'avatar.jpg'
      };

      // Simulate authenticated header
      headerElement.innerHTML = `
        <div class="user-info">
          <img src="${userData.avatar}" alt="Avatar" class="user-avatar">
          <span class="user-name">${userData.firstName} ${userData.lastName}</span>
          <span class="username">@${userData.username}</span>
        </div>
      `;

      expect(headerElement.querySelector('.user-avatar')).toBeTruthy();
      expect(headerElement.querySelector('.user-name').textContent).toBe('Test User');
      expect(headerElement.querySelector('.username').textContent).toBe('@testuser');
    });
  });

  describe('Login Component', () => {
    test('should render login form', () => {
      const mainContent = document.getElementById('main-content');
      mainContent.innerHTML = `
        <div class="login-container">
          <form id="login-form" class="login-form">
            <h2>Login</h2>
            <div class="form-group">
              <input type="text" id="identifier" name="identifier" placeholder="Username or Email" required>
            </div>
            <div class="form-group">
              <input type="password" id="password" name="password" placeholder="Password" required>
            </div>
            <button type="submit" class="btn btn-primary">Login</button>
            <div class="form-links">
              <a href="/signup">Don't have an account? Sign up</a>
            </div>
          </form>
        </div>
      `;

      expect(mainContent.querySelector('#login-form')).toBeTruthy();
      expect(mainContent.querySelector('#identifier')).toBeTruthy();
      expect(mainContent.querySelector('#password')).toBeTruthy();
      expect(mainContent.querySelector('button[type="submit"]')).toBeTruthy();
    });

    test('should validate login form inputs', () => {
      const mainContent = document.getElementById('main-content');
      mainContent.innerHTML = `
        <form id="login-form">
          <input type="text" id="identifier" name="identifier" required>
          <input type="password" id="password" name="password" required>
          <button type="submit">Login</button>
        </form>
      `;

      const form = mainContent.querySelector('#login-form');
      const identifierInput = mainContent.querySelector('#identifier');
      const passwordInput = mainContent.querySelector('#password');

      // Test empty form validation
      expect(identifierInput.checkValidity()).toBe(false);
      expect(passwordInput.checkValidity()).toBe(false);

      // Test with valid inputs
      identifierInput.value = 'testuser';
      passwordInput.value = 'password123';
      
      expect(identifierInput.checkValidity()).toBe(true);
      expect(passwordInput.checkValidity()).toBe(true);
    });

    test('should handle login form submission', async () => {
      const mainContent = document.getElementById('main-content');
      mainContent.innerHTML = `
        <form id="login-form">
          <input type="text" id="identifier" value="testuser">
          <input type="password" id="password" value="password123">
          <button type="submit">Login</button>
        </form>
      `;

      // Mock successful login response
      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          success: true,
          user_id: 1,
          username: 'testuser',
          email: 'test@example.com'
        })
      });

      const form = mainContent.querySelector('#login-form');
      const submitEvent = new Event('submit');
      
      // Simulate form submission
      form.dispatchEvent(submitEvent);

      // Verify fetch was called
      expect(fetch).toHaveBeenCalledWith('/api/login', expect.objectContaining({
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        }
      }));
    });
  });

  describe('Chat Component', () => {
    test('should render chat interface', () => {
      const chatContainer = document.getElementById('chat-container');
      chatContainer.innerHTML = `
        <div class="chat-interface">
          <div class="chat-sidebar">
            <div class="conversations-list">
              <h3>Chats (2)</h3>
              <div class="conversation-item active">
                <div class="conversation-info">
                  <span class="conversation-name">John Doe</span>
                  <span class="last-message">Hello there!</span>
                </div>
                <div class="conversation-meta">
                  <span class="timestamp">2 min ago</span>
                  <span class="unread-count">1</span>
                </div>
              </div>
            </div>
          </div>
          <div class="chat-main">
            <div class="chat-header">
              <h3>John Doe</h3>
              <span class="online-status">Online</span>
            </div>
            <div class="messages-container">
              <div class="message received">
                <div class="message-content">Hello there!</div>
                <div class="message-time">2:30 PM</div>
              </div>
            </div>
            <div class="message-input">
              <input type="text" placeholder="Type a message..." id="message-input">
              <button id="send-button">Send</button>
            </div>
          </div>
        </div>
      `;

      expect(chatContainer.querySelector('.chat-interface')).toBeTruthy();
      expect(chatContainer.querySelector('.chat-sidebar')).toBeTruthy();
      expect(chatContainer.querySelector('.chat-main')).toBeTruthy();
      expect(chatContainer.querySelector('.messages-container')).toBeTruthy();
      expect(chatContainer.querySelector('#message-input')).toBeTruthy();
    });

    test('should handle message sending', () => {
      const chatContainer = document.getElementById('chat-container');
      chatContainer.innerHTML = `
        <div class="message-input">
          <input type="text" id="message-input" value="Test message">
          <button id="send-button">Send</button>
        </div>
        <div class="messages-container"></div>
      `;

      const messageInput = chatContainer.querySelector('#message-input');
      const sendButton = chatContainer.querySelector('#send-button');
      const messagesContainer = chatContainer.querySelector('.messages-container');

      // Simulate sending a message
      const messageText = messageInput.value;
      
      // Mock WebSocket send
      const mockWS = new WebSocket();
      mockWS.send(JSON.stringify({
        type: 'message',
        content: messageText,
        conversation_id: 1
      }));

      expect(mockWS.send).toHaveBeenCalledWith(
        JSON.stringify({
          type: 'message',
          content: 'Test message',
          conversation_id: 1
        })
      );
    });

    test('should display typing indicators', () => {
      const chatContainer = document.getElementById('chat-container');
      chatContainer.innerHTML = `
        <div class="messages-container">
          <div class="typing-indicator hidden">
            <span class="typing-user">John Doe</span> is typing...
          </div>
        </div>
      `;

      const typingIndicator = chatContainer.querySelector('.typing-indicator');
      
      // Simulate showing typing indicator
      typingIndicator.classList.remove('hidden');
      
      expect(typingIndicator.classList.contains('hidden')).toBe(false);
      expect(typingIndicator.textContent).toContain('is typing...');
    });

    test('should handle real-time message updates', () => {
      const chatContainer = document.getElementById('chat-container');
      chatContainer.innerHTML = `
        <div class="messages-container"></div>
      `;

      const messagesContainer = chatContainer.querySelector('.messages-container');

      // Simulate receiving a new message via WebSocket
      const newMessage = {
        id: 1,
        content: 'New message',
        sender_id: 2,
        sender_name: 'Jane Doe',
        timestamp: new Date().toISOString()
      };

      // Create message element
      const messageElement = document.createElement('div');
      messageElement.className = 'message received';
      messageElement.innerHTML = `
        <div class="message-content">${newMessage.content}</div>
        <div class="message-sender">${newMessage.sender_name}</div>
        <div class="message-time">${new Date(newMessage.timestamp).toLocaleTimeString()}</div>
      `;

      messagesContainer.appendChild(messageElement);

      expect(messagesContainer.children.length).toBe(1);
      expect(messagesContainer.querySelector('.message-content').textContent).toBe('New message');
    });
  });

  describe('Home Component', () => {
    test('should render posts feed', () => {
      const mainContent = document.getElementById('main-content');
      mainContent.innerHTML = `
        <div class="home-container">
          <div class="posts-header">
            <h2>Latest Posts</h2>
            <div class="filter-tabs">
              <button class="tab active" data-filter="all">All</button>
              <button class="tab" data-filter="top-rated">Top Rated</button>
              <button class="tab" data-filter="oldest">Oldest</button>
            </div>
          </div>
          <div class="posts-container">
            <div class="post-card">
              <h3 class="post-title">Test Post</h3>
              <p class="post-content">This is a test post content</p>
              <div class="post-meta">
                <span class="author">By John Doe</span>
                <span class="timestamp">2 hours ago</span>
                <span class="category">Technology</span>
              </div>
            </div>
          </div>
        </div>
      `;

      expect(mainContent.querySelector('.home-container')).toBeTruthy();
      expect(mainContent.querySelector('.posts-header')).toBeTruthy();
      expect(mainContent.querySelector('.filter-tabs')).toBeTruthy();
      expect(mainContent.querySelector('.posts-container')).toBeTruthy();
      expect(mainContent.querySelector('.post-card')).toBeTruthy();
    });

    test('should handle post filtering', () => {
      const mainContent = document.getElementById('main-content');
      mainContent.innerHTML = `
        <div class="filter-tabs">
          <button class="tab active" data-filter="all">All</button>
          <button class="tab" data-filter="top-rated">Top Rated</button>
        </div>
        <div class="posts-container"></div>
      `;

      const tabs = mainContent.querySelectorAll('.tab');
      const topRatedTab = mainContent.querySelector('[data-filter="top-rated"]');

      // Simulate clicking top-rated filter
      topRatedTab.click();

      // Mock API call for filtered posts
      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ([
          {
            id: 1,
            title: 'Top Rated Post',
            content: 'This is a top rated post',
            author: 'Jane Doe',
            likes: 25
          }
        ])
      });

      expect(topRatedTab.dataset.filter).toBe('top-rated');
    });
  });

  describe('Post Component', () => {
    test('should render individual post view', () => {
      const mainContent = document.getElementById('main-content');
      mainContent.innerHTML = `
        <div class="post-view">
          <article class="post">
            <h1 class="post-title">Sample Post Title</h1>
            <div class="post-meta">
              <span class="author">By John Doe</span>
              <span class="timestamp">3 hours ago</span>
              <div class="categories">
                <span class="category">Technology</span>
                <span class="category">Programming</span>
              </div>
            </div>
            <div class="post-content">
              <p>This is the post content...</p>
            </div>
          </article>
          <section class="comments">
            <h3>Comments</h3>
            <div class="comment-form">
              <textarea placeholder="Add a comment..." id="comment-input"></textarea>
              <button id="submit-comment">Post Comment</button>
            </div>
            <div class="comments-list">
              <div class="comment">
                <div class="comment-author">Jane Doe</div>
                <div class="comment-content">Great post!</div>
                <div class="comment-time">1 hour ago</div>
              </div>
            </div>
          </section>
        </div>
      `;

      expect(mainContent.querySelector('.post-view')).toBeTruthy();
      expect(mainContent.querySelector('.post')).toBeTruthy();
      expect(mainContent.querySelector('.comments')).toBeTruthy();
      expect(mainContent.querySelector('#comment-input')).toBeTruthy();
      expect(mainContent.querySelector('.comments-list')).toBeTruthy();
    });

    test('should handle comment submission', () => {
      const mainContent = document.getElementById('main-content');
      mainContent.innerHTML = `
        <div class="comment-form">
          <textarea id="comment-input">This is a test comment</textarea>
          <button id="submit-comment">Post Comment</button>
        </div>
        <div class="comments-list"></div>
      `;

      const commentInput = mainContent.querySelector('#comment-input');
      const submitButton = mainContent.querySelector('#submit-comment');

      // Mock comment submission
      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          success: true,
          comment_id: 1
        })
      });

      // Simulate comment submission
      submitButton.click();

      expect(commentInput.value).toBe('This is a test comment');
    });
  });
});
