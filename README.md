# Product Management API

## Overview
This is a Product Management API built using Go, PostgreSQL, Redis, RabbitMQ, and Gorilla Mux. The API allows users to manage products, filter them based on various parameters, and handle image processing tasks using a message queue.

## Prerequisites

Before starting, ensure you have the following installed on your system:
- [Go](https://golang.org/dl/) (at least version 1.19)
- [PostgreSQL](https://www.postgresql.org/download/)
- [Redis](https://redis.io/download)
- [RabbitMQ](https://www.rabbitmq.com/download.html)
- [Git](https://git-scm.com/downloads)

## Project Setup

### Step 1: Clone the Project
Open a terminal and run:
```bash
git clone https://github.com/yourusername/product-management.git
cd product-management
```

### Step 2: Install Dependencies
Run the following command to install Go dependencies:
```bash
go mod tidy
```

### Step 3: Set Up PostgreSQL Database
1. Open pgAdmin or your preferred PostgreSQL client.
2. Create a new database named `product_management`.
3. Run the following SQL commands to create the products table:
```sql
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    product_name TEXT NOT NULL,
    product_description TEXT,
    product_price FLOAT NOT NULL,
    product_images TEXT[],
    compressed_product_images TEXT[]
);
```

### Step 4: Set Up Redis
Start the Redis server.
By default, Redis runs on `localhost:6379`. Ensure Redis is up and running.

### Step 5: Set Up RabbitMQ
1. Start RabbitMQ.
2. Enable the RabbitMQ Management Plugin:
```bash
rabbitmq-plugins enable rabbitmq_management
```
3. Access the RabbitMQ dashboard at http://localhost:15672
   - Username: guest
   - Password: guest

### Step 6: Configure Environment Variables
Create a `.env` file in the project root with the following content:
```env
DATABASE_URL=postgres://<username>:<password>@localhost:5432/product_management
REDIS_ADDR=localhost:6379
```
Replace `<username>` and `<password>` with your PostgreSQL username and password.

### Step 7: Run the Application
Start the application by running:
```bash
go run cmd/main.go
```
The server will start on http://localhost:8080

## Features
- Product management (Create, Read, Update, Delete)
- Product filtering
- Image processing with message queue
- Caching with Redis

## Technologies Used
- Go
- PostgreSQL
- Redis
- RabbitMQ
- Gorilla Mux

## Contributing
1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request


## Contact
Your Name - your.email@example.com

Project Link: [https://github.com/yourusername/product-management](https://github.com/yourusername/product-management)

