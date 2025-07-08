/**
 * Jest Setup Configuration for Frontend Tests
 * Sets up the testing environment for DOM manipulation and API mocking
 */

// Mock console methods to reduce noise in tests (jest is globally available)
global.console = {
  ...console,
  debug: jest.fn(),
  info: jest.fn(),
  warn: jest.fn(),
  error: jest.fn(),
};

// Mock fetch API for API testing
global.fetch = jest.fn();

// Mock WebSocket for real-time testing
global.WebSocket = jest.fn().mockImplementation(() => ({
  send: jest.fn(),
  close: jest.fn(),
  addEventListener: jest.fn(),
  removeEventListener: jest.fn(),
  readyState: 1, // OPEN
  CONNECTING: 0,
  OPEN: 1,
  CLOSING: 2,
  CLOSED: 3,
}));

// Mock localStorage
const localStorageMock = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
  clear: jest.fn(),
};
global.localStorage = localStorageMock;

// Mock sessionStorage
const sessionStorageMock = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
  clear: jest.fn(),
};
global.sessionStorage = sessionStorageMock;

// Mock window.location
delete window.location;
window.location = {
  href: 'http://localhost:8080',
  origin: 'http://localhost:8080',
  pathname: '/',
  search: '',
  hash: '',
  assign: jest.fn(),
  replace: jest.fn(),
  reload: jest.fn(),
};

// Mock window.history
window.history = {
  pushState: jest.fn(),
  replaceState: jest.fn(),
  back: jest.fn(),
  forward: jest.fn(),
  go: jest.fn(),
  length: 1,
  state: null,
};

// Mock performance API
window.performance = {
  timing: {
    navigationStart: Date.now() - 1000,
    loadEventEnd: Date.now(),
    domComplete: Date.now() - 100,
    domLoading: Date.now() - 500,
  },
  now: jest.fn(() => Date.now()),
};

// Mock IntersectionObserver
global.IntersectionObserver = jest.fn().mockImplementation(() => ({
  observe: jest.fn(),
  unobserve: jest.fn(),
  disconnect: jest.fn(),
}));

// Mock ResizeObserver
global.ResizeObserver = jest.fn().mockImplementation(() => ({
  observe: jest.fn(),
  unobserve: jest.fn(),
  disconnect: jest.fn(),
}));

// Mock matchMedia
window.matchMedia = jest.fn().mockImplementation(query => ({
  matches: false,
  media: query,
  onchange: null,
  addListener: jest.fn(),
  removeListener: jest.fn(),
  addEventListener: jest.fn(),
  removeEventListener: jest.fn(),
  dispatchEvent: jest.fn(),
}));

// Setup DOM environment
document.body.innerHTML = `
  <div id="app">
    <div id="main-content"></div>
  </div>
`;

// Global test utilities
global.testUtils = {
  // Create a mock response for fetch
  createMockResponse: (data, status = 200) => ({
    ok: status >= 200 && status < 300,
    status,
    statusText: status === 200 ? 'OK' : 'Error',
    json: jest.fn().mockResolvedValue(data),
    text: jest.fn().mockResolvedValue(JSON.stringify(data)),
    headers: new Map(),
  }),

  // Wait for DOM updates
  waitForDOM: () => new Promise(resolve => setTimeout(resolve, 0)),

  // Simulate user events
  simulateClick: (element) => {
    const event = new MouseEvent('click', {
      bubbles: true,
      cancelable: true,
      view: window,
    });
    element.dispatchEvent(event);
  },

  simulateInput: (element, value) => {
    element.value = value;
    const event = new Event('input', {
      bubbles: true,
      cancelable: true,
    });
    element.dispatchEvent(event);
  },

  simulateSubmit: (form) => {
    const event = new Event('submit', {
      bubbles: true,
      cancelable: true,
    });
    form.dispatchEvent(event);
  },

  // Clean up DOM between tests
  cleanupDOM: () => {
    document.body.innerHTML = `
      <div id="app">
        <div id="main-content"></div>
      </div>
    `;
  },
};

// Reset mocks before each test
beforeEach(() => {
  jest.clearAllMocks();
  global.testUtils.cleanupDOM();
  
  // Reset fetch mock
  global.fetch.mockClear();
  
  // Reset WebSocket mock
  global.WebSocket.mockClear();
  
  // Reset storage mocks
  localStorageMock.getItem.mockClear();
  localStorageMock.setItem.mockClear();
  localStorageMock.removeItem.mockClear();
  localStorageMock.clear.mockClear();
  
  sessionStorageMock.getItem.mockClear();
  sessionStorageMock.setItem.mockClear();
  sessionStorageMock.removeItem.mockClear();
  sessionStorageMock.clear.mockClear();
});

// Global error handling for tests
process.on('unhandledRejection', (reason, promise) => {
  console.error('Unhandled Rejection at:', promise, 'reason:', reason);
});

console.log('[Test Setup] Jest environment configured for frontend testing');
