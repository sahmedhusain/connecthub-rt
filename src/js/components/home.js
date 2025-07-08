import { fetchPosts, fetchCategories } from '../utils/api.js';
import { renderHeader, attachHeaderEvents } from './header.js';
import { renderSidebar } from './sidebar.js';
import { renderChatSidebarHTML, initChatSidebar } from './chat.js';
import { handleApiError, createInlineError } from './error.js';
import '../utils/scrollPhysics.js';

let selectedTab = 'posts';
let selectedFilter = 'all';
let categories = [];
let isFilterClick = false; // Add this flag at the top of the file with your other variables
let isExternalNavigation = false; // Flag to track navigation from external sources (like post category tags)

const defaultAvatarPath = '/static/assets/default-avatar.png';

export async function renderHome() {
    const appContainer = document.getElementById('app');
    if (!appContainer) {
        console.error("[Home] App container not found in DOM");
        return;
    }

    appContainer.innerHTML = '';

    const urlParams = new URLSearchParams(window.location.search);
    let tab = urlParams.get('tab') || 'posts';

    tab = tab.replace(/\s+/g, '+');
    selectedTab = tab;

    let filter = urlParams.get('filter') ||
                 ((tab === 'your+posts' || tab === 'your+replies') ? 'newest' : 'all');

    if ((tab === 'your+posts' || tab === 'your+replies') && filter === 'all') {
        filter = 'newest';
    }

    selectedFilter = filter;

    // Detect external navigation (e.g., from post category tags)
    // This happens when we're on tags tab with a specific category filter
    isExternalNavigation = (tab === 'tags' && filter && filter !== 'all');
    if (isExternalNavigation) {
        console.debug(`[Home] Detected external navigation to tags tab with filter: ${filter}`);
    }
    
    const currentUrl = new URL(window.location);
    let urlUpdated = false;
    
    if (currentUrl.searchParams.get('tab') !== tab) {
        currentUrl.searchParams.set('tab', tab);
        urlUpdated = true;
    }
    
    if (currentUrl.searchParams.get('filter') !== filter) {
        currentUrl.searchParams.set('filter', filter);
        urlUpdated = true;
    }
    
    if (urlUpdated) {
        window.history.replaceState({}, '', currentUrl);
    }

    console.debug("[Home] Rendering with params:", { tab, filter });
    
    const isUserLoggedIn = localStorage.getItem('user') !== null;
    const personalTabs = ['your+posts', 'your+replies'];
    if (personalTabs.includes(selectedTab) && !isUserLoggedIn) {
        console.info("[Home] User not logged in but requested personal tab, redirecting to posts tab");
        selectedTab = 'posts';
        selectedFilter = 'all';
        history.replaceState(null, '', '/home?tab=posts&filter=all');
    }

    const headerEl = document.createElement('header');
    headerEl.innerHTML = renderHeader ? renderHeader() : '<p>Header loading error...</p>';
    appContainer.appendChild(headerEl);
    if (attachHeaderEvents) attachHeaderEvents();

    try {
        const categoriesResponse = await fetchCategories();
        if (categoriesResponse.success && Array.isArray(categoriesResponse.data)) {
            categories = categoriesResponse.data;
            console.debug(`[Home] Categories loaded: ${categories.length}`);
            if (selectedTab === 'tags' && selectedFilter !== 'all' && !categories.some(cat => cat.name === selectedFilter)) {
                console.warn(`[Home] Invalid category filter "${selectedFilter}" detected. Resetting to "all".`);
                selectedFilter = 'all';
                history.replaceState(null, '', `/home?tab=${selectedTab}&filter=all`);
            }
        } else {
            console.error('[Home] Failed to load categories:', categoriesResponse.error);
            categories = [];
        }
    } catch (error) {
        console.error('[Home] Exception during category fetch:', error.message || error);
        categories = [];
    }

    const feedContainerHTML = `
        <div class="feed-container">
            ${renderSidebar ? renderSidebar() : '<aside class="sidebar"><p>Loading sidebar error...</p></aside>'}

            <main>
                <section class="feed">
                    <div class="filters">
                        ${renderFilters(selectedTab, selectedFilter, categories)}
                    </div>
                    <div id="posts-list" class="posts-list md3-stagger-container">
                        <div class="loading-indicator md3-loading">
                            <div class="md3-skeleton-card md3-skeleton"></div>
                            <div class="md3-skeleton-card md3-skeleton"></div>
                            <div class="md3-skeleton-card md3-skeleton"></div>
                            <p>Loading posts...</p>
                        </div>
                    </div>
                </section>
            </main>

            <aside class="chat-sidebar-right" id="chat-sidebar-right-container">
                 ${renderChatSidebarHTML ? renderChatSidebarHTML() : '<p>Loading chat error...</p>'}
            </aside>
        </div>
    `;
    appContainer.insertAdjacentHTML('beforeend', feedContainerHTML);

    setTimeout(() => {
        setupFilterEventListeners();
        setupSidebarLinks();
        loadPosts(selectedTab, selectedFilter);

        const chatSidebarElement = document.getElementById('chat-sidebar-right-container');
        if (chatSidebarElement && typeof initChatSidebar === 'function') {
            initChatSidebar(chatSidebarElement);
        } else {
            console.error("[Home] Chat sidebar container or initChatSidebar function not found after rendering.");
        }
    }, 0);
}

