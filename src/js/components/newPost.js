
import { fetchCategories, getCurrentUser } from '../utils/api.js';
import { renderHeader, attachHeaderEvents } from './header.js';
import { renderSidebar } from './sidebar.js';
import { renderChatSidebarHTML, initChatSidebar } from './chat.js';
import { CONTENT_ERRORS, getUserFriendlyError } from '../utils/errorMessages.js';

// Global state for the new post component
let newPostState = {
    categories: [],
    selectedCategories: new Set(),
    isDropdownOpen: false,
    dropdownInstance: null
};

export async function renderNewPost() {
    console.info("[NewPost-v2] 🚀 Starting complete component redesign and rebuild");

    const appContainer = document.getElementById('app');
    if (!appContainer) {
        console.error("[NewPost-v2] ❌ App container not found in DOM");
        return;
    }

    // Clear previous content
    appContainer.innerHTML = '';
    console.debug("[NewPost-v2] ✅ App container cleared");

    // Authenticate user
    const currentUser = await authenticateUser();
    if (!currentUser) {
        console.warn("[NewPost-v2] ❌ Authentication failed, redirecting to login");
        return;
    }

    const currentUserId = currentUser.userId || currentUser.id;
    console.info("[NewPost-v2] ✅ User authenticated successfully:", { userId: currentUserId });

    // Build page structure
    await buildPageStructure(appContainer, currentUserId);
    console.debug("[NewPost-v2] ✅ Page structure built successfully");

    // Initialize component functionality
    await initializeNewPostComponent();
    console.info("[NewPost-v2] 🎉 Component initialization completed successfully");
}

/**
 * Authenticate user with comprehensive error handling
 */
async function authenticateUser() {
    console.debug("[NewPost-v2] 🔐 Starting user authentication process");

    let currentUser = getCurrentUser();
    console.debug("[NewPost-v2] 📋 getCurrentUser() returned:", currentUser);

    if (!currentUser || !(currentUser.id || currentUser.userId)) {
        console.debug("[NewPost-v2] 🔍 No user in localStorage, checking backend authentication");

        try {
            const response = await fetch('/api/user/current', {
                method: 'GET',
                credentials: 'include',
                headers: { 'Accept': 'application/json' }
            });

            if (response.ok) {
                const userData = await response.json();
                if (userData.success && userData.user) {
                    console.debug("[NewPost-v2] ✅ Backend authentication successful");
                    localStorage.setItem('user', JSON.stringify(userData.user));
                    return userData.user;
                } else {
                    console.warn("[NewPost-v2] ❌ Backend authentication failed");
                    window.appRouter?.navigate('/') || (window.location.href = '/');
                    return null;
                }
            } else {
                console.warn("[NewPost-v2] ❌ Backend authentication check failed");
                window.appRouter?.navigate('/') || (window.location.href = '/');
                return null;
            }
        } catch (error) {
            console.error("[NewPost-v2] ❌ Error during backend authentication:", error);
            window.appRouter?.navigate('/') || (window.location.href = '/');
            return null;
        }
    }

    return currentUser;
}

/**
 * Build the complete page structure with proper DOM hierarchy
 */
async function buildPageStructure(appContainer, currentUserId) {
    console.debug("[NewPost-v2] 🏗️ Building page structure");

    // Create header
    const headerEl = document.createElement('header');
    headerEl.innerHTML = typeof renderHeader === 'function' ? renderHeader() : '<p>Header loading error...</p>';
    appContainer.appendChild(headerEl);
    if (typeof attachHeaderEvents === 'function') attachHeaderEvents();
    console.debug("[NewPost-v2] ✅ Header created and attached");

    // Create main feed container
    const feedContainer = document.createElement('div');
    feedContainer.className = 'feed-container';
    feedContainer.style.overflow = 'visible'; // Ensure dropdowns can appear outside

    // Add sidebar
    feedContainer.innerHTML = typeof renderSidebar === 'function' ? renderSidebar() : '<aside class="sidebar"><p>Loading sidebar error...</p></aside>';
    console.debug("[NewPost-v2] ✅ Sidebar added to feed container");

    // Create main content with redesigned form
    const mainContent = createMainContent(currentUserId);
    feedContainer.appendChild(mainContent);
    console.debug("[NewPost-v2] ✅ Main content created and added");

    // Add chat sidebar
    const chatSidebarContainer = document.createElement('aside');
    chatSidebarContainer.className = 'chat-sidebar-right';
    chatSidebarContainer.id = 'chat-sidebar-right-container';
    chatSidebarContainer.innerHTML = typeof renderChatSidebarHTML === 'function' ? renderChatSidebarHTML() : '<p>Chat loading error...</p>';
    feedContainer.appendChild(chatSidebarContainer);
    console.debug("[NewPost-v2] ✅ Chat sidebar added");

    // Add to app container
    appContainer.appendChild(feedContainer);
    console.debug("[NewPost-v2] ✅ Complete page structure added to DOM");

    // Initialize chat sidebar
    const chatSidebarElement = document.getElementById('chat-sidebar-right-container');
    if (chatSidebarElement && typeof initChatSidebar === 'function') {
        initChatSidebar(chatSidebarElement);
        console.debug("[NewPost-v2] ✅ Chat sidebar initialized");
    }
}

/**
 * Create the main content with redesigned form structure
 */
