/**
 * Performance Tests
 * Tests for page load times, WebSocket performance, memory usage, etc.
 */

import { test, expect } from '@playwright/test';

test.describe('Performance Tests', () => {
  test.describe('Page Load Performance', () => {
    test('should load home page within acceptable time', async ({ page }) => {
      const startTime = Date.now();
      
      await page.goto('/');
      await page.waitForLoadState('networkidle');
      
      const loadTime = Date.now() - startTime;
      
      // Page should load within 3 seconds
      expect(loadTime).toBeLessThan(3000);
      
      // Verify critical elements are loaded
      await expect(page.locator('header')).toBeVisible();
      await expect(page.locator('main')).toBeVisible();
      await expect(page.locator('#login-form')).toBeVisible();
    });

    test('should load chat page efficiently', async ({ page }) => {
      // Login first
      await page.goto('/');
      await page.fill('#identifier', 'johndoe');
      await page.fill('#password', 'password123');
      await page.click('button[type="submit"]');
      
      const startTime = Date.now();
      
      await page.goto('/chat');
      await page.waitForLoadState('networkidle');
      
      // Wait for WebSocket connection
      await page.waitForFunction(() => {
        return window.websocket && window.websocket.readyState === WebSocket.OPEN;
      });
      
      const loadTime = Date.now() - startTime;
      
      // Chat page should load within 2 seconds
      expect(loadTime).toBeLessThan(2000);
      
      // Verify chat interface is ready
      await expect(page.locator('.chat-interface')).toBeVisible();
      await expect(page.locator('.conversations-list')).toBeVisible();
    });

    test('should handle large conversation lists efficiently', async ({ page }) => {
      await page.goto('/');
      await page.fill('#identifier', 'johndoe');
      await page.fill('#password', 'password123');
      await page.click('button[type="submit"]');
      
      const startTime = Date.now();
      
      await page.goto('/chat');
      await page.waitForSelector('.conversation-item');
      
      const loadTime = Date.now() - startTime;
      
      // Should load even with many conversations within 3 seconds
      expect(loadTime).toBeLessThan(3000);
      
      // Verify conversations are loaded
      const conversationCount = await page.locator('.conversation-item').count();
      expect(conversationCount).toBeGreaterThan(0);
    });
  });

  test.describe('WebSocket Performance', () => {
    test('should establish WebSocket connection quickly', async ({ page }) => {
      await page.goto('/');
      await page.fill('#identifier', 'johndoe');
      await page.fill('#password', 'password123');
      await page.click('button[type="submit"]');
      
      const startTime = Date.now();
      
      await page.goto('/chat');
      
      // Wait for WebSocket connection
      await page.waitForFunction(() => {
        return window.websocket && window.websocket.readyState === WebSocket.OPEN;
      });
      
      const connectionTime = Date.now() - startTime;
      
      // WebSocket should connect within 1 second
      expect(connectionTime).toBeLessThan(1000);
    });

    test('should handle message sending with low latency', async ({ page, context }) => {
      const page1 = page;
      const page2 = await context.newPage();
      
      // Setup both users
      await page1.goto('/');
      await page1.fill('#identifier', 'johndoe');
      await page1.fill('#password', 'password123');
      await page1.click('button[type="submit"]');
      await page1.goto('/chat');
      
      await page2.goto('/');
      await page2.fill('#identifier', 'janesmith');
      await page2.fill('#password', 'password123');
      await page2.click('button[type="submit"]');
      await page2.goto('/chat');
      
      // Start conversation
      await page1.click('[data-testid="conversation-janesmith"]');
      await page2.click('[data-testid="conversation-johndoe"]');
      
      const startTime = Date.now();
      
      // Send message from page1
      await page1.fill('[data-testid="message-input"]', 'Performance test message');
      await page1.click('[data-testid="send-button"]');
      
      // Wait for message to appear on page2
      await page2.waitForSelector('.message.received:last-child');
      
      const latency = Date.now() - startTime;
      
      // Message should appear within 500ms
      expect(latency).toBeLessThan(500);
    });

    test('should handle rapid message sending', async ({ page }) => {
      await page.goto('/');
      await page.fill('#identifier', 'johndoe');
      await page.fill('#password', 'password123');
      await page.click('button[type="submit"]');
      await page.goto('/chat');
      await page.click('[data-testid="conversation-janesmith"]');
      
      const startTime = Date.now();
      
      // Send 20 messages rapidly
      for (let i = 0; i < 20; i++) {
        await page.fill('[data-testid="message-input"]', `Rapid message ${i + 1}`);
        await page.click('[data-testid="send-button"]');
      }
      
      // Wait for all messages to be sent
      await page.waitForSelector('.message.sent:nth-child(20)');
      
      const totalTime = Date.now() - startTime;
      
      // Should handle 20 messages within 5 seconds
      expect(totalTime).toBeLessThan(5000);
      
      // Verify all messages are displayed
      const messageCount = await page.locator('.message.sent').count();
      expect(messageCount).toBeGreaterThanOrEqual(20);
    });
  });

  test.describe('Memory and Resource Usage', () => {
    test('should not have memory leaks during extended chat usage', async ({ page }) => {
      await page.goto('/');
      await page.fill('#identifier', 'johndoe');
      await page.fill('#password', 'password123');
      await page.click('button[type="submit"]');
      await page.goto('/chat');
      
      // Get initial memory usage
      const initialMemory = await page.evaluate(() => {
        return performance.memory ? performance.memory.usedJSHeapSize : 0;
      });
      
      // Simulate extended chat usage
      for (let i = 0; i < 50; i++) {
        await page.fill('[data-testid="message-input"]', `Memory test message ${i + 1}`);
        await page.click('[data-testid="send-button"]');
        
        if (i % 10 === 0) {
          // Switch conversations periodically
          await page.click('[data-testid="conversation-list"] .conversation-item:nth-child(2)');
          await page.waitForTimeout(100);
          await page.click('[data-testid="conversation-list"] .conversation-item:first-child');
        }
      }
      
      // Force garbage collection if available
      await page.evaluate(() => {
        if (window.gc) {
          window.gc();
        }
      });
      
      const finalMemory = await page.evaluate(() => {
        return performance.memory ? performance.memory.usedJSHeapSize : 0;
      });
      
      // Memory usage should not increase dramatically (allow for 50MB increase)
      const memoryIncrease = finalMemory - initialMemory;
      expect(memoryIncrease).toBeLessThan(50 * 1024 * 1024); // 50MB
    });

    test('should handle DOM efficiently with many messages', async ({ page }) => {
      await page.goto('/');
      await page.fill('#identifier', 'johndoe');
      await page.fill('#password', 'password123');
      await page.click('button[type="submit"]');
      await page.goto('/chat');
      await page.click('[data-testid="conversation-janesmith"]');
      
      // Send many messages to test DOM performance
      for (let i = 0; i < 100; i++) {
        await page.fill('[data-testid="message-input"]', `DOM test message ${i + 1}`);
        await page.click('[data-testid="send-button"]');
        
        if (i % 20 === 0) {
          // Check scroll performance periodically
          const scrollStart = Date.now();
          await page.evaluate(() => {
            document.querySelector('.messages-container').scrollTop = 0;
          });
          await page.evaluate(() => {
            const container = document.querySelector('.messages-container');
            container.scrollTop = container.scrollHeight;
          });
          const scrollTime = Date.now() - scrollStart;
          
          // Scrolling should be smooth (under 100ms)
          expect(scrollTime).toBeLessThan(100);
        }
      }
      
      // Verify DOM node count is reasonable
      const nodeCount = await page.evaluate(() => {
        return document.querySelectorAll('*').length;
      });
      
      // Should not have excessive DOM nodes (under 5000)
      expect(nodeCount).toBeLessThan(5000);
    });
  });

  test.describe('Network Performance', () => {
    test('should handle slow network conditions', async ({ page, context }) => {
      // Simulate slow 3G network
      await context.route('**/*', async route => {
        await new Promise(resolve => setTimeout(resolve, 100)); // Add 100ms delay
        await route.continue();
      });
      
      const startTime = Date.now();
      
      await page.goto('/');
      await page.waitForLoadState('networkidle');
      
      const loadTime = Date.now() - startTime;
      
      // Should still load within reasonable time on slow network
      expect(loadTime).toBeLessThan(10000); // 10 seconds
      
      // Verify page is functional
      await expect(page.locator('#login-form')).toBeVisible();
    });

    test('should handle network interruptions gracefully', async ({ page, context }) => {
      await page.goto('/');
      await page.fill('#identifier', 'johndoe');
      await page.fill('#password', 'password123');
      await page.click('button[type="submit"]');
      await page.goto('/chat');
      
      // Wait for initial connection
      await page.waitForFunction(() => {
        return window.websocket && window.websocket.readyState === WebSocket.OPEN;
      });
      
      // Simulate network interruption
      await context.setOffline(true);
      
      // Verify offline handling
      await expect(page.locator('.connection-status')).toHaveText('Disconnected');
      
      // Restore network
      await context.setOffline(false);
      
      // Verify reconnection
      await page.waitForFunction(() => {
        return window.websocket && window.websocket.readyState === WebSocket.OPEN;
      }, { timeout: 10000 });
      
      await expect(page.locator('.connection-status')).toHaveText('Connected');
    });
  });

  test.describe('Rendering Performance', () => {
    test('should render large post lists efficiently', async ({ page }) => {
      await page.goto('/');
      await page.fill('#identifier', 'johndoe');
      await page.fill('#password', 'password123');
      await page.click('button[type="submit"]');
      
      const startTime = Date.now();
      
      await page.goto('/home');
      await page.waitForSelector('.post-card');
      
      const renderTime = Date.now() - startTime;
      
      // Should render post list within 2 seconds
      expect(renderTime).toBeLessThan(2000);
      
      // Test scroll performance
      const scrollStart = Date.now();
      
      await page.evaluate(() => {
        window.scrollTo(0, document.body.scrollHeight);
      });
      
      await page.waitForTimeout(100);
      
      const scrollTime = Date.now() - scrollStart;
      
      // Scrolling should be smooth
      expect(scrollTime).toBeLessThan(200);
    });

    test('should handle dynamic content updates efficiently', async ({ page, context }) => {
      const page1 = page;
      const page2 = await context.newPage();
      
      await page1.goto('/');
      await page1.fill('#identifier', 'johndoe');
      await page1.fill('#password', 'password123');
      await page1.click('button[type="submit"]');
      await page1.goto('/home');
      
      await page2.goto('/');
      await page2.fill('#identifier', 'janesmith');
      await page2.fill('#password', 'password123');
      await page2.click('button[type="submit"]');
      
      // Create multiple posts rapidly from page2
      for (let i = 0; i < 5; i++) {
        await page2.goto('/create-post');
        await page2.fill('[data-testid="post-title"]', `Performance Post ${i + 1}`);
        await page2.fill('[data-testid="post-content"]', `Content for performance test post ${i + 1}`);
        await page2.check('[data-testid="category-technology"]');
        await page2.click('[data-testid="submit-post"]');
        await page2.waitForURL('/home');
      }
      
      // Verify page1 updates efficiently
      await page1.waitForSelector('.post-card:first-child');
      
      const postCount = await page1.locator('.post-card').count();
      expect(postCount).toBeGreaterThanOrEqual(5);
      
      // Verify latest post appears first
      await expect(page1.locator('.post-card:first-child .post-title')).toContainText('Performance Post 5');
    });
  });

  test.describe('Bundle Size and Loading', () => {
    test('should have reasonable JavaScript bundle size', async ({ page }) => {
      // Monitor network requests
      const jsRequests = [];
      
      page.on('response', response => {
        if (response.url().endsWith('.js')) {
          jsRequests.push({
            url: response.url(),
            size: response.headers()['content-length']
          });
        }
      });
      
      await page.goto('/');
      await page.waitForLoadState('networkidle');
      
      // Calculate total JS size
      const totalJSSize = jsRequests.reduce((total, request) => {
        return total + (parseInt(request.size) || 0);
      }, 0);
      
      // Total JS should be under 1MB
      expect(totalJSSize).toBeLessThan(1024 * 1024);
    });

    test('should load critical resources first', async ({ page }) => {
      const resourceLoadOrder = [];
      
      page.on('response', response => {
        if (response.url().includes('.css') || response.url().includes('.js')) {
          resourceLoadOrder.push({
            url: response.url(),
            timestamp: Date.now()
          });
        }
      });
      
      await page.goto('/');
      await page.waitForLoadState('networkidle');
      
      // CSS should load before non-critical JS
      const cssLoaded = resourceLoadOrder.find(r => r.url.includes('.css'));
      const jsLoaded = resourceLoadOrder.find(r => r.url.includes('.js') && !r.url.includes('critical'));
      
      if (cssLoaded && jsLoaded) {
        expect(cssLoaded.timestamp).toBeLessThanOrEqual(jsLoaded.timestamp);
      }
    });
  });
});
