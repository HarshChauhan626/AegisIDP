# AegisIDP Local Running Guide

This guide provides step-by-step instructions for running Platform Orchestrator (AegisIDP) on your local machine.

There are two primary ways to run this project:
1. **Using Docker Compose** (Recommended for the easiest setup)
2. **Running Manually** (Best for development and debugging)

---

## Prerequisites

Depending on how you choose to run the project, you will need:
- **Docker** and **Docker Compose** (if using Option 1)
- **Go 1.26+** (if running the backend manually)
- **Node.js** (v20+ recommended) and **npm** (if running the frontend manually)
- **Make** (optional, but convenient as the project provides a `Makefile`)

---

## Option 1: Running via Docker Compose

This is the simplest method as it builds and orchestrates both the backend and frontend in containers.

1. **Open a terminal** and navigate to the project root directory.
2. **Run the following command**:
   ```bash
   make dev
   ```
   *(Alternatively, if you don't have Make installed, run: `docker-compose up --build`)*
3. **Wait for the containers to build and start.** You should see logs for both the `frontend` and `backend` services.
4. **Access the application:**
   - Frontend UI: [http://localhost:3000](http://localhost:3000)
   - Backend API: [http://localhost:8080](http://localhost:8080)

To stop the services, press `Ctrl+C` in the terminal, or run `docker-compose down` if running in detached mode.

---

## Option 2: Running Manually (For Development)

If you are actively developing and want to run the backend and frontend natively on your machine, follow these steps. You will need two separate terminal windows.

### Step 1: Start the Backend

1. Open your first terminal and navigate to the project root.
2. If this is your first time, you may want to download Go dependencies:
   ```bash
   make tidy
   ```
   *(Or manually: `cd backend && go mod tidy`)*
3. Start the Go backend:
   ```bash
   make backend-dev
   ```
   *(Or manually: `cd backend && go run ./cmd/main.go`)*
4. The backend will start on `http://localhost:8080`. SQLite will automatically be created in `backend/data/platform.db`.

### Step 2: Start the Frontend

1. Open your second terminal and navigate to the project root.
2. Navigate to the frontend directory and install dependencies:
   ```bash
   cd frontend
   npm install
   ```
3. Start the Next.js development server:
   ```bash
   npm run dev
   ```
   *(Or from the project root using Make: `make frontend-dev`)*
4. The frontend will start on `http://localhost:3000`.

---

## Verification

Once the services are running, open your web browser and go to:
**[http://localhost:3000](http://localhost:3000)**

You should see the Platform Orchestrator dashboard. You can begin interacting with it to trigger workflows and provision environments.

### Useful Commands Reference (Makefile)

The project includes a `Makefile` with several helpful commands:
- `make dev` - Start all services via Docker Compose
- `make backend-dev` - Run backend locally (without Docker)
- `make frontend-dev` - Run frontend locally (without Docker)
- `make build` - Build backend binary + frontend production bundle
- `make test` - Run backend unit tests
- `make migrate` - Run DB migrations (auto-migrate via GORM on startup)
- `make lint` - Lint backend + frontend
- `make tidy` - Tidy go modules

## Troubleshooting

- **Port conflicts:** If you get an error that a port is already in use, make sure you don't have other services running on port `3000` (Frontend) or `8080` (Backend).
- **Database issues:** If you want to start fresh with a clean state, simply delete the SQLite database file at `backend/data/platform.db` and restart the backend.
