# Workshop Management Service

## Introduction

This repository contains the Workshop Management Service, an application built with Gorilla Mux in Go. It is designed to handle the CRUD (Create, Read, Update, Delete) operations for workshop listings. The application also integrates with RabbitMQ for logging purposes and uses Amazon DynamoDB for data storage.

## Features

- **CRUD Operations**: Manage workshop listings with full create, read, update, and delete capabilities.
- **RabbitMQ Integration**: Sends logs to RabbitMQ for processing or storage by other services.
- **DynamoDB Storage**: Utilizes Amazon DynamoDB for persistent storage of workshop data.

## Getting Started

Follow these instructions to get a copy of the project up and running on your local machine for development and testing purposes.

### Prerequisites

- Go (Golang)
- Access to Amazon DynamoDB
- Access to a RabbitMQ instance

### Installation

1. Clone the repository:
   ```
   git clone https://github.com/GreenHarbor/workshop.git
   ```
2. Navigate to the project directory:
   ```
   cd workshop
   ```
3. Install the necessary Go modules (if any):
   ```
   go mod tidy
   ```

### Configuration

- Set up your AWS credentials to allow access to DynamoDB.
- Configure the connection details for your RabbitMQ instance.

### Running the Application

1. To start the service, run:
   ```
   go run main.go
   ```
2. The service will start and begin listening for HTTP requests to handle workshop listings.
