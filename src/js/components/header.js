import { logout, getCurrentUser } from '../utils/api.js';

function calculateAge(dateOfBirth) {
    if (!dateOfBirth) {
        console.debug("[Header] No dateOfBirth provided for age calculation");
        return null;
    }
    try {
        console.debug("[Header] Calculating age for dateOfBirth:", dateOfBirth);
        const birthDate = new Date(dateOfBirth);

        // Check if the date is valid
        if (isNaN(birthDate.getTime())) {
            console.warn("[Header] Invalid date format for dateOfBirth:", dateOfBirth);
            return null;
        }

        const today = new Date();
        let age = today.getFullYear() - birthDate.getFullYear();
        const m = today.getMonth() - birthDate.getMonth();
        if (m < 0 || (m === 0 && today.getDate() < birthDate.getDate())) {
            age--;
        }
        console.debug("[Header] Calculated age:", age);
        return age;
    } catch (e) {
        console.error("[Header] Error calculating age:", e.message || e);
        return null;
    }
}

export function renderHeader() {
    const user = getCurrentUser();
    const userLoggedIn = !!user;

    console.debug("[Header] Rendering header", {
        loggedIn: userLoggedIn,
        username: user?.username,
        email: user?.email,
        firstName: user?.firstName,
        lastName: user?.lastName,
        gender: user?.gender,
        dateOfBirth: user?.dateOfBirth,
        avatar: user?.avatar,
        fullUserData: user
    });

    // Validate critical user fields
    if (userLoggedIn) {
        console.debug("[Header] User data validation:", {
            hasUsername: !!user?.username,
            hasEmail: !!user?.email,
            hasFirstName: !!user?.firstName,
            hasLastName: !!user?.lastName,
            hasDateOfBirth: !!user?.dateOfBirth,
            hasGender: !!user?.gender,
            emailValue: user?.email,
            dateOfBirthValue: user?.dateOfBirth
        });
    }

    const defaultAvatar = '/static/assets/default-avatar.png';

    // Handle both simple string avatar and complex object avatar formats
    const avatarSrc = user?.avatar?.Valid ? user.avatar.String :
        (user?.avatar?.String ? user.avatar.String :
            (user?.avatar ? user.avatar : defaultAvatar));

    const headerHTML = `
        <a href="/home" class="logo-container">
            <img src="/static/assets/logo.png" alt="Connect Hub Logo">
            <span>Connect</span><span>Hub</span>
        </a>
        <div class="user-actions">
            ${userLoggedIn ? `
                <button id="new-post-btn" class="btn btn-primary"><i class="fa-regular fa-square-plus"></i> New post</button>
                <div class="user-dropdown-container">
                    <button id="avatar-dropdown-btn" class="avatar-btn">
                        <img src="${avatarSrc}" alt="User Avatar" onerror="this.onerror=null; this.src='${defaultAvatar}';">
                    </button>
                    <div id="user-dropdown-menu" class="dropdown-menu">
                        <!-- User Profile Section -->
                        <div class="dropdown-profile-section">
                            <div class="dropdown-avatar-container">
                                <img src="${avatarSrc}" alt="User Avatar" class="dropdown-avatar-large" onerror="this.onerror=null; this.src='${defaultAvatar}';">
                            </div>
                            <div class="dropdown-user-info">
                                <div class="user-fullname">${user.firstName && user.lastName ? `${user.firstName} ${user.lastName}` : user.username}</div>
                            </div>
                        </div>


                        <!-- User Details Section -->
                        <div class="dropdown-details-section">
                            <div class="dropdown-item user-detail">
                                <i class="fas fa-user"></i>
                                <span class="detail-label">Username:</span>
                                <span class="detail-value">${user.username}</span>
                            </div>
                            ${user.firstName ? `
                            <div class="dropdown-item user-detail">
                                <i class="fas fa-id-card"></i>
                                <span class="detail-label">Full Name:</span>
                                <span class="detail-value">${user.firstName} ${user.lastName || ''}</span>
                            </div>
                            ` : ''}
                            <div class="dropdown-item user-detail">
                                <i class="fas fa-envelope"></i>
                                <span class="detail-label">Email:</span>
                                <span class="detail-value">${user.email || 'Not provided'}</span>
                            </div>
                            ${user.gender ? `
                            <div class="dropdown-item user-detail">
                                <i class="fas fa-venus-mars"></i>
                                <span class="detail-label">Gender:</span>
                                <span class="detail-value">${user.gender.charAt(0).toUpperCase() + user.gender.slice(1)}</span>
                            </div>
                            ` : ''}
                            ${user.dateOfBirth ? `
                            <div class="dropdown-item user-detail">
                                <i class="fas fa-birthday-cake"></i>
                                <span class="detail-label">Age:</span>
                                <span class="detail-value">${calculateAge(user.dateOfBirth) || 'N/A'} years</span>
                            </div>
                            ` : ''}
                        </div>

                        <div class="dropdown-divider"></div>

                        <!-- Notification Settings Section -->
                        <div class="dropdown-notifications-section">
                            <div class="dropdown-section-header">
                                <i class="fas fa-bell"></i>
                                <span>Notification Settings</span>
                                <button id="dropdown-notifications-toggle" class="dropdown-toggle-btn" aria-label="Toggle notification settings">
                                    <i class="fas fa-chevron-down"></i>
                                </button>
                            </div>
                            <div id="dropdown-notifications-content" class="dropdown-notifications-content" style="display: none;">
                                <div class="notification-setting">
                                    <label class="notification-setting-label">
                                        <input type="checkbox" id="dropdown-notification-sound" class="notification-checkbox">
                                        <span class="notification-setting-text">
                                            <i class="fas fa-volume-up"></i>
                                            Sound notifications
                                        </span>
                                    </label>
                                </div>
                                <div class="notification-setting">
                                    <label class="notification-setting-label">
                                        <input type="checkbox" id="dropdown-notification-desktop" class="notification-checkbox">
                                        <span class="notification-setting-text">
                                            <i class="fas fa-desktop"></i>
                                            Desktop notifications
                                        </span>
                                    </label>
                                </div>
                                <div class="notification-setting">
                                    <label class="notification-setting-label">
                                        <input type="checkbox" id="dropdown-notification-visual" class="notification-checkbox">
                                        <span class="notification-setting-text">
                                            <i class="fas fa-eye"></i>
                                            Visual notifications
                                        </span>
                                    </label>
                                </div>
                                <div class="notification-setting volume-setting">
                                    <div class="volume-setting-header">
                                        <i class="fas fa-volume-down"></i>
                                        <span>Volume</span>
                                        <span id="dropdown-volume-display" class="volume-display">30%</span>
                                    </div>
                                    <input type="range" id="dropdown-notification-volume" class="volume-slider" min="0" max="100" value="30">
                                </div>

                                <!-- Advanced Notification Controls -->
                                <div class="notification-setting">
                                    <label class="notification-setting-label">
                                        <span class="notification-setting-text">
                                            <i class="fas fa-clock"></i>
                                            Do Not Disturb
                                        </span>
                                        <select id="dropdown-dnd-schedule" class="notification-select">
                                            <option value="off">Off</option>
                                            <option value="1h">1 Hour</option>
                                            <option value="4h">4 Hours</option>
                                            <option value="8h">8 Hours</option>
                                            <option value="custom">Custom</option>
                                        </select>
                                    </label>
                                </div>

                                <div class="notification-setting">
                                    <label class="notification-setting-label">
                                        <span class="notification-setting-text">
                                            <i class="fas fa-exclamation-triangle"></i>
                                            Priority Messages Only
                                        </span>
                                        <input type="checkbox" id="dropdown-priority-only" class="notification-checkbox">
                                    </label>
                                </div>

                                <div class="notification-setting">
                                    <label class="notification-setting-label">
                                        <span class="notification-setting-text">
                                            <i class="fas fa-music"></i>
                                            Notification Sound
                                        </span>
                                        <select id="dropdown-notification-sound-type" class="notification-select">
                                            <option value="default">Default</option>
                                            <option value="chime">Chime</option>
                                            <option value="bell">Bell</option>
                                            <option value="pop">Pop</option>
                                            <option value="custom">Custom</option>
                                        </select>
                                    </label>
                                </div>
                            </div>
                        </div>

                        <div class="dropdown-divider"></div>

                        <!-- Logout Section -->
                        <button id="dropdown-logout-btn" class="dropdown-item logout-item">
                            <i class="fas fa-sign-out-alt"></i>
                            <span>Logout</span>
                        </button>
                    </div>
                </div>
            ` : `
                <a href="/signup"><button class="btn btn-secondary"><i class="fas fa-user-plus"></i> Register</button></a>
                <a href="/"><button class="btn btn-primary"><i class="fas fa-sign-in-alt"></i> Login</button></a>
            `}
        </div>
    `;

    return headerHTML;
}

