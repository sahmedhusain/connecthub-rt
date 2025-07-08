import { isLoggedIn } from '../utils/api.js';

export function renderSidebar() {
    console.debug("[Sidebar] Rendering sidebar component");
    
    const urlParams = new URLSearchParams(window.location.search);
    const currentTab = urlParams.get('tab') || 'posts';
    
    console.debug(`[Sidebar] Current active tab: ${currentTab}`);
    
    // Check if user is logged in to determine whether to show personal sections
    const userLoggedIn = isLoggedIn();
    
    const sidebar = `
        <aside class="sidebar md3-enhanced">
            <nav class="md3-stagger-container">
                <h3 class="menu-heading md3-enhanced">Menu</h3>
                <ul>
                    <li>
                        <button type="button" class="nav-button md3-enhanced ${currentTab === 'posts' ? 'selected' : ''}" data-tab="posts" data-filter="all" data-tooltip="View all posts">
                            <i class="fas fa-newspaper"></i>
                            <span>Posts</span>
                            <div class="nav-ripple"></div>
                        </button>
                    </li>
                    <li>
                        <button type="button" class="nav-button md3-enhanced ${currentTab === 'tags' ? 'selected' : ''}" data-tab="tags" data-filter="all" data-tooltip="Browse by tags">
                            <i class="fas fa-tags"></i>
                            <span>Tags</span>
                            <div class="nav-ripple"></div>
                        </button>
                    </li>
                </ul>
                ${userLoggedIn ? `
                <h3 class="menu-heading md3-enhanced">Activity centre</h3>
                <ul>
                    <li>
                        <button type="button" class="nav-button md3-enhanced ${currentTab === 'your+posts' ? 'selected' : ''}" data-tab="your+posts" data-filter="newest" data-tooltip="Your posts">
                            <i class="fa-solid fa-grip-lines"></i>
                            <span>Your posts</span>
                            <div class="nav-ripple"></div>
                        </button>
                    </li>
                    <li>
                        <button type="button" class="nav-button md3-enhanced ${currentTab === 'your+replies' ? 'selected' : ''}" data-tab="your+replies" data-filter="newest" data-tooltip="Your comments">
                            <i class="fa-solid fa-reply"></i>
                            <span>Your comments</span>
                            <div class="nav-ripple"></div>
                        </button>
                    </li>
                </ul>
                ` : ''}
            </nav>
            <div class="social-icons md3-enhanced">
                <a href="https://instagram.com/reboot01.coding" target="_blank" class="social-link md3-enhanced" data-tooltip="Follow us on Instagram">
                    <i class="fab fa-instagram"></i>
                    <div class="social-ripple"></div>
                </a>
                <a href="https://linkedin.com/school/reboot-coding-institute" target="_blank" class="social-link md3-enhanced" data-tooltip="Connect on LinkedIn">
                    <i class="fab fa-linkedin"></i>
                    <div class="social-ripple"></div>
                </a>
                <a href="https://reboot01.com" target="_blank" class="social-link md3-enhanced" data-tooltip="Visit our website">
                    <i class="fa-solid fa-globe"></i>
                    <div class="social-ripple"></div>
                </a>
            </div>
            <div class="rights md3-enhanced">
                <p>© 2025 ConnectHub | All rights reserved.</p>
            </div>
        </aside>
    `;

    console.debug("[Sidebar] Sidebar HTML generated, scheduling event listener attachment");
    setTimeout(() => addSidebarEventListeners(), 0);
    return sidebar;
}