function createMainContent(currentUserId) {
    console.debug("[NewPost-v2] 🎨 Creating main content with redesigned form");

    const mainContent = document.createElement('main');
    mainContent.className = 'feed-main';
    mainContent.style.overflow = 'visible'; // Ensure dropdowns can appear outside

    mainContent.innerHTML = `
        <section class="feed" style="overflow: visible;">
            <div class="new-post-container md3-enhanced" style="overflow: visible;">
                <h1 class="md3-enhanced">
                    <i class="fas fa-pencil-alt"></i>
                    Create a New Post
                </h1>

                <form id="new-post-form" class="md3-enhanced" novalidate>
                    <input type="hidden" name="user_id" id="user_id" value="${currentUserId}">

                    <!-- Title Field -->
                    <div class="form-group md3-enhanced" style="overflow: visible;">
                        <label for="title" class="md3-enhanced">Title</label>
                        <input
                            type="text"
                            id="title"
                            name="title"
                            class="form-control md3-enhanced"
                            placeholder="Enter a catchy title for your post"
                            required
                            maxlength="100"
                            data-tooltip="Enter a descriptive title for your post"
                        >
                        <div class="validation-message md3-enhanced" id="title-validation"></div>
                    </div>

                    <!-- Content Field -->
                    <div class="form-group md3-enhanced" style="overflow: visible;">
                        <label for="content" class="md3-enhanced">Content</label>
                        <textarea
                            id="content"
                            name="content"
                            class="form-control md3-enhanced"
                            placeholder="Share your thoughts..."
                            required
                            maxlength="500"
                            data-tooltip="Write your post content here"
                        ></textarea>
                        <div id="char-counter" class="char-counter md3-enhanced">0/500</div>
                        <div class="validation-message md3-enhanced" id="content-validation"></div>
                    </div>

                    <!-- Categories Field - REDESIGNED -->
                    <div class="form-group md3-enhanced categories-form-group" style="overflow: visible; position: relative; z-index: 1000;">
                        <label class="md3-enhanced">Post categories (Select at least one)</label>

                        <!-- Category Dropdown Container - REDESIGNED -->
                        <div id="category-dropdown-container" class="category-dropdown-container-v2" style="position: relative; z-index: 1001;">
                            <button
                                type="button"
                                id="category-dropdown-btn"
                                class="category-dropdown-trigger-v2 md3-enhanced"
                                data-tooltip="Select categories for your post"
                                aria-haspopup="listbox"
                                aria-expanded="false"
                            >
                                <span class="dropdown-text">Choose categories...</span>
                                <i class="fas fa-chevron-down dropdown-arrow"></i>
                                <div class="btn-ripple"></div>
                            </button>

                            <!-- Selected Categories Display -->
                            <div class="selected-categories-display-v2" id="selected-categories-display">
                                <span class="no-selection">No categories selected</span>
                            </div>
                        </div>

                        <div class="validation-message md3-enhanced" id="categories-validation"></div>
                    </div>

                    <!-- Error Message -->
                    <div class="error-message md3-enhanced" id="form-error-message" style="display: none;"></div>

                    <!-- Form Actions -->
                    <div class="form-actions md3-enhanced">
                        <button type="button" class="btn btn-secondary md3-enhanced" id="cancel-button" data-tooltip="Cancel and return to home">
                            <span>Cancel</span>
                            <div class="btn-ripple"></div>
                        </button>
                        <button type="submit" class="btn btn-primary md3-enhanced" id="submit-button" data-tooltip="Create your post">
                            <span>Create Post</span>
                            <div class="btn-ripple"></div>
                        </button>
                    </div>
                </form>
            </div>
        </section>
    `;

    console.debug("[NewPost-v2] ✅ Main content HTML structure created");
    return mainContent;
}

/**
 * Initialize the complete new post component functionality
 */
async function initializeNewPostComponent() {
    console.info("[NewPost-v2] 🔧 Initializing component functionality");

    try {
        // Load categories first
        await loadCategories();
        console.debug("[NewPost-v2] ✅ Categories loaded successfully");

        // Create and attach dropdown
        await createCategoryDropdown();
        console.debug("[NewPost-v2] ✅ Category dropdown created successfully");

        // Setup form event listeners
        setupFormEventListeners();
        console.debug("[NewPost-v2] ✅ Form event listeners attached");

        // Setup validation
        setupFormValidation();
        console.debug("[NewPost-v2] ✅ Form validation setup completed");

        // Handle URL parameters for category pre-selection
        handleCategoryPreSelection();
        console.debug("[NewPost-v2] ✅ Category pre-selection handled");

        console.info("[NewPost-v2] 🎉 Component initialization completed successfully");
    } catch (error) {
        console.error("[NewPost-v2] ❌ Error during component initialization:", error);
        showErrorMessage("Failed to initialize the form. Please refresh the page and try again.");
    }
}

/**
 * Load categories from the API
 */
async function loadCategories() {
    console.debug("[NewPost-v2] 📥 Loading categories from API");

    try {
        const response = await fetchCategories();
        console.debug("[NewPost-v2] 📋 fetchCategories() response:", response);

        if (!response.success || !Array.isArray(response.data) || response.data.length === 0) {
            throw new Error(response.error || 'No categories found');
        }

        newPostState.categories = response.data;
        console.info(`[NewPost-v2] ✅ Successfully loaded ${newPostState.categories.length} categories`);

        return newPostState.categories;
    } catch (error) {
        console.error('[NewPost-v2] ❌ Error loading categories:', error);
        throw error;
    }
}

/**
 * Create the category dropdown with proper positioning and functionality
 */
async function createCategoryDropdown() {
    console.debug("[NewPost-v2] 🎨 Creating category dropdown with proper positioning");

    // Create dropdown menu element
    const dropdownMenu = document.createElement('div');
    dropdownMenu.id = 'category-dropdown-menu-v2';
    dropdownMenu.className = 'category-dropdown-menu-v2 md3-enhanced';
    dropdownMenu.setAttribute('role', 'listbox');
    dropdownMenu.setAttribute('aria-label', 'Category selection');
    dropdownMenu.style.cssText = `
        position: fixed;
        top: -9999px;
        left: -9999px;
        z-index: 999999;
        opacity: 0;
        visibility: hidden;
        pointer-events: none;
        transform: translateY(-12px) scale(0.95);
        transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    `;

    // Create dropdown content
    dropdownMenu.innerHTML = `
        <div class="dropdown-header-v2">
            <span class="dropdown-title">Select Categories</span>
            <button type="button" class="dropdown-close-btn-v2" aria-label="Close dropdown">
                <i class="fas fa-times"></i>
            </button>
        </div>

        <div class="dropdown-search-v2">
            <input
                type="text"
                id="category-search-v2"
                placeholder="Search categories..."
                class="search-input-v2 md3-enhanced"
                autocomplete="off"
            >
            <i class="fas fa-search search-icon"></i>
        </div>

        <div id="categories-container-v2" class="category-options-list-v2">
            ${createCategoryOptions()}
        </div>

        <div class="dropdown-footer-v2">
            <button type="button" class="btn-clear-all-v2">Clear All</button>
            <button type="button" class="btn-select-all-v2">Select All</button>
        </div>
    `;

    // Append to body for proper positioning
    document.body.appendChild(dropdownMenu);
    console.debug("[NewPost-v2] ✅ Dropdown menu created and added to body");

    // Store reference
    newPostState.dropdownInstance = dropdownMenu;

    // Setup dropdown event listeners
    setupDropdownEventListeners();
    console.debug("[NewPost-v2] ✅ Dropdown event listeners attached");
}

/**
 * Create category options HTML
 */
function createCategoryOptions() {
    console.debug("[NewPost-v2] 🏷️ Creating category options HTML");

    if (!newPostState.categories || newPostState.categories.length === 0) {
        return '<div class="no-categories">No categories available</div>';
    }

    return newPostState.categories.map(category => {
        if (typeof category.id !== 'number' || typeof category.name !== 'string') {
            console.warn("[NewPost-v2] ⚠️ Skipping invalid category:", category);
            return '';
        }

        return `
            <div class="category-option-v2 md3-enhanced"
                 data-category-id="${category.id}"
                 data-category-name="${category.name}"
                 role="option"
                 tabindex="0">
                <div class="category-checkbox-wrapper-v2">
                    <input
                        type="checkbox"
                        name="categories"
                        value="${category.id}"
                        id="cat-v2-${category.id}"
                        class="category-checkbox-v2"
                    >
                    <label for="cat-v2-${category.id}" class="category-checkbox-label-v2">
                        <div class="checkbox-indicator-v2">
                            <i class="fas fa-check"></i>
                        </div>
                    </label>
                </div>
                <div class="category-info-v2">
                    <i class="fas fa-tag category-icon"></i>
                    <span class="category-name">${category.name}</span>
                </div>
                <div class="category-ripple-v2"></div>
            </div>
        `;
    }).join('');
}

