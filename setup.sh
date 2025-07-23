#!/bin/bash

echo "🚀 Setting up First Draft In Progress (FDIP)"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21+ first."
    exit 1
fi

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo "❌ Node.js is not installed. Please install Node.js 18+ first."
    exit 1
fi

# Check if MariaDB/MySQL is installed
if ! command -v mysql &> /dev/null; then
    echo "⚠️  MariaDB/MySQL is not installed. Please install it first."
    echo "   You can still set up the Go backend and React frontend."
fi

echo "📦 Setting up backend..."

# Setup backend
cd backend
go mod tidy

# Create .env file if it doesn't exist
if [ ! -f .env ]; then
    cp env.example .env
    echo "✅ Created .env file. Please configure your database and Stripe settings."
fi

cd ..

echo "📦 Setting up frontend..."

# Setup frontend
cd frontend
npm install

# Create .env file if it doesn't exist
if [ ! -f .env ]; then
    cp env.example .env
    echo "✅ Created .env file. Please configure your API URL and Stripe settings."
fi

cd ..

echo ""
echo "🎉 Setup complete!"
echo ""
echo "Next steps:"
echo "1. Configure your database settings in backend/.env"
echo "2. Configure your Stripe settings in both backend/.env and frontend/.env"
echo "3. Create a MariaDB database named 'fdip'"
echo "4. Run the database migrations: cd backend && go run main.go"
echo "5. Start the backend: cd backend && go run main.go"
echo "6. Start the frontend: cd frontend && npm start"
echo ""
echo "Default admin credentials:"
echo "Username: admin"
echo "Password: admin123"
echo ""
echo "Happy coding! 📚✍️" 