function addSidebarEventListeners() {
    console.debug("[Sidebar] Adding sidebar event listeners");
    
    const buttons = document.querySelectorAll('.nav-button');
    if (!buttons.length) {
        console.warn("[Sidebar] No sidebar navigation buttons found in DOM");
        return;
    }
    
    buttons.forEach(button => {
        // Add MD3 ripple effect on click
        button.addEventListener('click', (e) => {
            e.preventDefault();

            // Create ripple effect
            createRippleEffect(button, e);

            const tab = button.dataset.tab;
            const filter = button.dataset.filter;

            console.debug(`[Sidebar] Navigation button clicked:`, { tab, filter });

            // Add loading state
            button.classList.add('md3-loading');

            // Navigate after brief delay for visual feedback
            setTimeout(() => {
                if (window.appRouter) {
                    console.debug(`[Sidebar] Using appRouter to navigate to tab=${tab}, filter=${filter}`);
                    window.appRouter.navigate(`/home?tab=${tab}&filter=${filter}`);
                } else {
                    console.debug(`[Sidebar] appRouter not available, using history API to navigate`);
                    const newUrl = `/home?tab=${tab}&filter=${filter}`;
                    window.history.pushState(null, null, newUrl);
                    window.dispatchEvent(new PopStateEvent('popstate'));
                }

                // Remove loading state
                button.classList.remove('md3-loading');
            }, 150);
        });

        // Add hover effects
        button.addEventListener('mouseenter', () => {
            if (window.PerformanceOptimizer) {
                window.PerformanceOptimizer.forceGPUAcceleration(button);
            }
        });

        button.addEventListener('mouseleave', () => {
            if (window.PerformanceOptimizer) {
                setTimeout(() => {
                    if (!button.matches(':hover, :focus, :active')) {
                        window.PerformanceOptimizer.resetGPUAcceleration(button);
                    }
                }, 300);
            }
        });
    });

    console.debug(`[Sidebar] Added click handlers to ${buttons.length} navigation buttons`);

    // Add MD3 interactions to social links
    const socialLinks = document.querySelectorAll('.social-link');
    socialLinks.forEach(link => {
        link.addEventListener('click', (e) => {
            createRippleEffect(link, e);
        });

        link.addEventListener('mouseenter', () => {
            if (window.PerformanceOptimizer) {
                window.PerformanceOptimizer.forceGPUAcceleration(link);
            }
        });

        link.addEventListener('mouseleave', () => {
            if (window.PerformanceOptimizer) {
                setTimeout(() => {
                    if (!link.matches(':hover, :focus, :active')) {
                        window.PerformanceOptimizer.resetGPUAcceleration(link);
                    }
                }, 300);
            }
        });
    });

    const searchInput = document.querySelector('.search-bar');
    if (searchInput) {
        // Debounce function for search
        function debounce(func, wait) {
            let timeout;
            return function (...args) {
                const context = this;
                clearTimeout(timeout);
                timeout = setTimeout(() => func.apply(context, args), wait);
            };
        }

        // Debounced search function
        const debouncedSearch = debounce((searchTerm) => {
            if (searchTerm.trim()) {
                console.info(`[Sidebar] Debounced search initiated for: "${searchTerm}"`);
                // Search functionality would be implemented here
                // For now, just log the search term
                console.debug(`[Sidebar] Would search for: "${searchTerm}"`);
            }
        }, 300); // 300ms debounce delay

        // Live search with debouncing
        searchInput.addEventListener('input', (e) => {
            const searchTerm = e.target.value.trim();
            if (searchTerm.length >= 2) { // Only search if 2+ characters
                debouncedSearch(searchTerm);
            }
        });

        // Still support Enter key for immediate search
        searchInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                const searchTerm = e.target.value.trim();
                if (searchTerm) {
                    console.info(`[Sidebar] Immediate search initiated for: "${searchTerm}"`);
                    // Search functionality would be implemented here
                    alert(`Search for "${searchTerm}" not implemented yet.`);
                } else {
                    console.debug(`[Sidebar] Empty search term ignored`);
                }
            }
        });
        console.debug("[Sidebar] Search input event listeners attached (with debouncing)");
    } else {
        console.debug("[Sidebar] Search input not found in DOM");
    }
    
    console.info("[Sidebar] All sidebar event listeners attached successfully");
}

/**
 * Highlights the currently active tab in the sidebar
 * @param {string} activeTab - The currently active tab
 */
export function updateActiveSidebarTab(activeTab) {
    console.debug(`[Sidebar] Updating active sidebar tab: ${activeTab}`);
    
    const buttons = document.querySelectorAll('.sidebar .nav-button');
    if (!buttons.length) {
        console.warn("[Sidebar] Cannot update active tab: No sidebar buttons found");
        return;
    }
    
    let tabFound = false;
    
    buttons.forEach(button => {
        const isActive = button.dataset.tab === activeTab;
        button.classList.toggle('selected', isActive);
        if (isActive) tabFound = true;
    });
    
    if (!tabFound) {
        console.warn(`[Sidebar] No matching sidebar tab found for: ${activeTab}`);
    } else {
        console.debug(`[Sidebar] Active tab updated to: ${activeTab}`);
    }
}

/**
 * Create MD3 ripple effect for interactive elements
 * @param {HTMLElement} element - The element to add ripple to
 * @param {Event} event - The click event
 */
function createRippleEffect(element, event) {
    const rippleContainer = element.querySelector('.nav-ripple, .social-ripple');
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