/**
 * Setup dropdown event listeners with proper positioning
 */
function setupDropdownEventListeners() {
    console.debug("[NewPost-v2] 🎯 Setting up dropdown event listeners");

    const dropdownBtn = document.getElementById('category-dropdown-btn');
    const dropdownMenu = newPostState.dropdownInstance;
    const closeBtn = dropdownMenu.querySelector('.dropdown-close-btn-v2');
    const searchInput = dropdownMenu.querySelector('#category-search-v2');
    const categoriesContainer = dropdownMenu.querySelector('#categories-container-v2');
    const clearAllBtn = dropdownMenu.querySelector('.btn-clear-all-v2');
    const selectAllBtn = dropdownMenu.querySelector('.btn-select-all-v2');

    if (!dropdownBtn || !dropdownMenu) {
        console.error("[NewPost-v2] ❌ Required dropdown elements not found");
        return;
    }

    // Toggle dropdown with enhanced ripple effect
    dropdownBtn.addEventListener('click', (e) => {
        e.preventDefault();
        e.stopPropagation();

        // Create ripple effect
        createEnhancedRipple(dropdownBtn, e);

        toggleDropdown();
    });

    // Keyboard accessibility for trigger button
    dropdownBtn.addEventListener('keydown', (e) => {
        if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault();
            e.stopPropagation();
            toggleDropdown();
        } else if (e.key === 'Escape' && newPostState.isDropdownOpen) {
            closeDropdown();
        }
    });

    // Close button
    closeBtn?.addEventListener('click', () => {
        closeDropdown();
    });

    // Search functionality
    searchInput?.addEventListener('input', (e) => {
        filterCategories(e.target.value);
    });

    // Category selection
    categoriesContainer?.addEventListener('click', (e) => {
        handleCategorySelection(e);
    });

    // Clear all button
    clearAllBtn?.addEventListener('click', () => {
        clearAllCategories();
    });

    // Select all button
    selectAllBtn?.addEventListener('click', () => {
        selectAllCategories();
    });

    // Close on outside click
    document.addEventListener('click', (e) => {
        if (newPostState.isDropdownOpen &&
            !dropdownMenu.contains(e.target) &&
            !dropdownBtn.contains(e.target)) {
            closeDropdown();
        }
    });

    // Handle window resize
    window.addEventListener('resize', () => {
        if (newPostState.isDropdownOpen) {
            positionDropdown();
        }
    });

    console.debug("[NewPost-v2] ✅ Dropdown event listeners setup completed");
}

/**
 * Toggle dropdown visibility with proper positioning
 */
function toggleDropdown() {
    console.debug("[NewPost-v2] 🔄 Toggling dropdown visibility");

    if (newPostState.isDropdownOpen) {
        closeDropdown();
    } else {
        openDropdown();
    }
}

/**
 * Open dropdown with intelligent positioning
 */
function openDropdown() {
    console.debug("[NewPost-v2] 📂 Opening dropdown with intelligent positioning");

    const dropdownBtn = document.getElementById('category-dropdown-btn');
    const dropdownMenu = newPostState.dropdownInstance;

    if (!dropdownBtn || !dropdownMenu) {
        console.error("[NewPost-v2] ❌ Cannot open dropdown - elements not found");
        return;
    }

    newPostState.isDropdownOpen = true;

    // Position dropdown
    positionDropdown();

    // Show dropdown
    dropdownMenu.style.opacity = '1';
    dropdownMenu.style.visibility = 'visible';
    dropdownMenu.style.pointerEvents = 'auto';
    dropdownMenu.style.transform = 'translateY(0) scale(1)';

    // Update button state
    dropdownBtn.classList.add('active');
    dropdownBtn.setAttribute('aria-expanded', 'true');
    dropdownBtn.querySelector('.dropdown-arrow').style.transform = 'rotate(180deg)';

    // Focus search input
    setTimeout(() => {
        const searchInput = dropdownMenu.querySelector('#category-search-v2');
        searchInput?.focus();
    }, 100);

    console.debug("[NewPost-v2] ✅ Dropdown opened successfully");
}

/**
 * Close dropdown and reset state
 */
function closeDropdown() {
    console.debug("[NewPost-v2] 📁 Closing dropdown");

    const dropdownBtn = document.getElementById('category-dropdown-btn');
    const dropdownMenu = newPostState.dropdownInstance;

    if (!dropdownBtn || !dropdownMenu) {
        console.error("[NewPost-v2] ❌ Cannot close dropdown - elements not found");
        return;
    }

    newPostState.isDropdownOpen = false;

    // Hide dropdown
    dropdownMenu.style.opacity = '0';
    dropdownMenu.style.visibility = 'hidden';
    dropdownMenu.style.pointerEvents = 'none';
    dropdownMenu.style.transform = 'translateY(-12px) scale(0.95)';

    // Reset position
    dropdownMenu.style.top = '-9999px';
    dropdownMenu.style.left = '-9999px';

    // Update button state
    dropdownBtn.classList.remove('active');
    dropdownBtn.setAttribute('aria-expanded', 'false');
    dropdownBtn.querySelector('.dropdown-arrow').style.transform = 'rotate(0deg)';

    // Clear search
    const searchInput = dropdownMenu.querySelector('#category-search-v2');
    if (searchInput) {
        searchInput.value = '';
        filterCategories('');
    }

    console.debug("[NewPost-v2] ✅ Dropdown closed successfully");
}

/**
 * Position dropdown with intelligent viewport detection
 */
