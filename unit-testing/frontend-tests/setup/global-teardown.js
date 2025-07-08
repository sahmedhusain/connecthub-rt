/**
 * Playwright Global Teardown
 * Cleans up after all E2E tests have completed
 */

import { exec } from 'child_process';
import { promisify } from 'util';
import fs from 'fs/promises';

const execAsync = promisify(exec);

async function globalTeardown() {
  console.log('[Global Teardown] Starting Playwright global teardown...');

  try {
    // Clean up authentication state files
    console.log('[Global Teardown] Cleaning up authentication state...');
    try {
      await fs.unlink('frontend-tests/setup/auth-state.json');
      console.log('[Global Teardown] Authentication state file removed');
    } catch (error) {
      console.log('[Global Teardown] No authentication state file to remove');
    }

    // Clean up temporary test files
    console.log('[Global Teardown] Cleaning up temporary test files...');
    try {
      await fs.rm('frontend-tests/setup/temp', { recursive: true, force: true });
      console.log('[Global Teardown] Temporary files cleaned up');
    } catch (error) {
      console.log('[Global Teardown] No temporary files to clean up');
    }

    // Archive test results if in CI environment
    if (process.env.CI) {
      console.log('[Global Teardown] Archiving test results for CI...');
      try {
        const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
        await execAsync(`tar -czf test-results-${timestamp}.tar.gz test-reports/`);
        console.log('[Global Teardown] Test results archived');
      } catch (error) {
        console.warn('[Global Teardown] Could not archive test results:', error.message);
      }
    }

    // Generate test summary
    console.log('[Global Teardown] Generating test summary...');
    try {
      const summary = {
        timestamp: new Date().toISOString(),
        environment: {
          ci: !!process.env.CI,
          node_version: process.version,
          platform: process.platform,
        },
        test_run_completed: true,
      };

      await fs.writeFile(
        'test-reports/test-summary.json',
        JSON.stringify(summary, null, 2)
      );
      console.log('[Global Teardown] Test summary generated');
    } catch (error) {
      console.warn('[Global Teardown] Could not generate test summary:', error.message);
    }

    console.log('[Global Teardown] Global teardown completed successfully');

  } catch (error) {
    console.error('[Global Teardown] Global teardown failed:', error);
    // Don't throw error in teardown to avoid masking test failures
  }
}

export default globalTeardown;
