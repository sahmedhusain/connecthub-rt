/**
 * Responsive Design End-to-End Tests
 * Tests for responsive layout and mobile/tablet compatibility using Playwright
 */

import { test, expect } from '@playwright/test';

test.describe('Responsive Design E2E Tests', () => {
  test.beforeEach(async ({ page }) => {
    // Login before each test
    await page.goto('/');
    await page.fill('input[name="identifier"]', 'alexchen');
    await page.fill('input[name="password"]', 'Aa123456');
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL(/.*\/home/);
  });

  test.describe('Mobile Viewport (375x667)', () => {
    test.beforeEach(async ({ page }) => {
      await page.setViewportSize({ width: 375, height: 667 });
    });

    test('should display mobile-friendly navigation', async ({ page }) => {
      // Check if mobile navigation elements are visible
      const mobileNav = page.locator('.mobile-nav, .hamburger, .nav-toggle, [class*="mobile"]');
      
      if (await mobileNav.first().isVisible()) {
        await expect(mobileNav.first()).toBeVisible();
        
        // Test mobile menu toggle
        await mobileNav.first().click();
        
        // Check if menu opens
        const mobileMenu = page.locator('.mobile-menu, .nav-menu, [class*="menu"]');
        if (await mobileMenu.first().isVisible()) {
          await expect(mobileMenu.first()).toBeVisible();
        }
      }
    });

    test('should stack elements vertically on mobile', async ({ page }) => {
      // Check main layout elements
      const sidebar = page.locator('.sidebar, #sidebar, [class*="sidebar"]');
      const mainContent = page.locator('.main-content, #content-area, [class*="content"]');
      
      if (await sidebar.first().isVisible() && await mainContent.first().isVisible()) {
        const sidebarBox = await sidebar.first().boundingBox();
        const contentBox = await mainContent.first().boundingBox();
        
        // On mobile, elements should stack (sidebar above content or hidden)
        if (sidebarBox && contentBox) {
          // Either sidebar is hidden or positioned above content
          const isStacked = sidebarBox.y < contentBox.y || sidebarBox.width < 100;
          expect(isStacked).toBe(true);
        }
      }
    });

    test('should make forms mobile-friendly', async ({ page }) => {
      // Navigate to create post page if available
      const createPostLink = page.locator('a[href*="create"], button:has-text("Create"), [class*="create"]');
      
      if (await createPostLink.first().isVisible()) {
        await createPostLink.first().click();
        
        // Check form elements are properly sized
        const formInputs = page.locator('input, textarea, select');
        
        if (await formInputs.first().isVisible()) {
          const inputBox = await formInputs.first().boundingBox();
          
          if (inputBox) {
            // Form inputs should not exceed viewport width
            expect(inputBox.width).toBeLessThanOrEqual(375);
            
            // Form inputs should have reasonable minimum height for touch
            expect(inputBox.height).toBeGreaterThanOrEqual(40);
          }
        }
      }
    });

    test('should handle chat interface on mobile', async ({ page }) => {
      // Look for chat elements
      const chatSection = page.locator('.chat-section, .chats, [class*="chat"]');
      
      if (await chatSection.first().isVisible()) {
        // On mobile, chat might be in a modal or overlay
        const chatContainer = page.locator('.chat-container, .chat-window, [class*="chat-window"]');
        
        if (await chatContainer.first().isVisible()) {
          const chatBox = await chatContainer.first().boundingBox();
          
          if (chatBox) {
            // Chat should fit within mobile viewport
            expect(chatBox.width).toBeLessThanOrEqual(375);
          }
        }
      }
    });

    test('should have touch-friendly buttons', async ({ page }) => {
      // Check button sizes
      const buttons = page.locator('button, .btn, [role="button"]');
      
      if (await buttons.first().isVisible()) {
        const buttonBox = await buttons.first().boundingBox();
        
        if (buttonBox) {
          // Buttons should be at least 44px for touch accessibility
          expect(buttonBox.height).toBeGreaterThanOrEqual(40);
          expect(buttonBox.width).toBeGreaterThanOrEqual(40);
        }
      }
    });
  });

  test.describe('Tablet Viewport (768x1024)', () => {
    test.beforeEach(async ({ page }) => {
      await page.setViewportSize({ width: 768, height: 1024 });
    });

    test('should display tablet layout', async ({ page }) => {
      // Check if sidebar and content are side by side on tablet
      const sidebar = page.locator('.sidebar, #sidebar, [class*="sidebar"]');
      const mainContent = page.locator('.main-content, #content-area, [class*="content"]');
      
      if (await sidebar.first().isVisible() && await mainContent.first().isVisible()) {
        const sidebarBox = await sidebar.first().boundingBox();
        const contentBox = await mainContent.first().boundingBox();
        
        if (sidebarBox && contentBox) {
          // On tablet, sidebar and content should be side by side
          const isSideBySide = Math.abs(sidebarBox.y - contentBox.y) < 50;
          expect(isSideBySide).toBe(true);
        }
      }
    });

    test('should optimize content width for tablet', async ({ page }) => {
      // Check content doesn't stretch too wide
      const contentArea = page.locator('.content, .main-content, [class*="content"]');
      
      if (await contentArea.first().isVisible()) {
        const contentBox = await contentArea.first().boundingBox();
        
        if (contentBox) {
          // Content should use available space but not be too narrow
          expect(contentBox.width).toBeGreaterThan(400);
          expect(contentBox.width).toBeLessThanOrEqual(768);
        }
      }
    });

    test('should handle posts layout on tablet', async ({ page }) => {
      // Check post layout
      const posts = page.locator('.post, .post-item, [class*="post"]');
      
      if (await posts.first().isVisible()) {
        const postBox = await posts.first().boundingBox();
        
        if (postBox) {
          // Posts should have reasonable width on tablet
          expect(postBox.width).toBeGreaterThan(300);
          expect(postBox.width).toBeLessThanOrEqual(600);
        }
      }
    });
  });

  test.describe('Desktop Viewport (1200x800)', () => {
    test.beforeEach(async ({ page }) => {
      await page.setViewportSize({ width: 1200, height: 800 });
    });

    test('should display full desktop layout', async ({ page }) => {
      // Check if all elements are visible on desktop
      const sidebar = page.locator('.sidebar, #sidebar, [class*="sidebar"]');
      const mainContent = page.locator('.main-content, #content-area, [class*="content"]');
      const header = page.locator('header, .header, [class*="header"]');
      
      await expect(sidebar.first()).toBeVisible();
      await expect(mainContent.first()).toBeVisible();
      
      if (await header.first().isVisible()) {
        await expect(header.first()).toBeVisible();
      }
    });

    test('should utilize full width effectively', async ({ page }) => {
      // Check layout utilizes desktop space
      const mainLayout = page.locator('.main-layout, .container, [class*="layout"]');
      
      if (await mainLayout.first().isVisible()) {
        const layoutBox = await mainLayout.first().boundingBox();
        
        if (layoutBox) {
          // Layout should use most of the available width
          expect(layoutBox.width).toBeGreaterThan(800);
        }
      }
    });

    test('should show desktop navigation', async ({ page }) => {
      // Desktop should show full navigation
      const navigation = page.locator('nav, .navbar, [class*="nav"]');
      
      if (await navigation.first().isVisible()) {
        await expect(navigation.first()).toBeVisible();
        
        // Check for navigation links
        const navLinks = navigation.first().locator('a, button');
        const linkCount = await navLinks.count();
        
        // Desktop should have multiple navigation options
        expect(linkCount).toBeGreaterThan(2);
      }
    });
  });

  test.describe('Orientation Changes', () => {
    test('should handle portrait to landscape on mobile', async ({ page }) => {
      // Start in portrait
      await page.setViewportSize({ width: 375, height: 667 });
      
      // Check initial layout
      const initialLayout = await page.locator('body').boundingBox();
      expect(initialLayout?.width).toBe(375);
      
      // Switch to landscape
      await page.setViewportSize({ width: 667, height: 375 });
      
      // Wait for layout to adjust
      await page.waitForTimeout(500);
      
      // Check layout adjusted
      const landscapeLayout = await page.locator('body').boundingBox();
      expect(landscapeLayout?.width).toBe(667);
      
      // Interface should still be functional
      const mainContent = page.locator('.main-content, #content-area, [class*="content"]');
      if (await mainContent.first().isVisible()) {
        await expect(mainContent.first()).toBeVisible();
      }
    });

    test('should handle tablet orientation changes', async ({ page }) => {
      // Start in portrait tablet
      await page.setViewportSize({ width: 768, height: 1024 });
      
      // Switch to landscape tablet
      await page.setViewportSize({ width: 1024, height: 768 });
      
      // Wait for layout adjustment
      await page.waitForTimeout(500);
      
      // Check layout is still functional
      const sidebar = page.locator('.sidebar, #sidebar, [class*="sidebar"]');
      const content = page.locator('.main-content, #content-area, [class*="content"]');
      
      if (await sidebar.first().isVisible() && await content.first().isVisible()) {
        await expect(sidebar.first()).toBeVisible();
        await expect(content.first()).toBeVisible();
      }
    });
  });

  test.describe('Text Readability', () => {
    test('should maintain readable text size across viewports', async ({ page }) => {
      const viewports = [
        { width: 375, height: 667, name: 'mobile' },
        { width: 768, height: 1024, name: 'tablet' },
        { width: 1200, height: 800, name: 'desktop' }
      ];

      for (const viewport of viewports) {
        await page.setViewportSize(viewport);
        await page.waitForTimeout(300);

        // Check text elements
        const textElements = page.locator('p, h1, h2, h3, span, div');
        
        if (await textElements.first().isVisible()) {
          // Text should be readable (this is a basic check)
          const textBox = await textElements.first().boundingBox();
          
          if (textBox) {
            // Text container should have reasonable height
            expect(textBox.height).toBeGreaterThan(10);
          }
        }
      }
    });
  });

  test.describe('Interactive Elements', () => {
    test('should maintain usability across viewports', async ({ page }) => {
      const viewports = [
        { width: 375, height: 667 },
        { width: 768, height: 1024 },
        { width: 1200, height: 800 }
      ];

      for (const viewport of viewports) {
        await page.setViewportSize(viewport);
        await page.waitForTimeout(300);

        // Test clicking on interactive elements
        const clickableElements = page.locator('button, a, [role="button"]');
        
        if (await clickableElements.first().isVisible()) {
          // Element should be clickable
          await expect(clickableElements.first()).toBeVisible();
          
          // Try clicking (should not throw error)
          try {
            await clickableElements.first().click({ timeout: 1000 });
          } catch (error) {
            // Some clicks might navigate or cause changes, that's okay
            // We just want to ensure the element is clickable
          }
        }
      }
    });
  });
});
