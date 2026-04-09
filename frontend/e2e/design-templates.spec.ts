/**
 * E2E tests for design templates (issue #50).
 *
 * These tests require a running fabrik backend with a clean test database.
 * Run with: npx playwright test
 *
 * The base URL is configured via the PLAYWRIGHT_BASE_URL environment variable
 * (default: http://localhost:4200).
 */
import { test, expect } from '@playwright/test';

const BASE_URL = process.env.PLAYWRIGHT_BASE_URL ?? 'http://localhost:4200';

test.describe('New Design dialog — templates', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(BASE_URL);
  });

  test('shows 4 template options when dialog opens', async ({ page }) => {
    await page.getByRole('button', { name: /new design/i }).click();
    await expect(page.getByText('Start from template')).toBeVisible();
    await expect(page.getByRole('button', { name: /blank/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /2-stage clos/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /3-stage clos/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /pod-based fabric/i })).toBeVisible();
  });

  test('Blank is selected by default', async ({ page }) => {
    await page.getByRole('button', { name: /new design/i }).click();
    const blankBtn = page.getByRole('button', { name: /blank/i });
    await expect(blankBtn).toHaveAttribute('aria-pressed', 'true');
  });

  test('2-stage Clos: block exists in hierarchy after creation', async ({ page }) => {
    await page.getByRole('button', { name: /new design/i }).click();

    // Select 2-stage Clos template
    await page.getByRole('button', { name: /2-stage clos/i }).click();
    await expect(page.getByRole('button', { name: /2-stage clos/i })).toHaveAttribute(
      'aria-pressed',
      'true',
    );

    // Fill name and submit
    await page.getByLabel(/name/i).fill('E2E 2-stage test');
    await page.getByRole('button', { name: /create design/i }).click();

    // Should navigate to /design
    await page.waitForURL('**/design');

    // Hierarchy tree should contain Block 1
    await expect(page.getByText('Block 1')).toBeVisible({ timeout: 10_000 });
    // And the site and pod should also be visible
    await expect(page.getByText('Site 1')).toBeVisible();
    await expect(page.getByText('Pod A')).toBeVisible();
  });

  test('Blank: no hierarchy is created — hierarchy section is empty', async ({ page }) => {
    await page.getByRole('button', { name: /new design/i }).click();

    // Blank is default — no need to click
    await page.getByLabel(/name/i).fill('E2E blank test');
    await page.getByRole('button', { name: /create design/i }).click();

    await page.waitForURL('**/design');

    // No "Site 1" should appear (blank canvas)
    await expect(page.getByText('Site 1')).not.toBeVisible();
    await expect(page.getByText('Pod A')).not.toBeVisible();
    await expect(page.getByText('Block 1')).not.toBeVisible();
  });

  test('pod-based fabric: 4 blocks exist after creation', async ({ page }) => {
    await page.getByRole('button', { name: /new design/i }).click();

    await page.getByRole('button', { name: /pod-based fabric/i }).click();
    await page.getByLabel(/name/i).fill('E2E pod-based test');
    await page.getByRole('button', { name: /create design/i }).click();

    await page.waitForURL('**/design');

    // All four blocks should appear in the hierarchy
    await expect(page.getByText('Block 1')).toBeVisible({ timeout: 10_000 });
    await expect(page.getByText('Block 2')).toBeVisible();
    await expect(page.getByText('Block 3')).toBeVisible();
    await expect(page.getByText('Block 4')).toBeVisible();
    await expect(page.getByText('Pod A')).toBeVisible();
    await expect(page.getByText('Pod B')).toBeVisible();
  });
});