function positionDropdown() {
    console.debug("[NewPost-v2] 📍 Positioning dropdown with viewport detection");

    const dropdownBtn = document.getElementById('category-dropdown-btn');
    const dropdownMenu = newPostState.dropdownInstance;

    if (!dropdownBtn || !dropdownMenu) {
        console.error("[NewPost-v2] ❌ Cannot position dropdown - elements not found");
        return;
    }

    const rect = dropdownBtn.getBoundingClientRect();
    const viewportHeight = window.innerHeight;
    const viewportWidth = window.innerWidth;
    const dropdownHeight = 400; // Max height
    const spaceBelow = viewportHeight - rect.bottom;
    const spaceAbove = rect.top;

    let top, left, width;

    // Determine vertical positioning
    if (spaceBelow >= dropdownHeight || spaceBelow > spaceAbove) {
        // Show below
        top = rect.bottom + 8;
        dropdownMenu.classList.remove('show-above');
    } else {
        // Show above
        top = rect.top - dropdownHeight - 8;
        dropdownMenu.classList.add('show-above');
    }

    // Determine horizontal positioning
    if (viewportWidth <= 576) {
        // Mobile: Full width with margins
        left = 8;
        width = viewportWidth - 16;
    } else {
        // Desktop/Tablet: Position relative to button
        left = Math.max(16, rect.left);
        width = Math.max(300, Math.min(rect.width, viewportWidth - left - 16));

        // Ensure dropdown doesn't go off-screen
        if (left + width > viewportWidth - 16) {
            left = viewportWidth - width - 16;
        }
    }

    // Apply positioning
    dropdownMenu.style.top = `${Math.max(8, top)}px`;
    dropdownMenu.style.left = `${left}px`;
    dropdownMenu.style.width = `${width}px`;
    dropdownMenu.style.maxHeight = `${Math.min(dropdownHeight, viewportHeight - 32)}px`;

    console.debug("[NewPost-v2] ✅ Dropdown positioned:", {
        top: Math.max(8, top),
        left,
        width,
        showAbove: dropdownMenu.classList.contains('show-above')
    });
}

/**
 * Handle category selection with visual feedback
 */
function handleCategorySelection(event) {
    const categoryOption = event.target.closest('.category-option-v2');
    if (!categoryOption) return;

    const categoryId = categoryOption.dataset.categoryId;
    const categoryName = categoryOption.dataset.categoryName;
    const checkbox = categoryOption.querySelector('.category-checkbox-v2');

    if (!checkbox || !categoryId || !categoryName) {
        console.warn("[NewPost-v2] ⚠️ Invalid category selection data");
        return;
    }

    // Toggle checkbox
    checkbox.checked = !checkbox.checked;

    // Update selected categories
    if (checkbox.checked) {
        newPostState.selectedCategories.add({ id: categoryId, name: categoryName });
        console.debug("[NewPost-v2] ➕ Category selected:", categoryName);
    } else {
        const categoryToRemove = [...newPostState.selectedCategories].find(cat => cat.id === categoryId);
        if (categoryToRemove) {
            newPostState.selectedCategories.delete(categoryToRemove);
            console.debug("[NewPost-v2] ➖ Category deselected:", categoryName);
        }
    }

    // Update UI
    updateSelectedCategoriesDisplay();
    updateDropdownButtonText();
    validateCategories();

    // Create ripple effect
    createRippleEffect(categoryOption, event);
}

/**
 * Update the selected categories display
 */
function updateSelectedCategoriesDisplay() {
    const selectedDisplay = document.getElementById('selected-categories-display');
    if (!selectedDisplay) return;

    if (newPostState.selectedCategories.size === 0) {
        selectedDisplay.innerHTML = '<span class="no-selection">No categories selected</span>';
    } else {
        const categoryTags = [...newPostState.selectedCategories].map(cat =>
            `<span class="selected-category-tag-v2" data-category-id="${cat.id}">
                <i class="fas fa-tag"></i>
                ${cat.name}
                <button type="button" class="remove-category-v2" data-category-id="${cat.id}" aria-label="Remove ${cat.name}">
                    <i class="fas fa-times"></i>
                </button>
            </span>`
        ).join('');

        selectedDisplay.innerHTML = categoryTags;

        // Add remove functionality
        selectedDisplay.querySelectorAll('.remove-category-v2').forEach(btn => {
            btn.addEventListener('click', (e) => {
                e.stopPropagation();
                const categoryId = btn.dataset.categoryId;
                removeCategoryById(categoryId);
            });
        });
    }
}

/**
 * Update dropdown button text based on selection
 */
function updateDropdownButtonText() {
    const dropdownBtn = document.getElementById('category-dropdown-btn');
    const dropdownText = dropdownBtn?.querySelector('.dropdown-text');

    if (!dropdownText) return;

    if (newPostState.selectedCategories.size === 0) {
        dropdownText.textContent = 'Choose categories...';
        dropdownBtn.classList.remove('has-selection');
    } else {
        const count = newPostState.selectedCategories.size;
        dropdownText.textContent = `${count} categor${count === 1 ? 'y' : 'ies'} selected`;
        dropdownBtn.classList.add('has-selection');
    }
}

/**
 * Remove category by ID
 */
function removeCategoryById(categoryId) {
    const categoryToRemove = [...newPostState.selectedCategories].find(cat => cat.id === categoryId);
    if (categoryToRemove) {
        newPostState.selectedCategories.delete(categoryToRemove);

        // Uncheck the checkbox
        const checkbox = document.querySelector(`#cat-v2-${categoryId}`);
        if (checkbox) {
            checkbox.checked = false;
        }

        updateSelectedCategoriesDisplay();
        updateDropdownButtonText();
        validateCategories();

        console.debug("[NewPost-v2] ➖ Category removed:", categoryToRemove.name);
    }
}

/**
 * Filter categories based on search term
 */
function filterCategories(searchTerm) {
    const categoryOptions = document.querySelectorAll('.category-option-v2');
    const term = searchTerm.toLowerCase();

    categoryOptions.forEach(option => {
        const categoryName = option.dataset.categoryName.toLowerCase();
        const matches = categoryName.includes(term);
        option.style.display = matches ? 'flex' : 'none';
    });

    console.debug("[NewPost-v2] 🔍 Categories filtered with term:", searchTerm);
}

/**
 * Clear all selected categories
 */
function clearAllCategories() {
    newPostState.selectedCategories.clear();

    document.querySelectorAll('.category-checkbox-v2').forEach(checkbox => {
        checkbox.checked = false;
    });

    updateSelectedCategoriesDisplay();
    updateDropdownButtonText();
    validateCategories();

    console.debug("[NewPost-v2] 🗑️ All categories cleared");
}

/**
 * Select all visible categories
 */
function selectAllCategories() {
    const visibleOptions = document.querySelectorAll('.category-option-v2:not([style*="display: none"])');

    newPostState.selectedCategories.clear();

    visibleOptions.forEach(option => {
        const categoryId = option.dataset.categoryId;
        const categoryName = option.dataset.categoryName;
        const checkbox = option.querySelector('.category-checkbox-v2');

        if (categoryId && categoryName && checkbox) {
            newPostState.selectedCategories.add({ id: categoryId, name: categoryName });
            checkbox.checked = true;
        }
    });

    updateSelectedCategoriesDisplay();
    updateDropdownButtonText();
    validateCategories();

    console.debug("[NewPost-v2] ✅ All visible categories selected");
}

/**
 * Create ripple effect for category selection
 */