function renderFilters(tab, currentFilter, categories) {
    console.debug(`[Home] renderFilters called with tab=${tab}, currentFilter=${currentFilter}, categories=${categories?.length || 0}`);

    let filterOptions = [
        { value: 'all', label: 'All Posts', icon: 'fas fa-globe' },
        { value: 'oldest', label: 'Oldest', icon: 'fas fa-history' }
    ];

    // Determine if this is the tags tab
    const isTagsTab = tab === 'tags';
    
    if (tab === 'your+posts' || tab === 'your+replies') {
        filterOptions = [
            { value: 'newest', label: 'Newest First', icon: 'fas fa-sort-amount-down' },
            { value: 'oldest', label: 'Oldest First', icon: 'fas fa-sort-amount-up' }
        ];
        
        if (currentFilter === 'all') {
            currentFilter = 'newest';
        }
    } 
    else if (isTagsTab) {
        // Start with "All Categories" as the first option
        filterOptions = [
            { value: 'all', label: 'All Categories', icon: 'fas fa-tags' }
        ];

        // Add each category as a filter option
        if (Array.isArray(categories) && categories.length > 0) {
            categories.forEach(category => {
                if (category && category.name) {
                    filterOptions.push({
                        value: category.name,
                        label: category.name,
                        icon: 'fas fa-tag'
                    });
                }
            });
            console.debug(`[Home] Generated ${filterOptions.length - 1} category filter options for tags tab`);
        } else {
            console.warn(`[Home] No categories available for tags tab filters`);
        }

        // Check if the current filter exists in the generated options
        const filterExists = filterOptions.some(option => option.value === currentFilter);
        console.debug(`[Home] Current filter "${currentFilter}" exists in options: ${filterExists}`);

        if (currentFilter === 'oldest' || currentFilter === 'newest') {
            currentFilter = 'all';
        }
    }
    
    return `
        <div class="filter-options md3-enhanced scrollable-filters" data-scroll-physics="true" data-filter-tab="${tab}">
            ${filterOptions.map((filter) => `
                <button
                    class="filter-btn ${currentFilter === filter.value ? 'selected' : ''}"
                    data-filter="${filter.value}"
                    data-tooltip="${filter.label}">
                    <i class="${filter.icon}"></i>
                    <span class="filter-label">${filter.label}</span>
                    <div class="filter-selection-indicator"></div>
                </button>
            `).join('')}
        </div>
    `;
}

// Global scroll position memory for filters
const filterScrollPositions = new Map();

