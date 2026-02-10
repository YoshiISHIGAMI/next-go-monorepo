import { expect, test } from "@playwright/test";

test.describe("Authentication", () => {
  test("login page displays correctly", async ({ page }) => {
    await page.goto("/login");

    await expect(page.getByRole("heading", { name: "ログイン" })).toBeVisible();
    await expect(
      page.getByRole("button", { name: "Sign in with GitHub" }),
    ).toBeVisible();
  });

  test("unauthenticated user is redirected from /me to /login", async ({
    page,
  }) => {
    await page.goto("/me");

    // Should be redirected to login page
    await expect(page).toHaveURL("/login");
    await expect(page.getByRole("heading", { name: "ログイン" })).toBeVisible();
  });
});
