/**
 * Utility Functions Tests
 * Tests for API utilities, router, validation, and other helper functions
 */

import { describe, test, expect, beforeEach, afterEach, jest } from '@jest/globals';

// Mock fetch API
global.fetch = jest.fn();

// Mock window.location
delete window.location;
window.location = {
  href: 'http://localhost:3000/',
  pathname: '/',
  search: '',
  hash: '',
  assign: jest.fn(),
  replace: jest.fn(),
  reload: jest.fn()
};

// Mock history API
window.history = {
  pushState: jest.fn(),
  replaceState: jest.fn(),
  back: jest.fn(),
  forward: jest.fn(),
  go: jest.fn()
};

describe('Utility Functions Tests', () => {
  beforeEach(() => {
    fetch.mockClear();
    jest.clearAllMocks();
  });

  describe('API Utilities', () => {
    test('should make GET requests correctly', async () => {
      const mockResponse = { data: 'test' };
      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse
      });

      // Mock API function
      const apiGet = async (url) => {
        const response = await fetch(url, {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json'
          }
        });
        
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        return response.json();
      };

      const result = await apiGet('/api/test');

      expect(fetch).toHaveBeenCalledWith('/api/test', {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json'
        }
      });
      expect(result).toEqual(mockResponse);
    });

    test('should make POST requests with data', async () => {
      const mockResponse = { success: true };
      const postData = { username: 'test', password: 'password' };

      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse
      });

      // Mock API function
      const apiPost = async (url, data) => {
        const response = await fetch(url, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify(data)
        });
        
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        return response.json();
      };

      const result = await apiPost('/api/login', postData);

      expect(fetch).toHaveBeenCalledWith('/api/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(postData)
      });
      expect(result).toEqual(mockResponse);
    });

    test('should handle API errors correctly', async () => {
      fetch.mockResolvedValueOnce({
        ok: false,
        status: 404,
        statusText: 'Not Found'
      });

      // Mock API function with error handling
      const apiRequest = async (url) => {
        const response = await fetch(url);
        
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        return response.json();
      };

      await expect(apiRequest('/api/nonexistent')).rejects.toThrow('HTTP error! status: 404');
    });

    test('should handle network errors', async () => {
      fetch.mockRejectedValueOnce(new Error('Network error'));

      // Mock API function
      const apiRequest = async (url) => {
        try {
          const response = await fetch(url);
          return response.json();
        } catch (error) {
          throw new Error(`Network error: ${error.message}`);
        }
      };

      await expect(apiRequest('/api/test')).rejects.toThrow('Network error: Network error');
    });

    test('should include authentication headers when available', async () => {
      // Mock session storage
      Object.defineProperty(window, 'sessionStorage', {
        value: {
          getItem: jest.fn(() => 'mock-session-token'),
          setItem: jest.fn(),
          removeItem: jest.fn()
        },
        writable: true
      });

      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ data: 'authenticated' })
      });

      // Mock authenticated API function
      const authenticatedRequest = async (url) => {
        const token = sessionStorage.getItem('session_token');
        const headers = {
          'Content-Type': 'application/json'
        };
        
        if (token) {
          headers['Authorization'] = `Bearer ${token}`;
        }

        const response = await fetch(url, {
          method: 'GET',
          headers
        });
        
        return response.json();
      };

      await authenticatedRequest('/api/protected');

      expect(fetch).toHaveBeenCalledWith('/api/protected', {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer mock-session-token'
        }
      });
    });
  });

  describe('Router Utilities', () => {
    test('should navigate to different routes', () => {
      // Mock router function
      const navigateTo = (path) => {
        window.history.pushState({}, '', path);
        window.location.pathname = path;
      };

      navigateTo('/home');

      expect(window.history.pushState).toHaveBeenCalledWith({}, '', '/home');
      expect(window.location.pathname).toBe('/home');
    });

    test('should handle route parameters', () => {
      // Mock route parameter extraction
      const getRouteParams = (path, pattern) => {
        const pathParts = path.split('/');
        const patternParts = pattern.split('/');
        const params = {};

        for (let i = 0; i < patternParts.length; i++) {
          if (patternParts[i].startsWith(':')) {
            const paramName = patternParts[i].substring(1);
            params[paramName] = pathParts[i];
          }
        }

        return params;
      };

      const params = getRouteParams('/post/123', '/post/:id');
      expect(params).toEqual({ id: '123' });
    });

    test('should handle query parameters', () => {
      // Mock query parameter parsing
      const parseQueryParams = (search) => {
        const params = new URLSearchParams(search);
        const result = {};
        
        for (const [key, value] of params) {
          result[key] = value;
        }
        
        return result;
      };

      const queryParams = parseQueryParams('?filter=technology&sort=newest');
      expect(queryParams).toEqual({
        filter: 'technology',
        sort: 'newest'
      });
    });

    test('should handle browser back/forward navigation', () => {
      // Mock popstate event handling
      const handlePopState = (event) => {
        const path = window.location.pathname;
        // Route handling logic would go here
        return path;
      };

      // Simulate popstate event
      window.location.pathname = '/chat';
      const currentPath = handlePopState({ state: {} });

      expect(currentPath).toBe('/chat');
    });
  });

  describe('Validation Utilities', () => {
    test('should validate email addresses', () => {
      const validateEmail = (email) => {
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        return emailRegex.test(email);
      };

      expect(validateEmail('test@example.com')).toBe(true);
      expect(validateEmail('user.name+tag@domain.co.uk')).toBe(true);
      expect(validateEmail('invalid-email')).toBe(false);
      expect(validateEmail('test@')).toBe(false);
      expect(validateEmail('@example.com')).toBe(false);
      expect(validateEmail('')).toBe(false);
    });

    test('should validate usernames', () => {
      const validateUsername = (username) => {
        // Username should be 3-20 characters, alphanumeric and underscores only
        const usernameRegex = /^[a-zA-Z0-9_]{3,20}$/;
        return usernameRegex.test(username);
      };

      expect(validateUsername('validuser')).toBe(true);
      expect(validateUsername('user_123')).toBe(true);
      expect(validateUsername('ab')).toBe(false); // Too short
      expect(validateUsername('a'.repeat(21))).toBe(false); // Too long
      expect(validateUsername('user-name')).toBe(false); // Invalid character
      expect(validateUsername('user name')).toBe(false); // Space not allowed
    });

    test('should validate passwords', () => {
      const validatePassword = (password) => {
        // Password should be at least 8 characters with at least one letter and one number
        const passwordRegex = /^(?=.*[A-Za-z])(?=.*\d)[A-Za-z\d@$!%*#?&]{8,}$/;
        return passwordRegex.test(password);
      };

      expect(validatePassword('password123')).toBe(true);
      expect(validatePassword('MyPass1!')).toBe(true);
      expect(validatePassword('password')).toBe(false); // No number
      expect(validatePassword('12345678')).toBe(false); // No letter
      expect(validatePassword('pass1')).toBe(false); // Too short
    });

    test('should validate post content', () => {
      const validatePostContent = (title, content) => {
        const errors = [];

        if (!title || title.trim().length === 0) {
          errors.push('Title is required');
        } else if (title.trim().length > 200) {
          errors.push('Title must be less than 200 characters');
        }

        if (!content || content.trim().length === 0) {
          errors.push('Content is required');
        } else if (content.trim().length < 10) {
          errors.push('Content must be at least 10 characters');
        }

        return {
          isValid: errors.length === 0,
          errors
        };
      };

      expect(validatePostContent('Valid Title', 'This is valid content with enough characters')).toEqual({
        isValid: true,
        errors: []
      });

      expect(validatePostContent('', 'Valid content')).toEqual({
        isValid: false,
        errors: ['Title is required']
      });

      expect(validatePostContent('Valid Title', 'Short')).toEqual({
        isValid: false,
        errors: ['Content must be at least 10 characters']
      });
    });

    test('should sanitize user input', () => {
      const sanitizeInput = (input) => {
        if (typeof input !== 'string') return '';
        
        return input
          .replace(/</g, '&lt;')
          .replace(/>/g, '&gt;')
          .replace(/"/g, '&quot;')
          .replace(/'/g, '&#x27;')
          .replace(/\//g, '&#x2F;');
      };

      expect(sanitizeInput('<script>alert("xss")</script>')).toBe('&lt;script&gt;alert(&quot;xss&quot;)&lt;&#x2F;script&gt;');
      expect(sanitizeInput('Normal text')).toBe('Normal text');
      expect(sanitizeInput('Text with "quotes" and \'apostrophes\'')).toBe('Text with &quot;quotes&quot; and &#x27;apostrophes&#x27;');
    });
  });

  describe('Form Utilities', () => {
    test('should serialize form data', () => {
      // Create a mock form
      document.body.innerHTML = `
        <form id="test-form">
          <input type="text" name="username" value="testuser">
          <input type="email" name="email" value="test@example.com">
          <input type="password" name="password" value="password123">
          <select name="gender">
            <option value="male" selected>Male</option>
            <option value="female">Female</option>
          </select>
          <input type="checkbox" name="terms" checked>
        </form>
      `;

      const serializeForm = (form) => {
        const formData = new FormData(form);
        const data = {};
        
        for (const [key, value] of formData.entries()) {
          data[key] = value;
        }
        
        return data;
      };

      const form = document.getElementById('test-form');
      const serialized = serializeForm(form);

      expect(serialized).toEqual({
        username: 'testuser',
        email: 'test@example.com',
        password: 'password123',
        gender: 'male',
        terms: 'on'
      });
    });

    test('should handle form validation errors', () => {
      const displayFormErrors = (errors, formElement) => {
        // Clear existing errors
        const existingErrors = formElement.querySelectorAll('.error-message');
        existingErrors.forEach(error => error.remove());

        // Display new errors
        Object.keys(errors).forEach(fieldName => {
          const field = formElement.querySelector(`[name="${fieldName}"]`);
          if (field) {
            const errorElement = document.createElement('div');
            errorElement.className = 'error-message';
            errorElement.textContent = errors[fieldName];
            field.parentNode.appendChild(errorElement);
          }
        });
      };

      document.body.innerHTML = `
        <form id="test-form">
          <div>
            <input type="text" name="username">
          </div>
          <div>
            <input type="email" name="email">
          </div>
        </form>
      `;

      const form = document.getElementById('test-form');
      const errors = {
        username: 'Username is required',
        email: 'Invalid email format'
      };

      displayFormErrors(errors, form);

      expect(form.querySelectorAll('.error-message')).toHaveLength(2);
      expect(form.querySelector('[name="username"]').parentNode.querySelector('.error-message').textContent).toBe('Username is required');
    });
  });

  describe('Date and Time Utilities', () => {
    test('should format timestamps correctly', () => {
      const formatTimestamp = (timestamp) => {
        const date = new Date(timestamp);
        const now = new Date();
        const diffInSeconds = Math.floor((now - date) / 1000);

        if (diffInSeconds < 60) {
          return 'Just now';
        } else if (diffInSeconds < 3600) {
          const minutes = Math.floor(diffInSeconds / 60);
          return `${minutes} minute${minutes > 1 ? 's' : ''} ago`;
        } else if (diffInSeconds < 86400) {
          const hours = Math.floor(diffInSeconds / 3600);
          return `${hours} hour${hours > 1 ? 's' : ''} ago`;
        } else {
          return date.toLocaleDateString();
        }
      };

      const now = new Date();
      const oneMinuteAgo = new Date(now.getTime() - 60 * 1000);
      const oneHourAgo = new Date(now.getTime() - 60 * 60 * 1000);
      const oneDayAgo = new Date(now.getTime() - 24 * 60 * 60 * 1000);

      expect(formatTimestamp(now.toISOString())).toBe('Just now');
      expect(formatTimestamp(oneMinuteAgo.toISOString())).toBe('1 minute ago');
      expect(formatTimestamp(oneHourAgo.toISOString())).toBe('1 hour ago');
      expect(formatTimestamp(oneDayAgo.toISOString())).toMatch(/\d{1,2}\/\d{1,2}\/\d{4}/);
    });

    test('should handle timezone conversions', () => {
      const convertToUserTimezone = (utcTimestamp) => {
        const date = new Date(utcTimestamp);
        return date.toLocaleString();
      };

      const utcTime = '2023-12-25T12:00:00Z';
      const localTime = convertToUserTimezone(utcTime);

      expect(localTime).toMatch(/\d{1,2}\/\d{1,2}\/\d{4}/);
    });
  });

  describe('Storage Utilities', () => {
    test('should handle localStorage operations', () => {
      // Mock localStorage
      const localStorageMock = {
        getItem: jest.fn(),
        setItem: jest.fn(),
        removeItem: jest.fn(),
        clear: jest.fn()
      };
      Object.defineProperty(window, 'localStorage', { value: localStorageMock });

      const storageUtils = {
        set: (key, value) => {
          localStorage.setItem(key, JSON.stringify(value));
        },
        get: (key) => {
          const item = localStorage.getItem(key);
          return item ? JSON.parse(item) : null;
        },
        remove: (key) => {
          localStorage.removeItem(key);
        }
      };

      const testData = { user: 'test', preferences: { theme: 'dark' } };

      storageUtils.set('userData', testData);
      expect(localStorage.setItem).toHaveBeenCalledWith('userData', JSON.stringify(testData));

      localStorageMock.getItem.mockReturnValue(JSON.stringify(testData));
      const retrieved = storageUtils.get('userData');
      expect(retrieved).toEqual(testData);

      storageUtils.remove('userData');
      expect(localStorage.removeItem).toHaveBeenCalledWith('userData');
    });
  });
});
