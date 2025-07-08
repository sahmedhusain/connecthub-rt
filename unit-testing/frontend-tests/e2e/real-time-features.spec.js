/**
 * Real-Time Features E2E Tests
 * Tests for WebSocket functionality, real-time messaging, typing indicators, etc.
 */

import { test, expect } from '@playwright/test';

test.describe('Real-Time Features', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to the application
    await page.goto('/');
    
    // Login as test user
    await page.fill('#identifier', 'johndoe');
    await page.fill('#password', 'password123');
    await page.click('button[type="submit"]');
    
    // Wait for successful login and navigation
    await page.waitForURL('/home');
  });

  test.describe('Real-Time Messaging', () => {
    test('should establish WebSocket connection', async ({ page }) => {
      // Navigate to chat page
      await page.goto('/chat');
      
      // Wait for WebSocket connection to be established
      await page.waitForFunction(() => {
        return window.websocket && window.websocket.readyState === WebSocket.OPEN;
      });

      // Verify connection status indicator
      const connectionStatus = page.locator('.connection-status');
      await expect(connectionStatus).toHaveText('Connected');
    });

    test('should send and receive messages in real-time', async ({ page, context }) => {
      // Open two pages to simulate two users
      const page1 = page;
      const page2 = await context.newPage();

      // Login as first user on page1
      await page1.goto('/chat');
      await page1.waitForLoadState('networkidle');

      // Login as second user on page2
      await page2.goto('/');
      await page2.fill('#identifier', 'janesmith');
      await page2.fill('#password', 'password123');
      await page2.click('button[type="submit"]');
      await page2.waitForURL('/home');
      await page2.goto('/chat');
      await page2.waitForLoadState('networkidle');

      // Start a conversation between the two users
      await page1.click('[data-testid="start-conversation"]');
      await page1.fill('[data-testid="user-search"]', 'janesmith');
      await page1.click('[data-testid="user-janesmith"]');

      // Send a message from page1
      const messageText = 'Hello from user 1!';
      await page1.fill('[data-testid="message-input"]', messageText);
      await page1.click('[data-testid="send-button"]');

      // Verify message appears on page1
      await expect(page1.locator('.message.sent').last()).toContainText(messageText);

      // Verify message appears on page2 in real-time
      await expect(page2.locator('.message.received').last()).toContainText(messageText);

      // Send a reply from page2
      const replyText = 'Hello back from user 2!';
      await page2.fill('[data-testid="message-input"]', replyText);
      await page2.click('[data-testid="send-button"]');

      // Verify reply appears on both pages
      await expect(page2.locator('.message.sent').last()).toContainText(replyText);
      await expect(page1.locator('.message.received').last()).toContainText(replyText);
    });

    test('should show typing indicators', async ({ page, context }) => {
      const page1 = page;
      const page2 = await context.newPage();

      // Setup both users in a conversation
      await page1.goto('/chat');
      await page2.goto('/');
      await page2.fill('#identifier', 'janesmith');
      await page2.fill('#password', 'password123');
      await page2.click('button[type="submit"]');
      await page2.goto('/chat');

      // Start conversation
      await page1.click('[data-testid="conversation-janesmith"]');
      await page2.click('[data-testid="conversation-johndoe"]');

      // Start typing on page1
      await page1.fill('[data-testid="message-input"]', 'I am typing...');

      // Verify typing indicator appears on page2
      await expect(page2.locator('.typing-indicator')).toBeVisible();
      await expect(page2.locator('.typing-indicator')).toContainText('johndoe is typing...');

      // Stop typing (clear input)
      await page1.fill('[data-testid="message-input"]', '');

      // Verify typing indicator disappears
      await expect(page2.locator('.typing-indicator')).toBeHidden();
    });

    test('should update message status in real-time', async ({ page, context }) => {
      const page1 = page;
      const page2 = await context.newPage();

      // Setup conversation
      await page1.goto('/chat');
      await page2.goto('/');
      await page2.fill('#identifier', 'janesmith');
      await page2.fill('#password', 'password123');
      await page2.click('button[type="submit"]');
      await page2.goto('/chat');

      // Send message from page1
      await page1.click('[data-testid="conversation-janesmith"]');
      await page1.fill('[data-testid="message-input"]', 'Test message status');
      await page1.click('[data-testid="send-button"]');

      // Initially message should show single tick (sent)
      await expect(page1.locator('.message.sent .status-icon')).toHaveText('✓');

      // When page2 opens the conversation, message should show double tick (read)
      await page2.click('[data-testid="conversation-johndoe"]');
      
      // Verify read status on page1
      await expect(page1.locator('.message.sent .status-icon')).toHaveText('✓✓');
    });

    test('should handle connection interruptions gracefully', async ({ page }) => {
      await page.goto('/chat');
      
      // Wait for initial connection
      await page.waitForFunction(() => window.websocket?.readyState === WebSocket.OPEN);

      // Simulate connection loss
      await page.evaluate(() => {
        if (window.websocket) {
          window.websocket.close();
        }
      });

      // Verify disconnection status
      await expect(page.locator('.connection-status')).toHaveText('Disconnected');

      // Verify reconnection attempt
      await page.waitForFunction(() => window.websocket?.readyState === WebSocket.OPEN, {
        timeout: 10000
      });

      await expect(page.locator('.connection-status')).toHaveText('Connected');
    });
  });

  test.describe('Real-Time Notifications', () => {
    test('should show new message notifications', async ({ page, context }) => {
      const page1 = page;
      const page2 = await context.newPage();

      // Setup users
      await page1.goto('/home'); // User 1 on home page
      await page2.goto('/');
      await page2.fill('#identifier', 'janesmith');
      await page2.fill('#password', 'password123');
      await page2.click('button[type="submit"]');
      await page2.goto('/chat');

      // Send message from page2 to page1
      await page2.click('[data-testid="start-conversation"]');
      await page2.fill('[data-testid="user-search"]', 'johndoe');
      await page2.click('[data-testid="user-johndoe"]');
      await page2.fill('[data-testid="message-input"]', 'New message notification test');
      await page2.click('[data-testid="send-button"]');

      // Verify notification appears on page1
      await expect(page1.locator('.notification.new-message')).toBeVisible();
      await expect(page1.locator('.notification.new-message')).toContainText('New message from janesmith');

      // Verify unread count updates
      await expect(page1.locator('.chat-link .unread-count')).toHaveText('1');
    });

    test('should update conversation list in real-time', async ({ page, context }) => {
      const page1 = page;
      const page2 = await context.newPage();

      await page1.goto('/chat');
      await page2.goto('/');
      await page2.fill('#identifier', 'janesmith');
      await page2.fill('#password', 'password123');
      await page2.click('button[type="submit"]');
      await page2.goto('/chat');

      // Send message from page2
      await page2.click('[data-testid="start-conversation"]');
      await page2.fill('[data-testid="user-search"]', 'johndoe');
      await page2.click('[data-testid="user-johndoe"]');
      await page2.fill('[data-testid="message-input"]', 'Latest message');
      await page2.click('[data-testid="send-button"]');

      // Verify conversation appears at top of list on page1
      await expect(page1.locator('.conversation-item').first()).toContainText('janesmith');
      await expect(page1.locator('.conversation-item').first().locator('.last-message')).toContainText('Latest message');

      // Verify timestamp is updated
      await expect(page1.locator('.conversation-item').first().locator('.timestamp')).toContainText('now');
    });
  });

  test.describe('Real-Time Post Updates', () => {
    test('should show new posts in real-time', async ({ page, context }) => {
      const page1 = page;
      const page2 = await context.newPage();

      // User 1 on home page
      await page1.goto('/home');

      // User 2 creates a new post
      await page2.goto('/');
      await page2.fill('#identifier', 'janesmith');
      await page2.fill('#password', 'password123');
      await page2.click('button[type="submit"]');
      await page2.goto('/create-post');

      await page2.fill('[data-testid="post-title"]', 'Real-time Post Test');
      await page2.fill('[data-testid="post-content"]', 'This post should appear in real-time');
      await page2.check('[data-testid="category-technology"]');
      await page2.click('[data-testid="submit-post"]');

      // Verify new post appears on page1 without refresh
      await expect(page1.locator('.post-card').first().locator('.post-title')).toContainText('Real-time Post Test');
    });

    test('should update comment counts in real-time', async ({ page, context }) => {
      const page1 = page;
      const page2 = await context.newPage();

      // Both users view the same post
      await page1.goto('/post?id=1');
      await page2.goto('/');
      await page2.fill('#identifier', 'janesmith');
      await page2.fill('#password', 'password123');
      await page2.click('button[type="submit"]');
      await page2.goto('/post?id=1');

      // User 2 adds a comment
      await page2.fill('[data-testid="comment-input"]', 'Real-time comment test');
      await page2.click('[data-testid="submit-comment"]');

      // Verify comment appears on page1
      await expect(page1.locator('.comment').last()).toContainText('Real-time comment test');

      // Verify comment count updates
      await expect(page1.locator('.comment-count')).toContainText(/\d+ comments?/);
    });
  });

  test.describe('Online Status', () => {
    test('should show user online status', async ({ page, context }) => {
      const page1 = page;
      const page2 = await context.newPage();

      await page1.goto('/chat');

      // User 2 comes online
      await page2.goto('/');
      await page2.fill('#identifier', 'janesmith');
      await page2.fill('#password', 'password123');
      await page2.click('button[type="submit"]');

      // Verify online status indicator appears
      await expect(page1.locator('[data-user="janesmith"] .online-indicator')).toBeVisible();
      await expect(page1.locator('[data-user="janesmith"] .status-text')).toHaveText('Online');

      // User 2 goes offline
      await page2.close();

      // Verify offline status
      await expect(page1.locator('[data-user="janesmith"] .online-indicator')).toBeHidden();
      await expect(page1.locator('[data-user="janesmith"] .status-text')).toHaveText('Offline');
    });

    test('should show last seen timestamp when offline', async ({ page, context }) => {
      const page1 = page;
      const page2 = await context.newPage();

      await page1.goto('/chat');

      // User 2 comes online then goes offline
      await page2.goto('/');
      await page2.fill('#identifier', 'janesmith');
      await page2.fill('#password', 'password123');
      await page2.click('button[type="submit"]');
      await page2.close();

      // Verify last seen timestamp
      await expect(page1.locator('[data-user="janesmith"] .last-seen')).toContainText(/Last seen \d+ (second|minute|hour)s? ago/);
    });
  });

  test.describe('Performance and Reliability', () => {
    test('should handle high message volume', async ({ page }) => {
      await page.goto('/chat');
      await page.click('[data-testid="conversation-janesmith"]');

      // Send multiple messages rapidly
      for (let i = 0; i < 10; i++) {
        await page.fill('[data-testid="message-input"]', `Message ${i + 1}`);
        await page.click('[data-testid="send-button"]');
        await page.waitForTimeout(100); // Small delay between messages
      }

      // Verify all messages are displayed
      const messages = page.locator('.message.sent');
      await expect(messages).toHaveCount(10);

      // Verify messages are in correct order
      await expect(messages.last()).toContainText('Message 10');
    });

    test('should maintain scroll position during real-time updates', async ({ page, context }) => {
      const page1 = page;
      const page2 = await context.newPage();

      // Setup conversation with many messages
      await page1.goto('/chat');
      await page1.click('[data-testid="conversation-janesmith"]');

      // Scroll to middle of conversation
      await page1.evaluate(() => {
        const messagesContainer = document.querySelector('.messages-container');
        messagesContainer.scrollTop = messagesContainer.scrollHeight / 2;
      });

      const initialScrollPosition = await page1.evaluate(() => {
        return document.querySelector('.messages-container').scrollTop;
      });

      // User 2 sends a message
      await page2.goto('/');
      await page2.fill('#identifier', 'janesmith');
      await page2.fill('#password', 'password123');
      await page2.click('button[type="submit"]');
      await page2.goto('/chat');
      await page2.click('[data-testid="conversation-johndoe"]');
      await page2.fill('[data-testid="message-input"]', 'New message while scrolled');
      await page2.click('[data-testid="send-button"]');

      // Verify scroll position is maintained (not auto-scrolled to bottom)
      const currentScrollPosition = await page1.evaluate(() => {
        return document.querySelector('.messages-container').scrollTop;
      });

      expect(currentScrollPosition).toBe(initialScrollPosition);

      // Verify new message indicator appears
      await expect(page1.locator('.new-message-indicator')).toBeVisible();
    });
  });
});
