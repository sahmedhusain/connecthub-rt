/* real-time-forum/src/js/utils/router.js */

// Import component render functions
import { renderLogin } from '../components/login.js';
import { renderSignup } from '../components/signup.js';
import { renderHome } from '../components/home.js';
import { renderPost } from '../components/post.js';
import { renderNewPost } from '../components/newPost.js';
import { renderError } from '../components/error.js';
import { initNotificationSystem } from '../components/chatView.js';

/**
 * Simple SPA router to handle navigation between components
 * without full page reloads.
 */
class Router {
  constructor(routes) {
    this.routes = routes;
    this.currentRoute = null;

    console.debug("[Router] Initializing with routes:", routes.map(r => r.path));

    // Handle browser navigation events (back/forward buttons)
    window.addEventListener('popstate', (event) => {
      console.debug("[Router] Popstate event triggered", event.state);
      this.resolve(); // Re-resolve the route when history changes
    });

    // Handle hash change events for hash-based routing
    window.addEventListener('hashchange', (event) => {
      console.debug("[Router] Hash change event triggered", { oldURL: event.oldURL, newURL: event.newURL });
      this.resolve(); // Re-resolve the route when hash changes
    });

    // Initial route resolution on page load
    this.resolve();
  }

  /**
   * Navigates to a new path within the SPA.
   * Updates the browser history and resolves the new route.
   * @param {string} path - The path to navigate to (e.g., '/home', '/post?id=123').
   * @param {Object} [options] - Navigation options
   * @param {boolean} [options.replace=false] - Replace current history entry instead of pushing new one
   */
  navigate(path, options = {}) {
    // Only update state if the path is different from the current one
    const currentPath = window.location.pathname + window.location.search + window.location.hash;
    
    if (currentPath !== path) {
      console.info(`[Router] Navigating to: ${path}`);
      
      const state = { path: path, timestamp: Date.now() };
      
      // Push or replace the browser history state
      if (options.replace) {
        console.debug(`[Router] Replacing history state with: ${path}`);
        window.history.replaceState(state, '', path);
      } else {
        console.debug(`[Router] Pushing new history state: ${path}`);
        window.history.pushState(state, '', path);
      }
      
      // Resolve and render the component for the new path
      this.resolve();
    } else {
      console.debug(`[Router] Already at path: ${path}, not navigating.`);
    }
  }

  /**
   * Resolves the current URL path against the defined routes
   * and renders the corresponding component.
   */
  resolve() {
    // Get the current path, search query, and hash from the window location
    const fullPath = window.location.pathname + window.location.search + window.location.hash;
    let currentPath = window.location.pathname; // Path part for matching routes

    // Handle hash-based routing (e.g., /#/error)
    if (window.location.hash && window.location.hash.startsWith('#/')) {
      const hashPath = window.location.hash.substring(1); // Remove the '#'
      const hashPathOnly = hashPath.split('?')[0]; // Remove query parameters from hash
      currentPath = hashPathOnly;
      console.debug(`[Router] Hash detected: ${window.location.hash}`);
      console.debug(`[Router] Hash path: ${hashPath}`);
      console.debug(`[Router] Using hash-based path: ${currentPath}`);
    }

    console.debug(`[Router] Resolving route for: ${fullPath}, currentPath: ${currentPath}`);

    // Note: WebSocket connections are managed by the WebSocketConnectionManager
    // and will persist across route changes unless the user logs out or leaves the site

    let route = null;
    let matchType = 'none';

    // Find a matching route
    for (const r of this.routes) {
      if (typeof r.path === 'string' && r.path === currentPath) {
        route = r;
        matchType = 'exact';
        break;
      } else if (r.path instanceof RegExp && r.path.test(currentPath)) {
        route = r;
        matchType = 'regex';
        break;
      }
    }

    // If no specific route matches, use the 404 route
    if (!route) {
      console.warn(`[Router] No matching route found for ${currentPath}, using 404 fallback.`);
      route = this.routes.find(r => r.path === '404');
      matchType = '404';
    } else {
      console.debug(`[Router] Found ${matchType} match for route:`, typeof route.path === 'string' ? route.path : 'RegExp');
    }

    // Check if a valid route component exists
    if (route && typeof route.component === 'function') {
      this.currentRoute = route;
      console.info(`[Router] Rendering component for route: ${matchType === 'regex' ? 'RegExp' : route.path}`);

      // Clear the app container and render the new component
      const appContainer = document.getElementById('app');
      if (appContainer) {
        if (route.path === '/home') {
          const urlParams = new URLSearchParams(window.location.search);
          let tab = urlParams.get('tab') || 'posts';

          // Normalize tab parameter - ensure we have the correct format for server
          if (tab === 'your posts') {
              tab = 'your+posts';
              console.debug(`[Router] Normalized tab 'your posts' to 'your+posts'`);
          } else if (tab === 'your replies') {
              tab = 'your+replies';
              console.debug(`[Router] Normalized tab 'your replies' to 'your+replies'`);
          }

          // Set appropriate default filter based on tab
          let filter = urlParams.get('filter');
          if (!filter) {
              if (tab === 'your+posts' || tab === 'your+replies') {
                  filter = 'newest';
                  console.debug(`[Router] Setting default filter 'newest' for personal tab: ${tab}`);
              } else {
                  filter = 'all';
                  console.debug(`[Router] Setting default filter 'all' for tab: ${tab}`);
              }

              // Update URL with normalized parameters
              const newUrl = new URL(window.location);
              newUrl.searchParams.set('tab', tab);
              newUrl.searchParams.set('filter', filter);
              window.history.replaceState({}, '', newUrl);
              console.debug(`[Router] Updated URL parameters: tab=${tab}, filter=${filter}`);
          }

          renderHome();
        } else {
          route.component();
        }

        // Initialize notification system for all routes after rendering
        try {
          console.debug(`[Router] Initializing notification system for route: ${typeof route.path === 'string' ? route.path : 'RegExp'}`);
          initNotificationSystem();
          console.debug(`[Router] Notification system initialized successfully`);
        } catch (error) {
          console.warn(`[Router] Failed to initialize notification system:`, error);
        }
      } else {
        console.error("[Router] App container '#app' not found in DOM");
      }
    } else {
      console.error(`[Router] No valid component found for route: ${currentPath}`);
      // Optionally render a generic error page if even 404 route is missing
      const appContainer = document.getElementById('app');
      if (appContainer) {
        console.warn("[Router] Rendering fallback error page due to missing/invalid component");
        renderError('500', 'Router Configuration Error');
      } else {
        console.error("[Router] Cannot render error page: App container not found");
      }
    }
  }
  
  /**
   * Gets information about the current route
   * @returns {Object} Information about the current route
   */
  getCurrentRouteInfo() {
    if (!this.currentRoute) {
      console.debug("[Router] getCurrentRouteInfo called but no current route exists");
      return { exists: false };
    }
    
    const pathString = typeof this.currentRoute.path === 'string' 
      ? this.currentRoute.path 
      : '[RegExp]';
      
    console.debug(`[Router] Current route info: ${pathString}`);
    
    return {
      exists: true,
      path: this.currentRoute.path,
      fullPath: window.location.pathname + window.location.search + window.location.hash,
      params: new URLSearchParams(window.location.search)
    };
  }
}

export default Router;
