import { getUserFriendlyError, getErrorIcon, getErrorActions } from '../utils/errorMessages.js';

/**
 * Renders an error page with the given error code and message
 * @param {string} code - The error code to display
 * @param {string} message - The error message to display
 * @param {Error|null} error - Optional Error object for logging details
 * @param {string} errorType - Type of error: "navigation", "authentication", "api", "critical"
 * @param {string} originalPath - The original path that caused the error
 */
export function renderError(code = '404', message = 'Page Not Found', error = null, errorType = 'navigation', originalPath = '') {
    console.info(`[Error] Rendering error page: ${code} - ${message} (Type: ${errorType})`);
    console.debug('[Error] renderError called with parameters:', { code, message, error, errorType, originalPath });

    const appContainer = document.getElementById('app');
    if (!appContainer) {
        console.error('[Error] Cannot render error page: #app container not found in DOM');
        return;
    }

    console.debug('[Error] App container found, proceeding with error page rendering');

    const { title, description, actions } = getErrorPageContent(code, message, errorType, originalPath);
    console.debug('[Error] Generated error page content:', { title, description, actions });

    const errorHtml = `
    <main class="error-page-main">
        <div class="error-content md3-enhanced">
            <div class="error-icon-container">
                <img src="/static/assets/error.png" alt="Error Icon" onerror="this.style.display='none'; this.nextElementSibling.style.display='block';">
                <div class="error-icon-fallback" style="display: none;">
                    <i class="fas fa-exclamation-triangle"></i>
                </div>
            </div>
            <h1 class="error-code">${code}</h1>
            <h2 class="error-title">${title}</h2>
            <p class="error-description">${description}</p>
            <div class="error-actions">
                ${actions}
            </div>
        </div>
    </main>
    <footer class="error-page-footer">
        <p>© 2024 ConnectHub | All rights reserved.</p>
    </footer>
    `;

    console.debug('[Error] Generated HTML length:', errorHtml.length);
    appContainer.innerHTML = errorHtml;
    console.debug('[Error] Error page HTML set to app container');

    // Log additional error details if provided
    if (error) {
        console.error('[Error] Detailed error information:', {
            code,
            message,
            errorType,
            originalPath,
            errorObject: error,
            stack: error.stack,
            timestamp: new Date().toISOString()
        });
    }

    // Track error occurrence for analytics (if needed)
    if (typeof window.trackEvent === 'function') {
        try {
            window.trackEvent('error_displayed', { code, message, errorType });
        } catch (e) {
            console.debug('[Error] Could not track error event:', e.message);
        }
    }

    // Add smooth scroll to top for better UX
    setTimeout(() => {
        window.scrollTo({ top: 0, behavior: 'smooth' });
    }, 100);
}

/**
 * Gets appropriate error page content based on error type and code
 * @param {string} code - Error code
 * @param {string} message - Error message
 * @param {string} errorType - Type of error
 * @param {string} originalPath - Original path that caused error
 * @returns {Object} Object with title, description, and actions HTML
 */
function getErrorPageContent(code, message, errorType, originalPath) {
    let title = message;
    let description = 'Oops! Something went wrong.';
    let actions = '';

    switch (errorType) {
        case 'authentication':
            title = 'Authentication Required';
            description = 'You need to be logged in to access this content.';
            actions = `
                <a href="/" class="btn btn-primary md3-enhanced">
                    <i class="fas fa-sign-in-alt"></i> Go to Login
                </a>
                <button onclick="window.location.reload()" class="btn btn-secondary md3-enhanced">
                    <i class="fas fa-rotate-right"></i> Try Again
                </button>
            `;
            break;

        case 'navigation':
            if (code === '404') {
                title = 'Page Not Found';
                description = 'The page you\'re looking for doesn\'t exist or has been moved.';
                actions = `
                    <a href="/" class="btn btn-primary md3-enhanced">
                        <i class="fas fa-home"></i> Go to Home
                    </a>
                    <button onclick="window.history.back()" class="btn btn-secondary md3-enhanced">
                        <i class="fas fa-arrow-left"></i> Go Back
                    </button>
                `;
            } else {
                actions = `
                    <a href="/" class="btn btn-primary md3-enhanced">
                        <i class="fas fa-home"></i> Go to Home
                    </a>
                    <button onclick="window.location.reload()" class="btn btn-secondary md3-enhanced">
                        <i class="fas fa-rotate-right"></i> Try Again
                    </button>
                `;
            }
            break;

        case 'api':
            title = 'Service Unavailable';
            description = 'We\'re having trouble connecting to our servers. Please try again in a moment.';
            actions = `
                <button onclick="window.location.reload()" class="btn btn-primary md3-enhanced">
                    <i class="fas fa-rotate-right"></i> Retry
                </button>
                <a href="/" class="btn btn-secondary md3-enhanced">
                    <i class="fas fa-home"></i> Go to Home
                </a>
            `;
            break;

        case 'critical':
            title = 'System Error';
            description = 'A critical error has occurred. Please contact support if this problem persists.';
            actions = `
                <button onclick="window.location.reload()" class="btn btn-primary md3-enhanced">
                    <i class="fas fa-rotate-right"></i> Reload Page
                </button>
                <a href="/" class="btn btn-secondary md3-enhanced">
                    <i class="fas fa-home"></i> Start Over
                </a>
            `;
            break;

        default:
            actions = `
                <a href="/" class="btn btn-primary md3-enhanced">
                    <i class="fas fa-home"></i> Go to Home
                </a>
                <button onclick="window.location.reload()" class="btn btn-secondary md3-enhanced">
                    <i class="fas fa-rotate-right"></i> Try Again
                </button>
            `;
    }

    return { title, description, actions };
}

