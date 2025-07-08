/* real-time-forum/src/js/components/post.js */

import { fetchPostById, addComment, isLoggedIn, getCurrentUser } from '../utils/api.js';
import { renderHeader, attachHeaderEvents } from './header.js';
import { renderSidebar } from './sidebar.js';
import { renderChatSidebarHTML, initChatSidebar } from './chat.js';
import { handleApiError, createInlineError } from './error.js';

const defaultAvatarPath = '/static/assets/default-avatar.png';

export function renderPost() {
    console.debug("[Post] Rendering post page");
    
    const appContainer = document.getElementById('app');
    if (!appContainer) {
        console.error("[Post] App container not found in DOM");
        return;
    }
    appContainer.innerHTML = '';

    if (!isLoggedIn()) {
        console.warn("[Post] User not logged in, redirecting to login page");
        window.appRouter?.navigate('/') || (window.location.href = '/');
        return;
    }

    const urlParams = new URLSearchParams(window.location.search);
    const postId = urlParams.get('id');

    if (!postId || isNaN(parseInt(postId))) {
        console.error("[Post] Invalid or missing post ID in URL");
        handleApiError({
            status: '400',
            error: 'Invalid Post ID',
            message: 'Invalid Post ID'
        }, {
            context: 'page_load',
            originalPath: window.location.pathname
        });
        return;
    }

    console.debug(`[Post] Rendering post with ID: ${postId}`);
    
    const headerEl = document.createElement('header');
    headerEl.innerHTML = typeof renderHeader === 'function' ? renderHeader() : '<p>Header loading error...</p>';
    appContainer.appendChild(headerEl);
    if (typeof attachHeaderEvents === 'function') attachHeaderEvents();

    const feedContainer = document.createElement('div');
    feedContainer.className = 'feed-container';

    feedContainer.innerHTML += typeof renderSidebar === 'function' ? renderSidebar() : '<aside class="sidebar"><p>Loading sidebar error...</p></aside>';

    const mainContent = document.createElement('main');
    mainContent.innerHTML = `
        <section class="feed">
            <div class="post-view-container" id="post-view-content">
                <div class="loading-indicator md3-loading">
                    <div class="md3-skeleton-card md3-skeleton"></div>
                    <div class="md3-skeleton-text md3-skeleton"></div>
                    <div class="md3-skeleton-text md3-skeleton"></div>
                    <p>Loading post...</p>
                </div>
            </div>
        </section>
    `;
    feedContainer.appendChild(mainContent);

    const chatSidebarContainer = document.createElement('aside');
    chatSidebarContainer.className = 'chat-sidebar-right';
    chatSidebarContainer.id = 'chat-sidebar-right-container';
    chatSidebarContainer.innerHTML = typeof renderChatSidebarHTML === 'function' ? renderChatSidebarHTML() : '<p>Chat loading error...</p>';
    feedContainer.appendChild(chatSidebarContainer);

    appContainer.appendChild(feedContainer);
    console.debug("[Post] DOM structure created");

    fetchPostData(postId);

    const chatSidebarElement = document.getElementById('chat-sidebar-right-container');
    if (chatSidebarElement && typeof initChatSidebar === 'function') {
        initChatSidebar(chatSidebarElement);
        console.debug("[Post] Chat sidebar initialized");
    } else {
        console.error("[Post] Chat sidebar container or initChatSidebar function not found after rendering");
    }
}

async function fetchPostData(postId) {
    console.debug(`[Post] Fetching data for post ID: ${postId}`);
    
    const container = document.getElementById('post-view-content');
    if (!container) {
        console.error("[Post] Post view content container (#post-view-content) not found in DOM");
        return;
    }
    container.innerHTML = `<div class="loading-indicator"><i class="fas fa-spinner fa-spin fa-2x"></i><p>Loading post details...</p></div>`;

    try {
        const result = await fetchPostById(postId);

        if (result.success && result.data && result.data.post && typeof result.data.post.PostID === 'number') {
            const currentUserData = getCurrentUser();
            const data = {
                post: result.data.post,
                comments: Array.isArray(result.data.comments) ? result.data.comments : [],
                categories: Array.isArray(result.data.categories) ? result.data.categories : [],
                userId: currentUserData?.id || currentUserData?.userId || null,
                user_reaction: result.data.user_reaction
            };
            console.info(`[Post] Successfully fetched post data (ID: ${postId})`);
            console.debug(`[Post] Post has ${data.comments.length} comments and ${data.categories.length} categories`);
            renderPostContent(container, data);
        } else {
            throw new Error(result?.error || 'Post not found or failed to fetch');
        }
    } catch (error) {
        console.error('[Post] Error fetching post data:', error.message || error);
        handleApiError({
            status: error.status || '500',
            error: error.message || 'Failed to load post',
            message: error.message || 'Failed to load post'
        }, {
            context: 'page_load',
            originalPath: window.location.pathname
        });
    }
}

