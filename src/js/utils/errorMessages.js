/**
 * Enhanced Error Messages - User-friendly, actionable messages
 * Following Material Design 3 principles for clear, helpful communication
 */

// Authentication Error Messages
export const AUTH_ERRORS = {
    INVALID_CREDENTIALS: "The username/email or password you entered is incorrect. Please check your credentials and try again.",
    USER_NOT_FOUND: "We couldn't find an account with that username or email. Would you like to sign up instead?",
    INCORRECT_PASSWORD: "The password you entered is incorrect. Please try again or reset your password if you've forgotten it.",
    ACCOUNT_EXISTS: "An account with this email already exists. Try logging in instead, or use a different email address.",
    USERNAME_EXISTS: "This username is already taken. Please choose a different username.",
    SESSION_EXPIRED: "Your session has expired for security reasons. Please log in again to continue.",
    SESSION_INVALID: "Your session is no longer valid. Please log in again to access this content.",
    AUTH_REQUIRED: "You need to be logged in to access this content. Please log in to continue.",
    ACCESS_DENIED: "You don't have permission to access this content. Contact support if you believe this is an error."
};

// Form Validation Error Messages
export const VALIDATION_ERRORS = {
    EMAIL_REQUIRED: "Email address is required. Please enter your email address.",
    EMAIL_INVALID: "Please enter a valid email address (e.g., user@example.com).",
    PASSWORD_REQUIRED: "Password is required. Please enter your password.",
    PASSWORD_TOO_SHORT: "Password must be at least 8 characters long. Please choose a stronger password.",
    PASSWORD_WEAK: "Password must contain at least one uppercase letter, one lowercase letter, and one number.",
    USERNAME_REQUIRED: "Username is required. Please choose a username.",
    USERNAME_INVALID: "Username must be 3-20 characters long and contain only letters, numbers, and underscores.",
    FIRST_NAME_REQUIRED: "First name is required. Please enter your first name.",
    FIRST_NAME_INVALID: "First name should contain only letters and be 2-30 characters long.",
    LAST_NAME_REQUIRED: "Last name is required. Please enter your last name.",
    LAST_NAME_INVALID: "Last name should contain only letters and be 2-30 characters long.",
    PASSWORD_MISMATCH: "Passwords don't match. Please make sure both password fields are identical.",
    DATE_INVALID: "Please enter a valid date.",
    GENDER_REQUIRED: "Please select your gender."
};

// Content Validation Error Messages
export const CONTENT_ERRORS = {
    TITLE_REQUIRED: "Post title is required. Please enter a title for your post.",
    TITLE_TOO_LONG: "Post title is too long. Please keep it under 100 characters.",
    CONTENT_REQUIRED: "Post content is required. Please write some content for your post.",
    CONTENT_TOO_LONG: "Post content is too long. Please keep it under 500 characters.",
    CATEGORIES_REQUIRED: "Please select at least one category for your post.",
    COMMENT_REQUIRED: "Comment content is required. Please write your comment.",
    COMMENT_TOO_LONG: "Comment is too long. Please keep it under 1,000 characters."
};

// System Error Messages
export const SYSTEM_ERRORS = {
    DATABASE_CONNECTION: "We're experiencing technical difficulties. Please try again in a moment.",
    DATABASE_OPERATION: "Something went wrong while processing your request. Please try again.",
    SERVER_ERROR: "An unexpected error occurred. Our team has been notified. Please try again later.",
    NETWORK_ERROR: "Network connection failed. Please check your internet connection and try again.",
    FILE_UPLOAD_FAILED: "File upload failed. Please check the file size and format, then try again.",
    FILE_TOO_BIG: "File is too large. Please choose a file smaller than 5MB.",
    FILE_INVALID_FORMAT: "Invalid file format. Please upload a JPEG, PNG, or GIF image."
};

// API Error Messages
export const API_ERRORS = {
    NOT_FOUND: "The requested resource was not found. Please check the URL and try again.",
    METHOD_NOT_ALLOWED: "This action is not allowed. Please use the correct method for this request.",
    RATE_LIMIT: "Too many requests. Please wait a moment before trying again.",
    INVALID_REQUEST: "Invalid request format. Please check your input and try again.",
    UNAUTHORIZED: "Authentication required. Please log in to access this resource.",
    FORBIDDEN: "Access denied. You don't have permission to perform this action."
};

// Chat and Messaging Error Messages
export const CHAT_ERRORS = {
    MESSAGE_SEND_FAILED: "Failed to send message. Please check your connection and try again.",
    RECIPIENT_OFFLINE: "The recipient is currently offline. Your message will be delivered when they come online.",
    CONVERSATION_NOT_FOUND: "Conversation not found. It may have been deleted or you don't have access to it.",
    MESSAGE_TOO_LONG: "Message is too long. Please keep it under 1,000 characters.",
    INVALID_RECIPIENT: "Invalid recipient. Please select a valid user to send the message to."
};

// General User-Friendly Messages
export const GENERAL_ERRORS = {
    BAD_REQUEST: "There was a problem with your request. Please check your input and try again.",
    NOT_FOUND: "The page or resource you're looking for doesn't exist. Please check the URL.",
    SERVER_ERROR: "Something went wrong on our end. Please try again in a few minutes.",
    TIMEOUT: "The request took too long to process. Please try again.",
    UNKNOWN: "An unexpected error occurred. Please try again or contact support if the problem persists."
};

// Success Messages
export const SUCCESS_MESSAGES = {
    LOGIN_SUCCESS: "Welcome back! You've been successfully logged in.",
    SIGNUP_SUCCESS: "Account created successfully! Welcome to our community.",
    POST_CREATED: "Your post has been published successfully!",
    COMMENT_ADDED: "Your comment has been added successfully!",
    MESSAGE_SENT: "Message sent successfully!",
    PROFILE_UPDATED: "Your profile has been updated successfully!",
    PASSWORD_CHANGED: "Your password has been changed successfully!",
    FILE_UPLOADED: "File uploaded successfully!"
};

