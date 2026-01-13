# Development Workflow

This document outlines the standard development workflow for Cloud Sentinel.

## Docker-based Development

As of current development sessions, we use **Docker Compose** as the primary method for running and building the application. Avoid using local `npm` or `go` commands directly unless troubleshooting specific local environment issues.

### Key Commands

- **Build and Start (Full Stack)**:
  ```bash
  docker compose up -d --build
  ```

- **Stop Services**:
  ```bash
  docker compose down
  ```

- **View Logs**:
  ```bash
  docker compose logs -f
  ```

- **Restart a Specific Service**:
  ```bash
  docker compose up -d --build frontend
  # or
  docker compose up -d --build backend
  ```

### Development Environment

- **Frontend**: Accessible at `http://localhost:3000`
- **Backend**: Proxied through the frontend at `http://localhost:3000/api`
- **Configuration**: Managed via `.env` in the root directory.

### Session Persistence

When performing code changes, always use `docker compose up -d --build` to ensure changes are reflected in the running containers.
