import { login } from '../utils/api.js';
import { handleApiError } from './error.js';
import { getUserFriendlyError, getErrorIcon, getErrorActions } from '../utils/errorMessages.js';

export function renderLogin() {
    console.debug("[Login] Rendering login page");
    
    const appContainer = document.getElementById('app');
    if (!appContainer) {
        console.error("[Login] App container not found in DOM");
        return;
    }

    appContainer.innerHTML = '';

    const loginHTML = `
        <header>
            <a href="/" class="logo-container">
                <img src="/static/assets/logo.png" alt="Connect Hub Logo">
                <span>Connect</span><span>Hub</span>
            </a>
            <div class="user-actions">
                <a href="/signup"><button class="btn btn-secondary"><i class="fas fa-user-plus"></i> Register</button></a>
            </div>
        </header>
        <main>
            <div class="auth-container">
                <div class="login-form-container">
                    <div class="logo">
                        <div class="logotext">
                            <span>Welcome Back!</span>
                        </div>
                    </div>
                    <div id="login-error" class="error-message" style="display: none;"></div>
                    <form id="login-form" action="/" method="POST">
                        <div class="form-group">
                            <label for="email">Email or Username</label>
                            <input type="text" id="email" name="email" class="form-control" placeholder="Enter your email or username" required>
                        </div>

                        <div class="form-group">
                            <label for="password">Password</label>
                            <input type="password" id="password" name="password" class="form-control" placeholder="Enter your password" required>
                        </div>

                        <div class="form-actions">
                            <button type="submit" class="btn btn-primary login-btn md3-enhanced">Sign In</button>
                        </div>
                    </form>
                    <div class="auth-link">
                        Don't have an account? <a href="/signup">Sign Up</a>
                    </div>
                </div>
            </div>
        </main>
        <footer>
            <p>© 2025 ConnectHub | All rights reserved.</p>
        </footer>
    `;

    appContainer.innerHTML = loginHTML;
    
    const loginForm = document.getElementById('login-form');
    if (loginForm) {
        loginForm.addEventListener('submit', handleLoginSubmit);
        console.debug("[Login] Login form event listener attached");
    } else {
        console.error("[Login] Login form not found in DOM after rendering");
    }
}

async function handleLoginSubmit(event) {
    event.preventDefault();
    console.debug("[Login] Login form submitted");

    const errorElement = document.getElementById('login-error');
    if (!errorElement) {
        console.error("[Login] Error element not found in DOM");
        return;
    }
    
    errorElement.style.display = 'none';

    const emailField = document.getElementById('email');
    const passwordField = document.getElementById('password');
    
    if (!emailField || !passwordField) {
        console.error("[Login] Form fields not found in DOM");
        return;
    }

    const identifier = emailField.value.trim();
    const password = passwordField.value;

    if (!identifier || !password) {
        console.debug("[Login] Validation failed - missing identifier or password");
        showError('Please enter both identifier and password');
        return;
    }

    try {
        console.info("[Login] Attempting login with identifier:", identifier.substring(0, 2) + '***');
        const result = await login(identifier, password);

        if (result.success) {
            console.info("[Login] Login successful, redirecting to home page");
            window.location.href = '/home';
        } else {
            console.warn("[Login] Login failed:", result.error || 'Unknown error');

            // Get user-friendly error message
            const userFriendlyMessage = getUserFriendlyError(
                result.code || '401',
                result.error || 'Invalid credentials',
                'login'
            );

            const errorResult = handleApiError({
                status: result.code || '401',
                error: userFriendlyMessage,
                message: userFriendlyMessage
            }, {
                renderPage: false,
                context: 'form_validation',
                fallbackToInline: true
            });

            if (errorResult.shouldShowInline) {
                showEnhancedError(userFriendlyMessage, result.code || '401', 'login');
            }
            // If not inline, the error page would have been rendered
        }
    } catch (error) {
        console.error("[Login] Login request failed:", error.message || error);
        const networkErrorMessage = getUserFriendlyError('NETWORK_ERROR', error.message, 'login');
        showEnhancedError(networkErrorMessage, 'NETWORK_ERROR', 'login');
    }
}

function showError(message) {
    console.debug("[Login] Showing error message:", message);

    const errorElement = document.getElementById('login-error');
    if (!errorElement) {
        console.error("[Login] Error element not found when trying to show error");
        return;
    }

    errorElement.textContent = message;
    errorElement.style.display = 'block';
    errorElement.classList.add('visible');

    setTimeout(() => {
        errorElement.classList.remove('visible');
    }, 500);
}

function showEnhancedError(message, errorCode, context) {
    console.debug("[Login] Showing enhanced error message:", message);

    const errorElement = document.getElementById('login-error');
    if (!errorElement) {
        console.error("[Login] Error element not found when trying to show error");
        return;
    }

    // Get appropriate icon and actions for this error
    const icon = getErrorIcon(errorCode, context);
    const actions = getErrorActions(errorCode, context);

    // Create enhanced error content with icon and actions
    let actionsHTML = '';
    if (actions && actions.length > 0) {
        actionsHTML = '<div class="error-actions">';
        actions.forEach(action => {
            const buttonClass = action.primary ? 'btn btn-primary btn-sm' : 'btn btn-secondary btn-sm';
            const onclick = getActionHandler(action.action, context);
            actionsHTML += `<button class="${buttonClass}" onclick="${onclick}">${action.text}</button>`;
        });
        actionsHTML += '</div>';
    }

    // Clear existing classes and add enhanced styling to the existing error-message container
    errorElement.className = 'error-message error-inline visible';

    errorElement.innerHTML = `
        <i class="${icon}"></i>
        <div class="error-content">
            <span class="error-text">${message}</span>
            ${actionsHTML}
        </div>
    `;

    errorElement.style.display = 'block';

    // Add accessibility announcement
    errorElement.setAttribute('role', 'alert');
    errorElement.setAttribute('aria-live', 'polite');

    setTimeout(() => {
        errorElement.classList.remove('visible');
    }, 500);
}

function getActionHandler(action, context) {
    switch (action) {
        case 'retry':
            return 'clearLoginError(); document.getElementById("email").focus();';
        case 'forgot-password':
            return 'alert("Password reset functionality coming soon!")';
        case 'signup':
            return 'window.location.href = "/#/signup"';
        case 'home':
            return 'window.location.href = "/#/home"';
        case 'back':
            return 'history.back()';
        default:
            return 'console.log("Action not implemented:", "' + action + '")';
    }
}

// Helper function to clear login error
function clearLoginError() {
    const errorElement = document.getElementById('login-error');
    if (errorElement) {
        errorElement.style.display = 'none';
        errorElement.classList.remove('visible');
        errorElement.innerHTML = '';
    }
}

// Make clearLoginError globally accessible for onclick handlers
window.clearLoginError = clearLoginError;

// Add a helper function to validate email format if needed
function isValidEmail(email) {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
}