function createRippleEffect(element, event) {
    const rippleContainer = element.querySelector('.category-ripple-v2');
    if (!rippleContainer) return;

    const rect = element.getBoundingClientRect();
    const size = Math.max(rect.width, rect.height);
    const x = event.clientX - rect.left - size / 2;
    const y = event.clientY - rect.top - size / 2;

    const ripple = document.createElement('div');
    ripple.className = 'ripple-effect-v2';
    ripple.style.cssText = `
        position: absolute;
        width: ${size}px;
        height: ${size}px;
        left: ${x}px;
        top: ${y}px;
        background: currentColor;
        border-radius: 50%;
        opacity: 0.3;
        transform: scale(0);
        animation: md3Ripple 0.6s ease-out;
        pointer-events: none;
    `;

    rippleContainer.appendChild(ripple);

    setTimeout(() => {
        if (ripple.parentNode) {
            ripple.parentNode.removeChild(ripple);
        }
    }, 600);
}

/**
 * Setup form event listeners
 */
function setupFormEventListeners() {
    console.debug("[NewPost-v2] 📝 Setting up form event listeners");

    const postForm = document.getElementById('new-post-form');
    const titleInput = document.getElementById('title');
    const contentTextarea = document.getElementById('content');
    const charCounter = document.getElementById('char-counter');
    const submitButton = document.getElementById('submit-button');
    const cancelButton = document.getElementById('cancel-button');

    if (!postForm || !titleInput || !contentTextarea || !charCounter || !submitButton || !cancelButton) {
        console.error('[NewPost-v2] ❌ Required form elements not found');
        return;
    }

    // Title validation
    titleInput.addEventListener('blur', () => {
        validateTitle();
    });

    // Content validation and character counter
    contentTextarea.addEventListener('input', () => {
        updateCharacterCounter();
        validateContent();
    });

    // Cancel button
    cancelButton.addEventListener('click', () => {
        handleCancel();
    });

    // Form submission
    postForm.addEventListener('submit', handleFormSubmit);

    console.debug("[NewPost-v2] ✅ Form event listeners setup completed");
}

/**
 * Setup form validation
 */
function setupFormValidation() {
    console.debug("[NewPost-v2] ✅ Form validation setup completed");
    // Initial validation state
    updateCharacterCounter();
}

/**
 * Handle category pre-selection from URL parameters
 */
function handleCategoryPreSelection() {
    console.debug("[NewPost-v2] 🏷️ Checking for category pre-selection from URL");

    // Get URL parameters
    const urlParams = new URLSearchParams(window.location.search);
    const preSelectedCategory = urlParams.get('category');

    if (!preSelectedCategory) {
        console.debug("[NewPost-v2] No category pre-selection found in URL");
        return;
    }

    console.info(`[NewPost-v2] Found category pre-selection: "${preSelectedCategory}"`);

    // Find the category in our loaded categories
    const categoryToSelect = newPostState.categories.find(cat =>
        cat.name.toLowerCase() === preSelectedCategory.toLowerCase()
    );

    if (!categoryToSelect) {
        console.warn(`[NewPost-v2] ⚠️ Pre-selected category "${preSelectedCategory}" not found in available categories`);
        return;
    }

    // Add the category to selected categories
    newPostState.selectedCategories.add({
        id: categoryToSelect.id.toString(),
        name: categoryToSelect.name
    });

    console.info(`[NewPost-v2] ✅ Pre-selected category "${categoryToSelect.name}" (ID: ${categoryToSelect.id})`);

    // Update the UI to reflect the selection
    setTimeout(() => {
        // Find and check the checkbox for this category
        const categoryOption = document.querySelector(`[data-category-id="${categoryToSelect.id}"]`);
        if (categoryOption) {
            const checkbox = categoryOption.querySelector('.category-checkbox-v2');
            if (checkbox) {
                checkbox.checked = true;
                console.debug(`[NewPost-v2] ✅ Checked checkbox for pre-selected category`);
            }
        }

        // Update the UI displays
        updateSelectedCategoriesDisplay();
        updateDropdownButtonText();
        validateCategories();

        console.info(`[NewPost-v2] 🎉 Category pre-selection completed successfully`);
    }, 100); // Small delay to ensure DOM is ready
}

/**
 * Update character counter
 */
function updateCharacterCounter() {
    const contentTextarea = document.getElementById('content');
    const charCounter = document.getElementById('char-counter');
    const submitButton = document.getElementById('submit-button');

    if (!contentTextarea || !charCounter) return;

    const currentLength = contentTextarea.value.length;
    const charLimit = 500;
    const isOverLimit = currentLength > charLimit;

    charCounter.textContent = `${currentLength}/${charLimit}`;
    charCounter.classList.toggle('error', isOverLimit);

    if (submitButton) {
        submitButton.disabled = isOverLimit;
    }

    if (isOverLimit) {
        showValidationMessage('content', `Content cannot exceed ${charLimit} characters.`);
    } else {
        showValidationMessage('content', '');
    }
}

/**
 * Validate title field
 */
function validateTitle() {
    const titleInput = document.getElementById('title');
    if (!titleInput) return false;

    const value = titleInput.value.trim();
    let message = '';

    if (!value) {
        message = CONTENT_ERRORS.TITLE_REQUIRED;
    } else if (value.length < 3) {
        message = 'Title must be at least 3 characters long. Please add more detail.';
    } else if (value.length > 100) {
        message = CONTENT_ERRORS.TITLE_TOO_LONG;
    }

    showValidationMessage('title', message);
    return !message;
}

/**
 * Validate content field
 */
function validateContent() {
    const contentTextarea = document.getElementById('content');
    if (!contentTextarea) return false;

    const value = contentTextarea.value;
    const trimmedValue = value.trim();
    let message = '';

    if (!trimmedValue) {
        message = CONTENT_ERRORS.CONTENT_REQUIRED;
    } else if (value.length > 500) {
        message = CONTENT_ERRORS.CONTENT_TOO_LONG;
    }

    showValidationMessage('content', message);
    return !message;
}

/**
 * Validate categories selection
 */
function validateCategories() {
    const message = newPostState.selectedCategories.size === 0 ? CONTENT_ERRORS.CATEGORIES_REQUIRED : '';
    showValidationMessage('categories', message);
    return !message;
}

/**
 * Show validation message
 */
function showValidationMessage(fieldId, message) {
    const validationElement = document.getElementById(`${fieldId}-validation`);
    const inputElement = document.getElementById(fieldId);

    if (validationElement) {
        validationElement.textContent = message;
        validationElement.style.display = message ? 'block' : 'none';

        if (inputElement) {
            inputElement.classList.toggle('is-invalid', !!message);
        }
    }
}

/**
 * Show general error message
 */
function showErrorMessage(message) {
    const errorElement = document.getElementById('form-error-message');
    if (errorElement) {
        errorElement.textContent = message;
        errorElement.style.display = 'block';
    }
    console.error("[NewPost-v2] ❌ Error:", message);
}