/**
 * Notification preferences management
 */
const NotificationPreferences = {
    get: () => {
        const prefs = localStorage.getItem('chatNotificationPreferences');
        return prefs ? JSON.parse(prefs) : {
            sound: true,
            desktop: true,
            visual: true,
            volume: 0.3
        };
    },

    set: (preferences) => {
        localStorage.setItem('chatNotificationPreferences', JSON.stringify(preferences));
        console.debug("[Header] Notification preferences updated:", preferences);
    }
};

export function attachHeaderEvents() {
    console.debug("[Header] Attaching header event listeners");
    const newPostBtn = document.getElementById('new-post-btn');
    const avatarDropdownBtn = document.getElementById('avatar-dropdown-btn');
    const userDropdownMenu = document.getElementById('user-dropdown-menu');
    const dropdownLogoutBtn = document.getElementById('dropdown-logout-btn');

    if (newPostBtn) {
        console.debug("[Header] Setting up new post button listener");
        newPostBtn.addEventListener('click', (e) => {
            e.preventDefault();
            console.debug("[Header] New post button clicked");
            window.appRouter?.navigate('/create-post') || (window.location.href = '/create-post');
        });
    } else {
        console.debug("[Header] New post button not found in DOM");
    }

    if (avatarDropdownBtn && userDropdownMenu) {
        console.debug("[Header] Setting up user dropdown listeners");
        avatarDropdownBtn.addEventListener('click', (e) => {
            e.stopPropagation();
            const isShowing = userDropdownMenu.classList.toggle('show');
            console.debug(`[Header] User dropdown ${isShowing ? 'opened' : 'closed'}`);

            // Initialize notification settings when dropdown opens
            if (isShowing) {
                initializeNotificationSettings();
            }
        });

        // Close dropdown when clicking outside
        document.addEventListener('click', (e) => {
            if (!avatarDropdownBtn.contains(e.target) && !userDropdownMenu.contains(e.target)) {
                if (userDropdownMenu.classList.contains('show')) {
                    userDropdownMenu.classList.remove('show');
                    console.debug("[Header] User dropdown closed via outside click");
                }
            }
        });

        // Keyboard navigation for dropdown
        avatarDropdownBtn.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                avatarDropdownBtn.click();
            } else if (e.key === 'Escape') {
                userDropdownMenu.classList.remove('show');
            }
        });
    } else {
        console.debug("[Header] User dropdown elements not found in DOM");
    }

    if (dropdownLogoutBtn) {
        console.debug("[Header] Setting up logout button listener");
        dropdownLogoutBtn.addEventListener('click', async (e) => {
            e.preventDefault();
            console.info("[Header] User logout initiated");
            try {
                await logout();
                console.info("[Header] User logged out successfully");
            } catch (error) {
                console.error("[Header] Logout failed:", error.message || error);
                console.info("[Header] Redirecting to homepage after failed logout");
                window.location.href = '/';
            }
        });
    } else {
        console.debug("[Header] Logout button not found in DOM");
    }

    // Setup notification settings event listeners
    setupNotificationSettingsListeners();
}

