# Apartment Service

A comprehensive apartment management system that allows property managers to manage apartments, residents, and bills efficiently. The service includes Telegram integration for seamless resident invitations and communication.

## Features

### For Managers
- **User Management**: View, retrieve, and delete users
- **Apartment Management**: Create, update, delete apartments and manage residents
- **Bill Management**: Create bills with image attachments, set due dates, and track payments
- **Resident Invitations**: Invite residents via Telegram username
- **Comprehensive Oversight**: View all apartments and their associated residents

### For Residents
- **Profile Management**: View and update personal information
- **Apartment Participation**: Join and leave apartments
- **Bill Handling**: View unpaid bills, make individual or batch payments
- **Payment History**: Track complete payment history

### Key Capabilities
- JWT-based authentication with role-based access control
- Telegram bot integration for user invitations
- Multi-unit apartment support
- Bill categorization (water, electricity, etc.)
- Image upload support for bills(with minio)
- Automatic bill division among residents
- Payment tracking and history

## Prerequisites

- Docker
- Docker Compose
- Telegram Bot Token

## Setup

### 1. Clone the Repository
```bash
git clone <repository-url>
cd apartment-service
```

### 2. Configure Telegram Bot
1. Create a new bot on Telegram by messaging [@BotFather](https://t.me/botfather)
2. Get your bot token
3. Replace the placeholder telegram token in your configuration files with your actual token

### 3. Build and Run

```bash
# Build the Docker image
docker build -t apartment .

# Start the services
docker compose up -d
```

## API Documentation

The service provides a comprehensive REST API with the following main endpoints:

### Authentication
- `POST /user/signup` - User registration
- `POST /user/login` - User authentication

### Manager Endpoints
- User management: `/manager/user/*`
- Apartment management: `/manager/apartment/*`
- Bill management: `/manager/bill/*`
- Resident invitations: `/manager/apartment/{apartment-id}/invite/resident/{telegram-username}`

### Resident Endpoints
- Profile management: `/resident/profile`
- Apartment participation: `/resident/apartment/join`, `/resident/apartment/leave`
- Bill operations: `/resident/bills/*`

## User Types

### Manager
- Can create and manage multiple apartments
- Full control over residents and bills
- Can invite residents via Telegram
- Access to all apartment data and analytics

### Resident
- Can join apartments (via invitation or request)
- View and pay bills
- Manage personal profile
- Access payment history

## Bill Management

The system supports:
- Multiple bill types (water, electricity, gas, etc.)
- Image attachments for bill documentation
- Due date tracking
- Automatic division among apartment residents
- Batch payment processing
- Payment history tracking

## Authentication

The service uses JWT (JSON Web Tokens) for authentication:
- Include the token in the `Authorization` header as `Bearer <token>`
- Tokens contain user ID and user type for role-based access control
- Different endpoints require different user types (manager vs resident)

## Development

### Project Structure
```
apartment-service/
├── docker-compose.yaml
├── Dockerfile
├── main.go
├── cmd/
├── config/
├── internal/
└── README.md
```