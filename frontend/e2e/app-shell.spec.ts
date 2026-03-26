import { test, expect } from '@playwright/test';

test.describe('App Shell', () => {
  test('has correct title', async ({ page }) => {
    await page.goto('/');
    await expect(page).toHaveTitle(/fabrik/i);
  });

  test('displays the app toolbar with fabrik branding', async ({ page }) => {
    await page.goto('/');
    await expect(page.getByText('fabrik').first()).toBeVisible();
  });

  test('sidebar shows all navigation links', async ({ page }) => {
    await page.goto('/');
    await expect(page.getByText('Home')).toBeVisible();
    await expect(page.getByText('Design')).toBeVisible();
    await expect(page.getByText('Catalog')).toBeVisible();
    await expect(page.getByText('Metrics')).toBeVisible();
    await expect(page.getByText('Knowledge Base')).toBeVisible();
  });

  test('navigates to Design page', async ({ page }) => {
    await page.goto('/');
    await page.getByText('Design').click();
    await expect(page).toHaveURL(/\/design/);
    await expect(page.getByRole('heading', { name: 'Design' })).toBeVisible();
  });

  test('navigates to Catalog page', async ({ page }) => {
    await page.goto('/');
    await page.getByText('Catalog').click();
    await expect(page).toHaveURL(/\/catalog/);
    await expect(page.getByRole('heading', { name: 'Catalog' })).toBeVisible();
  });

  test('navigates to Metrics page', async ({ page }) => {
    await page.goto('/');
    await page.getByText('Metrics').click();
    await expect(page).toHaveURL(/\/metrics/);
    await expect(page.getByRole('heading', { name: 'Metrics' })).toBeVisible();
  });

  test('navigates to Knowledge Base page', async ({ page }) => {
    await page.goto('/');
    await page.getByText('Knowledge Base').click();
    await expect(page).toHaveURL(/\/knowledge/);
    await expect(page.getByRole('heading', { name: 'Knowledge Base' })).toBeVisible();
  });

  test('hamburger menu toggle works', async ({ page }) => {
    await page.goto('/');
    const menuBtn = page.getByRole('button', { name: /toggle navigation menu/i });
    await expect(menuBtn).toBeVisible();
    await menuBtn.click();
    // Toggle again
    await menuBtn.click();
  });

  test('dark mode toggle persists theme', async ({ page }) => {
    await page.goto('/');
    const toggleBtn = page.getByRole('button', { name: /toggle dark mode/i });
    await expect(toggleBtn).toBeVisible();
    await toggleBtn.click();
    // Verify dark theme applied
    const theme = await page.evaluate(() =>
      document.body.getAttribute('data-theme'),
    );
    expect(theme).toBe('dark');
    // Reload and verify persistence
    await page.reload();
    const themeAfterReload = await page.evaluate(() =>
      document.body.getAttribute('data-theme'),
    );
    expect(themeAfterReload).toBe('dark');
  });
});

test.describe('Dashboard', () => {
  test('shows welcome heading', async ({ page }) => {
    await page.goto('/');
    await expect(page.getByRole('heading', { name: 'Welcome to fabrik' })).toBeVisible();
  });

  test('shows quick actions', async ({ page }) => {
    await page.goto('/');
    await expect(page.getByText('New Design')).toBeVisible();
    await expect(page.getByText('Browse Catalog')).toBeVisible();
    await expect(page.getByText('View Metrics')).toBeVisible();
  });

  test('shows recent designs section', async ({ page }) => {
    await page.goto('/');
    await expect(page.getByText('Recent Designs')).toBeVisible();
  });
});