/**
 * Determines if an error should show full error page vs inline error
 * @param {Object} errorInfo - Error information
 * @returns {boolean} True if should show full error page
 */
export function shouldShowFullErrorPage(errorInfo) {
    const { code, errorType, context } = errorInfo;

    // Always show full error page for these cases
    if (errorType === 'navigation' || errorType === 'authentication' || errorType === 'critical') {
        return true;
    }

    // Show full error page for specific error codes
    if (['401', '403', '404', '500', '503'].includes(code)) {
        return true;
    }

    // Show full error page if entire page content cannot be loaded
    if (context === 'page_load' || context === 'initial_data') {
        return true;
    }

    // Use inline error for form validation, partial content failures, etc.
    return false;
}

/**
 * Handles API errors and renders appropriate error pages or inline errors
 * @param {Object} apiError - Error response from API
 * @param {Object} options - Options for error handling
 */
export function handleApiError(apiError, options = {}) {
    const {
        renderPage = true,
        context = 'api',
        originalPath = '',
        fallbackToInline = false
    } = options;

    // Default values
    let errorCode = '500';
    let errorMessage = 'Server Error';
    let errorType = 'api';

    // Extract specific error info if available
    if (apiError.status) {
        errorCode = apiError.status.toString();
    }

    if (apiError.message) {
        errorMessage = apiError.message;
    } else if (apiError.error) {
        errorMessage = apiError.error;
    }

    switch (errorCode) {
        case '401':
            errorType = 'authentication';
            errorMessage = 'You need to be logged in to access this content';
            break;
        case '403':
            errorType = 'authentication';
            errorMessage = 'You don\'t have permission to access this content';
            break;
        case '404':
            errorType = context === 'page_load' ? 'navigation' : 'api';
            errorMessage = context === 'page_load' ? 'Page not found' : 'The requested resource was not found';
            break;
        case '429':
            errorMessage = 'Too many requests. Please try again later';
            break;
        case '500':
        case '502':
        case '503':
        case '504':
            errorType = 'critical';
            errorMessage = 'Server error. Please try again later';
            break;
    }

    console.warn(`[Error] API error occurred: ${errorCode} - ${errorMessage} (Type: ${errorType})`);

    const errorInfo = { code: errorCode, errorType, context };

    if (renderPage && shouldShowFullErrorPage(errorInfo)) {
        renderError(errorCode, errorMessage, apiError, errorType, originalPath);
    } else if (fallbackToInline) {
        return {
            success: false,
            error: errorMessage,
            code: errorCode,
            type: errorType,
            shouldShowInline: true
        };
    }

    return {
        success: false,
        error: errorMessage,
        code: errorCode,
        type: errorType,
        shouldShowInline: false
    };
}

/**
 * Creates an enhanced inline error display with Material Design 3 styling
 * @param {string} message - The error message to display
 * @param {string} errorCode - The error code for styling and actions
 * @param {string} context - The context where the error occurred
 * @param {HTMLElement} container - The container element to display the error in
 * @param {Object} options - Additional options for error display
 */
export function showEnhancedInlineError(message, errorCode, context, container, options = {}) {
    if (!container) {
        console.error('[Error] No container provided for inline error display');
        return;
    }

    // Get user-friendly message if not already provided
    const userFriendlyMessage = getUserFriendlyError(errorCode, message, context);
    const icon = getErrorIcon(errorCode, context);
    const actions = getErrorActions(errorCode, context);

    // Create enhanced error HTML
    let actionsHTML = '';
    if (actions && actions.length > 0 && !options.hideActions) {
        actionsHTML = '<div class="error-actions">';
        actions.forEach(action => {
            const buttonClass = action.primary ? 'btn btn-primary btn-sm md3-enhanced' : 'btn btn-secondary btn-sm md3-enhanced';
            const onclick = options.actionHandlers?.[action.action] || `console.log('Action: ${action.action}')`;
            actionsHTML += `<button class="${buttonClass}" onclick="${onclick}" aria-label="${action.text}">${action.text}</button>`;
        });
        actionsHTML += '</div>';
    }

    const errorHTML = `
        <div class="error-inline md3-enhanced" role="alert" aria-live="polite">
            <i class="${icon}" aria-hidden="true"></i>
            <div class="error-content">
                <div class="error-message">${userFriendlyMessage}</div>
                ${actionsHTML}
            </div>
        </div>
    `;

    container.innerHTML = errorHTML;
    container.style.display = 'block';
    container.classList.add('visible');

    // Add accessibility features
    container.setAttribute('tabindex', '-1');

    // Auto-hide after specified time if requested
    if (options.autoHide && options.autoHideDelay) {
        setTimeout(() => {
            hideInlineError(container);
        }, options.autoHideDelay);
    }

    // Scroll to error if requested
    if (options.scrollToError) {
        container.scrollIntoView({ behavior: 'smooth', block: 'center' });
    }

    console.debug('[Error] Enhanced inline error displayed:', { message: userFriendlyMessage, code: errorCode, context });
}

