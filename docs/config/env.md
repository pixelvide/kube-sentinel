# Environment Variables

Kube Sentinel supports several environment variables to customize its behavior.

## Core Configuration

- **PORT**: Port on which Kube Sentinel runs, default value is `8080`.
- **HOST**: Used for generating OAuth 2.0 authorization callback addresses. Usually detected from request headers, but can be set manually (e.g., `https://kube-sentinel.example.com`).
- **KUBE_SENTINEL_BASE**: Base path for the application. If set to `/kube-sentinel`, the application will be accessible at `domain.com/kube-sentinel`.
- **JWT_SECRET**: Secret key used for signing and verifying JWT tokens. **Must be changed in production!**
- **KUBE_SENTINEL_ENCRYPT_KEY**: Secret key used for encrypting sensitive data (user passwords, tokens, kubeconfigs). **Must be changed in production!**

## Database Configuration

- **DB_TYPE**: The type of database to use. Supported: `sqlite`, `mysql`, `postgres`. Default is `sqlite`.
- **DB_DSN**: Database connection DSN.
    - For `sqlite`: Path to the database file (e.g., `data/kube-sentinel.db`). Default is `dev.db`.
    - For `mysql`/`postgres`: Standard connection string (e.g., `user:pass@tcp(host:3306)/dbname`).
- **DB_SCHEMA_CORE**: PostgreSQL schema name for core tables (e.g., `apps`, `app_configs`). Defaults to `public`.
- **DB_SCHEMA_APP**: PostgreSQL schema name for application-specific tables (e.g., `users`, `clusters`, `audit_logs`). Defaults to `public`.

## Authentication & Authorization

- **KUBE_SENTINEL_USERNAME**: Set the initial administrator username during bootstrap.
- **KUBE_SENTINEL_PASSWORD**: Set the initial administrator password during bootstrap.
- **KUBECONFIG**: Path to the initial Kubernetes configuration file. Default is `~/.kube/config`. Clusters from this config will be discovered and imported on the first run.

## Third-party Integrations

- **GITLAB_HOSTS**: Comma-separated list of GitLab hosts to pre-configure (e.g., `https://gitlab.com,http://my-gitlab.local`). These will be seeded into the database on startup.

## AI & LLM Configuration

- **AI_ALLOW_USER_KEYS**: Allow users to provide their own API keys for AI services. Default is `true`.
- **AI_FORCE_USER_KEYS**: Force users to provide their own API keys; system-wide keys will not be used. Default is `false`.

## Specialized Settings

- **NODE_TERMINAL_IMAGE**: Docker image used for the Node Terminal Agent. Default is `busybox:latest`.
- **DISABLE_GZIP**: Disable GZIP compression for API responses. Default is `true`.
- **DISABLE_VERSION_CHECK**: Disable the automatic check for new application versions. Default is `false`.
- **DISABLE_CACHE**: Disable the Kubernetes client-side cache. Default is `false`.
- **INSECURE_SKIP_VERIFY**: Disable SSL certificate verification for OAuth providers. Dangerous! Use only in development or if you trust the network. Default is `false`.