/**
 * Initialize notification settings with current preferences
 */
function initializeNotificationSettings() {
    const preferences = NotificationPreferences.get();

    // Set checkbox states
    const soundCheckbox = document.getElementById('dropdown-notification-sound');
    const desktopCheckbox = document.getElementById('dropdown-notification-desktop');
    const visualCheckbox = document.getElementById('dropdown-notification-visual');
    const volumeSlider = document.getElementById('dropdown-notification-volume');
    const volumeDisplay = document.getElementById('dropdown-volume-display');

    if (soundCheckbox) soundCheckbox.checked = preferences.sound;
    if (desktopCheckbox) desktopCheckbox.checked = preferences.desktop;
    if (visualCheckbox) visualCheckbox.checked = preferences.visual;
    if (volumeSlider) {
        volumeSlider.value = Math.round(preferences.volume * 100);
        if (volumeDisplay) {
            volumeDisplay.textContent = `${Math.round(preferences.volume * 100)}%`;
        }
    }

    // Initialize advanced controls
    const dndSchedule = document.getElementById('dropdown-dnd-schedule');
    const priorityOnly = document.getElementById('dropdown-priority-only');
    const soundType = document.getElementById('dropdown-notification-sound-type');

    if (dndSchedule) dndSchedule.value = preferences.dndSchedule || 'off';
    if (priorityOnly) priorityOnly.checked = preferences.priorityOnly || false;
    if (soundType) soundType.value = preferences.soundType || 'default';

    console.debug("[Header] Notification settings initialized:", preferences);
}

