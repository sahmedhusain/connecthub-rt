import { VALIDATION_ERRORS, CONTENT_ERRORS } from './errorMessages.js';

/**
 * Email validation using regex pattern
 * @param {string} email - Email to validate
 * @returns {boolean} - True if email is valid
 */
export function isValidEmail(email) {
    console.debug(`[Validation] Validating email: ${email ? (email.substring(0, 3) + '***@' + email.split('@')[1]) : 'undefined'}`);
    
    if (!email) {
        console.debug("[Validation] Email validation failed: email is empty");
        return false;
    }
    
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    const isValid = emailRegex.test(email);
    
    console.debug(`[Validation] Email validation result: ${isValid}`);
    return isValid;
}

/**
 * Password validation
 * - At least 8 characters
 * - Contains at least one uppercase letter
 * - Contains at least one lowercase letter
 * - Contains at least one number
 * @param {string} password - Password to validate
 * @returns {boolean} - True if password meets requirements
 */
export function isValidPassword(password) {
    console.debug("[Validation] Validating password");
    
    if (!password) {
        console.debug("[Validation] Password validation failed: password is empty");
        return false;
    }
    
    if (password.length < 8) {
        console.debug("[Validation] Password validation failed: length < 8");
        return false;
    }
    
    const hasUpperCase = /[A-Z]/.test(password);
    const hasLowerCase = /[a-z]/.test(password);
    const hasNumbers = /\d/.test(password);
    
    const isValid = hasUpperCase && hasLowerCase && hasNumbers;
    
    console.debug(`[Validation] Password validation result: ${isValid}`, {
        hasUpperCase,
        hasLowerCase,
        hasNumbers,
        length: password.length
    });
    
    return isValid;
}

/**
 * Validate form input length
 * @param {string} input - Input to validate
 * @param {number} minLength - Minimum length
 * @param {number} maxLength - Maximum length
 * @returns {boolean} - True if input length is valid
 */
export function isValidLength(input, minLength, maxLength) {
    if (input === undefined || input === null) {
        console.debug(`[Validation] Length validation failed: input is ${input === undefined ? 'undefined' : 'null'}`);
        return false;
    }
    
    const length = input.length;
    const isValid = length >= minLength && length <= maxLength;
    
    console.debug(`[Validation] Length validation (${minLength}-${maxLength}): ${isValid}`, {
        actualLength: length,
        minLength,
        maxLength
    });
    
    return isValid;
}

/**
 * Validate username
 * - Only alphanumeric characters and underscores
 * - Length between 3 and 20 characters
 * @param {string} username - Username to validate
 * @returns {boolean} - True if username is valid
 */
export function isValidUsername(username) {
    console.debug(`[Validation] Validating username: ${username || 'undefined'}`);
    
    if (!username) {
        console.debug("[Validation] Username validation failed: username is empty");
        return false;
    }
    
    const usernameRegex = /^[a-zA-Z0-9_]{3,20}$/;
    const isValid = usernameRegex.test(username);
    
    console.debug(`[Validation] Username validation result: ${isValid}`);
    return isValid;
}

/**
 * Validate name (first name, last name)
 * - Only letters and spaces
 * - Length between 2 and 30 characters
 * @param {string} name - Name to validate
 * @returns {boolean} - True if name is valid
 */
export function isValidName(name) {
    console.debug(`[Validation] Validating name: ${name || 'undefined'}`);
    
    if (!name) {
        console.debug("[Validation] Name validation failed: name is empty");
        return false;
    }
    
    const nameRegex = /^[a-zA-Z\s]{2,30}$/;
    const isValid = nameRegex.test(name);
    
    console.debug(`[Validation] Name validation result: ${isValid}`);
    return isValid;
}

/**
 * Get form validation errors
 * @param {Object} formData - Form data to validate
 * @returns {Object} - Object with validation errors
 */
export function getSignupFormErrors(formData) {
    console.debug("[Validation] Validating signup form data");
    
    if (!formData) {
        console.error("[Validation] Form validation failed: formData is null or undefined");
        return { general: 'Invalid form data provided' };
    }
    
    const errors = {};
    
    if (!isValidName(formData.firstName)) {
        errors.firstName = VALIDATION_ERRORS.FIRST_NAME_INVALID;
    }

    if (!isValidName(formData.lastName)) {
        errors.lastName = VALIDATION_ERRORS.LAST_NAME_INVALID;
    }

    if (!isValidUsername(formData.username)) {
        errors.username = VALIDATION_ERRORS.USERNAME_INVALID;
    }

    if (!isValidEmail(formData.email)) {
        errors.email = VALIDATION_ERRORS.EMAIL_INVALID;
    }

    if (!isValidPassword(formData.password)) {
        errors.password = VALIDATION_ERRORS.PASSWORD_WEAK;
    }

    if (formData.password !== formData.confirmPassword) {
        console.debug("[Validation] Password confirmation validation failed: passwords don't match");
        errors.confirmPassword = VALIDATION_ERRORS.PASSWORD_MISMATCH;
    }
    
    const errorCount = Object.keys(errors).length;
    if (errorCount > 0) {
        console.warn(`[Validation] Form validation found ${errorCount} errors:`, Object.keys(errors));
    } else {
        console.info("[Validation] Form validation passed successfully");
    }
    
    return errors;
}