// Helper function to get user-friendly error message based on error code/type
export function getUserFriendlyError(errorCode, errorMessage, context = 'general') {
    // First try to match specific error codes
    switch (errorCode) {
        case '400':
        case 'VALIDATION_ERROR':
            return GENERAL_ERRORS.BAD_REQUEST;
        case '401':
        case 'INVALID_CREDENTIALS':
            return AUTH_ERRORS.INVALID_CREDENTIALS;
        case 'INVALID_SESSION':
            return AUTH_ERRORS.SESSION_INVALID;
        case '403':
        case 'FORBIDDEN':
            return AUTH_ERRORS.ACCESS_DENIED;
        case '404':
        case 'NOT_FOUND':
            return context === 'api' ? API_ERRORS.NOT_FOUND : GENERAL_ERRORS.NOT_FOUND;
        case '409':
        case 'DUPLICATE_ENTRY':
            if (errorMessage && errorMessage.toLowerCase().includes('email')) {
                return AUTH_ERRORS.ACCOUNT_EXISTS;
            }
            if (errorMessage && errorMessage.toLowerCase().includes('username')) {
                return AUTH_ERRORS.USERNAME_EXISTS;
            }
            return GENERAL_ERRORS.BAD_REQUEST;
        case '429':
        case 'RATE_LIMITED':
            return API_ERRORS.RATE_LIMIT;
        case '500':
        case 'DATABASE_ERROR':
            return SYSTEM_ERRORS.DATABASE_CONNECTION;
        case 'SESSION_ERROR':
            return AUTH_ERRORS.SESSION_EXPIRED;
        case 'NETWORK_ERROR':
            return SYSTEM_ERRORS.NETWORK_ERROR;
        default:
            // Try to extract meaningful information from the error message
            if (errorMessage) {
                const lowerMessage = errorMessage.toLowerCase();
                
                // Authentication related
                if (lowerMessage.includes('already exists') && lowerMessage.includes('email')) {
                    return AUTH_ERRORS.ACCOUNT_EXISTS;
                }
                if (lowerMessage.includes('already exists') && lowerMessage.includes('username')) {
                    return AUTH_ERRORS.USERNAME_EXISTS;
                }
                if (lowerMessage.includes('invalid credentials') || lowerMessage.includes('authentication failed')) {
                    return AUTH_ERRORS.INVALID_CREDENTIALS;
                }
                if (lowerMessage.includes('not found') && (lowerMessage.includes('user') || lowerMessage.includes('account'))) {
                    return AUTH_ERRORS.USER_NOT_FOUND;
                }
                
                // Network/Connection related
                if (lowerMessage.includes('network') || lowerMessage.includes('connection')) {
                    return SYSTEM_ERRORS.NETWORK_ERROR;
                }
                
                // Database related
                if (lowerMessage.includes('database') || lowerMessage.includes('sql')) {
                    return SYSTEM_ERRORS.DATABASE_CONNECTION;
                }
            }
            
            return errorMessage || GENERAL_ERRORS.UNKNOWN;
    }
}

// Helper function to get appropriate icon for error type
export function getErrorIcon(errorCode, context = 'general') {
    switch (errorCode) {
        case '401':
        case '403':
        case 'INVALID_CREDENTIALS':
        case 'INVALID_SESSION':
            return 'fas fa-lock';
        case '404':
        case 'NOT_FOUND':
            return 'fas fa-search';
        case '429':
        case 'RATE_LIMITED':
            return 'fas fa-clock';
        case '500':
        case 'DATABASE_ERROR':
        case 'SERVER_ERROR':
            return 'fas fa-server';
        case 'NETWORK_ERROR':
            return 'fas fa-wifi';
        case 'VALIDATION_ERROR':
            return 'fas fa-exclamation-triangle';
        default:
            return 'fas fa-exclamation-circle';
    }
}

// Helper function to determine if error should show retry button
export function shouldShowRetry(errorCode) {
    const retryableErrors = ['500', 'DATABASE_ERROR', 'NETWORK_ERROR', 'TIMEOUT', '429'];
    return retryableErrors.includes(errorCode);
}

// Helper function to get suggested actions for error
export function getErrorActions(errorCode, context = 'general') {
    switch (errorCode) {
        case '401':
        case 'INVALID_CREDENTIALS':
            return [
                { text: 'Try Again', action: 'retry', primary: true },
                { text: 'Forgot Password?', action: 'forgot-password', primary: false }
            ];
        case 'INVALID_SESSION':
            return [
                { text: 'Log In', action: 'login', primary: true }
            ];
        case '404':
        case 'NOT_FOUND':
            return [
                { text: 'Go Home', action: 'home', primary: true },
                { text: 'Go Back', action: 'back', primary: false }
            ];
        case '429':
        case 'RATE_LIMITED':
            return [
                { text: 'Try Again Later', action: 'retry-later', primary: true }
            ];
        case 'DUPLICATE_ENTRY':
            if (context === 'signup') {
                return [
                    { text: 'Try Different Email', action: 'retry', primary: true },
                    { text: 'Log In Instead', action: 'login', primary: false }
                ];
            }
            return [{ text: 'Try Again', action: 'retry', primary: true }];
        default:
            if (shouldShowRetry(errorCode)) {
                return [
                    { text: 'Try Again', action: 'retry', primary: true },
                    { text: 'Contact Support', action: 'support', primary: false }
                ];
            }
            return [{ text: 'Try Again', action: 'retry', primary: true }];
    }
}
