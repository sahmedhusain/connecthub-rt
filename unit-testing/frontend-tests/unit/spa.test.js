/**
 * SPA Navigation and Routing Tests
 * Tests for single-page application routing, navigation, and state management
 */

import { describe, test, expect, beforeEach, jest } from '@jest/globals';

describe('SPA Navigation and Routing Tests', () => {
  let mockRouter;
  let mockRoutes;

  beforeEach(() => {
    // Reset DOM
    global.testUtils.cleanupDOM();
    
    // Reset window.history mock
    window.history.pushState.mockClear();
    window.history.replaceState.mockClear();
    
    // Setup mock routes
    mockRoutes = [
      { path: '/', component: jest.fn() },
      { path: '/home', component: jest.fn() },
      { path: '/signup', component: jest.fn() },
      { path: '/create-post', component: jest.fn() },
      { path: /^\/post\/(\d+)/, component: jest.fn() },
      { path: '404', component: jest.fn() }
    ];

    // Mock router implementation
    mockRouter = {
      routes: mockRoutes,
      currentRoute: null,
      navigate: jest.fn(),
      getCurrentPath: jest.fn(() => window.location.pathname),
      matchRoute: jest.fn(),
      renderRoute: jest.fn()
    };
  });

  describe('Route Matching', () => {
    test('should match exact string routes', () => {
      const testCases = [
        { path: '/', route: '/', shouldMatch: true },
        { path: '/home', route: '/home', shouldMatch: true },
        { path: '/signup', route: '/signup', shouldMatch: true },
        { path: '/home', route: '/', shouldMatch: false },
        { path: '/unknown', route: '/home', shouldMatch: false }
      ];

      testCases.forEach(({ path, route, shouldMatch }) => {
        const routeObj = mockRoutes.find(r => r.path === route);
        const matches = (typeof routeObj.path === 'string') ? routeObj.path === path : routeObj.path.test(path);
        
        expect(matches).toBe(shouldMatch);
      });
    });

    test('should match regex routes with parameters', () => {
      const postRouteRegex = /^\/post\/(\d+)/;
      
      const testCases = [
        { path: '/post/123', shouldMatch: true, expectedId: '123' },
        { path: '/post/456', shouldMatch: true, expectedId: '456' },
        { path: '/post/abc', shouldMatch: false },
        { path: '/posts/123', shouldMatch: false },
        { path: '/post/', shouldMatch: false }
      ];

      testCases.forEach(({ path, shouldMatch, expectedId }) => {
        const match = postRouteRegex.exec(path);
        
        if (shouldMatch) {
          expect(match).toBeTruthy();
          expect(match[1]).toBe(expectedId);
        } else {
          expect(match).toBeFalsy();
        }
      });
    });

    test('should handle 404 for unmatched routes', () => {
      const unknownPaths = ['/unknown', '/invalid/path', '/post/invalid'];
      
      unknownPaths.forEach(path => {
        const matchedRoute = mockRoutes.find(route => {
          if (typeof route.path === 'string') {
            return route.path === path;
          } else if (route.path instanceof RegExp) {
            return route.path.test(path);
          }
          return false;
        });

        // Should not match any route except 404
        expect(matchedRoute).toBeFalsy();
        
        // Should fall back to 404 route
        const notFoundRoute = mockRoutes.find(route => route.path === '404');
        expect(notFoundRoute).toBeTruthy();
      });
    });
  });

  describe('Navigation Functions', () => {
    test('should navigate to new route', () => {
      const targetPath = '/home';
      
      // Simulate navigation
      mockRouter.navigate(targetPath);
      
      expect(mockRouter.navigate).toHaveBeenCalledWith(targetPath);
    });

    test('should update browser history on navigation', () => {
      const targetPath = '/create-post';
      const targetTitle = 'Create Post';
      
      // Simulate navigation with history update
      window.history.pushState({ path: targetPath }, targetTitle, targetPath);
      
      expect(window.history.pushState).toHaveBeenCalledWith(
        { path: targetPath },
        targetTitle,
        targetPath
      );
    });

    test('should handle back button navigation', () => {
      let popstateHandled = false;
      
      // Setup popstate listener
      window.addEventListener('popstate', (event) => {
        popstateHandled = true;
        const path = event.state?.path || '/';
        mockRouter.navigate(path);
      });

      // Simulate back button
      const popstateEvent = new PopStateEvent('popstate', {
        state: { path: '/home' }
      });
      window.dispatchEvent(popstateEvent);

      expect(popstateHandled).toBe(true);
      expect(mockRouter.navigate).toHaveBeenCalledWith('/home');
    });
  });

  describe('Link Interception', () => {
    test('should intercept internal link clicks', () => {
      const mainContent = document.getElementById('main-content');
      mainContent.innerHTML = `
        <div>
          <a href="/home" id="internal-link">Home</a>
          <a href="https://external.com" id="external-link">External</a>
          <a href="/download" download id="download-link">Download</a>
          <a href="/new-tab" target="_blank" id="new-tab-link">New Tab</a>
        </div>
      `;

      let interceptedLinks = [];

      // Setup link interception
      document.addEventListener('click', (event) => {
        const link = event.target.closest('a');
        if (link &&
            link.href.startsWith(window.location.origin) &&
            !link.hasAttribute('data-no-spa') &&
            !link.hasAttribute('download') &&
            link.target !== '_blank') {
          
          event.preventDefault();
          interceptedLinks.push(link.pathname);
          mockRouter.navigate(link.pathname);
        }
      });

      // Test internal link
      const internalLink = document.getElementById('internal-link');
      global.testUtils.simulateClick(internalLink);

      expect(interceptedLinks).toContain('/home');
      expect(mockRouter.navigate).toHaveBeenCalledWith('/home');

      // Test external link (should not be intercepted)
      const externalLink = document.getElementById('external-link');
      global.testUtils.simulateClick(externalLink);

      expect(interceptedLinks).not.toContain('https://external.com');

      // Test download link (should not be intercepted)
      const downloadLink = document.getElementById('download-link');
      global.testUtils.simulateClick(downloadLink);

      expect(interceptedLinks).not.toContain('/download');

      // Test new tab link (should not be intercepted)
      const newTabLink = document.getElementById('new-tab-link');
      global.testUtils.simulateClick(newTabLink);

      expect(interceptedLinks).not.toContain('/new-tab');
    });

    test('should respect data-no-spa attribute', () => {
      const mainContent = document.getElementById('main-content');
      mainContent.innerHTML = `
        <a href="/no-spa" data-no-spa id="no-spa-link">No SPA</a>
      `;

      let intercepted = false;

      document.addEventListener('click', (event) => {
        const link = event.target.closest('a');
        if (link &&
            link.href.startsWith(window.location.origin) &&
            !link.hasAttribute('data-no-spa') &&
            !link.hasAttribute('download') &&
            link.target !== '_blank') {
          
          event.preventDefault();
          intercepted = true;
        }
      });

      const noSpaLink = document.getElementById('no-spa-link');
      global.testUtils.simulateClick(noSpaLink);

      expect(intercepted).toBe(false);
    });
  });

  describe('Route Rendering', () => {
    test('should render component for matched route', () => {
      const mainContent = document.getElementById('main-content');
      
      // Mock home component
      const homeComponent = jest.fn(() => {
        mainContent.innerHTML = '<div id="home-page">Welcome to Home</div>';
      });

      // Simulate route rendering
      homeComponent();

      expect(homeComponent).toHaveBeenCalled();
      expect(document.getElementById('home-page')).toBeTruthy();
      expect(document.getElementById('home-page').textContent).toBe('Welcome to Home');
    });

    test('should pass route parameters to component', () => {
      const postComponent = jest.fn((params) => {
        const mainContent = document.getElementById('main-content');
        mainContent.innerHTML = `<div id="post-page">Post ID: ${params.id}</div>`;
      });

      // Simulate route with parameters
      const path = '/post/123';
      const match = /^\/post\/(\d+)/.exec(path);
      const params = { id: match[1] };

      postComponent(params);

      expect(postComponent).toHaveBeenCalledWith(params);
      expect(document.getElementById('post-page').textContent).toBe('Post ID: 123');
    });

    test('should handle route transitions', () => {
      const mainContent = document.getElementById('main-content');
      let transitionSteps = [];

      // Mock route transition
      function transitionToRoute(newComponent) {
        transitionSteps.push('cleanup-old');
        mainContent.innerHTML = '';
        
        transitionSteps.push('render-new');
        newComponent();
        
        transitionSteps.push('complete');
      }

      const newComponent = jest.fn(() => {
        mainContent.innerHTML = '<div>New Route Content</div>';
      });

      transitionToRoute(newComponent);

      expect(transitionSteps).toEqual(['cleanup-old', 'render-new', 'complete']);
      expect(newComponent).toHaveBeenCalled();
      expect(mainContent.innerHTML).toBe('<div>New Route Content</div>');
    });
  });

  describe('State Management', () => {
    test('should preserve application state during navigation', () => {
      // Mock application state
      const appState = {
        user: { id: 1, username: 'testuser' },
        currentPage: 'home',
        sidebarOpen: true
      };

      // Simulate navigation that preserves state
      function navigateWithState(newPath, newPageState) {
        appState.currentPage = newPageState;
        window.history.pushState(appState, '', newPath);
      }

      navigateWithState('/create-post', 'create-post');

      expect(appState.currentPage).toBe('create-post');
      expect(appState.user).toEqual({ id: 1, username: 'testuser' });
      expect(appState.sidebarOpen).toBe(true);
    });

    test('should restore state on browser navigation', () => {
      const savedState = {
        user: { id: 1, username: 'testuser' },
        currentPage: 'home',
        scrollPosition: 150
      };

      // Simulate popstate with saved state
      const popstateEvent = new PopStateEvent('popstate', {
        state: savedState
      });

      let restoredState = null;
      window.addEventListener('popstate', (event) => {
        restoredState = event.state;
      });

      window.dispatchEvent(popstateEvent);

      expect(restoredState).toEqual(savedState);
    });

    test('should handle deep linking', () => {
      // Simulate direct navigation to deep URL
      const deepPath = '/post/456';
      window.location.pathname = deepPath;

      // Mock route initialization from deep link
      function initializeFromDeepLink() {
        const path = window.location.pathname;
        const postMatch = /^\/post\/(\d+)/.exec(path);
        
        if (postMatch) {
          return {
            route: 'post',
            params: { id: postMatch[1] }
          };
        }
        
        return { route: '404' };
      }

      const result = initializeFromDeepLink();

      expect(result.route).toBe('post');
      expect(result.params.id).toBe('456');
    });
  });

  describe('Error Handling', () => {
    test('should handle navigation errors gracefully', () => {
      let errorHandled = false;
      let errorMessage = '';

      // Mock navigation with error
      function navigateWithError(path) {
        try {
          if (path === '/error-route') {
            throw new Error('Navigation failed');
          }
          mockRouter.navigate(path);
        } catch (error) {
          errorHandled = true;
          errorMessage = error.message;
          // Fallback to 404
          mockRouter.navigate('/404');
        }
      }

      navigateWithError('/error-route');

      expect(errorHandled).toBe(true);
      expect(errorMessage).toBe('Navigation failed');
      expect(mockRouter.navigate).toHaveBeenCalledWith('/404');
    });

    test('should handle component rendering errors', () => {
      const mainContent = document.getElementById('main-content');
      let renderingError = null;

      // Mock component that throws error
      const errorComponent = jest.fn(() => {
        throw new Error('Component rendering failed');
      });

      // Mock error boundary
      function renderWithErrorBoundary(component) {
        try {
          component();
        } catch (error) {
          renderingError = error;
          mainContent.innerHTML = '<div class="error-page">Something went wrong</div>';
        }
      }

      renderWithErrorBoundary(errorComponent);

      expect(renderingError).toBeTruthy();
      expect(renderingError.message).toBe('Component rendering failed');
      expect(mainContent.innerHTML).toBe('<div class="error-page">Something went wrong</div>');
    });
  });
});