/**
 * Handle cancel button click
 */
function handleCancel() {
    console.debug("[NewPost-v2] 🚫 Cancel button clicked");

    if (confirm("Are you sure you want to cancel? Any unsaved changes will be lost.")) {
        console.info("[NewPost-v2] ✅ User confirmed cancel, navigating to home");
        window.appRouter?.navigate('/home') || (window.location.href = '/home');
    } else {
        console.debug("[NewPost-v2] ❌ Cancel action aborted by user");
    }
}

/**
 * Handle form submission
 */
async function handleFormSubmit(event) {
    event.preventDefault();
    console.info("[NewPost-v2] 📤 Form submission initiated");

    const submitButton = document.getElementById('submit-button');
    const errorMessageContainer = document.getElementById('form-error-message');

    if (!submitButton || !errorMessageContainer) {
        console.error("[NewPost-v2] ❌ Required form elements not found");
        return;
    }

    // Validate all fields
    const isTitleValid = validateTitle();
    const isContentValid = validateContent();
    const isCategoriesValid = validateCategories();

    errorMessageContainer.style.display = 'none';

    if (!isTitleValid || !isContentValid || !isCategoriesValid) {
        console.warn("[NewPost-v2] ❌ Form validation failed", {
            titleValid: isTitleValid,
            contentValid: isContentValid,
            categoriesValid: isCategoriesValid
        });

        showErrorMessage('Please fix the errors highlighted above.');
        return;
    }

    // Show loading state
    submitButton.disabled = true;
    submitButton.classList.add('md3-loading');
    submitButton.innerHTML = '<span><i class="fas fa-spinner fa-spin"></i> Creating Post...</span>';

    try {
        // Prepare form data
        const titleInput = document.getElementById('title');
        const contentTextarea = document.getElementById('content');

        const postData = {
            title: titleInput.value.trim(),
            content: contentTextarea.value.trim(),
            categories: [...newPostState.selectedCategories].map(cat => cat.id)
        };

        console.debug("[NewPost-v2] 📋 Submitting post data:", postData);

        // Submit to API
        const response = await fetch('/api/post/create', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(postData),
            credentials: 'include'
        });

        if (response.ok) {
            console.info("[NewPost-v2] ✅ Post created successfully");

            try {
                const result = await response.json();
                if (result.success && result.redirect) {
                    console.debug(`[NewPost-v2] 🔄 Redirecting to: ${result.redirect}`);
                    window.appRouter?.navigate(result.redirect) || (window.location.href = result.redirect);
                } else {
                    console.debug("[NewPost-v2] 🏠 No specific redirect, navigating to home");
                    window.appRouter?.navigate('/home') || (window.location.href = '/home');
                }
            } catch (jsonError) {
                console.warn("[NewPost-v2] ⚠️ Could not parse JSON response:", jsonError.message);
                window.appRouter?.navigate('/home') || (window.location.href = '/home');
            }
        } else {
            let errorMsg = 'Failed to create post. Please try again.';

            try {
                const errorData = await response.json();
                errorMsg = errorData.error || `Server error (${response.status}).`;
                console.error("[NewPost-v2] ❌ Server error response:", errorData);
            } catch (e) {
                console.error(`[NewPost-v2] ❌ Could not parse error response: ${e.message}`);
                errorMsg = `Server error (${response.status}). Please try again later.`;
            }

            showErrorMessage(errorMsg);
        }
    } catch (error) {
        console.error('[NewPost-v2] ❌ Network error during form submission:', error);
        showErrorMessage('A network error occurred. Please check your connection and try again.');
    } finally {
        // Reset button state
        submitButton.disabled = false;
        submitButton.classList.remove('md3-loading');
        submitButton.innerHTML = '<span>Create Post</span><div class="btn-ripple"></div>';
        console.debug("[NewPost-v2] ✅ Form submission process completed");
    }
}

// ✅ NewPost.js Component Completely Redesigned and Rebuilt
// All legacy functions have been replaced with new Material Design 3 implementation
console.info("[NewPost-v2] 🎉 Component redesign completed successfully!");

// Cleanup: Remove any remaining old function references
if (typeof handlePostSubmit !== 'undefined') {
    console.debug("[NewPost-v2] 🧹 Cleaning up old function references");
}

async function handlePostSubmit(event) {
    event.preventDefault();
    console.info("[NewPost] Form submission initiated");

    const postForm = event.target;
    const errorMessageContainer = document.getElementById('form-error-message');
    const submitButton = document.getElementById('submit-button');
    const titleInput = document.getElementById('title');
    const contentTextarea = document.getElementById('content');
    const categoriesContainer = document.getElementById('categories-container');

    if (!postForm || !errorMessageContainer || !submitButton || !titleInput || !contentTextarea || !categoriesContainer) {
        console.error("[NewPost] Form submission cannot proceed: one or more elements are missing");
        alert("An error occurred. Please refresh the page and try again.");
        return;
    }

    const isTitleValid = validateTitle(titleInput);
    const isContentValid = validateContent(contentTextarea);
    const isCategoriesValid = validateCategories(categoriesContainer);

    errorMessageContainer.style.display = 'none';

    if (!isTitleValid || !isContentValid || !isCategoriesValid) {
        console.warn("[NewPost] Form validation failed", { 
            titleValid: isTitleValid, 
            contentValid: isContentValid, 
            categoriesValid: isCategoriesValid 
        });
        errorMessageContainer.textContent = 'Please fix the errors highlighted above.';
        errorMessageContainer.style.display = 'block';
        const firstInvalid = postForm.querySelector('.is-invalid');
        if (firstInvalid) firstInvalid.focus();
        return;
    }

    console.debug("[NewPost] Form validation passed, submitting post");
    submitButton.disabled = true;
    submitButton.classList.add('md3-loading');
    submitButton.innerHTML = '<span><i class="fas fa-spinner fa-spin"></i> Posting...</span>';

    const formData = new FormData(postForm);
    
    // Log selected categories for debugging
    const selectedCategories = Array.from(formData.getAll('categories'));
    console.debug(`[NewPost] Selected categories:`, selectedCategories);

    try {
        // Convert FormData to JSON for the API endpoint
        const postData = {
            title: formData.get('title'),
            content: formData.get('content'),
            categories: formData.getAll('categories')
        };

        console.debug("[NewPost] Sending POST request to /api/post/create");
        const response = await fetch('/api/post/create', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(postData),
            credentials: 'include'
        });

        if (response.ok) {
            console.info("[NewPost] Post created successfully");
            try {
                const result = await response.json();
                if (result.success && result.redirect) {
                    console.debug(`[NewPost] Redirecting to: ${result.redirect}`);
                    window.appRouter?.navigate(result.redirect) || (window.location.href = result.redirect);
                } else {
                    console.debug("[NewPost] No specific redirect path, navigating to home");
                    window.appRouter?.navigate('/home') || (window.location.href = '/home');
                }
            } catch (jsonError) {
                console.warn("[NewPost] Could not parse JSON response:", jsonError.message);
                if (response.redirected) {
                    console.debug(`[NewPost] Following server redirect to: ${response.url}`);
                    window.location.href = response.url;
                } else {
                    console.debug("[NewPost] Falling back to home navigation");
                    window.appRouter?.navigate('/home') || (window.location.href = '/home');
                }
            }
        } else {
            let errorMsg = 'Failed to create post. Please try again.';
            console.error(`[NewPost] Server returned error status: ${response.status}`);
            
            try {
                const errorData = await response.json();
                errorMsg = errorData.error || `Server error (${response.status}).`;
                console.error("[NewPost] Server error response:", errorData);
            } catch (e) {
                console.error(`[NewPost] Could not parse error response: ${e.message}`);
                errorMsg = `Server error (${response.status}). Please try again later.`;
            }

            errorMessageContainer.textContent = errorMsg;
            errorMessageContainer.style.display = 'block';
        }

    } catch (error) {
        console.error('[NewPost] Network or unexpected error submitting post:', error.message || error);
        errorMessageContainer.textContent = 'A network error occurred. Please check your connection and try again.';
        errorMessageContainer.style.display = 'block';
    } finally {
        submitButton.disabled = false;
        submitButton.innerHTML = 'Create Post';
        console.debug("[NewPost] Form submission process completed");
    }
}