/**
 * Hides an inline error display
 * @param {HTMLElement} container - The container element with the error
 */
export function hideInlineError(container) {
    if (!container) return;

    container.classList.remove('visible');
    setTimeout(() => {
        container.style.display = 'none';
        container.innerHTML = '';
    }, 300); // Match CSS transition duration
}

/**
 * Creates a toast notification for errors
 * @param {string} message - The error message
 * @param {string} errorCode - The error code
 * @param {string} context - The context
 * @param {Object} options - Toast options
 */
export function showErrorToast(message, errorCode, context, options = {}) {
    const userFriendlyMessage = getUserFriendlyError(errorCode, message, context);
    const icon = getErrorIcon(errorCode, context);

    // Create toast container if it doesn't exist
    let toastContainer = document.getElementById('toast-container');
    if (!toastContainer) {
        toastContainer = document.createElement('div');
        toastContainer.id = 'toast-container';
        toastContainer.className = 'toast-container';
        toastContainer.setAttribute('aria-live', 'polite');
        toastContainer.setAttribute('aria-atomic', 'true');
        document.body.appendChild(toastContainer);
    }

    // Create toast element
    const toastId = `toast-${Date.now()}`;
    const toast = document.createElement('div');
    toast.id = toastId;
    toast.className = 'toast error-toast md3-enhanced';
    toast.setAttribute('role', 'alert');
    toast.innerHTML = `
        <div class="toast-content">
            <i class="${icon}" aria-hidden="true"></i>
            <span class="toast-message">${userFriendlyMessage}</span>
            <button class="toast-close" onclick="hideErrorToast('${toastId}')" aria-label="Close notification">
                <i class="fas fa-times"></i>
            </button>
        </div>
    `;

    toastContainer.appendChild(toast);

    // Show toast with animation
    setTimeout(() => {
        toast.classList.add('show');
    }, 100);

    // Auto-hide after delay
    const hideDelay = options.duration || 5000;
    setTimeout(() => {
        hideErrorToast(toastId);
    }, hideDelay);

    console.debug('[Error] Error toast displayed:', { message: userFriendlyMessage, code: errorCode, context });
}

/**
 * Hides an error toast
 * @param {string} toastId - The ID of the toast to hide
 */
export function hideErrorToast(toastId) {
    const toast = document.getElementById(toastId);
    if (!toast) return;

    toast.classList.remove('show');
    setTimeout(() => {
        toast.remove();
    }, 300);
}

// Make hideErrorToast available globally for onclick handlers
window.hideErrorToast = hideErrorToast;

/**
 * Renders error page from URL parameters (used by router)
 */
export function renderErrorFromURL() {
    console.debug('[Error] renderErrorFromURL called, current URL:', window.location.href);
    console.debug('[Error] Hash:', window.location.hash);

    // Parse URL parameters from hash fragment
    // URL format: /#/error?code=404&message=Test&type=navigation
    const hash = window.location.hash;
    const queryString = hash.includes('?') ? hash.split('?')[1] : '';

    console.debug('[Error] Query string:', queryString);

    const urlParams = new URLSearchParams(queryString);
    const code = urlParams.get('code') || '404';
    const message = urlParams.get('message') || 'Page Not Found';
    const errorType = urlParams.get('type') || 'navigation';
    const originalPath = urlParams.get('path') || '';

    console.info(`[Error] Rendering error from URL: ${code} - ${message} (Type: ${errorType})`);
    console.debug('[Error] Parsed parameters:', { code, message, errorType, originalPath });

    renderError(code, message, null, errorType, originalPath);
}

/**
 * Creates an inline error display element
 * @param {string} message - Error message to display
 * @param {string} code - Error code
 * @param {Object} options - Display options
 * @returns {string} HTML string for inline error
 */
export function createInlineError(message, code = '', options = {}) {
    const {
        showRetry = true,
        retryCallback = null,
        showCode = false,
        className = 'error-inline md3-enhanced',
        icon = 'fas fa-exclamation-triangle'
    } = options;

    const codeDisplay = showCode && code ? `<span class="error-code">[${code}]</span> ` : '';
    const retryButton = showRetry ? `
        <button class="error-retry md3-enhanced" ${retryCallback ? `onclick="${retryCallback}"` : 'onclick="window.location.reload()"'}>
            <i class="fas fa-rotate-right"></i> Try Again
        </button>
    ` : '';

    return `
        <div class="${className}" role="alert" aria-live="polite">
            <i class="${icon}" aria-hidden="true"></i>
            <span class="error-message">${codeDisplay}${message}</span>
            ${retryButton}
        </div>
    `;
}
