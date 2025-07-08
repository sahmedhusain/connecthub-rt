/**
 * Playwright Global Setup
 * Prepares the test environment before running E2E tests
 */

import { chromium } from '@playwright/test';
import { exec } from 'child_process';
import { promisify } from 'util';

const execAsync = promisify(exec);

async function globalSetup() {
  console.log('[Global Setup] Starting Playwright global setup...');

  try {
    // Ensure test database is clean and has test data
    console.log('[Global Setup] Preparing test database...');
    
    // Remove existing database to ensure clean state
    try {
      await execAsync('rm -f ../database/main.db');
      console.log('[Global Setup] Removed existing database');
    } catch (error) {
      console.log('[Global Setup] No existing database to remove');
    }

    // Wait a moment for file system
    await new Promise(resolve => setTimeout(resolve, 1000));

    console.log('[Global Setup] Database preparation completed');

    // Create a browser instance for authentication state setup
    console.log('[Global Setup] Setting up browser authentication states...');
    
    const browser = await chromium.launch();
    const context = await browser.newContext();
    const page = await context.newPage();

    // Wait for the server to be ready
    let serverReady = false;
    let attempts = 0;
    const maxAttempts = 30;

    while (!serverReady && attempts < maxAttempts) {
      try {
        await page.goto('http://localhost:8080', { timeout: 5000 });
        serverReady = true;
        console.log('[Global Setup] Server is ready');
      } catch (error) {
        attempts++;
        console.log(`[Global Setup] Waiting for server... (attempt ${attempts}/${maxAttempts})`);
        await new Promise(resolve => setTimeout(resolve, 2000));
      }
    }

    if (!serverReady) {
      throw new Error('Server failed to start within timeout period');
    }

    // Create authenticated user state for tests
    console.log('[Global Setup] Creating authenticated user state...');
    
    try {
      // Navigate to login page
      await page.goto('http://localhost:8080/');
      
      // Wait for login form to be visible
      await page.waitForSelector('form', { timeout: 10000 });
      
      // Fill in test credentials
      await page.fill('input[name="identifier"]', 'alexchen');
      await page.fill('input[name="password"]', 'Aa123456');
      
      // Submit login form
      await page.click('button[type="submit"]');
      
      // Wait for successful login (redirect to home)
      await page.waitForURL('**/home', { timeout: 10000 });
      
      // Save authenticated state
      await context.storageState({ path: 'frontend-tests/setup/auth-state.json' });
      console.log('[Global Setup] Authenticated user state saved');
      
    } catch (error) {
      console.warn('[Global Setup] Could not create authenticated state:', error.message);
      // Continue without authenticated state - some tests can still run
    }

    await browser.close();

    // Create test data snapshots for consistent testing
    console.log('[Global Setup] Creating test data snapshots...');
    
    // Take a screenshot of the initial state for visual regression testing
    const snapshotBrowser = await chromium.launch();
    const snapshotContext = await snapshotBrowser.newContext();
    const snapshotPage = await snapshotContext.newPage();
    
    try {
      await snapshotPage.goto('http://localhost:8080/');
      await snapshotPage.screenshot({ 
        path: 'frontend-tests/setup/baseline-login.png',
        fullPage: true 
      });
      console.log('[Global Setup] Baseline login screenshot captured');
    } catch (error) {
      console.warn('[Global Setup] Could not capture baseline screenshot:', error.message);
    }
    
    await snapshotBrowser.close();

    console.log('[Global Setup] Global setup completed successfully');

  } catch (error) {
    console.error('[Global Setup] Global setup failed:', error);
    throw error;
  }
}

export default globalSetup;
