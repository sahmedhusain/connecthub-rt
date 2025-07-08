import { getSignupFormErrors } from '../utils/validation.js';
import { signup } from '../utils/api.js';
import { handleApiError } from './error.js';
import { getUserFriendlyError, getErrorIcon, getErrorActions } from '../utils/errorMessages.js';

export function renderSignup() {
    console.debug("[Signup] Rendering signup page");
    
    const appContainer = document.getElementById('app');
    if (!appContainer) {
        console.error("[Signup] App container not found in DOM");
        return;
    }

    appContainer.innerHTML = '';

    const signupHtml = `
    <header>
        <a href="/" class="logo-container">
            <img src="/static/assets/logo.png" alt="Connect Hub Logo">
            <span>Connect</span><span>Hub</span>
        </a>
        <div class="user-actions">
            <a href="/"><button class="btn btn-secondary"><i class="fas fa-sign-in-alt"></i> Sign In</button></a>
        </div>
    </header>
    <main>
        <div class="auth-container">
            <div class="register-form-container">
                <div class="logo">
                    <div class="logotext">
                        <span>Create Your Account</span>
                    </div>
                </div>
                
                <div id="error-message" class="error-message" style="display: none;"></div>

                <form id="signup-form">
                    <div class="form-row">
                        <div class="form-group">
                            <label for="first_name">First Name</label>
                            <input type="text" id="first_name" name="first_name" class="form-control" placeholder="Enter your first name" required>
                        </div>

                        <div class="form-group">
                            <label for="last_name">Last Name</label>
                            <input type="text" id="last_name" name="last_name" class="form-control" placeholder="Enter your last name" required>
                        </div>
                    </div>

                    <div class="form-row">
                        <div class="form-group">
                            <label for="username">Username</label>
                            <input type="text" id="username" name="username" class="form-control" placeholder="Choose a username" required>
                        </div>

                        <div class="form-group">
                            <label for="email">Email</label>
                            <input type="email" id="email" name="email" class="form-control" placeholder="Enter your email address" required>
                        </div>
                    </div>

                    <div class="form-row">
                        <div class="form-group">
                            <label for="gender">Gender</label>
                            <select id="gender" name="gender" class="form-control" required>
                                <option value="" disabled selected>Select gender</option>
                                <option value="male">Male</option>
                                <option value="female">Female</option>
                            </select>
                        </div>

                        <div class="form-group">
                            <label for="date_of_birth">Date of Birth</label>
                            <input type="date" id="date_of_birth" name="date_of_birth" class="form-control" required>
                        </div>
                    </div>

                    <div class="password-row">
                        <div class="form-group">
                            <label for="password">Password</label>
                            <input type="password" id="password" name="password" class="form-control" placeholder="Create a password" required>
                        </div>

                        <div class="form-group">
                            <label for="confirm-password">Confirm Password</label>
                            <input type="password" id="confirm-password" name="confirm-password" class="form-control" placeholder="Re-enter your password" required>
                        </div>
                    </div>

                    <div class="form-actions">
                        <button type="submit" class="btn btn-primary register-btn md3-enhanced">Sign Up</button>
                    </div>
                </form>
                
                <div class="auth-link">
                    Already have an account? <a href="/">Log In</a>
                </div>
            </div>
        </div>
    </main>
    <footer>
        <p>© 2025 ConnectHub | All rights reserved.</p>
    </footer>
    `;

    appContainer.innerHTML = signupHtml;

    const signupForm = document.getElementById('signup-form');
    if (signupForm) {
        signupForm.addEventListener('submit', handleSignupSubmit);
        console.debug("[Signup] Form event listener attached");
    } else {
        console.error("[Signup] Signup form not found in DOM after rendering");
    }
    
    console.info("[Signup] Signup page rendered successfully");
}