function setupFilterEventListeners() {
    const filtersContainer = document.querySelector('main .filters');
    if (!filtersContainer) {
        console.warn("[Home] Filters container not found in DOM");
        return;
    }

    // Optional: Remove any existing listeners to prevent duplicates
    if (filtersContainer.dataset.listenersAttached === 'true') {
        console.debug("[Home] Filter listeners already attached, skipping");
        return; // Listeners already attached, skip to avoid duplicates
    }

    // Mark the container as having listeners attached
    filtersContainer.dataset.listenersAttached = 'true';

    // Add scroll event listener for handling gradient masks and position memory
    const scrollableFilters = filtersContainer.querySelector('.scrollable-filters');
    if (scrollableFilters) {
        const filterTab = scrollableFilters.dataset.filterTab || 'default';

        // Check if scrolling is actually needed
        const isScrollNeeded = checkIfScrollNeeded(scrollableFilters);

        if (isScrollNeeded) {
            // Check if we should scroll to selected filter (prioritize external navigation)
            const urlParams = new URLSearchParams(window.location.search);
            const currentFilter = urlParams.get('filter') || selectedFilter;
            const shouldScrollToSelected = isExternalNavigation && currentFilter && currentFilter !== 'all' &&
                                         scrollableFilters.querySelector(`[data-filter="${currentFilter}"]`);

            if (shouldScrollToSelected) {
                // Scroll to the selected filter instead of restoring saved position
                // This handles navigation from post category tags or direct URL access
                setTimeout(() => {
                    scrollToSelectedFilter(scrollableFilters, currentFilter);
                    // Reset the flag after handling external navigation
                    isExternalNavigation = false;
                }, 150); // Increased delay to ensure selection state is applied
                console.debug(`[Home] Auto-scrolling to filter for external navigation: ${currentFilter}`);
            } else {
                // Restore scroll position if it exists
                const savedPosition = filterScrollPositions.get(filterTab);
                if (savedPosition !== undefined) {
                    scrollableFilters.scrollLeft = savedPosition;
                    console.debug(`[Home] Restored scroll position for tab ${filterTab}: ${savedPosition}px`);
                }
            }

            // Check initial scroll position
            updateGradientMask(scrollableFilters);

            // Listen for scroll events with throttling
            let scrollTimeout;
            scrollableFilters.addEventListener('scroll', function() {
                updateGradientMask(this);

                // Save scroll position with throttling
                clearTimeout(scrollTimeout);
                scrollTimeout = setTimeout(() => {
                    filterScrollPositions.set(filterTab, this.scrollLeft);
                    console.debug(`[Home] Saved scroll position for tab ${filterTab}: ${this.scrollLeft}px`);
                }, 100);
            });

            // Enhanced window resize handling with debouncing
            let resizeTimeout;
            const handleResize = () => {
                clearTimeout(resizeTimeout);
                resizeTimeout = setTimeout(() => {
                    if (scrollableFilters && scrollableFilters.parentNode) {
                        const stillNeedsScroll = checkIfScrollNeeded(scrollableFilters);
                        if (stillNeedsScroll) {
                            updateGradientMask(scrollableFilters);
                            console.debug('[Home] Scroll features re-enabled after resize');
                        } else {
                            disableScrollFeatures(scrollableFilters);
                            console.debug('[Home] Scroll features disabled after resize');
                        }
                    }
                }, 150); // Debounce resize events
            };

            window.addEventListener('resize', handleResize);

            console.debug("[Home] Scrollable filters with position memory initialized");
        } else {
            // Disable scroll features if not needed
            disableScrollFeatures(scrollableFilters);
            console.debug("[Home] Scroll features disabled - content fits within container");
        }
    }

    filtersContainer.addEventListener('click', async (event) => {
        const targetButton = event.target.closest('button[data-filter]');

        if (targetButton) {
            event.preventDefault();

            // Store the current scroll position before making changes
            let scrollPosition = 0;
            const filterTab = scrollableFilters?.dataset.filterTab || 'default';
            if (scrollableFilters) {
                scrollPosition = scrollableFilters.scrollLeft;
                filterScrollPositions.set(filterTab, scrollPosition);
                console.debug(`[Home] Stored scroll position before filter change: ${scrollPosition}px`);
            }
            
            const urlParams = new URLSearchParams(window.location.search);
            let currentTab = urlParams.get('tab') || selectedTab;
            
            currentTab = currentTab.replace(/\s+/g, '+');
            
            const newFilter = targetButton.dataset.filter;
            console.debug(`[Home] Filter clicked:`, { tab: currentTab, filter: newFilter });
            
            if (currentTab === 'tags' && (newFilter === 'oldest' || newFilter === 'newest')) {
                let category = urlParams.get('filter') || 'all';
                const url = `/home?tab=${currentTab}&filter=${category}&sort=${newFilter}`;
                window.appRouter?.navigate(url);
            } else {
                // Normal behavior for other tabs
                const url = `/home?tab=${currentTab}&filter=${newFilter}`;
                window.appRouter?.navigate(url);
            }
            
            // Set filter click flag to prevent re-rendering filters
            isFilterClick = true;
            
            selectedTab = currentTab;
            selectedFilter = newFilter;
            
            // Simple visual selection
            const previousSelected = filtersContainer.querySelector('button[data-filter].selected');
            if (previousSelected && previousSelected !== targetButton) {
                previousSelected.classList.remove('selected');
            }

            // Simple new selection
            targetButton.classList.add('selected');

            try {
                await loadPosts(currentTab, newFilter);
            } finally {
                // Reset the flag after loading completes (whether successful or not)
                isFilterClick = false;
            }
            
            // Enhanced scroll position restoration after DOM updates and data loading
            if (scrollableFilters) {
                // More robust approach to ensure the scroll position is restored
                const restoreScroll = () => {
                    // Check if the element is still in the DOM and scrollable
                    if (scrollableFilters.parentNode && checkIfScrollNeeded(scrollableFilters)) {
                        // Use requestAnimationFrame for proper timing
                        requestAnimationFrame(() => {
                            scrollableFilters.scrollLeft = scrollPosition;
                            updateGradientMask(scrollableFilters);
                            console.debug(`[Home] Restored scroll position: ${scrollPosition}px`);
                        });
                    }
                };

                // Try multiple times with increasing delays and RAF
                restoreScroll();
                setTimeout(restoreScroll, 16); // ~1 frame
                setTimeout(restoreScroll, 50);
                setTimeout(restoreScroll, 150);
                setTimeout(restoreScroll, 300);

                // Final attempt to ensure scroll position is restored
                setTimeout(restoreScroll, 100);
            }
        }
    });
    
    console.debug("[Home] Filter event listeners attached");
}