/**
 * Validate post content
 * @param {Object} postData - Post data to validate
 * @returns {Object} - Object with validation errors
 */
export function validatePostContent(postData) {
    console.debug("[Validation] Validating post content");
    
    if (!postData) {
        console.error("[Validation] Post validation failed: postData is null or undefined");
        return { general: 'Invalid post data provided' };
    }
    
    const errors = {};
    
    if (!postData.title || postData.title.trim() === '') {
        errors.title = CONTENT_ERRORS.TITLE_REQUIRED;
    } else if (!isValidLength(postData.title, 3, 200)) {
        errors.title = CONTENT_ERRORS.TITLE_TOO_LONG;
    }

    if (!postData.content || postData.content.trim() === '') {
        errors.content = CONTENT_ERRORS.CONTENT_REQUIRED;
    } else if (!isValidLength(postData.content, 10, 10000)) {
        errors.content = CONTENT_ERRORS.CONTENT_TOO_LONG;
    }
    
    if (!postData.categories || postData.categories.length === 0) {
        console.debug("[Validation] Post categories validation failed: no categories selected");
        errors.categories = 'Please select at least one category';
    }
    
    const errorCount = Object.keys(errors).length;
    if (errorCount > 0) {
        console.warn(`[Validation] Post validation found ${errorCount} errors:`, Object.keys(errors));
    } else {
        console.info("[Validation] Post validation passed successfully");
    }
    
    return errors;
}

/**
 * Validate comment content
 * @param {string} content - Comment content
 * @returns {Object} - Object with validation result
 */
export function validateCommentContent(content) {
    console.debug("[Validation] Validating comment content");
    
    if (!content || content.trim() === '') {
        console.debug("[Validation] Comment validation failed: content is empty");
        return {
            valid: false,
            error: 'Comment cannot be empty'
        };
    }
    
    if (content.length > 200) {
        console.debug(`[Validation] Comment validation failed: content exceeds max length (${content.length}/200)`);
        return {
            valid: false,
            error: 'Comment cannot exceed 200 characters'
        };
    }
    
    console.debug("[Validation] Comment validation passed");
    return { valid: true };
}

/**
 * Validate age from date of birth
 * @param {string} dateOfBirth - Date of birth in ISO format
 * @param {number} minAge - Minimum age required
 * @returns {boolean} - True if age is valid
 */
export function isValidAge(dateOfBirth, minAge = 12) {
    console.debug(`[Validation] Validating age for DOB: ${dateOfBirth || 'undefined'}, min age: ${minAge}`);
    
    if (!dateOfBirth) {
        console.debug("[Validation] Age validation failed: date of birth is empty");
        return false;
    }
    
    try {
        const birthDate = new Date(dateOfBirth);
        if (isNaN(birthDate.getTime())) {
            console.debug("[Validation] Age validation failed: invalid date format");
            return false;
        }
        
        const today = new Date();
        let age = today.getFullYear() - birthDate.getFullYear();
        const monthDiff = today.getMonth() - birthDate.getMonth();
        
        if (monthDiff < 0 || (monthDiff === 0 && today.getDate() < birthDate.getDate())) {
            age--;
        }
        
        const isValid = age >= minAge;
        console.debug(`[Validation] Age validation result: ${isValid}, calculated age: ${age}`);
        return isValid;
        
    } catch (e) {
        console.error('[Validation] Error calculating age:', e.message || e);
        return false;
    }
}

/**
 * Validate username length
 * @param {Object} formData - Form data to validate
 * @returns {Object} - Object with validation result
 */
export function validateUsernameLength(formData) {
    console.debug("[Validation] Validating username length");
    
    if (!formData) {
        console.error("[Validation] Username validation failed: formData is null or undefined");
        return { success: false, error: 'Invalid form data provided.' };
    }
    
    if (!formData.username) {
        console.debug("[Validation] Username validation failed: username is empty");
        return { success: false, error: 'Username is required.' };
    }
    
    if (formData.username.length < 3) {
        console.debug(`[Validation] Username validation failed: length (${formData.username.length}) < 3`);
        return { success: false, error: 'Username must be at least 3 characters long.' };
    }
    
    console.debug("[Validation] Username length validation passed");
    return { success: true };
}