function renderPostContent(container, data) {
    if (!container || !data || !data.post || typeof data.post.PostID !== 'number') {
        console.error("[Post] renderPostContent called with invalid data:", { 
            hasContainer: !!container, 
            hasData: !!data, 
            hasPostData: !!(data && data.post), 
            hasValidPostID: !!(data && data.post && typeof data.post.PostID === 'number')
        });
        if (container) container.innerHTML = `<div class="error"><p>Could not display post details.</p><a href="/home" class="btn btn-secondary">Back to Home</a></div>`;
        return;
    }

    const { post, comments = [], categories = [], userId } = data;
    console.debug(`[Post] Rendering post content for post ID: ${post.PostID}`);

    const validCategories = Array.isArray(categories) ? categories : [];
    const categoryButtonsHTML = validCategories.length > 0
        ? validCategories.map(category =>
            category.name ? `<button class="category-tag" data-category="${category.name}">${category.name}</button>` : ''
        ).join('')
        : `<button class="category-tag uncategorized" disabled>Uncategorized</button>`;

    let formattedDate = 'Date unavailable';
    try {
        const postDate = new Date(post.PostAt || '');
        if (!isNaN(postDate.getTime())) {
            formattedDate = postDate.toLocaleString('en-US', {
                year: 'numeric', month: 'short', day: 'numeric', hour: 'numeric', minute: '2-digit'
            });
        }
    } catch (e) { 
        console.error("[Post] Error formatting post date:", { 
            date: post.PostAt, 
            error: e.message || e 
        }); 
    }

    const avatarSrc = post.Avatar?.String || post.avatar || defaultAvatarPath;
    const authorDisplayName = `${post.FirstName || ''} ${post.LastName || ''}`.trim() || post.Username || 'Anonymous';

    const postHtml = `
        <div class="post-view-sections">
            <!-- Main Post Content (No Card Styling) -->
            <article class="post-content-main md3-enhanced" data-post-id="${post.PostID}">
                <div class="post-header md3-enhanced">
                    <img src="${avatarSrc}" alt="${post.Username || 'User'}'s Avatar" class="avatar large md3-enhanced" onerror="this.onerror=null; this.src='${defaultAvatarPath}';">
                    <div class="post-info">
                        <span class="post-author-name md3-enhanced">${authorDisplayName}</span>
                        <span class="post-username">@${post.Username || 'anonymous'}</span>
                        <time class="md3-enhanced"><i class="far fa-clock"></i> ${formattedDate}</time>
                    </div>
                </div>
                <div class="post-content-body md3-enhanced">
                    <h1 class="post-title-full md3-enhanced">${post.Title || 'Untitled Post'}</h1>
                    <div class="post-text-full md3-enhanced">${(post.Content || '').replace(/\n/g, '<br>')}</div>
                </div>
                <div class="post-categories md3-stagger-container">
                    ${categoryButtonsHTML}
                </div>
            </article>

            <!-- Comments Section Card -->
            <section class="comments-card md3-enhanced">
                <div class="comments-header">
                    <h2 class="md3-enhanced"><i class="fas fa-comments"></i> Comments (${comments.length})</h2>
                </div>
                <div class="comments-list">
                     ${renderComments(comments, userId)}
                </div>
            </section>

            <!-- Add Comment Card -->
            <section class="add-comment-card md3-enhanced">
                <div class="add-comment-header">
                    <h3 class="md3-enhanced"><i class="fas fa-plus-circle"></i> Leave a Comment</h3>
                </div>
                <form id="comment-form" class="comment-form md3-enhanced">
                    <div class="comment-input-wrapper">
                        <textarea name="content" rows="3" placeholder="Write your comment here..." required maxlength="200" aria-label="Comment content" class="md3-enhanced"></textarea>
                        <div id="char-limit-error" class="error-message md3-enhanced" style="display: none;">Comment cannot exceed 200 characters.</div>
                    </div>
                    <div class="comment-form-footer">
                        <div id="char-counter" class="char-counter">0/200</div>
                        <input type="hidden" name="post_id" value="${post.PostID}">
                        <input type="hidden" name="user_id" value="${userId || ''}">
                        <button type="submit" class="btn btn-primary md3-enhanced" data-tooltip="Post your comment">
                            <span>Post Comment</span>
                            <div class="btn-ripple"></div>
                        </button>
                    </div>
                </form>
            </section>
        </div>
    `;

    container.innerHTML = postHtml;
    addPostEventListeners(container, userId);
    setupCommentFormValidation(container);
    console.debug(`[Post] Post content rendered for post ID: ${post.PostID}`);
}