/**
 * Enhanced check if scrolling is actually needed for the filter container
 * @param {HTMLElement} element - The scrollable element
 * @returns {boolean} - True if scrolling is needed
 */
function checkIfScrollNeeded(element) {
    if (!element) return false;

    // Store original styles
    const originalOverflow = element.style.overflow;
    const originalOverflowX = element.style.overflowX;

    // Force a layout calculation to get accurate measurements
    element.style.overflow = 'hidden';
    element.style.overflowX = 'hidden';

    // Force reflow to ensure accurate measurements
    element.offsetHeight;

    const containerWidth = element.clientWidth;
    const contentWidth = element.scrollWidth;

    // Restore original styles
    element.style.overflow = originalOverflow || 'auto';
    element.style.overflowX = originalOverflowX || 'auto';

    // Add small buffer to account for rounding errors
    const isScrollNeeded = contentWidth > (containerWidth + 2);
    console.debug(`[Home] Enhanced scroll check: container=${containerWidth}px, content=${contentWidth}px, needed=${isScrollNeeded}`);

    return isScrollNeeded;
}

/**
 * Disable scroll features when not needed
 * @param {HTMLElement} element - The scrollable element
 */
function disableScrollFeatures(element) {
    if (!element) return;

    // Remove scroll-related classes
    element.classList.remove('at-start', 'at-end', 'middle-scroll');

    // Add disabled scroll class
    element.classList.add('scroll-disabled');

    console.debug('[Home] Scroll features disabled for filter container');
}

/**
 * Scroll to the selected filter button in the scrollable container
 * @param {HTMLElement} scrollableFilters - The scrollable filters container
 * @param {string} selectedFilter - The filter value to scroll to
 */
function scrollToSelectedFilter(scrollableFilters, selectedFilter) {
    if (!scrollableFilters || !selectedFilter) return;

    // Find the selected filter button
    const selectedButton = scrollableFilters.querySelector(`[data-filter="${selectedFilter}"].selected`);
    if (!selectedButton) {
        console.debug(`[Home] Selected filter button not found: ${selectedFilter}`);
        return;
    }

    // Calculate the scroll position to center the selected button
    const containerWidth = scrollableFilters.clientWidth;
    const buttonLeft = selectedButton.offsetLeft;
    const buttonWidth = selectedButton.offsetWidth;

    // Center the button in the container
    const targetScrollLeft = buttonLeft - (containerWidth / 2) + (buttonWidth / 2);

    // Ensure we don't scroll beyond the boundaries
    const maxScrollLeft = scrollableFilters.scrollWidth - containerWidth;
    const finalScrollLeft = Math.max(0, Math.min(targetScrollLeft, maxScrollLeft));

    // Smooth scroll to the position
    scrollableFilters.scrollTo({
        left: finalScrollLeft,
        behavior: 'smooth'
    });

    // Update the scroll position memory
    const filterTab = scrollableFilters.dataset.filterTab || 'default';
    filterScrollPositions.set(filterTab, finalScrollLeft);

    console.debug(`[Home] Scrolled to selected filter "${selectedFilter}" at position: ${finalScrollLeft}px`);
}

