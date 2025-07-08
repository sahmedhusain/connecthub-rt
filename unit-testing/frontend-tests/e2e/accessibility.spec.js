/**
 * Accessibility Tests
 * Tests for WCAG compliance, keyboard navigation, screen reader support, etc.
 */

import { test, expect } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';

test.describe('Accessibility Tests', () => {
  test.describe('WCAG Compliance', () => {
    test('should pass axe accessibility tests on login page', async ({ page }) => {
      await page.goto('/');
      
      const accessibilityScanResults = await new AxeBuilder({ page }).analyze();
      
      expect(accessibilityScanResults.violations).toEqual([]);
    });

    test('should pass axe accessibility tests on home page', async ({ page }) => {
      await page.goto('/');
      await page.fill('#identifier', 'johndoe');
      await page.fill('#password', 'password123');
      await page.click('button[type="submit"]');
      await page.waitForURL('/home');
      
      const accessibilityScanResults = await new AxeBuilder({ page }).analyze();
      
      expect(accessibilityScanResults.violations).toEqual([]);
    });

    test('should pass axe accessibility tests on chat page', async ({ page }) => {
      await page.goto('/');
      await page.fill('#identifier', 'johndoe');
      await page.fill('#password', 'password123');
      await page.click('button[type="submit"]');
      await page.goto('/chat');
      await page.waitForLoadState('networkidle');
      
      const accessibilityScanResults = await new AxeBuilder({ page }).analyze();
      
      expect(accessibilityScanResults.violations).toEqual([]);
    });

    test('should have proper heading hierarchy', async ({ page }) => {
      await page.goto('/');
      await page.fill('#identifier', 'johndoe');
      await page.fill('#password', 'password123');
      await page.click('button[type="submit"]');
      await page.goto('/home');
      
      // Check heading hierarchy
      const headings = await page.locator('h1, h2, h3, h4, h5, h6').allTextContents();
      
      // Should have at least one h1
      const h1Elements = await page.locator('h1').count();
      expect(h1Elements).toBeGreaterThanOrEqual(1);
      
      // Verify logical heading structure
      const headingLevels = await page.evaluate(() => {
        const headings = Array.from(document.querySelectorAll('h1, h2, h3, h4, h5, h6'));
        return headings.map(h => parseInt(h.tagName.charAt(1)));
      });
      
      // First heading should be h1
      expect(headingLevels[0]).toBe(1);
    });

    test('should have proper color contrast', async ({ page }) => {
      await page.goto('/');
      
      // Test with axe color-contrast rule specifically
      const accessibilityScanResults = await new AxeBuilder({ page })
        .withTags(['wcag2aa'])
        .include('body')
        .analyze();
      
      const colorContrastViolations = accessibilityScanResults.violations.filter(
        violation => violation.id === 'color-contrast'
      );
      
      expect(colorContrastViolations).toEqual([]);
    });
  });

  test.describe('Keyboard Navigation', () => {
    test('should navigate login form with keyboard', async ({ page }) => {
      await page.goto('/');
      
      // Tab to username field
      await page.keyboard.press('Tab');
      await expect(page.locator('#identifier')).toBeFocused();
      
      // Type username
      await page.keyboard.type('johndoe');
      
      // Tab to password field
      await page.keyboard.press('Tab');
      await expect(page.locator('#password')).toBeFocused();
      
      // Type password
      await page.keyboard.type('password123');
      
      // Tab to submit button
      await page.keyboard.press('Tab');
      await expect(page.locator('button[type="submit"]')).toBeFocused();
      
      // Submit form with Enter
      await page.keyboard.press('Enter');
      await page.waitForURL('/home');
    });

    test('should navigate main navigation with keyboard', async ({ page }) => {
      await page.goto('/');
      await page.fill('#identifier', 'johndoe');
      await page.fill('#password', 'password123');
      await page.click('button[type="submit"]');
      await page.waitForURL('/home');
      
      // Tab through navigation links
      await page.keyboard.press('Tab');
      await expect(page.locator('nav a:first-child')).toBeFocused();
      
      await page.keyboard.press('Tab');
      await expect(page.locator('nav a:nth-child(2)')).toBeFocused();
      
      // Navigate with Enter key
      await page.keyboard.press('Enter');
      
      // Should navigate to the linked page
      await page.waitForLoadState('networkidle');
    });

    test('should navigate chat interface with keyboard', async ({ page }) => {
      await page.goto('/');
      await page.fill('#identifier', 'johndoe');
      await page.fill('#password', 'password123');
      await page.click('button[type="submit"]');
      await page.goto('/chat');
      await page.waitForLoadState('networkidle');
      
      // Tab to conversation list
      await page.keyboard.press('Tab');
      await expect(page.locator('.conversation-item:first-child')).toBeFocused();
      
      // Navigate conversations with arrow keys
      await page.keyboard.press('ArrowDown');
      await expect(page.locator('.conversation-item:nth-child(2)')).toBeFocused();
      
      await page.keyboard.press('ArrowUp');
      await expect(page.locator('.conversation-item:first-child')).toBeFocused();
      
      // Select conversation with Enter
      await page.keyboard.press('Enter');
      
      // Tab to message input
      await page.keyboard.press('Tab');
      await expect(page.locator('[data-testid="message-input"]')).toBeFocused();
      
      // Type and send message
      await page.keyboard.type('Test keyboard message');
      await page.keyboard.press('Enter');
      
      // Verify message was sent
      await expect(page.locator('.message.sent').last()).toContainText('Test keyboard message');
    });

    test('should handle dropdown menus with keyboard', async ({ page }) => {
      await page.goto('/');
      await page.fill('#identifier', 'johndoe');
      await page.fill('#password', 'password123');
      await page.click('button[type="submit"]');
      
      // Tab to user dropdown
      await page.keyboard.press('Tab');
      // Continue tabbing until we reach the dropdown
      let attempts = 0;
      while (attempts < 10) {
        const focused = await page.evaluate(() => document.activeElement.className);
        if (focused.includes('dropdown-toggle')) break;
        await page.keyboard.press('Tab');
        attempts++;
      }
      
      // Open dropdown with Enter
      await page.keyboard.press('Enter');
      await expect(page.locator('.dropdown-menu')).toBeVisible();
      
      // Navigate dropdown items with arrow keys
      await page.keyboard.press('ArrowDown');
      await expect(page.locator('.dropdown-menu a:first-child')).toBeFocused();
      
      await page.keyboard.press('ArrowDown');
      await expect(page.locator('.dropdown-menu a:nth-child(2)')).toBeFocused();
      
      // Close dropdown with Escape
      await page.keyboard.press('Escape');
      await expect(page.locator('.dropdown-menu')).toBeHidden();
    });

    test('should support skip links', async ({ page }) => {
      await page.goto('/');
      
      // Tab to skip link (should be first focusable element)
      await page.keyboard.press('Tab');
      
      const skipLink = page.locator('.skip-link, [href="#main-content"]').first();
      if (await skipLink.count() > 0) {
        await expect(skipLink).toBeFocused();
        
        // Activate skip link
        await page.keyboard.press('Enter');
        
        // Should focus main content
        await expect(page.locator('#main-content, main')).toBeFocused();
      }
    });
  });

  test.describe('Screen Reader Support', () => {
    test('should have proper ARIA labels and roles', async ({ page }) => {
      await page.goto('/');
      await page.fill('#identifier', 'johndoe');
      await page.fill('#password', 'password123');
      await page.click('button[type="submit"]');
      await page.goto('/chat');
      
      // Check for proper ARIA roles
      await expect(page.locator('[role="main"], main')).toBeVisible();
      await expect(page.locator('[role="navigation"], nav')).toBeVisible();
      
      // Check for ARIA labels on interactive elements
      const messageInput = page.locator('[data-testid="message-input"]');
      const ariaLabel = await messageInput.getAttribute('aria-label');
      const placeholder = await messageInput.getAttribute('placeholder');
      
      expect(ariaLabel || placeholder).toBeTruthy();
      
      // Check for ARIA live regions for dynamic content
      const liveRegions = await page.locator('[aria-live]').count();
      expect(liveRegions).toBeGreaterThan(0);
    });

    test('should announce dynamic content changes', async ({ page, context }) => {
      const page1 = page;
      const page2 = await context.newPage();
      
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
      
      // Check for ARIA live region for new messages
      const liveRegion = page1.locator('[aria-live="polite"], [aria-live="assertive"]');
      await expect(liveRegion).toBeVisible();
      
      // Send message from page2
      await page2.click('[data-testid="conversation-johndoe"]');
      await page2.fill('[data-testid="message-input"]', 'Screen reader test message');
      await page2.click('[data-testid="send-button"]');
      
      // Verify live region is updated (content should change)
      await page1.waitForFunction(() => {
        const liveRegion = document.querySelector('[aria-live]');
        return liveRegion && liveRegion.textContent.includes('Screen reader test message');
      });
    });

    test('should have proper form labels', async ({ page }) => {
      await page.goto('/');
      
      // Check login form labels
      const usernameInput = page.locator('#identifier');
      const passwordInput = page.locator('#password');
      
      // Should have associated labels
      const usernameLabel = await usernameInput.getAttribute('aria-label') || 
                           await page.locator('label[for="identifier"]').textContent();
      const passwordLabel = await passwordInput.getAttribute('aria-label') || 
                           await page.locator('label[for="password"]').textContent();
      
      expect(usernameLabel).toBeTruthy();
      expect(passwordLabel).toBeTruthy();
      
      // Check for required field indicators
      const isUsernameRequired = await usernameInput.getAttribute('required');
      const isPasswordRequired = await passwordInput.getAttribute('required');
      
      expect(isUsernameRequired).toBeTruthy();
      expect(isPasswordRequired).toBeTruthy();
    });

    test('should provide status updates for form submissions', async ({ page }) => {
      await page.goto('/');
      
      // Try invalid login
      await page.fill('#identifier', 'invalid');
      await page.fill('#password', 'invalid');
      await page.click('button[type="submit"]');
      
      // Should have error message with proper ARIA attributes
      const errorMessage = page.locator('.error-message, [role="alert"]');
      await expect(errorMessage).toBeVisible();
      
      // Error should be announced to screen readers
      const ariaLive = await errorMessage.getAttribute('aria-live');
      const role = await errorMessage.getAttribute('role');
      
      expect(ariaLive === 'assertive' || role === 'alert').toBeTruthy();
    });
  });

  test.describe('Focus Management', () => {
    test('should manage focus in modal dialogs', async ({ page }) => {
      await page.goto('/');
      await page.fill('#identifier', 'johndoe');
      await page.fill('#password', 'password123');
      await page.click('button[type="submit"]');
      
      // Open a modal (e.g., create post modal)
      await page.click('[data-testid="create-post-button"]');
      
      // Focus should move to modal
      const modal = page.locator('.modal, [role="dialog"]');
      await expect(modal).toBeVisible();
      
      // First focusable element in modal should be focused
      const firstFocusable = modal.locator('input, button, textarea, select').first();
      await expect(firstFocusable).toBeFocused();
      
      // Tab should cycle within modal
      await page.keyboard.press('Tab');
      await page.keyboard.press('Tab');
      
      // Escape should close modal and restore focus
      await page.keyboard.press('Escape');
      await expect(modal).toBeHidden();
    });

    test('should maintain focus order in dynamic content', async ({ page }) => {
      await page.goto('/');
      await page.fill('#identifier', 'johndoe');
      await page.fill('#password', 'password123');
      await page.click('button[type="submit"]');
      await page.goto('/home');
      
      // Tab through posts
      await page.keyboard.press('Tab');
      const firstPost = page.locator('.post-card:first-child a, .post-card:first-child button').first();
      await expect(firstPost).toBeFocused();
      
      // Continue tabbing
      await page.keyboard.press('Tab');
      const secondFocusable = page.locator('.post-card:first-child a, .post-card:first-child button').nth(1);
      if (await secondFocusable.count() > 0) {
        await expect(secondFocusable).toBeFocused();
      }
    });

    test('should handle focus in infinite scroll', async ({ page }) => {
      await page.goto('/');
      await page.fill('#identifier', 'johndoe');
      await page.fill('#password', 'password123');
      await page.click('button[type="submit"]');
      await page.goto('/home');
      
      // Scroll to trigger infinite scroll
      await page.evaluate(() => {
        window.scrollTo(0, document.body.scrollHeight);
      });
      
      // Wait for new content to load
      await page.waitForTimeout(1000);
      
      // Focus should still be manageable
      await page.keyboard.press('Tab');
      const focusedElement = page.locator(':focus');
      await expect(focusedElement).toBeVisible();
    });
  });

  test.describe('Alternative Text and Media', () => {
    test('should have alt text for images', async ({ page }) => {
      await page.goto('/');
      await page.fill('#identifier', 'johndoe');
      await page.fill('#password', 'password123');
      await page.click('button[type="submit"]');
      
      // Check all images have alt text
      const images = page.locator('img');
      const imageCount = await images.count();
      
      for (let i = 0; i < imageCount; i++) {
        const img = images.nth(i);
        const alt = await img.getAttribute('alt');
        const ariaLabel = await img.getAttribute('aria-label');
        const role = await img.getAttribute('role');
        
        // Image should have alt text, aria-label, or be decorative
        expect(alt !== null || ariaLabel !== null || role === 'presentation').toBeTruthy();
      }
    });

    test('should handle avatar images properly', async ({ page }) => {
      await page.goto('/');
      await page.fill('#identifier', 'johndoe');
      await page.fill('#password', 'password123');
      await page.click('button[type="submit"]');
      await page.goto('/chat');
      
      // Check avatar images
      const avatars = page.locator('.avatar img, .user-avatar');
      const avatarCount = await avatars.count();
      
      if (avatarCount > 0) {
        const firstAvatar = avatars.first();
        const alt = await firstAvatar.getAttribute('alt');
        
        // Avatar should have descriptive alt text
        expect(alt).toBeTruthy();
        expect(alt).toMatch(/avatar|profile|user/i);
      }
    });
  });

  test.describe('Responsive Design Accessibility', () => {
    test('should be accessible on mobile viewport', async ({ page }) => {
      await page.setViewportSize({ width: 375, height: 667 });
      await page.goto('/');
      
      const accessibilityScanResults = await new AxeBuilder({ page }).analyze();
      expect(accessibilityScanResults.violations).toEqual([]);
      
      // Check touch targets are large enough (minimum 44px)
      const buttons = page.locator('button, a, input[type="submit"]');
      const buttonCount = await buttons.count();
      
      for (let i = 0; i < Math.min(buttonCount, 5); i++) {
        const button = buttons.nth(i);
        const box = await button.boundingBox();
        
        if (box) {
          expect(box.width).toBeGreaterThanOrEqual(44);
          expect(box.height).toBeGreaterThanOrEqual(44);
        }
      }
    });

    test('should handle zoom levels properly', async ({ page }) => {
      await page.goto('/');
      
      // Test at 200% zoom
      await page.evaluate(() => {
        document.body.style.zoom = '2';
      });
      
      await page.waitForTimeout(500);
      
      // Content should still be accessible
      const accessibilityScanResults = await new AxeBuilder({ page }).analyze();
      expect(accessibilityScanResults.violations).toEqual([]);
      
      // Reset zoom
      await page.evaluate(() => {
        document.body.style.zoom = '1';
      });
    });
  });
});