/**
 * Setup event listeners for notification settings
 */
function setupNotificationSettingsListeners() {
    // Notification settings toggle
    const notificationsToggle = document.getElementById('dropdown-notifications-toggle');
    const notificationsContent = document.getElementById('dropdown-notifications-content');

    if (notificationsToggle && notificationsContent) {
        notificationsToggle.addEventListener('click', (e) => {
            e.stopPropagation();
            const isExpanded = notificationsContent.style.display !== 'none';
            notificationsContent.style.display = isExpanded ? 'none' : 'block';

            // Update toggle icon
            const icon = notificationsToggle.querySelector('i');
            if (icon) {
                icon.className = isExpanded ? 'fas fa-chevron-down' : 'fas fa-chevron-up';
            }

            // Update ARIA attributes
            notificationsToggle.setAttribute('aria-expanded', !isExpanded);

            console.debug(`[Header] Notification settings ${isExpanded ? 'collapsed' : 'expanded'}`);
        });

        // Keyboard navigation for toggle
        notificationsToggle.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                notificationsToggle.click();
            }
        });
    }

    // Notification preference checkboxes
    const soundCheckbox = document.getElementById('dropdown-notification-sound');
    const desktopCheckbox = document.getElementById('dropdown-notification-desktop');
    const visualCheckbox = document.getElementById('dropdown-notification-visual');
    const volumeSlider = document.getElementById('dropdown-notification-volume');
    const volumeDisplay = document.getElementById('dropdown-volume-display');

    if (soundCheckbox) {
        soundCheckbox.addEventListener('change', (e) => {
            const preferences = NotificationPreferences.get();
            preferences.sound = e.target.checked;
            NotificationPreferences.set(preferences);
            console.debug("[Header] Sound notifications:", e.target.checked);
        });
    }

    if (desktopCheckbox) {
        desktopCheckbox.addEventListener('change', (e) => {
            const preferences = NotificationPreferences.get();
            preferences.desktop = e.target.checked;
            NotificationPreferences.set(preferences);
            console.debug("[Header] Desktop notifications:", e.target.checked);

            // Request permission if enabling desktop notifications
            if (e.target.checked && 'Notification' in window && Notification.permission === 'default') {
                Notification.requestPermission().then(permission => {
                    console.debug("[Header] Notification permission:", permission);
                    if (permission !== 'granted') {
                        e.target.checked = false;
                        preferences.desktop = false;
                        NotificationPreferences.set(preferences);
                    }
                });
            }
        });
    }

    if (visualCheckbox) {
        visualCheckbox.addEventListener('change', (e) => {
            const preferences = NotificationPreferences.get();
            preferences.visual = e.target.checked;
            NotificationPreferences.set(preferences);
            console.debug("[Header] Visual notifications:", e.target.checked);
        });
    }

    if (volumeSlider && volumeDisplay) {
        volumeSlider.addEventListener('input', (e) => {
            const volume = parseInt(e.target.value) / 100;
            const preferences = NotificationPreferences.get();
            preferences.volume = volume;
            NotificationPreferences.set(preferences);
            volumeDisplay.textContent = `${e.target.value}%`;
            console.debug("[Header] Notification volume:", volume);
        });
    }

    // Advanced notification controls
    const dndSchedule = document.getElementById('dropdown-dnd-schedule');
    const priorityOnly = document.getElementById('dropdown-priority-only');
    const soundType = document.getElementById('dropdown-notification-sound-type');

    if (dndSchedule) {
        dndSchedule.addEventListener('change', (e) => {
            const preferences = NotificationPreferences.get();
            preferences.dndSchedule = e.target.value;
            NotificationPreferences.set(preferences);
            console.debug("[Header] DND schedule:", e.target.value);
        });
    }

    if (priorityOnly) {
        priorityOnly.addEventListener('change', (e) => {
            const preferences = NotificationPreferences.get();
            preferences.priorityOnly = e.target.checked;
            NotificationPreferences.set(preferences);
            console.debug("[Header] Priority only:", e.target.checked);
        });
    }

    if (soundType) {
        soundType.addEventListener('change', (e) => {
            const preferences = NotificationPreferences.get();
            preferences.soundType = e.target.value;
            NotificationPreferences.set(preferences);
            console.debug("[Header] Sound type:", e.target.value);
        });
    }
}


