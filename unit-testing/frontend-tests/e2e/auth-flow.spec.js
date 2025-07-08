/**
 * End-to-End Authentication Flow Tests
 * Tests complete user authentication workflows using Playwright
 */

import { test, expect } from '@playwright/test';

test.describe('Authentication Flow E2E Tests', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to the application
    await page.goto('/');
  });

  test.describe('Login Flow', () => {
    test('should display login form on initial load', async ({ page }) => {
      // Check if login form is visible
      await expect(page.locator('form')).toBeVisible();
      await expect(page.locator('input[name="identifier"]')).toBeVisible();
      await expect(page.locator('input[name="password"]')).toBeVisible();
      await expect(page.locator('button[type="submit"]')).toBeVisible();
    });

    test('should show validation errors for empty form', async ({ page }) => {
      // Try to submit empty form
      await page.click('button[type="submit"]');
      
      // Check for validation messages
      const errorMessage = page.locator('.error, .error-message, [class*="error"]');
      await expect(errorMessage).toBeVisible();
    });

    test('should login successfully with valid credentials', async ({ page }) => {
      // Fill in valid test credentials
      await page.fill('input[name="identifier"]', 'alexchen');
      await page.fill('input[name="password"]', 'Aa123456');
      
      // Submit the form
      await page.click('button[type="submit"]');
      
      // Wait for navigation to home page
      await expect(page).toHaveURL(/.*\/home/);
      
      // Check for authenticated user interface
      await expect(page.locator('.user-dropdown, .user-menu, [class*="user"]')).toBeVisible();
    });

    test('should show error for invalid credentials', async ({ page }) => {
      // Fill in invalid credentials
      await page.fill('input[name="identifier"]', 'invaliduser');
      await page.fill('input[name="password"]', 'wrongpassword');
      
      // Submit the form
      await page.click('button[type="submit"]');
      
      // Check for error message
      const errorMessage = page.locator('.error, .error-message, [class*="error"]');
      await expect(errorMessage).toBeVisible();
      await expect(errorMessage).toContainText(/invalid|incorrect|wrong/i);
    });

    test('should login with email address', async ({ page }) => {
      // Fill in email and password
      await page.fill('input[name="identifier"]', 'alexandra.chen@techcorp.com');
      await page.fill('input[name="password"]', 'Aa123456');
      
      // Submit the form
      await page.click('button[type="submit"]');
      
      // Wait for successful login
      await expect(page).toHaveURL(/.*\/home/);
    });
  });

  test.describe('Signup Flow', () => {
    test('should navigate to signup page', async ({ page }) => {
      // Look for signup link and click it
      const signupLink = page.locator('a[href*="signup"], button:has-text("Sign Up"), .signup-link');
      await signupLink.click();
      
      // Check if signup form is visible
      await expect(page.locator('form')).toBeVisible();
      await expect(page.locator('input[name="firstName"]')).toBeVisible();
      await expect(page.locator('input[name="lastName"]')).toBeVisible();
      await expect(page.locator('input[name="username"]')).toBeVisible();
      await expect(page.locator('input[name="email"]')).toBeVisible();
      await expect(page.locator('input[name="password"]')).toBeVisible();
    });

    test('should validate required fields in signup form', async ({ page }) => {
      // Navigate to signup
      const signupLink = page.locator('a[href*="signup"], button:has-text("Sign Up"), .signup-link');
      await signupLink.click();
      
      // Try to submit empty form
      await page.click('button[type="submit"]');
      
      // Check for validation errors
      const errorMessages = page.locator('.error, .error-message, [class*="error"]');
      await expect(errorMessages.first()).toBeVisible();
    });

    test('should create new account successfully', async ({ page }) => {
      // Navigate to signup
      const signupLink = page.locator('a[href*="signup"], button:has-text("Sign Up"), .signup-link');
      await signupLink.click();
      
      // Fill in all required fields
      const timestamp = Date.now();
      await page.fill('input[name="firstName"]', 'Test');
      await page.fill('input[name="lastName"]', 'User');
      await page.fill('input[name="username"]', `testuser${timestamp}`);
      await page.fill('input[name="email"]', `test${timestamp}@example.com`);
      await page.fill('input[name="password"]', 'TestPassword123');
      
      // Select gender if dropdown exists
      const genderSelect = page.locator('select[name="gender"]');
      if (await genderSelect.isVisible()) {
        await genderSelect.selectOption('male');
      }
      
      // Fill date of birth if field exists
      const dobField = page.locator('input[name="dateOfBirth"], input[type="date"]');
      if (await dobField.isVisible()) {
        await dobField.fill('1990-01-01');
      }
      
      // Submit the form
      await page.click('button[type="submit"]');
      
      // Check for successful signup (might redirect to login or home)
      await page.waitForURL(/.*\/(home|login|\?)/);
    });
  });

  test.describe('Session Management', () => {
    test('should maintain session across page refreshes', async ({ page }) => {
      // Login first
      await page.fill('input[name="identifier"]', 'alexchen');
      await page.fill('input[name="password"]', 'Aa123456');
      await page.click('button[type="submit"]');
      await expect(page).toHaveURL(/.*\/home/);
      
      // Refresh the page
      await page.reload();
      
      // Should still be logged in
      await expect(page).toHaveURL(/.*\/home/);
      await expect(page.locator('.user-dropdown, .user-menu, [class*="user"]')).toBeVisible();
    });

    test('should logout successfully', async ({ page }) => {
      // Login first
      await page.fill('input[name="identifier"]', 'alexchen');
      await page.fill('input[name="password"]', 'Aa123456');
      await page.click('button[type="submit"]');
      await expect(page).toHaveURL(/.*\/home/);
      
      // Find and click logout button
      const userMenu = page.locator('.user-dropdown, .user-menu, [class*="user"]');
      await userMenu.click();
      
      const logoutButton = page.locator('button:has-text("Logout"), a:has-text("Logout"), .logout');
      await logoutButton.click();
      
      // Should redirect to login page
      await expect(page).toHaveURL(/.*\/(login|\?|$)/);
      await expect(page.locator('input[name="identifier"]')).toBeVisible();
    });

    test('should redirect to login when accessing protected pages without authentication', async ({ page }) => {
      // Try to access a protected page directly
      await page.goto('/home');
      
      // Should redirect to login
      await expect(page).toHaveURL(/.*\/(login|\?|$)/);
      await expect(page.locator('input[name="identifier"]')).toBeVisible();
    });
  });

  test.describe('Form Interactions', () => {
    test('should show/hide password visibility', async ({ page }) => {
      const passwordInput = page.locator('input[name="password"]');
      const toggleButton = page.locator('.password-toggle, [class*="show-password"], [class*="toggle-password"]');
      
      // Check if password toggle exists
      if (await toggleButton.isVisible()) {
        // Initially password should be hidden
        await expect(passwordInput).toHaveAttribute('type', 'password');
        
        // Click toggle to show password
        await toggleButton.click();
        await expect(passwordInput).toHaveAttribute('type', 'text');
        
        // Click toggle to hide password again
        await toggleButton.click();
        await expect(passwordInput).toHaveAttribute('type', 'password');
      }
    });

    test('should handle form submission with Enter key', async ({ page }) => {
      // Fill in credentials
      await page.fill('input[name="identifier"]', 'alexchen');
      await page.fill('input[name="password"]', 'Aa123456');
      
      // Press Enter in password field
      await page.press('input[name="password"]', 'Enter');
      
      // Should submit the form and navigate
      await expect(page).toHaveURL(/.*\/home/);
    });

    test('should validate email format in signup', async ({ page }) => {
      // Navigate to signup
      const signupLink = page.locator('a[href*="signup"], button:has-text("Sign Up"), .signup-link');
      if (await signupLink.isVisible()) {
        await signupLink.click();
        
        // Fill in invalid email
        await page.fill('input[name="email"]', 'invalid-email');
        await page.fill('input[name="firstName"]', 'Test');
        await page.fill('input[name="lastName"]', 'User');
        await page.fill('input[name="username"]', 'testuser');
        await page.fill('input[name="password"]', 'TestPassword123');
        
        // Try to submit
        await page.click('button[type="submit"]');
        
        // Should show email validation error
        const emailError = page.locator('.error, .error-message, [class*="error"]');
        await expect(emailError).toBeVisible();
      }
    });
  });

  test.describe('Accessibility', () => {
    test('should have proper form labels and accessibility attributes', async ({ page }) => {
      // Check for proper form labels
      const identifierInput = page.locator('input[name="identifier"]');
      const passwordInput = page.locator('input[name="password"]');
      
      // Check for labels or aria-labels
      await expect(identifierInput).toHaveAttribute('placeholder');
      await expect(passwordInput).toHaveAttribute('placeholder');
      
      // Check for required attributes
      await expect(identifierInput).toHaveAttribute('required');
      await expect(passwordInput).toHaveAttribute('required');
    });

    test('should support keyboard navigation', async ({ page }) => {
      // Tab through form elements
      await page.keyboard.press('Tab');
      await expect(page.locator('input[name="identifier"]')).toBeFocused();
      
      await page.keyboard.press('Tab');
      await expect(page.locator('input[name="password"]')).toBeFocused();
      
      await page.keyboard.press('Tab');
      await expect(page.locator('button[type="submit"]')).toBeFocused();
    });
  });

  test.describe('Visual Regression', () => {
    test('should match login page screenshot', async ({ page }) => {
      // Take screenshot of login page
      await expect(page).toHaveScreenshot('login-page.png');
    });

    test('should match signup page screenshot', async ({ page }) => {
      const signupLink = page.locator('a[href*="signup"], button:has-text("Sign Up"), .signup-link');
      if (await signupLink.isVisible()) {
        await signupLink.click();
        await expect(page).toHaveScreenshot('signup-page.png');
      }
    });
  });
});
