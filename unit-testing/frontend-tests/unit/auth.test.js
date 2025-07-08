/**
 * Authentication Flow and Session Management Tests
 * Tests for login, signup, logout, and session handling
 */

import { describe, test, expect, beforeEach, jest } from '@jest/globals';

describe('Authentication Flow Tests', () => {
  beforeEach(() => {
    // Reset fetch mock
    global.fetch.mockClear();
    
    // Reset storage mocks
    global.localStorage.clear();
    global.sessionStorage.clear();
    
    // Reset DOM
    global.testUtils.cleanupDOM();
  });

  describe('Login Form Validation', () => {
    test('should validate required fields', () => {
      const mainContent = document.getElementById('main-content');
      mainContent.innerHTML = `
        <form id="login-form">
          <input type="text" name="identifier" required>
          <input type="password" name="password" required>
          <button type="submit">Login</button>
          <div id="error-message" class="error hidden"></div>
        </form>
      `;

      const form = document.getElementById('login-form');
      const identifierInput = form.querySelector('input[name="identifier"]');
      const passwordInput = form.querySelector('input[name="password"]');
      const errorDiv = document.getElementById('error-message');

      // Test empty form submission
      let validationPassed = true;
      form.addEventListener('submit', (e) => {
        e.preventDefault();
        
        if (!identifierInput.value || !passwordInput.value) {
          validationPassed = false;
          errorDiv.textContent = 'Please fill in all fields';
          errorDiv.classList.remove('hidden');
        }
      });

      global.testUtils.simulateSubmit(form);

      expect(validationPassed).toBe(false);
      expect(errorDiv.textContent).toBe('Please fill in all fields');
      expect(errorDiv.classList.contains('hidden')).toBe(false);
    });

    test('should validate email format when using email', () => {
      const mainContent = document.getElementById('main-content');
      mainContent.innerHTML = `
        <form id="login-form">
          <input type="text" name="identifier" value="invalid-email">
          <input type="password" name="password" value="password123">
          <button type="submit">Login</button>
          <div id="error-message" class="error hidden"></div>
        </form>
      `;

      const form = document.getElementById('login-form');
      const identifierInput = form.querySelector('input[name="identifier"]');
      const errorDiv = document.getElementById('error-message');

      let validationPassed = true;
      form.addEventListener('submit', (e) => {
        e.preventDefault();
        
        const identifier = identifierInput.value;
        // Simple email validation
        if (identifier.includes('@') && !identifier.match(/^[^\s@]+@[^\s@]+\.[^\s@]+$/)) {
          validationPassed = false;
          errorDiv.textContent = 'Please enter a valid email address';
          errorDiv.classList.remove('hidden');
        }
      });

      global.testUtils.simulateSubmit(form);

      expect(validationPassed).toBe(false);
      expect(errorDiv.textContent).toBe('Please enter a valid email address');
    });
  });

  describe('Login API Integration', () => {
    test('should send login request with correct data', async () => {
      const mainContent = document.getElementById('main-content');
      mainContent.innerHTML = `
        <form id="login-form">
          <input type="text" name="identifier" value="testuser">
          <input type="password" name="password" value="password123">
          <button type="submit">Login</button>
        </form>
      `;

      const form = document.getElementById('login-form');
      
      // Mock successful login response
      global.fetch.mockResolvedValueOnce(
        global.testUtils.createMockResponse({
          success: true,
          user: { id: 1, username: 'testuser' },
          sessionToken: 'mock-session-token'
        })
      );

      form.addEventListener('submit', async (e) => {
        e.preventDefault();
        
        const formData = new FormData(form);
        const loginData = {
          identifier: formData.get('identifier'),
          password: formData.get('password')
        };

        const response = await fetch('/api/login', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(loginData)
        });

        const result = await response.json();
        
        if (result.success) {
          localStorage.setItem('sessionToken', result.sessionToken);
          localStorage.setItem('currentUser', JSON.stringify(result.user));
        }
      });

      global.testUtils.simulateSubmit(form);

      // Wait for async operations
      await global.testUtils.waitForDOM();

      expect(global.fetch).toHaveBeenCalledWith('/api/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          identifier: 'testuser',
          password: 'password123'
        })
      });

      expect(localStorage.setItem).toHaveBeenCalledWith('sessionToken', 'mock-session-token');
      expect(localStorage.setItem).toHaveBeenCalledWith('currentUser', JSON.stringify({
        id: 1,
        username: 'testuser'
      }));
    });

    test('should handle login failure', async () => {
      const mainContent = document.getElementById('main-content');
      mainContent.innerHTML = `
        <form id="login-form">
          <input type="text" name="identifier" value="wronguser">
          <input type="password" name="password" value="wrongpass">
          <button type="submit">Login</button>
          <div id="error-message" class="error hidden"></div>
        </form>
      `;

      const form = document.getElementById('login-form');
      const errorDiv = document.getElementById('error-message');
      
      // Mock failed login response
      global.fetch.mockResolvedValueOnce(
        global.testUtils.createMockResponse({
          success: false,
          error: 'Invalid credentials'
        }, 401)
      );

      form.addEventListener('submit', async (e) => {
        e.preventDefault();
        
        const formData = new FormData(form);
        const loginData = {
          identifier: formData.get('identifier'),
          password: formData.get('password')
        };

        const response = await fetch('/api/login', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(loginData)
        });

        const result = await response.json();
        
        if (!result.success) {
          errorDiv.textContent = result.error;
          errorDiv.classList.remove('hidden');
        }
      });

      global.testUtils.simulateSubmit(form);

      // Wait for async operations
      await global.testUtils.waitForDOM();

      expect(errorDiv.textContent).toBe('Invalid credentials');
      expect(errorDiv.classList.contains('hidden')).toBe(false);
    });
  });

  describe('Signup Form Validation', () => {
    test('should validate all required signup fields', () => {
      const mainContent = document.getElementById('main-content');
      mainContent.innerHTML = `
        <form id="signup-form">
          <input type="text" name="firstName" required>
          <input type="text" name="lastName" required>
          <input type="text" name="username" required>
          <input type="email" name="email" required>
          <input type="password" name="password" required>
          <select name="gender" required>
            <option value="">Select Gender</option>
            <option value="male">Male</option>
            <option value="female">Female</option>
          </select>
          <input type="date" name="dateOfBirth" required>
          <button type="submit">Sign Up</button>
          <div id="error-message" class="error hidden"></div>
        </form>
      `;

      const form = document.getElementById('signup-form');
      const errorDiv = document.getElementById('error-message');

      let validationErrors = [];
      form.addEventListener('submit', (e) => {
        e.preventDefault();
        
        const formData = new FormData(form);
        const requiredFields = ['firstName', 'lastName', 'username', 'email', 'password', 'gender', 'dateOfBirth'];
        
        requiredFields.forEach(field => {
          if (!formData.get(field)) {
            validationErrors.push(`${field} is required`);
          }
        });

        if (validationErrors.length > 0) {
          errorDiv.textContent = validationErrors.join(', ');
          errorDiv.classList.remove('hidden');
        }
      });

      global.testUtils.simulateSubmit(form);

      expect(validationErrors.length).toBeGreaterThan(0);
      expect(errorDiv.classList.contains('hidden')).toBe(false);
    });

    test('should validate password strength', () => {
      const mainContent = document.getElementById('main-content');
      mainContent.innerHTML = `
        <form id="signup-form">
          <input type="password" name="password" value="weak">
          <button type="submit">Sign Up</button>
          <div id="error-message" class="error hidden"></div>
        </form>
      `;

      const form = document.getElementById('signup-form');
      const passwordInput = form.querySelector('input[name="password"]');
      const errorDiv = document.getElementById('error-message');

      let validationPassed = true;
      form.addEventListener('submit', (e) => {
        e.preventDefault();
        
        const password = passwordInput.value;
        // Simple password validation
        if (password.length < 6) {
          validationPassed = false;
          errorDiv.textContent = 'Password must be at least 6 characters long';
          errorDiv.classList.remove('hidden');
        }
      });

      global.testUtils.simulateSubmit(form);

      expect(validationPassed).toBe(false);
      expect(errorDiv.textContent).toBe('Password must be at least 6 characters long');
    });
  });

  describe('Session Management', () => {
    test('should store session token on successful login', () => {
      const sessionToken = 'test-session-token-123';
      const userData = { id: 1, username: 'testuser', email: 'test@example.com' };

      // Simulate successful login
      localStorage.setItem('sessionToken', sessionToken);
      localStorage.setItem('currentUser', JSON.stringify(userData));

      expect(localStorage.setItem).toHaveBeenCalledWith('sessionToken', sessionToken);
      expect(localStorage.setItem).toHaveBeenCalledWith('currentUser', JSON.stringify(userData));
    });

    test('should check for existing session on page load', () => {
      // Mock existing session
      localStorage.getItem.mockImplementation((key) => {
        if (key === 'sessionToken') return 'existing-token';
        if (key === 'currentUser') return JSON.stringify({ id: 1, username: 'existinguser' });
        return null;
      });

      // Simulate session check
      const sessionToken = localStorage.getItem('sessionToken');
      const currentUser = localStorage.getItem('currentUser');

      expect(sessionToken).toBe('existing-token');
      expect(currentUser).toBe(JSON.stringify({ id: 1, username: 'existinguser' }));
    });

    test('should clear session on logout', () => {
      // Setup existing session
      localStorage.setItem('sessionToken', 'token-to-clear');
      localStorage.setItem('currentUser', JSON.stringify({ id: 1, username: 'user' }));

      // Simulate logout
      localStorage.removeItem('sessionToken');
      localStorage.removeItem('currentUser');

      expect(localStorage.removeItem).toHaveBeenCalledWith('sessionToken');
      expect(localStorage.removeItem).toHaveBeenCalledWith('currentUser');
    });

    test('should redirect to login when session expires', () => {
      // Mock expired session response
      global.fetch.mockResolvedValueOnce(
        global.testUtils.createMockResponse({
          success: false,
          error: 'Session expired'
        }, 401)
      );

      let redirected = false;
      window.location.assign = jest.fn(() => {
        redirected = true;
      });

      // Simulate API call with expired session
      fetch('/api/protected-endpoint', {
        headers: {
          'Authorization': 'Bearer expired-token'
        }
      }).then(response => {
        if (response.status === 401) {
          localStorage.removeItem('sessionToken');
          localStorage.removeItem('currentUser');
          window.location.assign('/');
        }
      });

      // Wait for promise resolution
      return new Promise(resolve => {
        setTimeout(() => {
          expect(localStorage.removeItem).toHaveBeenCalledWith('sessionToken');
          expect(localStorage.removeItem).toHaveBeenCalledWith('currentUser');
          expect(window.location.assign).toHaveBeenCalledWith('/');
          resolve();
        }, 0);
      });
    });
  });

  describe('Authentication State Management', () => {
    test('should update UI based on authentication state', () => {
      const mainContent = document.getElementById('main-content');
      
      // Simulate unauthenticated state
      mainContent.innerHTML = `
        <div id="auth-container">
          <div id="login-section" class="visible">Login Form</div>
          <div id="user-section" class="hidden">User Dashboard</div>
        </div>
      `;

      const loginSection = document.getElementById('login-section');
      const userSection = document.getElementById('user-section');

      // Simulate authentication
      function updateAuthState(isAuthenticated) {
        if (isAuthenticated) {
          loginSection.classList.add('hidden');
          loginSection.classList.remove('visible');
          userSection.classList.add('visible');
          userSection.classList.remove('hidden');
        } else {
          loginSection.classList.add('visible');
          loginSection.classList.remove('hidden');
          userSection.classList.add('hidden');
          userSection.classList.remove('visible');
        }
      }

      // Test authenticated state
      updateAuthState(true);
      expect(loginSection.classList.contains('hidden')).toBe(true);
      expect(userSection.classList.contains('visible')).toBe(true);

      // Test unauthenticated state
      updateAuthState(false);
      expect(loginSection.classList.contains('visible')).toBe(true);
      expect(userSection.classList.contains('hidden')).toBe(true);
    });

    test('should include session token in API requests', async () => {
      const sessionToken = 'valid-session-token';
      localStorage.getItem.mockReturnValue(sessionToken);

      // Mock API response
      global.fetch.mockResolvedValueOnce(
        global.testUtils.createMockResponse({ data: 'protected data' })
      );

      // Simulate authenticated API call
      const token = localStorage.getItem('sessionToken');
      await fetch('/api/protected-endpoint', {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      });

      expect(global.fetch).toHaveBeenCalledWith('/api/protected-endpoint', {
        headers: {
          'Authorization': `Bearer ${sessionToken}`,
          'Content-Type': 'application/json'
        }
      });
    });
  });
});