/**
 * Enhanced gradient mask updates for scrollable filters with better edge case handling
 * @param {HTMLElement} element - The scrollable element
 */
function updateGradientMask(element) {
    if (!element || !element.parentNode) {
        console.debug('[Home] Element not found or not in DOM, skipping gradient mask update');
        return;
    }

    // Only apply gradient masks if scrolling is actually needed
    if (!checkIfScrollNeeded(element)) {
        disableScrollFeatures(element);
        return;
    }

    // Enable scroll features if they were disabled
    element.classList.remove('scroll-disabled');

    // Get accurate scroll measurements with error handling
    let scrollLeft, clientWidth, scrollWidth;
    try {
        scrollLeft = element.scrollLeft || 0;
        clientWidth = element.clientWidth || 0;
        scrollWidth = element.scrollWidth || 0;
    } catch (error) {
        console.warn('[Home] Error getting scroll measurements:', error);
        return;
    }

    // Calculate if we're at the start or end of scrolling with tolerance
    const scrollTolerance = 2; // pixels
    const isAtStart = scrollLeft <= scrollTolerance;
    const isAtEnd = Math.ceil(scrollLeft + clientWidth) >= (scrollWidth - scrollTolerance);

    // Apply appropriate classes based on scroll position
    element.classList.toggle('at-start', isAtStart);
    element.classList.toggle('at-end', isAtEnd);
    element.classList.toggle('middle-scroll', !isAtStart && !isAtEnd);

    console.debug(`[Home] Gradient mask updated: start=${isAtStart}, end=${isAtEnd}, scroll=${scrollLeft}/${scrollWidth}`);
}

function setupSidebarLinks() {
    const sidebarLinks = document.querySelectorAll('.nav-button');
    if (!sidebarLinks.length) {
        console.warn("[Home] No sidebar links found in DOM");
        return;
    }
    
    sidebarLinks.forEach(link => {
        link.addEventListener('click', (e) => {
            e.preventDefault();
            
            let tab = link.dataset.tab || 'posts';
            
            if (tab === 'your posts') {
                tab = 'your+posts';
            } else if (tab === 'your replies') {
                tab = 'your+replies';
            }
            
            let filter = (tab === 'your+posts' || tab === 'your+replies') ? 'newest' : 'all';
            
            console.debug(`[Home] Sidebar link clicked:`, { tab, filter });
            window.appRouter.navigate(`/home?tab=${tab}&filter=${filter}`);
        });
    });
    
    console.debug(`[Home] Attached click handlers to ${sidebarLinks.length} sidebar links`);
}