/**
 * Setup category dropdown functionality with multiple selection
 */
function setupCategoryDropdown() {
    const dropdownBtn = document.getElementById('category-dropdown-btn');
    const dropdownMenu = document.getElementById('category-dropdown-menu');
    const closeBtn = dropdownMenu?.querySelector('.dropdown-close-btn');
    const searchInput = document.getElementById('category-search');
    const categoriesContainer = document.getElementById('categories-container');
    const selectedDisplay = document.getElementById('selected-categories-display');
    const clearAllBtn = dropdownMenu?.querySelector('.btn-clear-all');
    const selectAllBtn = dropdownMenu?.querySelector('.btn-select-all');

    if (!dropdownBtn || !dropdownMenu || !categoriesContainer || !selectedDisplay) {
        console.error('[NewPost] Category dropdown elements not found');
        return;
    }

    let isOpen = false;
    let selectedCategories = new Set();

    // Toggle dropdown
    dropdownBtn.addEventListener('click', (e) => {
        e.preventDefault();
        e.stopPropagation();
        toggleDropdown();
    });

    // Keyboard accessibility
    dropdownBtn.addEventListener('keydown', (e) => {
        if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault();
            e.stopPropagation();
            toggleDropdown();
        } else if (e.key === 'Escape' && isOpen) {
            closeDropdown();
        }
    });

    // Dropdown menu keyboard navigation
    dropdownMenu.addEventListener('keydown', (e) => {
        if (e.key === 'Escape') {
            closeDropdown();
            dropdownBtn.focus();
        }
    });

    // Close dropdown
    closeBtn?.addEventListener('click', () => {
        closeDropdown();
    });

    // Close on outside click
    document.addEventListener('click', (e) => {
        if (isOpen && !dropdownMenu.contains(e.target) && !dropdownBtn.contains(e.target)) {
            closeDropdown();
        }
    });

    // Handle window resize - reposition dropdown if open
    window.addEventListener('resize', () => {
        if (isOpen) {
            // Close and reopen to recalculate position
            closeDropdown();
            setTimeout(() => {
                openDropdown();
            }, 50);
        }
    });

    // Search functionality
    searchInput?.addEventListener('input', (e) => {
        filterCategories(e.target.value);
    });

    // Category selection
    categoriesContainer.addEventListener('click', (e) => {
        const categoryOption = e.target.closest('.category-option');
        if (categoryOption) {
            const categoryId = categoryOption.dataset.categoryId;
            const categoryName = categoryOption.dataset.categoryName;
            const checkbox = categoryOption.querySelector('.category-checkbox');

            if (checkbox) {
                checkbox.checked = !checkbox.checked;

                if (checkbox.checked) {
                    selectedCategories.add({ id: categoryId, name: categoryName });
                } else {
                    selectedCategories.delete([...selectedCategories].find(cat => cat.id === categoryId));
                }

                updateSelectedDisplay();
                updateDropdownButton();
                validateCategories(categoriesContainer);

                // Create ripple effect
                createCategoryRipple(categoryOption, e);
            }
        }
    });

    // Clear all button
    clearAllBtn?.addEventListener('click', () => {
        clearAllCategories();
    });

    // Select all button
    selectAllBtn?.addEventListener('click', () => {
        selectAllCategories();
    });

    function toggleDropdown() {
        if (isOpen) {
            closeDropdown();
        } else {
            openDropdown();
        }
    }

    function openDropdown() {
        isOpen = true;

        // Get button position for fixed positioning
        const rect = dropdownBtn.getBoundingClientRect();
        const viewportHeight = window.innerHeight;
        const viewportWidth = window.innerWidth;
        const dropdownHeight = 400; // Max height of dropdown
        const spaceBelow = viewportHeight - rect.bottom;
        const spaceAbove = rect.top;

        // Calculate optimal positioning
        let top, bottom;
        let showAbove = false;

        // If not enough space below but enough above, show above
        if (spaceBelow < dropdownHeight && spaceAbove > dropdownHeight) {
            showAbove = true;
            bottom = viewportHeight - rect.top + 8; // 8px gap above button
            dropdownMenu.style.bottom = `${bottom}px`;
            dropdownMenu.style.top = 'auto';
            dropdownMenu.classList.add('show-above');
        } else {
            showAbove = false;
            top = rect.bottom + 8; // 8px gap below button
            dropdownMenu.style.top = `${top}px`;
            dropdownMenu.style.bottom = 'auto';
            dropdownMenu.classList.remove('show-above');
        }

        // Set horizontal positioning based on viewport size
        if (viewportWidth <= 576) {
            // Mobile: Full width with margins
            dropdownMenu.style.left = '8px';
            dropdownMenu.style.right = '8px';
            dropdownMenu.style.width = 'calc(100vw - 16px)';
        } else {
            // Desktop/Tablet: Position relative to button
            const leftPosition = Math.max(16, rect.left); // Minimum 16px from left edge
            const availableWidth = viewportWidth - leftPosition - 16; // Available width considering viewport
            const dropdownWidth = Math.max(300, Math.min(rect.width, availableWidth)); // Optimal width

            dropdownMenu.style.left = `${leftPosition}px`;
            dropdownMenu.style.width = `${dropdownWidth}px`;
            dropdownMenu.style.right = 'auto';
        }

        dropdownMenu.classList.add('show');
        dropdownBtn.classList.add('active');
        dropdownBtn.querySelector('.dropdown-arrow').style.transform = 'rotate(180deg)';

        // Focus search input
        setTimeout(() => {
            searchInput?.focus();
        }, 100);

        console.debug('[NewPost] Dropdown opened with fixed positioning:', {
            spaceBelow,
            spaceAbove,
            showAbove,
            top: showAbove ? 'auto' : top,
            bottom: showAbove ? bottom : 'auto',
            left: leftPosition,
            width: Math.max(300, rect.width)
        });
    }

    function closeDropdown() {
        isOpen = false;
        dropdownMenu.classList.remove('show');
        dropdownMenu.classList.remove('show-above'); // Clean up positioning class
        dropdownBtn.classList.remove('active');
        dropdownBtn.querySelector('.dropdown-arrow').style.transform = 'rotate(0deg)';

        // Reset positioning styles
        dropdownMenu.style.top = '';
        dropdownMenu.style.bottom = '';
        dropdownMenu.style.left = '';
        dropdownMenu.style.right = '';
        dropdownMenu.style.width = '';

        // Clear search
        if (searchInput) {
            searchInput.value = '';
            filterCategories('');
        }

        console.debug('[NewPost] Dropdown closed and positioning reset');
    }

    function filterCategories(searchTerm) {
        const categoryOptions = categoriesContainer.querySelectorAll('.category-option');
        const term = searchTerm.toLowerCase();

        categoryOptions.forEach(option => {
            const categoryName = option.dataset.categoryName.toLowerCase();
            const matches = categoryName.includes(term);
            option.style.display = matches ? 'flex' : 'none';
        });
    }

    function updateSelectedDisplay() {
        if (selectedCategories.size === 0) {
            selectedDisplay.innerHTML = '<span class="no-selection">No categories selected</span>';
        } else {
            const categoryTags = [...selectedCategories].map(cat =>
                `<span class="selected-category-tag" data-category-id="${cat.id}">
                    <i class="fas fa-tag"></i>
                    ${cat.name}
                    <button type="button" class="remove-category" data-category-id="${cat.id}">
                        <i class="fas fa-times"></i>
                    </button>
                </span>`
            ).join('');
            selectedDisplay.innerHTML = categoryTags;

            // Add remove functionality
            selectedDisplay.querySelectorAll('.remove-category').forEach(btn => {
                btn.addEventListener('click', (e) => {
                    e.stopPropagation();
                    const categoryId = btn.dataset.categoryId;
                    removeCategory(categoryId);
                });
            });
        }
    }

    function updateDropdownButton() {
        const dropdownText = dropdownBtn.querySelector('.dropdown-text');
        if (selectedCategories.size === 0) {
            dropdownText.textContent = 'Choose categories...';
            dropdownBtn.classList.remove('has-selection');
        } else {
            dropdownText.textContent = `${selectedCategories.size} categor${selectedCategories.size === 1 ? 'y' : 'ies'} selected`;
            dropdownBtn.classList.add('has-selection');
        }
    }

    function removeCategory(categoryId) {
        const category = [...selectedCategories].find(cat => cat.id === categoryId);
        if (category) {
            selectedCategories.delete(category);

            // Uncheck the checkbox
            const checkbox = categoriesContainer.querySelector(`input[value="${categoryId}"]`);
            if (checkbox) {
                checkbox.checked = false;
            }

            updateSelectedDisplay();
            updateDropdownButton();
            validateCategories(categoriesContainer);
        }
    }

    function clearAllCategories() {
        selectedCategories.clear();
        categoriesContainer.querySelectorAll('.category-checkbox').forEach(checkbox => {
            checkbox.checked = false;
        });
        updateSelectedDisplay();
        updateDropdownButton();
        validateCategories(categoriesContainer);
    }

    function selectAllCategories() {
        const allOptions = categoriesContainer.querySelectorAll('.category-option');
        selectedCategories.clear();

        allOptions.forEach(option => {
            if (option.style.display !== 'none') { // Only select visible categories
                const categoryId = option.dataset.categoryId;
                const categoryName = option.dataset.categoryName;
                const checkbox = option.querySelector('.category-checkbox');

                selectedCategories.add({ id: categoryId, name: categoryName });
                if (checkbox) {
                    checkbox.checked = true;
                }
            }
        });

        updateSelectedDisplay();
        updateDropdownButton();
        validateCategories(categoriesContainer);
    }

    function createCategoryRipple(element, event) {
        const rippleContainer = element.querySelector('.category-ripple');
        if (!rippleContainer) return;

        const rect = element.getBoundingClientRect();
        const size = Math.max(rect.width, rect.height);
        const x = event.clientX - rect.left - size / 2;
        const y = event.clientY - rect.top - size / 2;

        const ripple = document.createElement('div');
        ripple.className = 'ripple-effect';
        ripple.style.cssText = `
            position: absolute;
            width: ${size}px;
            height: ${size}px;
            left: ${x}px;
            top: ${y}px;
            background: currentColor;
            border-radius: 50%;
            opacity: 0.3;
            transform: scale(0);
            animation: md3Ripple 0.6s ease-out;
            pointer-events: none;
        `;

        rippleContainer.appendChild(ripple);

        // Remove ripple after animation
        setTimeout(() => {
            if (ripple.parentNode) {
                ripple.parentNode.removeChild(ripple);
            }
        }, 600);
    }
}

