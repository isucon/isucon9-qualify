# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ISUCON9-qualify is the qualification round application for ISUCON9 (Iikanjini Speed Up Contest), a performance tuning competition. The application "ISUCARI" (椅子カリ) is a marketplace for buying and selling chairs.

## Key Commands

### Benchmarker
```bash
# Build benchmarker and external services
make

# Run benchmarker against target
./bin/benchmarker -target-url https://203.0.113.1 -target-host isucari.t.isucon.pw \
  -data-dir initial-data/ -static-dir webapp/public/static/ \
  -payment-url https://bp.t.isucon.pw -shipment-url https://bs.t.isucon.pw
```

### Application Setup
```bash
# Initialize data (run from root)
make init

# Run application with Docker
cd webapp
docker compose up

# Run Go implementation directly
cd webapp/go
make
./isucari
```

### Database Access
```bash
# Connect to MySQL (when using Docker)
docker compose exec mysql mysql -uroot -proot isucari
```

## Architecture

### Core Components
1. **Main Application** (`webapp/`): E-commerce platform with multiple language implementations (Go, Ruby, Node.js, PHP, Perl, Python)
2. **Frontend** (`webapp/frontend/`): React TypeScript application
3. **Benchmarker** (`bench/`): Load testing tool that simulates user behavior
4. **External Services**: Payment (`cmd/payment/`) and Shipment (`cmd/shipment/`) services

### Database Schema
- `users`: User accounts with bcrypt passwords
- `items`: Listed items for sale
- `transaction_evidences`: Purchase records
- `shippings`: Shipping information
- `categories`: Item categories (hierarchical)

### Key API Endpoints
- `POST /initialize`: Reset application state for benchmarking
- `POST /sell`: List new item
- `POST /buy`: Purchase item  
- `POST /ship`: Request shipping
- `POST /ship_done`: Mark as shipped
- `POST /complete`: Complete transaction
- `GET /users/transactions.json`: User's transaction history
- `GET /items/{id}.json`: Item details
- `GET /new_items/{root_category_id}.json`: Latest items by category

### Performance Considerations
- The benchmarker evaluates response times, error rates, and data consistency
- External API calls to payment/shipment services are performance bottlenecks
- Image serving optimization is critical (1000+ chair images)
- Database queries need careful indexing and optimization
- Session management impacts concurrency

### Go Implementation Details
The Go webapp uses:
- Chi router for HTTP routing
- SQLx for database access with prepared statements
- Gorilla sessions for session management
- bcrypt for password hashing
- Standard library for JSON and image handling

When optimizing, focus on:
1. Database query optimization (N+1 queries, missing indexes)
2. Caching strategies for categories, items, and images
3. External API call optimization (batching, caching)
4. Static file serving optimization
5. Connection pooling and concurrency tuning