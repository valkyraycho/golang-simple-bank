# Simple Bank System

## System Architecture

The application follows a clean microservices architecture with:

-   gRPC API and gateway service for both HTTP and RPC communication
-   PostgreSQL for persistent data storage
-   Redis for asynchronous task handling
-   Docker for containerization
-   GitHub Actions for CI/CD

## Tech Stack

### Backend Services

-   Go (Golang) - Main programming language
-   SQLC - Type-safe SQL query generator
-   Postgres - Primary database
-   Redis - Caching layer
-   Docker & Docker Compose - Containerization
-   Swagger/OpenAPI - API documentation

### API Layer

#### HTTP REST API

-   Account management
-   User authentication
-   Transaction processing
-   RESTful endpoints documented via Swagger/OpenAPI

#### gRPC API

-   High-performance internal service communication
-   Protocol Buffers for serialization

## Features

### Transaction Processing

-   Money transfers between accounts
-   Transaction history tracking
-   Balance updates
-   Transaction validation

### Security

-   Password hashing and encryption
-   PASETO based authentication
-   Role-based access control
-   Session management
-   Secure API endpoints

### Additional Features

-   Real-time balance updates
-   Transaction logging
-   Data validation
-   Error handling