async function loadPosts(tab = 'posts', filter = 'all') {
    let feedContent = document.querySelector('main .posts-list');
    const filtersContainer = document.querySelector('main .filters');
    
    if (!feedContent) {
        console.warn("[Home] Posts list container not found initially, waiting for DOM updates...");
        await new Promise(resolve => setTimeout(resolve, 100));
        feedContent = document.querySelector('main .posts-list');
        
        if (!feedContent) {
            console.error("[Home] Posts list container not found in main section even after waiting.");
            return;
        }
    }
    
    const normalizedTab = tab.replace(/\s+/g, '+');
    
    let normalizedFilter = filter;
    if ((normalizedTab === 'your+posts' || normalizedTab === 'your+replies') && filter === 'all') {
        normalizedFilter = 'newest';
    }
    
    if (normalizedTab === 'tags' && (normalizedFilter === 'oldest' || normalizedFilter === 'newest')) {
        console.debug('[Home] Tags tab cannot use sort filters directly. Defaulting to "all" categories.');
        normalizedFilter = 'all';
        if (window.appRouter) {
            window.appRouter.navigate(`/home?tab=tags&filter=all`, { replace: true });
        }
    }
    
    console.debug(`[Home] Loading posts:`, {
        tab: normalizedTab,
        filter: normalizedFilter,
        isFilterClick
    });

    feedContent.innerHTML = `
        <div class="loading-indicator">
            <i class="fas fa-spinner fa-spin fa-2x"></i>
            <p>Loading posts...</p>
        </div>
    `;
    
    // IMPORTANT CHANGE: Only update filters when loading a new tab, not during filter clicks
    // This is a page initialization, not a filter click
    if (filtersContainer && !isFilterClick) {
        console.debug(`[Home] Re-rendering filters for tab=${normalizedTab}, filter=${normalizedFilter}, categories=${categories.length}`);
        filtersContainer.innerHTML = renderFilters(normalizedTab, normalizedFilter, categories);

        // Re-attach scroll event listeners since we rebuilt the filters container
        const newScrollableFilters = filtersContainer.querySelector('.scrollable-filters');
        if (newScrollableFilters) {
            updateGradientMask(newScrollableFilters);
            newScrollableFilters.addEventListener('scroll', function() {
                updateGradientMask(this);
            });
        }

        // Reattach the filter click handlers to the newly created elements
        setupFilterEventListeners();

        // CRITICAL FIX: Ensure the correct filter button is selected after re-rendering
        // This is especially important when navigating from post category tags
        setTimeout(() => {
            const correctFilterButton = filtersContainer.querySelector(`button[data-filter="${normalizedFilter}"]`);
            if (correctFilterButton && !correctFilterButton.classList.contains('selected')) {
                // Remove any existing selection
                const currentSelected = filtersContainer.querySelector('button[data-filter].selected');
                if (currentSelected) {
                    currentSelected.classList.remove('selected');
                }
                // Add selection to the correct button
                correctFilterButton.classList.add('selected');
                console.debug(`[Home] Fixed filter selection for "${normalizedFilter}"`);
            }

            // CRITICAL FIX: Scroll to the selected filter to ensure it's visible
            // This provides the same UX as clicking filters directly within the tags tab
            if (correctFilterButton && newScrollableFilters && normalizedFilter !== 'all') {
                scrollToSelectedFilter(newScrollableFilters, normalizedFilter);
                console.debug(`[Home] Scrolled to selected filter "${normalizedFilter}" after navigation`);
            }
        }, 100); // Increased delay to ensure DOM and scroll setup is complete

        console.debug("[Home] Filters container refreshed and event listeners reattached");
    }

    try {
        console.debug(`[Home] Fetching posts with tab=${normalizedTab}, filter=${normalizedFilter}`);
        const result = await fetchPosts(normalizedTab, normalizedFilter);

        if (!result.success) {
            console.error(`[Home] Failed to fetch posts:`, result.error);
            const errorResult = handleApiError({
                status: result.code || '500',
                error: result.error,
                message: result.error
            }, {
                renderPage: false,
                context: 'partial_content',
                fallbackToInline: true
            });

            if (errorResult.shouldShowInline) {
                feedContent.innerHTML = createInlineError(
                    result.error || 'Failed to load posts',
                    result.code,
                    {
                        showRetry: true,
                        retryCallback: `loadPosts('${normalizedTab}', '${normalizedFilter}')`
                    }
                );
            } else {
                // This would trigger a full error page
                handleApiError({
                    status: result.code || '500',
                    error: result.error,
                    message: result.error
                }, {
                    context: 'page_load',
                    originalPath: '/home'
                });
            }
            return;
        }

        if (!Array.isArray(result.data) || result.data.length === 0) {
            let message = 'No posts available for this view.';
            const normalizedTab = tab.replace(/[ +]/g, '+');

            if (normalizedTab === 'tags' && filter !== 'all') {
                message = `No posts found with the category "${filter}".`;
            } else if (normalizedTab === 'your+replies') {
                message = 'You haven\'t commented on any posts yet.';
            } else if (normalizedTab === 'your+posts') {
                message = 'You haven\'t created any posts yet.';
            }

            console.info(`[Home] No posts found for:`, { tab: normalizedTab, filter: normalizedFilter });

            // Enhanced MD3 empty state with contextual icons and messages
            const getEmptyStateConfig = (tab, filter) => {
                if (tab === 'tags' && filter !== 'all') {
                    return {
                        icon: 'fas fa-tag',
                        title: 'No Posts in Category',
                        message: `No posts found with the category "${filter}".`,
                        subtitle: 'Try browsing other categories or create the first post in this category.',
                        showCreateButton: true,
                        buttonText: 'Create First Post',
                        buttonIcon: 'fas fa-plus'
                    };
                } else if (tab === 'your+replies') {
                    return {
                        icon: 'fas fa-comments',
                        title: 'No Comments Yet',
                        message: 'You haven\'t commented on any posts yet.',
                        subtitle: 'Start engaging with the community by commenting on posts that interest you.',
                        showCreateButton: false
                    };
                } else if (tab === 'your+posts') {
                    return {
                        icon: 'fas fa-edit',
                        title: 'No Posts Created',
                        message: 'You haven\'t created any posts yet.',
                        subtitle: 'Share your thoughts, ideas, or questions with the community.',
                        showCreateButton: true,
                        buttonText: 'Create Your First Post',
                        buttonIcon: 'fas fa-plus'
                    };
                } else {
                    return {
                        icon: 'fas fa-inbox',
                        title: 'No Posts Available',
                        message: 'No posts available for this view.',
                        subtitle: 'Check back later or try a different filter.',
                        showCreateButton: true,
                        buttonText: 'Create a Post',
                        buttonIcon: 'fas fa-plus'
                    };
                }
            };

            const config = getEmptyStateConfig(normalizedTab, filter);

            feedContent.innerHTML = `
                <div class="no-posts-container md3-enhanced">
                    <div class="no-posts-content">
                        <div class="no-posts-icon-container">
                            <div class="no-posts-icon-background">
                                <i class="${config.icon}"></i>
                            </div>
                        </div>
                        <div class="no-posts-text">
                            <h3 class="no-posts-title">${config.title}</h3>
                            <p class="no-posts-message">${config.message}</p>
                            <p class="no-posts-subtitle">${config.subtitle}</p>
                        </div>
                        ${config.showCreateButton ? `
                            <div class="no-posts-actions">
                                <button class="create-post-btn md3-fab-extended" onclick="navigateToCreatePost('${normalizedTab}', '${filter}')">
                                    <div class="fab-icon">
                                        <i class="${config.buttonIcon}"></i>
                                    </div>
                                    <span class="fab-label">${config.buttonText}</span>
                                    <div class="fab-ripple"></div>
                                </button>
                            </div>
                        ` : ''}
                    </div>
                </div>
            `;
            return;
        }

        console.info(`[Home] Loaded ${result.data.length} posts for tab=${normalizedTab}, filter=${normalizedFilter}`);
        feedContent.innerHTML = result.data.map(post => createPostHTML(post)).join('');
        attachPostEventListeners();

    } catch (error) {
        console.error('[Home] Error loading posts:', error.message || error);
        feedContent.innerHTML = `
            <div class="error">
                <p>An error occurred while loading posts.</p>
                <button class="btn btn-secondary" onClick="window.location.reload()">Try Again</button>
            </div>
        `;
    }
}

