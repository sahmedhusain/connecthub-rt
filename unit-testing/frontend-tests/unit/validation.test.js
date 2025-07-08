/**
 * Form Validation and Input Handling Tests
 * Tests for form validation, input sanitization, and user input handling
 */

import { describe, test, expect, beforeEach, jest } from '@jest/globals';

describe('Form Validation Tests', () => {
  beforeEach(() => {
    // Reset DOM to clean state
    global.testUtils.cleanupDOM();
  });

  describe('Input Validation Functions', () => {
    test('should validate email format', () => {
      const validateEmail = (email) => {
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        return emailRegex.test(email);
      };

      // Valid emails
      expect(validateEmail('test@example.com')).toBe(true);
      expect(validateEmail('user.name@domain.co.uk')).toBe(true);
      expect(validateEmail('test+tag@example.org')).toBe(true);

      // Invalid emails
      expect(validateEmail('invalid-email')).toBe(false);
      expect(validateEmail('test@')).toBe(false);
      expect(validateEmail('@example.com')).toBe(false);
      expect(validateEmail('test..test@example.com')).toBe(false);
    });

    test('should validate password strength', () => {
      const validatePassword = (password) => {
        // At least 6 characters, contains letter and number
        const minLength = password.length >= 6;
        const hasLetter = /[a-zA-Z]/.test(password);
        const hasNumber = /\d/.test(password);
        
        return {
          valid: minLength && hasLetter && hasNumber,
          minLength,
          hasLetter,
          hasNumber
        };
      };

      // Valid passwords
      expect(validatePassword('Aa123456').valid).toBe(true);
      expect(validatePassword('password123').valid).toBe(true);
      expect(validatePassword('Test1234').valid).toBe(true);

      // Invalid passwords
      expect(validatePassword('123').valid).toBe(false); // Too short
      expect(validatePassword('password').valid).toBe(false); // No number
      expect(validatePassword('123456').valid).toBe(false); // No letter
      expect(validatePassword('').valid).toBe(false); // Empty
    });

    test('should validate username format', () => {
      const validateUsername = (username) => {
        // 3-20 characters, alphanumeric and underscore only
        const lengthValid = username.length >= 3 && username.length <= 20;
        const formatValid = /^[a-zA-Z0-9_]+$/.test(username);
        
        return lengthValid && formatValid;
      };

      // Valid usernames
      expect(validateUsername('testuser')).toBe(true);
      expect(validateUsername('user_123')).toBe(true);
      expect(validateUsername('Test_User_2024')).toBe(true);

      // Invalid usernames
      expect(validateUsername('ab')).toBe(false); // Too short
      expect(validateUsername('a'.repeat(21))).toBe(false); // Too long
      expect(validateUsername('test-user')).toBe(false); // Invalid character
      expect(validateUsername('test user')).toBe(false); // Space
      expect(validateUsername('test@user')).toBe(false); // Special character
    });

    test('should validate required fields', () => {
      const validateRequired = (value) => {
        return value !== null && value !== undefined && value.toString().trim() !== '';
      };

      // Valid values
      expect(validateRequired('test')).toBe(true);
      expect(validateRequired('0')).toBe(true);
      expect(validateRequired(0)).toBe(true);
      expect(validateRequired(false)).toBe(true);

      // Invalid values
      expect(validateRequired('')).toBe(false);
      expect(validateRequired('   ')).toBe(false);
      expect(validateRequired(null)).toBe(false);
      expect(validateRequired(undefined)).toBe(false);
    });
  });

  describe('Real-time Form Validation', () => {
    test('should validate form fields on input', () => {
      const mainContent = document.getElementById('main-content');
      mainContent.innerHTML = `
        <form id="test-form">
          <input type="email" id="email" name="email">
          <div id="email-error" class="error hidden"></div>
          <input type="password" id="password" name="password">
          <div id="password-error" class="error hidden"></div>
          <button type="submit">Submit</button>
        </form>
      `;

      const emailInput = document.getElementById('email');
      const emailError = document.getElementById('email-error');
      const passwordInput = document.getElementById('password');
      const passwordError = document.getElementById('password-error');

      // Setup validation listeners
      emailInput.addEventListener('input', (e) => {
        const email = e.target.value;
        const isValid = /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
        
        if (email && !isValid) {
          emailError.textContent = 'Please enter a valid email address';
          emailError.classList.remove('hidden');
        } else {
          emailError.classList.add('hidden');
        }
      });

      passwordInput.addEventListener('input', (e) => {
        const password = e.target.value;
        const isValid = password.length >= 6 && /[a-zA-Z]/.test(password) && /\d/.test(password);
        
        if (password && !isValid) {
          passwordError.textContent = 'Password must be at least 6 characters with letters and numbers';
          passwordError.classList.remove('hidden');
        } else {
          passwordError.classList.add('hidden');
        }
      });

      // Test invalid email
      global.testUtils.simulateInput(emailInput, 'invalid-email');
      expect(emailError.classList.contains('hidden')).toBe(false);
      expect(emailError.textContent).toBe('Please enter a valid email address');

      // Test valid email
      global.testUtils.simulateInput(emailInput, 'test@example.com');
      expect(emailError.classList.contains('hidden')).toBe(true);

      // Test invalid password
      global.testUtils.simulateInput(passwordInput, '123');
      expect(passwordError.classList.contains('hidden')).toBe(false);
      expect(passwordError.textContent).toContain('Password must be at least 6 characters');

      // Test valid password
      global.testUtils.simulateInput(passwordInput, 'Test123');
      expect(passwordError.classList.contains('hidden')).toBe(true);
    });

    test('should prevent form submission with invalid data', () => {
      const mainContent = document.getElementById('main-content');
      mainContent.innerHTML = `
        <form id="test-form">
          <input type="email" id="email" name="email" value="invalid-email">
          <input type="password" id="password" name="password" value="123">
          <button type="submit">Submit</button>
          <div id="form-error" class="error hidden"></div>
        </form>
      `;

      const form = document.getElementById('test-form');
      const formError = document.getElementById('form-error');
      let submissionPrevented = false;

      form.addEventListener('submit', (e) => {
        const email = form.querySelector('#email').value;
        const password = form.querySelector('#password').value;
        
        const emailValid = /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
        const passwordValid = password.length >= 6 && /[a-zA-Z]/.test(password) && /\d/.test(password);
        
        if (!emailValid || !passwordValid) {
          e.preventDefault();
          submissionPrevented = true;
          formError.textContent = 'Please fix the errors above';
          formError.classList.remove('hidden');
        }
      });

      global.testUtils.simulateSubmit(form);

      expect(submissionPrevented).toBe(true);
      expect(formError.classList.contains('hidden')).toBe(false);
      expect(formError.textContent).toBe('Please fix the errors above');
    });
  });

  describe('Input Sanitization', () => {
    test('should sanitize HTML input', () => {
      const sanitizeHTML = (input) => {
        const div = document.createElement('div');
        div.textContent = input;
        return div.innerHTML;
      };

      // Test HTML sanitization
      expect(sanitizeHTML('<script>alert("xss")</script>')).toBe('&lt;script&gt;alert("xss")&lt;/script&gt;');
      expect(sanitizeHTML('<img src="x" onerror="alert(1)">')).toBe('&lt;img src="x" onerror="alert(1)"&gt;');
      expect(sanitizeHTML('Normal text')).toBe('Normal text');
      expect(sanitizeHTML('Text with <b>bold</b>')).toBe('Text with &lt;b&gt;bold&lt;/b&gt;');
    });

    test('should trim whitespace from inputs', () => {
      const trimInput = (input) => {
        return input.toString().trim();
      };

      expect(trimInput('  test  ')).toBe('test');
      expect(trimInput('\n\ttest\n\t')).toBe('test');
      expect(trimInput('   ')).toBe('');
      expect(trimInput('no spaces')).toBe('no spaces');
    });

    test('should limit input length', () => {
      const limitLength = (input, maxLength) => {
        return input.length > maxLength ? input.substring(0, maxLength) : input;
      };

      expect(limitLength('short', 10)).toBe('short');
      expect(limitLength('this is a very long string', 10)).toBe('this is a ');
      expect(limitLength('exactly10!', 10)).toBe('exactly10!');
      expect(limitLength('', 10)).toBe('');
    });
  });

  describe('Form State Management', () => {
    test('should track form dirty state', () => {
      const mainContent = document.getElementById('main-content');
      mainContent.innerHTML = `
        <form id="test-form">
          <input type="text" id="field1" name="field1" value="initial">
          <input type="text" id="field2" name="field2" value="">
        </form>
      `;

      const form = document.getElementById('test-form');
      let isDirty = false;
      const initialValues = {};

      // Store initial values
      const inputs = form.querySelectorAll('input');
      inputs.forEach(input => {
        initialValues[input.name] = input.value;
      });

      // Track changes
      form.addEventListener('input', (e) => {
        const currentValue = e.target.value;
        const initialValue = initialValues[e.target.name];
        
        // Check if any field has changed
        isDirty = Array.from(inputs).some(input => 
          input.value !== initialValues[input.name]
        );
      });

      // Initially not dirty
      expect(isDirty).toBe(false);

      // Change a field
      const field1 = document.getElementById('field1');
      global.testUtils.simulateInput(field1, 'changed');
      expect(isDirty).toBe(true);

      // Revert to original
      global.testUtils.simulateInput(field1, 'initial');
      expect(isDirty).toBe(false);
    });

    test('should handle form reset', () => {
      const mainContent = document.getElementById('main-content');
      mainContent.innerHTML = `
        <form id="test-form">
          <input type="text" id="field1" name="field1" value="initial">
          <input type="text" id="field2" name="field2" value="">
          <button type="reset">Reset</button>
        </form>
      `;

      const form = document.getElementById('test-form');
      const field1 = document.getElementById('field1');
      const field2 = document.getElementById('field2');

      // Change values
      global.testUtils.simulateInput(field1, 'changed');
      global.testUtils.simulateInput(field2, 'new value');

      expect(field1.value).toBe('changed');
      expect(field2.value).toBe('new value');

      // Reset form
      form.reset();

      expect(field1.value).toBe('initial');
      expect(field2.value).toBe('');
    });
  });
});