function renderComments(comments, currentUserId) {
    if (!Array.isArray(comments)) {
        console.warn("[Post] renderComments received non-array comments:", comments);
        return '<p class="error">Error loading comments.</p>';
    }
    if (comments.length === 0) {
        console.debug("[Post] No comments to render");
        return '<p class="no-comments">Be the first to comment!</p>';
    }

    console.debug(`[Post] Rendering ${comments.length} comments`);
    
    return comments.map(comment => {
        if (!comment || typeof comment.ID !== 'number') {
            console.warn("[Post] Skipping invalid comment object:", comment);
            return '';
        }

        const avatarSrc = comment.Avatar?.String || comment.avatar || defaultAvatarPath;
        const firstName = comment.FirstName || '';
        const lastName = comment.LastName || '';
        const username = comment.Username || 'anonymous';
        const content = comment.Content || '';
        const createdAt = comment.CreatedAt || null;
        const commentUserId = comment.UserID || null;

        let formattedCommentDate = 'Date unavailable';
        try {
            const commentDate = new Date(createdAt || '');
            if (!isNaN(commentDate.getTime())) {
                formattedCommentDate = commentDate.toLocaleString('en-US', {
                    year: 'numeric', month: 'short', day: 'numeric', hour: 'numeric', minute: '2-digit'
                });
            }
        } catch (e) { 
            console.error("[Post] Error formatting comment date:", { 
                date: createdAt, 
                error: e.message || e 
            }); 
        }

        const authorDisplayName = `${firstName} ${lastName}`.trim() || username;

        return `
            <div class="comment md3-enhanced" data-comment-id="${comment.ID}">
                <div class="comment-header md3-enhanced">
                    <img src="${avatarSrc}" alt="${username}'s Avatar" class="avatar small md3-enhanced" onerror="this.onerror=null; this.src='${defaultAvatarPath}';">
                    <div class="comment-info">
                        <span class="comment-author-name md3-enhanced">${authorDisplayName}</span>
                        <span class="comment-username">@${username}</span>
                    </div>
                    <time class="comment-time md3-enhanced"><i class="far fa-clock"></i> ${formattedCommentDate}</time>
                </div>
                <div class="comment-content md3-enhanced">
                    <p>${content.replace(/\n/g, '<br>')}</p>
                </div>
            </div>
        `;
    }).join('');
}

function setupCommentFormValidation(container) {
    console.debug("[Post] Setting up comment form validation");
    
    const textarea = container.querySelector('#comment-form textarea[name="content"]');
    const charCounter = container.querySelector('#comment-form #char-counter');
    const errorDisplay = container.querySelector('#comment-form #char-limit-error');
    const submitButton = container.querySelector('#comment-form button[type="submit"]');
    const maxLength = 200;

    if (!textarea || !charCounter || !errorDisplay || !submitButton) {
        console.warn("[Post] Could not find all elements for comment form validation");
        return;
    }

    const updateValidation = () => {
        const currentLength = textarea.value.length;
        charCounter.textContent = `${currentLength}/${maxLength}`;
        const isOverLimit = currentLength > maxLength;

        errorDisplay.style.display = isOverLimit ? 'block' : 'none';
        charCounter.classList.toggle('error', isOverLimit);
        submitButton.disabled = isOverLimit;
        
        if (isOverLimit) {
            console.debug(`[Post] Comment exceeds maximum length: ${currentLength}/${maxLength}`);
        }
    };

    textarea.addEventListener('input', updateValidation);
    updateValidation();
    console.debug("[Post] Comment form validation initialized");
}

