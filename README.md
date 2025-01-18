# Doctor

A lightweight URL health checker with email and Telegram notifications.

## Overview

Doctor is a simple monitoring tool that checks if specified URLs return successful (2xx) responses. If a check fails, it can notify you via email and/or Telegram. It's designed to be minimal and straightforward.

## Features

- Periodic HTTP health checks
- Configurable check intervals and timeouts
- Notification options:
  - Email (SMTP)
  - Telegram
- Docker support
- REST API for dynamic target management
- Prometheus metrics export

## Usage

```bash
# Run with default config
doctor

# Run with custom config location
doctor -config /path/to/config.json
```

## Configuration

Doctor uses a JSON configuration file. Example:

```json
{
    "checkIntervalInSec": 30,
    "checkTimeoutInSec": 10,
    "smtp": {
        "from": "alert@company.com",
        "password": "secret",
        "smtpHost": "webmail.company.com",
        "smtpPort": "587",
        "toEmails": [
            "admin.iscool@company.com"
        ],
        "startTLSAuth": true
    },
    "telegram": {
        "botToken": "xxx",
        "chatId": xxx,
        "throttleInSecs": 300
    },
    "targets": [
        {
            "id": "google",
            "url": "https://google.com"
        }
    ]
}
```

### Configuration Options

- `checkIntervalInSec`: Time between checks in seconds
- `checkTimeoutInSec`: HTTP request timeout in seconds
- `smtp`: Email notification settings
  - `from`: Sender email address
  - `password`: SMTP password
  - `smtpHost`: SMTP server hostname
  - `smtpPort`: SMTP server port
  - `toEmails`: List of recipient email addresses
  - `startTLSAuth`: Enable STARTTLS authentication
- `telegram`: Telegram notification settings
  - `botToken`: Telegram bot token
  - `chatId`: Target chat ID
  - `throttleInSecs`: Minimum time between notifications
- `targets`: List of URLs to monitor
  - `id`: Unique identifier for the target
  - `url`: URL to check

## REST API

Doctor provides a REST API for dynamic target management. The API runs on port 8080 by default.

### Endpoints

There is an `openapi.yml` so you can generate your client stubs, but they are also simple enough that I can just describe them here.

#### Register Target
```http
POST /register
Content-Type: application/json

{
    "id": "my-service",
    "url": "https://my-service.com"
}
```

Response (200 OK):
```json
{
    "message": "Target registered successfully"
}
```

#### Get Status
```http
GET /status
```

Response (200 OK):
```json
[
    {
        "id": "my-service",
        "url": "https://my-service.com",
        "status": 200,
        "healthy": true,
        "timestamp": "2025-01-18T10:30:00Z",
        "duration_seconds": 0.432
    }
]
```

## Prometheus Metrics

Doctor exposes metrics at `/metrics` in Prometheus format. Available metrics include:

- `doctor_health_check_duration_seconds`: Duration of health checks (histogram)
- `doctor_health_check_status`: Current health status of targets (gauge)
- `doctor_health_check_total`: Total number of health checks performed (counter)

## Docker

Build and run using Docker:

```bash
# Build
docker build -t doctor .

# Run
docker run -v $(pwd)/config.json:/app/config.json -p 8080:8080 doctor
```