async function handleSignupSubmit(event) {
    event.preventDefault();
    console.debug("[Signup] Form submission initiated");

    const errorMessage = document.getElementById('error-message');
    if (!errorMessage) {
        console.error("[Signup] Error message element not found in DOM");
        return;
    }
    
    errorMessage.style.display = 'none';

    // Get form values
    const formElements = {
        firstName: document.getElementById('first_name'),
        lastName: document.getElementById('last_name'),
        username: document.getElementById('username'),
        email: document.getElementById('email'),
        gender: document.getElementById('gender'),
        dateOfBirth: document.getElementById('date_of_birth'),
        password: document.getElementById('password'),
        confirmPassword: document.getElementById('confirm-password')
    };
    
    // Check if all form elements exist
    for (const [key, element] of Object.entries(formElements)) {
        if (!element) {
            console.error(`[Signup] Form element not found: ${key}`);
            showError('Registration form is incomplete. Please refresh the page and try again.');
            return;
        }
    }

    const formData = {
        firstName: formElements.firstName.value.trim(),
        lastName: formElements.lastName.value.trim(),
        username: formElements.username.value.trim(),
        email: formElements.email.value.trim(),
        gender: formElements.gender.value,
        dateOfBirth: formElements.dateOfBirth.value,
        password: formElements.password.value,
        confirmPassword: formElements.confirmPassword.value
    };
    
    console.debug("[Signup] Form data collected, validating...");

    const errors = getSignupFormErrors(formData);
    if (Object.keys(errors).length > 0) {
        const firstError = Object.values(errors)[0];
        console.warn(`[Signup] Validation failed: ${firstError}`);
        
        // Log all validation errors at debug level
        console.debug("[Signup] All validation errors:", errors);
        
        showError(firstError);
        return;
    }
    
    console.info("[Signup] Form validation passed, proceeding with registration");

    const submitButton = document.querySelector('#signup-form button[type="submit"]');
    if (submitButton) {
        submitButton.disabled = true;
        submitButton.classList.add('md3-loading');
        submitButton.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Creating Account...';
        console.debug("[Signup] Submit button disabled during processing");
    }

    try {
        console.debug("[Signup] Sending registration request for username:", formData.username);
        const result = await signup(formData);

        if (result.success) {
            console.info("[Signup] Registration successful, redirecting to home page");
            window.location.href = '/home';
        } else {
            console.warn("[Signup] Registration failed:", result.error || "Unknown error");

            // Get user-friendly error message
            const userFriendlyMessage = getUserFriendlyError(
                result.code || '400',
                result.error || 'Registration failed',
                'signup'
            );

            const errorResult = handleApiError({
                status: result.code || '400',
                error: userFriendlyMessage,
                message: userFriendlyMessage
            }, {
                renderPage: false,
                context: 'form_validation',
                fallbackToInline: true
            });

            if (errorResult.shouldShowInline) {
                showEnhancedError(userFriendlyMessage, result.code || '400', 'signup');
            }
            // If not inline, the error page would have been rendered
        }
    } catch (error) {
        console.error('[Signup] Error during registration:', error.message || error);
        const networkErrorMessage = getUserFriendlyError('NETWORK_ERROR', error.message, 'signup');
        showEnhancedError(networkErrorMessage, 'NETWORK_ERROR', 'signup');
    } finally {
        if (submitButton) {
            submitButton.disabled = false;
            submitButton.innerHTML = 'Sign Up';
            console.debug("[Signup] Submit button re-enabled");
        }
    }
}

function showError(message) {
    console.debug("[Signup] Showing error message:", message);

    const errorElement = document.getElementById('error-message');
    if (!errorElement) {
        console.error("[Signup] Error element not found when trying to show error");
        return;
    }

    errorElement.textContent = message;
    errorElement.style.display = 'block';

    // Ensure the error is visible (scroll to it if needed)
    errorElement.scrollIntoView({ behavior: 'smooth', block: 'center' });
}

function showEnhancedError(message, errorCode, context) {
    console.debug("[Signup] Showing enhanced error message:", message);

    const errorElement = document.getElementById('error-message');
    if (!errorElement) {
        console.error("[Signup] Error element not found when trying to show error");
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

    // Ensure the error is visible (scroll to it if needed)
    errorElement.scrollIntoView({ behavior: 'smooth', block: 'center' });
}

function getActionHandler(action, context) {
    switch (action) {
        case 'retry':
            return 'clearSignupError(); document.getElementById("email").focus();';
        case 'login':
            return 'window.location.href = "/#/"';
        case 'home':
            return 'window.location.href = "/#/home"';
        case 'back':
            return 'history.back()';
        default:
            return 'console.log("Action not implemented:", "' + action + '")';
    }
}

// Helper function to clear signup error
function clearSignupError() {
    const errorElement = document.getElementById('error-message');
    if (errorElement) {
        errorElement.style.display = 'none';
        errorElement.classList.remove('visible');
        errorElement.innerHTML = '';
    }
}

// Make clearSignupError globally accessible for onclick handlers
window.clearSignupError = clearSignupError;

function validateEmail(email) {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
}

function validatePassword(password) {
    // At least 8 characters, at least one letter and one number
    const passwordRegex = /^(?=.*[A-Za-z])(?=.*\d)[A-Za-z\d]{8,}$/;
    return passwordRegex.test(password);
}

function validateAge(dateOfBirth) {
    if (!dateOfBirth) return false;
    
    try {
        const birthDate = new Date(dateOfBirth);
        const today = new Date();
        
        // Calculate age
        let age = today.getFullYear() - birthDate.getFullYear();
        const monthDiff = today.getMonth() - birthDate.getMonth();
        
        if (monthDiff < 0 || (monthDiff === 0 && today.getDate() < birthDate.getDate())) {
            age--;
        }
        
        console.debug(`[Signup] Calculated age for validation: ${age} years`);
        return age >= 12; // At least 12 years old
    } catch (e) {
        console.error('[Signup] Error calculating age:', e.message || e);
        return false;
    }
}
