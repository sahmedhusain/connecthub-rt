import Router from './utils/router.js';
import { renderLogin } from './components/login.js';
import { renderSignup } from './components/signup.js';
import { renderHome } from './components/home.js';
import { renderPost } from './components/post.js';
import { renderNewPost } from './components/newPost.js';
import { renderError, renderErrorFromURL } from './components/error.js';
import { renderChatSidebarHTML, initChatSidebar, renderChatContainerHTML, cleanupChatConnections, getWebSocketStatus, forceWebSocketReconnect } from './components/chat.js';
import { initChatView } from './components/chatView.js'; // Import chat view initializer
import { renderHeader } from './components/header.js';
import { initGlobalChatNotifications } from './components/globalChatNotifications.js';
import { renderSidebar } from './components/sidebar.js'; // Remove chatPopup import
import './utils/scrollPhysics.js'; // Import scroll physics system

// Application version and build information
const APP_VERSION = '1.0.0';
const BUILD_DATE = '2025-05-29';

console.info(`[App] ConnectHub v${APP_VERSION} (built: ${BUILD_DATE})`);
console.debug('[App] Initializing application');

// Register global components for easy access
window.appComponents = {
    renderHeader,
    renderSidebar,
    renderError,
    renderChatSidebarHTML,
    initChatSidebar,
    initChatView,
    renderChatContainerHTML,
    cleanupChatConnections,
    getWebSocketStatus,
    forceWebSocketReconnect
};

console.debug('[App] Registered global components:', Object.keys(window.appComponents));

// Initialize the application when DOM is fully loaded
document.addEventListener('DOMContentLoaded', () => {
    console.info('[App] DOM content loaded, starting application initialization');
    
    const appContainer = document.getElementById('app');
    if (!appContainer) {
        console.error('[App] Fatal: #app container not found in DOM');
        document.body.innerHTML = '<div class="fatal-error">Application initialization failed: app container not found.</div>';
        return;
    }
    
    // Initialize chat view using the dedicated component
    try {
        console.debug('[App] Initializing chat view component');
        window.appComponents.initChatView();
        console.debug('[App] Chat components initialization requested');
    } catch (error) {
        console.error('[App] Chat view initialization failed:', error.message || error);
    }
    
    console.debug('[App] Creating main application structure');
    appContainer.innerHTML = `
        <div id="main-content"></div>
    `;

    // Define application routes
    const routes = [
        { path: '/', component: renderLogin },
        { path: '/signup', component: renderSignup },
        { path: '/home', component: renderHome },
        { path: /^\/post/, component: renderPost },
        { path: '/create-post', component: renderNewPost },
        { path: '/error', component: renderErrorFromURL },
        { path: '404', component: () => renderError('404', 'Page Not Found') }
    ];

    console.debug('[App] Defining application routes:', routes.map(r => typeof r.path === 'string' ? r.path : 'RegExp'));

    // Initialize the global chat notification system
    initGlobalChatNotifications();
    console.debug('[App] Global chat notification system initialized');

    // Initialize router
    try {
        console.debug('[App] Initializing router');
        const router = new Router(routes);
        window.appRouter = router;
        console.info('[App] Router initialized successfully');
    } catch (error) {
        console.error('[App] Router initialization failed:', error.message || error);
        renderError('500', 'Router Initialization Failed');
        return;
    }

    // Intercept link clicks for SPA navigation
    document.addEventListener('click', (event) => {
        const link = event.target.closest('a');
        if (link &&
            link.href.startsWith(window.location.origin) &&
            !link.hasAttribute('data-no-spa') &&
            !link.hasAttribute('download') &&
            link.target !== '_blank') {
            
            const destination = link.pathname + link.search + link.hash;
            console.debug(`[App] Intercepted click on link to: ${destination}`);
            
            event.preventDefault();
            window.appRouter.navigate(destination);
        }
    });

    // Register service worker if supported
    if ('serviceWorker' in navigator) {
        console.debug('[App] Service Worker is supported, attempting to register');
    } else {
        console.debug('[App] Service Worker is not supported in this browser');
    }

    // Add window error handling
    window.addEventListener('error', (event) => {
        console.error('[App] Unhandled global error:', event.message, {
            filename: event.filename,
            lineno: event.lineno,
            colno: event.colno,
            error: event.error
        });
    });

    window.addEventListener('unhandledrejection', (event) => {
        console.error('[App] Unhandled promise rejection:', event.reason?.message || event.reason || 'Unknown reason');
    });

    // Add page visibility change handler for additional WebSocket management
    document.addEventListener('visibilitychange', () => {
        if (document.hidden) {
            console.debug('[App] Page hidden - WebSocket connections will be maintained');
        } else {
            console.debug('[App] Page visible - WebSocket connections will be checked');
        }
    });

    console.info('[App] Application initialization completed successfully');
});

// Track application lifecycle events
window.addEventListener('load', () => {
    console.info('[App] Window load event fired, application fully loaded');
    
    // Calculate and log performance metrics
    if (window.performance) {
        const perfData = window.performance.timing;
        const pageLoadTime = perfData.loadEventEnd - perfData.navigationStart;
        const domReadyTime = perfData.domComplete - perfData.domLoading;
        
        console.debug('[App] Performance metrics:', {
            pageLoadTime: `${pageLoadTime}ms`,
            domReadyTime: `${domReadyTime}ms`
        });
    }
});

window.addEventListener('beforeunload', () => {
    console.debug('[App] Window beforeunload event fired, application shutting down');
    // Clean up WebSocket connections
    cleanupChatConnections();
});