function createPostHTML(post) {
    if (!post || typeof post.PostID !== 'number') {
        console.warn("[Home] Invalid post object received:", post);
        return '<article class="post-card error"><p>Error loading post data.</p></article>';
    }

    const getProp = (obj, path, defaultValue = '') => {
        const value = path.split('.').reduce((o, p) => (o && typeof o === 'object' ? o[p] : undefined), obj);
        return (value === undefined || value === null) ? defaultValue : value;
    };

    const categories = Array.isArray(post.Categories) ? post.Categories : [];
    const avatarSrc = getProp(post, 'Avatar.String') || post.avatar || defaultAvatarPath;

    const categoryButtonsHTML = categories.length > 0
        ? categories.map(cat => {
            const catName = getProp(cat, 'name', '');
            return catName ? `<button class="category-tag md3-enhanced" data-category="${catName}">${catName}</button>` : '';
        }).join('')
        : `<button class="category-tag uncategorized" disabled>Uncategorized</button>`;

    let formattedDate = 'Date unavailable';
    try {
        const postDate = new Date(getProp(post, 'PostAt', ''));
        if (!isNaN(postDate.getTime())) {
            formattedDate = postDate.toLocaleString('en-US', {
                year: 'numeric', month: 'short', day: 'numeric', hour: 'numeric', minute: '2-digit'
            });
        }
    } catch (e) { 
        console.error("[Home] Error formatting date:", { 
            date: post.PostAt, 
            error: e.message || e 
        }); 
    }

    const content = getProp(post, 'Content', '');
    const excerpt = content.substring(0, 150) + (content.length > 150 ? '...' : '');

    const firstName = getProp(post, 'FirstName', '');
    const lastName = getProp(post, 'LastName', '');
    const authorDisplayName = `${firstName} ${lastName}`.trim() || getProp(post, 'Username', 'Anonymous');

    return `
        <article class="post-card md3-enhanced scroll-reveal scroll-momentum" data-post-id="${post.PostID}">
            <div class="post-header md3-enhanced">
                <div class="post-author">
                     <img src="${avatarSrc}"
                          alt="${getProp(post, 'Username', 'User')}'s Avatar"
                          class="avatar md3-enhanced scroll-scale"
                          onerror="this.onerror=null; this.src='${defaultAvatarPath}';">
                    <div class="post-author-info">
                        <span class="author-name md3-enhanced">${authorDisplayName}</span>
                        <span class="post-username">@${getProp(post, 'Username', 'anonymous')}</span>
                    </div>
                </div>
            </div>
            <div class="post-content md3-enhanced">
                 <a href="/post?id=${post.PostID}" class="post-title-link md3-enhanced scroll-fade">
                    <h2 class="post-title md3-enhanced">${getProp(post, 'Title', 'Untitled Post')}</h2>
                 </a>
                <p class="post-excerpt md3-enhanced scroll-fade">${excerpt}</p>
                ${getProp(post, 'Image.Valid') ? `<img src="${getProp(post, 'Image.String')}" alt="Post image" class="post-image md3-enhanced scroll-scale">` : ''}
            </div>
            <div class="post-footer md3-enhanced">
                <div class="post-categories md3-stagger-container">
                    ${categoryButtonsHTML}
                </div>
                <div class="post-actions md3-enhanced">
                     <span class="post-stats md3-enhanced scroll-fade">
                        <span><i class="far fa-comment"></i> ${getProp(post, 'Comments', 0)}</span>
                     </span>
                    <time class="post-time md3-enhanced scroll-fade"><i class="far fa-clock"></i> ${formattedDate}</time>
                </div>
            </div>
        </article>
    `;
}