function addPostEventListeners(container, userId) {
    console.debug("[Post] Setting up post event listeners");
    
    const commentForm = container.querySelector('#comment-form');
    if (commentForm) {
        const userIdInput = commentForm.querySelector('input[name="user_id"]');
        if (userIdInput && !userIdInput.value && userId) {
            userIdInput.value = userId;
        }

        commentForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            console.debug("[Post] Comment form submitted");
            
            const contentElement = commentForm.querySelector('textarea[name="content"]');
            const postIdInput = commentForm.querySelector('input[name="post_id"]');
            const currentUserIdInput = commentForm.querySelector('input[name="user_id"]');
            const submitButton = commentForm.querySelector('button[type="submit"]');

            if (!contentElement || !postIdInput || !currentUserIdInput || !submitButton) {
                console.error("[Post] Comment form elements missing");
                alert("An error occurred. Please try again.");
                return;
            }

            const content = contentElement.value.trim();
            const postId = postIdInput.value;
            const currentUserId = currentUserIdInput.value || getCurrentUser()?.id || getCurrentUser()?.userId;

            if (!currentUserId) {
                console.warn("[Post] User not logged in when trying to comment");
                alert("You must be logged in to comment.");
                window.appRouter?.navigate('/') || (window.location.href = '/');
                return;
            }
            if (!content) {
                console.debug("[Post] Empty comment submission rejected");
                alert("Comment cannot be empty.");
                contentElement.focus();
                return;
            }
            if (content.length > 200) {
                console.debug("[Post] Comment too long: " + content.length + " characters");
                alert("Comment cannot exceed 200 characters.");
                contentElement.focus();
                return;
            }

            submitButton.disabled = true;
            submitButton.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Posting...';
            console.debug(`[Post] Submitting comment for post ${postId}`);

            try {
                const result = await addComment(postId, content);

                if (result.success || result.redirectUrl) {
                    console.info("[Post] Comment added successfully");
                    fetchPostData(postId);
                } else {
                    console.warn("[Post] Comment submission failed:", result.error || "No error message provided");
                    alert(result.error || 'Failed to add comment. Please try again.');
                }
            } catch (error) {
                console.error('[Post] Error submitting comment:', error.message || error);
                alert('An error occurred while posting the comment.');
            } finally {
                const finalSubmitButton = container.querySelector('#comment-form button[type="submit"]');
                if (finalSubmitButton) {
                    finalSubmitButton.disabled = false;
                    finalSubmitButton.innerHTML = 'Post Comment';
                }
                console.debug("[Post] Comment submission process completed");
            }
        });
    } else {
        console.warn("[Post] Comment form not found in DOM");
    }

    const categoriesContainer = container.querySelector('.post-categories');
    if (categoriesContainer) {
        categoriesContainer.addEventListener('click', (event) => {
            const targetButton = event.target.closest('.category-tag:not(.uncategorized)');
            if (targetButton) {
                const categoryName = targetButton.dataset.category;
                if (categoryName && window.appRouter) {
                    console.debug(`[Post] Category tag clicked: ${categoryName}`);
                    window.appRouter.navigate(`/home?tab=tags&filter=${encodeURIComponent(categoryName)}`);
                }
            }
        });
    }

    const commentsList = container.querySelector('.comments-list');
    if (commentsList) {
        commentsList.addEventListener('click', (event) => {
            const editButton = event.target.closest('.edit-comment-btn');
            const deleteButton = event.target.closest('.delete-comment-btn');
            const postId = container.querySelector('article.post-full')?.dataset.postId;

            if (editButton) {
                const commentId = editButton.dataset.commentId;
                console.debug(`[Post] Edit comment button clicked for comment ID: ${commentId}`);
                alert(`Edit functionality for comment ${commentId} not implemented.`);
            } else if (deleteButton) {
                const commentId = deleteButton.dataset.commentId;
                console.debug(`[Post] Delete comment button clicked for comment ID: ${commentId}`);
                if (confirm("Are you sure you want to delete this comment?")) {
                    console.debug(`[Post] User confirmed deletion of comment ID: ${commentId}`);
                    alert(`Delete functionality for comment ${commentId} not implemented.`);
                }
            }
        });
    }

    const reactionButtons = container.querySelectorAll('.post-reactions button');
    reactionButtons.forEach(button => {
        button.addEventListener('click', async () => {
            const action = button.dataset.action;
            const postId = button.dataset.postId;
            console.debug(`[Post] Reaction button clicked: ${action} on post ${postId}`);
            alert(`Reaction functionality (${action}) not implemented.`);
        });
    });
    
    console.debug("[Post] Post event listeners attached successfully");
}

function formatDate(dateString) {
    if (!dateString) return 'Date unavailable';
    try {
        const date = new Date(dateString);
        if (isNaN(date.getTime())) {
            console.warn("[Post] Could not parse date string:", dateString);
            return 'Invalid date';
        }
        return date.toLocaleDateString('en-US', {
            year: 'numeric', month: 'short', day: 'numeric',
            hour: 'numeric', minute: '2-digit', hour12: true
        });
    } catch (e) {
        console.error("[Post] Error formatting date:", { 
            input: dateString, 
            error: e.message || e 
        });
        return 'Invalid date';
    }
}
