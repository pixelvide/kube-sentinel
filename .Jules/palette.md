## 2024-05-23 - [Settings Button Accessibility]
**Learning:** Icon-only buttons (like `Settings2` in `ResourceTable`) MUST have accessible names and visual tooltips for clarity.
**Action:** Always wrap icon buttons in `Tooltip` and provide `aria-label`. Use translation keys for consistency.

## 2024-05-23 - [Frontend Redirect Logic]
**Learning:** The frontend forces redirect to `/settings` if no clusters are configured for admin users, blocking access to other routes like `/pods`.
**Action:** When testing components on protected routes, either mock the cluster context or temporarily disable the redirect logic.