/**
 * Navigate to create post page with optional category pre-selection
 * @param {string} tab - The current tab
 * @param {string} filter - The current filter/category
 */
function navigateToCreatePost(tab, filter) {
    console.debug('[Home] Navigating to create post with pre-selection:', { tab, filter });

    // Check if we're coming from a specific category in the tags tab
    if (tab === 'tags' && filter && filter !== 'all') {
        // Pre-select the category by passing it as a URL parameter
        const createPostUrl = `/create-post?category=${encodeURIComponent(filter)}`;
        console.debug(`[Home] Pre-selecting category "${filter}" for create post`);

        if (window.appRouter) {
            window.appRouter.navigate(createPostUrl);
        } else {
            window.location.href = createPostUrl;
        }
    } else {
        // Normal navigation without pre-selection
        if (window.appRouter) {
            window.appRouter.navigate('/create-post');
        } else {
            window.location.href = '/create-post';
        }
    }
}

// Make the function globally available for onclick handlers
window.navigateToCreatePost = navigateToCreatePost;

function attachPostEventListeners() {
    const postsList = document.querySelector('main .posts-list');
    if (!postsList) {
        console.warn("[Home] Posts list container not found for attaching listeners.");
        return;
    }

    postsList.addEventListener('click', (event) => {
        const target = event.target;

        const categoryButton = target.closest('.category-tag:not(.uncategorized)');
        if (categoryButton) {
            event.stopPropagation();
            const category = categoryButton.dataset.category;
            if (category && window.appRouter) {
                console.debug(`[Home] Category tag clicked:`, { category });
                window.appRouter.navigate(`/home?tab=tags&filter=${encodeURIComponent(category)}`);
            }
            return;
        }

        const postCard = target.closest('.post-card');
        const isInteractiveElement = target.closest('button, .category-tag');

        if (postCard && !isInteractiveElement) {
            const postId = postCard.dataset.postId;
            if (postId && window.appRouter) {
                console.debug(`[Home] Post card clicked:`, { postId });
                window.appRouter.navigate(`/post?id=${postId}`);
            }
        }
    });
    
    console.debug("[Home] Post event listeners attached");
}
