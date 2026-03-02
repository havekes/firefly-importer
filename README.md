# Firefly III Statement Importer

A Go web application to act as a smart middleman for Firefly III. The app parses CSV files and screenshots, deduplicates transactions against your actual Firefly instance, and pushes them seamlessly via the API.

## Prerequisites

- A running [Firefly III](https://www.firefly-iii.org/) instance and a [Personal Access Token](https://docs.firefly-iii.org/how-to/firefly-iii/features/api/#personal-access-tokens)
- An OpenAI-compatible Vision API for receipt/screenshot parsing ()

## Getting Started

The recommended way to run the application is using Docker compose.

Here's a base compose file you can use as a starting point:

```yaml
services:
  firefly-importer:
    image: ghcr.io/havekes/firefly-importer:latest
    restart: unless-stopped
    env_file: .env
    depends_on:
      - postgres
    ports:
      - "8080:8080"

  postgres:
    image: postgres:18
    restart: unless-stopped
    env_file: .env
    volumes:
      - postgres-data:/var/lib/postgresql/18/docker
    healthcheck:
      test: ["CMD-SHELL", "pg_isready", "-d", "${POSTGRES_DB}"]
      interval: 5m
      timeout: 30s
      retries: 3
      start_period: 120s

volumes:
  postgres-data:
```

Before running the application, you need to set up the environment variables.
Replace all variables in braces with proper values.

```ini
# firefly-importer
PORT="8080"
DATABASE_URL="postgres://{{ firefly_importer_db_user }}:{{ firefly_importer_db_password }}@postgres:5432/firefly_importer?sslmode=disable"
# Strong random 32 bit key
CSRF_KEY=

FIREFLY_URL=https://firefly.example.com/api/v1
FIREFLY_TOKEN=

VISION_API_URL=https://api.openai.com/v1
VISION_API_KEY=
VISION_API_MODEL=gpt-5-mini

# postgres
POSTGRES_USER={{ firefly_importer_db_user }}
POSTGRES_PASSWORD={{ firefly_importer_db_password }}
POSTGRES_DB=firefly_importer
```

## Running locally during development

To start the application in a Docker container, run:

```bash
docker compose up --build -d
```

This will build the Go binary using a multi-stage Dockerfile to keep the image slim, and start the server. 
You can access the UI at `http://localhost:8080` (or whichever port you specified in `.env`).

To view application logs in real-time, run:
```bash
docker compose logs -f
```

To stop the web server:
```bash
docker compose down
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
