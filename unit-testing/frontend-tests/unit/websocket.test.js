/**
 * WebSocket Client and Real-time Features Tests
 * Tests for WebSocket connectivity, message handling, and real-time updates
 */

import { describe, test, expect, beforeEach, jest } from '@jest/globals';

describe('WebSocket Client Tests', () => {
  let mockWebSocket;
  let mockEventListeners;

  beforeEach(() => {
    // Reset WebSocket mock
    mockEventListeners = {};
    mockWebSocket = {
      send: jest.fn(),
      close: jest.fn(),
      addEventListener: jest.fn((event, handler) => {
        if (!mockEventListeners[event]) {
          mockEventListeners[event] = [];
        }
        mockEventListeners[event].push(handler);
      }),
      removeEventListener: jest.fn(),
      readyState: 1, // OPEN
      CONNECTING: 0,
      OPEN: 1,
      CLOSING: 2,
      CLOSED: 3,
    };

    global.WebSocket = jest.fn().mockImplementation(() => mockWebSocket);
  });

  describe('WebSocket Connection Management', () => {
    test('should create WebSocket connection', () => {
      const ws = new WebSocket('ws://localhost:8080/ws');
      
      expect(global.WebSocket).toHaveBeenCalledWith('ws://localhost:8080/ws');
      expect(ws).toBe(mockWebSocket);
    });

    test('should handle connection open event', () => {
      const ws = new WebSocket('ws://localhost:8080/ws');
      let connectionOpened = false;

      ws.addEventListener('open', () => {
        connectionOpened = true;
      });

      // Simulate connection open
      const openHandler = mockEventListeners.open[0];
      openHandler();

      expect(connectionOpened).toBe(true);
    });

    test('should handle connection close event', () => {
      const ws = new WebSocket('ws://localhost:8080/ws');
      let connectionClosed = false;
      let closeCode = null;

      ws.addEventListener('close', (event) => {
        connectionClosed = true;
        closeCode = event.code;
      });

      // Simulate connection close
      const closeHandler = mockEventListeners.close[0];
      closeHandler({ code: 1000 });

      expect(connectionClosed).toBe(true);
      expect(closeCode).toBe(1000);
    });

    test('should handle connection error', () => {
      const ws = new WebSocket('ws://localhost:8080/ws');
      let errorOccurred = false;
      let errorMessage = null;

      ws.addEventListener('error', (event) => {
        errorOccurred = true;
        errorMessage = event.message;
      });

      // Simulate connection error
      const errorHandler = mockEventListeners.error[0];
      errorHandler({ message: 'Connection failed' });

      expect(errorOccurred).toBe(true);
      expect(errorMessage).toBe('Connection failed');
    });
  });

  describe('Message Sending', () => {
    test('should send authentication message', () => {
      const ws = new WebSocket('ws://localhost:8080/ws');
      const authMessage = {
        type: 'auth',
        token: 'test-session-token'
      };

      ws.send(JSON.stringify(authMessage));

      expect(ws.send).toHaveBeenCalledWith(JSON.stringify(authMessage));
    });

    test('should send chat message', () => {
      const ws = new WebSocket('ws://localhost:8080/ws');
      const chatMessage = {
        type: 'message',
        recipientId: 123,
        content: 'Hello, how are you?'
      };

      ws.send(JSON.stringify(chatMessage));

      expect(ws.send).toHaveBeenCalledWith(JSON.stringify(chatMessage));
    });

    test('should send typing indicator', () => {
      const ws = new WebSocket('ws://localhost:8080/ws');
      const typingMessage = {
        type: 'typing_start',
        recipientId: 123
      };

      ws.send(JSON.stringify(typingMessage));

      expect(ws.send).toHaveBeenCalledWith(JSON.stringify(typingMessage));
    });

    test('should send online status update', () => {
      const ws = new WebSocket('ws://localhost:8080/ws');
      const statusMessage = {
        type: 'status_update',
        status: 'online'
      };

      ws.send(JSON.stringify(statusMessage));

      expect(ws.send).toHaveBeenCalledWith(JSON.stringify(statusMessage));
    });
  });

  describe('Message Receiving', () => {
    test('should handle incoming chat message', () => {
      const ws = new WebSocket('ws://localhost:8080/ws');
      let receivedMessage = null;

      ws.addEventListener('message', (event) => {
        const data = JSON.parse(event.data);
        if (data.type === 'message') {
          receivedMessage = data;
        }
      });

      // Simulate incoming message
      const messageHandler = mockEventListeners.message[0];
      const incomingMessage = {
        type: 'message',
        senderId: 456,
        content: 'Hello back!',
        timestamp: '2024-01-01T10:00:00Z'
      };

      messageHandler({ data: JSON.stringify(incomingMessage) });

      expect(receivedMessage).toEqual(incomingMessage);
    });

    test('should handle typing indicator', () => {
      const ws = new WebSocket('ws://localhost:8080/ws');
      let typingUser = null;

      ws.addEventListener('message', (event) => {
        const data = JSON.parse(event.data);
        if (data.type === 'typing_start') {
          typingUser = data.senderId;
        }
      });

      // Simulate typing indicator
      const messageHandler = mockEventListeners.message[0];
      const typingMessage = {
        type: 'typing_start',
        senderId: 789
      };

      messageHandler({ data: JSON.stringify(typingMessage) });

      expect(typingUser).toBe(789);
    });

    test('should handle user online status', () => {
      const ws = new WebSocket('ws://localhost:8080/ws');
      let onlineUsers = [];

      ws.addEventListener('message', (event) => {
        const data = JSON.parse(event.data);
        if (data.type === 'user_online') {
          onlineUsers.push(data.userId);
        }
      });

      // Simulate user online notification
      const messageHandler = mockEventListeners.message[0];
      const onlineMessage = {
        type: 'user_online',
        userId: 101
      };

      messageHandler({ data: JSON.stringify(onlineMessage) });

      expect(onlineUsers).toContain(101);
    });
  });

  describe('Real-time UI Updates', () => {
    test('should update chat interface with new message', () => {
      // Setup DOM
      document.body.innerHTML = `
        <div id="chat-container">
          <div id="messages-list"></div>
        </div>
      `;

      const ws = new WebSocket('ws://localhost:8080/ws');
      const messagesList = document.getElementById('messages-list');

      ws.addEventListener('message', (event) => {
        const data = JSON.parse(event.data);
        if (data.type === 'message') {
          // Simulate adding message to UI
          const messageElement = document.createElement('div');
          messageElement.className = 'message';
          messageElement.innerHTML = `
            <span class="sender">${data.senderName}</span>
            <span class="content">${data.content}</span>
          `;
          messagesList.appendChild(messageElement);
        }
      });

      // Simulate incoming message
      const messageHandler = mockEventListeners.message[0];
      const incomingMessage = {
        type: 'message',
        senderId: 456,
        senderName: 'John Doe',
        content: 'New message!',
        timestamp: '2024-01-01T10:00:00Z'
      };

      messageHandler({ data: JSON.stringify(incomingMessage) });

      const messages = messagesList.querySelectorAll('.message');
      expect(messages.length).toBe(1);
      expect(messages[0].querySelector('.sender').textContent).toBe('John Doe');
      expect(messages[0].querySelector('.content').textContent).toBe('New message!');
    });

    test('should show typing indicator in UI', () => {
      // Setup DOM
      document.body.innerHTML = `
        <div id="chat-container">
          <div id="typing-indicators"></div>
        </div>
      `;

      const ws = new WebSocket('ws://localhost:8080/ws');
      const typingIndicators = document.getElementById('typing-indicators');

      ws.addEventListener('message', (event) => {
        const data = JSON.parse(event.data);
        if (data.type === 'typing_start') {
          // Simulate showing typing indicator
          const indicator = document.createElement('div');
          indicator.className = 'typing-indicator';
          indicator.dataset.userId = data.senderId;
          indicator.textContent = `${data.senderName} is typing...`;
          typingIndicators.appendChild(indicator);
        } else if (data.type === 'typing_stop') {
          // Simulate hiding typing indicator
          const indicator = typingIndicators.querySelector(`[data-user-id="${data.senderId}"]`);
          if (indicator) {
            indicator.remove();
          }
        }
      });

      // Simulate typing start
      const messageHandler = mockEventListeners.message[0];
      const typingStart = {
        type: 'typing_start',
        senderId: 789,
        senderName: 'Jane Smith'
      };

      messageHandler({ data: JSON.stringify(typingStart) });

      let indicators = typingIndicators.querySelectorAll('.typing-indicator');
      expect(indicators.length).toBe(1);
      expect(indicators[0].textContent).toBe('Jane Smith is typing...');

      // Simulate typing stop
      const typingStop = {
        type: 'typing_stop',
        senderId: 789
      };

      messageHandler({ data: JSON.stringify(typingStop) });

      indicators = typingIndicators.querySelectorAll('.typing-indicator');
      expect(indicators.length).toBe(0);
    });

    test('should update online status indicators', () => {
      // Setup DOM
      document.body.innerHTML = `
        <div id="user-list">
          <div class="user-item" data-user-id="101">
            <span class="username">User1</span>
            <span class="status offline">Offline</span>
          </div>
          <div class="user-item" data-user-id="102">
            <span class="username">User2</span>
            <span class="status offline">Offline</span>
          </div>
        </div>
      `;

      const ws = new WebSocket('ws://localhost:8080/ws');

      ws.addEventListener('message', (event) => {
        const data = JSON.parse(event.data);
        if (data.type === 'user_online' || data.type === 'user_offline') {
          // Simulate updating user status
          const userElement = document.querySelector(`[data-user-id="${data.userId}"]`);
          if (userElement) {
            const statusElement = userElement.querySelector('.status');
            if (data.type === 'user_online') {
              statusElement.className = 'status online';
              statusElement.textContent = 'Online';
            } else {
              statusElement.className = 'status offline';
              statusElement.textContent = 'Offline';
            }
          }
        }
      });

      // Simulate user going online
      const messageHandler = mockEventListeners.message[0];
      const userOnline = {
        type: 'user_online',
        userId: 101
      };

      messageHandler({ data: JSON.stringify(userOnline) });

      const user1Status = document.querySelector('[data-user-id="101"] .status');
      expect(user1Status.className).toBe('status online');
      expect(user1Status.textContent).toBe('Online');

      // Simulate user going offline
      const userOffline = {
        type: 'user_offline',
        userId: 101
      };

      messageHandler({ data: JSON.stringify(userOffline) });

      expect(user1Status.className).toBe('status offline');
      expect(user1Status.textContent).toBe('Offline');
    });
  });

  describe('Connection Recovery', () => {
    test('should attempt reconnection on close', () => {
      const ws = new WebSocket('ws://localhost:8080/ws');
      let reconnectAttempted = false;

      ws.addEventListener('close', (event) => {
        if (event.code !== 1000) { // Not a normal close
          // Simulate reconnection attempt
          setTimeout(() => {
            reconnectAttempted = true;
          }, 1000);
        }
      });

      // Simulate unexpected close
      const closeHandler = mockEventListeners.close[0];
      closeHandler({ code: 1006 }); // Abnormal closure

      // Fast-forward time
      jest.advanceTimersByTime(1000);

      expect(reconnectAttempted).toBe(true);
    });

    test('should handle reconnection with authentication', () => {
      const ws = new WebSocket('ws://localhost:8080/ws');
      let authSent = false;

      ws.addEventListener('open', () => {
        // Simulate sending auth on reconnection
        const authMessage = {
          type: 'auth',
          token: 'stored-session-token'
        };
        ws.send(JSON.stringify(authMessage));
        authSent = true;
      });

      // Simulate connection open
      const openHandler = mockEventListeners.open[0];
      openHandler();

      expect(authSent).toBe(true);
      expect(ws.send).toHaveBeenCalledWith(JSON.stringify({
        type: 'auth',
        token: 'stored-session-token'
      }));
    });
  });
});