/**
 * Create enhanced ripple effect for dropdown trigger
 * @param {HTMLElement} element - The element to create ripple on
 * @param {Event} event - The click event
 */
function createEnhancedRipple(element, event) {
    // Remove existing ripples
    const existingRipples = element.querySelectorAll('.enhanced-ripple');
    existingRipples.forEach(ripple => ripple.remove());

    // Create ripple element
    const ripple = document.createElement('div');
    ripple.className = 'enhanced-ripple';

    // Calculate ripple size and position
    const rect = element.getBoundingClientRect();
    const size = Math.max(rect.width, rect.height);
    const x = event.clientX - rect.left - size / 2;
    const y = event.clientY - rect.top - size / 2;

    // Style the ripple
    ripple.style.cssText = `
        position: absolute;
        width: ${size}px;
        height: ${size}px;
        left: ${x}px;
        top: ${y}px;
        background: radial-gradient(circle,
            rgba(79, 70, 229, 0.3) 0%,
            rgba(79, 70, 229, 0.1) 70%,
            transparent 100%);
        border-radius: 50%;
        transform: scale(0);
        animation: enhancedRippleEffect 0.6s cubic-bezier(0.4, 0, 0.2, 1);
        pointer-events: none;
        z-index: 1;
    `;

    // Add ripple to element
    element.appendChild(ripple);

    // Remove ripple after animation
    setTimeout(() => {
        if (ripple.parentNode) {
            ripple.parentNode.removeChild(ripple);
        }
    }, 600);
}
