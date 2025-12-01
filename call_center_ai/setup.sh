#!/bin/bash

# Call Center AI - Automated Setup Script
# Hỗ trợ: Linux và macOS

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
print_header() {
    echo -e "${BLUE}================================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}================================================${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Main setup
print_header "🚀 CALL CENTER AI - AUTOMATED SETUP"

echo ""
print_info "This script will:"
echo "  1. Check system requirements"
echo "  2. Setup Python virtual environment"
echo "  3. Install dependencies"
echo "  4. Setup MySQL database"
echo "  5. Configure environment variables"
echo "  6. Initialize scenarios"
echo "  7. Test the system"
echo ""

read -p "Continue with setup? (y/n) " -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    print_warning "Setup cancelled"
    exit 0
fi

# Step 1: Check requirements
print_header "Step 1: Checking System Requirements"

# Check Python
if command_exists python3; then
    PYTHON_VERSION=$(python3 --version | cut -d' ' -f2)
    print_success "Python found: $PYTHON_VERSION"
else
    print_error "Python 3 not found. Please install Python 3.11 or higher"
    exit 1
fi

# Check pip
if command_exists pip3; then
    print_success "pip3 found"
else
    print_error "pip3 not found. Please install pip3"
    exit 1
fi

# Check MySQL
if command_exists mysql; then
    MYSQL_VERSION=$(mysql --version | cut -d' ' -f3)
    print_success "MySQL found: $MYSQL_VERSION"
else
    print_warning "MySQL not found. You'll need to install MySQL 8.0+"
    echo "  Ubuntu/Debian: sudo apt install mysql-server"
    echo "  macOS: brew install mysql"
    read -p "Continue anyway? (y/n) " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Step 2: Setup virtual environment
print_header "Step 2: Setting up Python Virtual Environment"

if [ -d "venv" ]; then
    print_warning "Virtual environment already exists"
    read -p "Remove and recreate? (y/n) " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm -rf venv
        python3 -m venv venv
        print_success "Virtual environment recreated"
    fi
else
    python3 -m venv venv
    print_success "Virtual environment created"
fi

# Activate virtual environment
source venv/bin/activate
print_success "Virtual environment activated"

# Step 3: Install dependencies
print_header "Step 3: Installing Python Dependencies"

print_info "Upgrading pip..."
pip install --upgrade pip >/dev/null 2>&1

print_info "Installing requirements (this may take a few minutes)..."
pip install -r requirements.txt

print_success "All dependencies installed"

# Step 4: Setup MySQL database
print_header "Step 4: Setting up MySQL Database"

if command_exists mysql; then
    echo ""
    print_info "MySQL Database Setup"
    echo "Please provide MySQL root credentials:"
    read -p "MySQL root password: " -s MYSQL_ROOT_PASSWORD
    echo ""
    
    DB_NAME="call_center_db"
    DB_USER="callcenter"
    DB_PASSWORD=$(openssl rand -base64 12 2>/dev/null || echo "CallCenter2024!")
    
    print_info "Creating database and user..."
    
    mysql -u root -p"$MYSQL_ROOT_PASSWORD" <<EOF 2>/dev/null
CREATE DATABASE IF NOT EXISTS $DB_NAME CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS '$DB_USER'@'localhost' IDENTIFIED BY '$DB_PASSWORD';
GRANT ALL PRIVILEGES ON $DB_NAME.* TO '$DB_USER'@'localhost';
FLUSH PRIVILEGES;
EOF
    
    if [ $? -eq 0 ]; then
        print_success "Database created successfully"
        print_info "Database: $DB_NAME"
        print_info "User: $DB_USER"
        print_info "Password: $DB_PASSWORD"
    else
        print_error "Failed to create database"
        print_warning "You may need to create it manually"
    fi
else
    print_warning "MySQL not available - skipping database setup"
    print_info "You'll need to setup MySQL manually"
fi

# Step 5: Configure environment variables
print_header "Step 5: Configuring Environment Variables"

if [ -f ".env" ]; then
    print_warning ".env file already exists"
    read -p "Overwrite? (y/n) " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "Keeping existing .env file"
    else
        cp .env.example .env
        print_success "Created new .env file from template"
    fi
else
    cp .env.example .env
    print_success "Created .env file from template"
fi

if [ ! -z "$DB_PASSWORD" ]; then
    # Update database credentials in .env
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        sed -i '' "s/DB_PASSWORD=.*/DB_PASSWORD=$DB_PASSWORD/" .env
        sed -i '' "s/DB_USER=.*/DB_USER=$DB_USER/" .env
        sed -i '' "s/DB_NAME=.*/DB_NAME=$DB_NAME/" .env
    else
        # Linux
        sed -i "s/DB_PASSWORD=.*/DB_PASSWORD=$DB_PASSWORD/" .env
        sed -i "s/DB_USER=.*/DB_USER=$DB_USER/" .env
        sed -i "s/DB_NAME=.*/DB_NAME=$DB_NAME/" .env
    fi
    print_success "Updated database credentials in .env"
fi

echo ""
print_warning "IMPORTANT: You need to configure these in .env:"
echo "  - TWILIO_ACCOUNT_SID"
echo "  - TWILIO_AUTH_TOKEN"
echo "  - TWILIO_PHONE_NUMBER"
echo "  - OPENAI_API_KEY or ANTHROPIC_API_KEY"
echo ""
read -p "Press Enter to open .env in editor (or Ctrl+C to skip)"
${EDITOR:-nano} .env

# Step 6: Initialize scenarios
print_header "Step 6: Initializing Database & Scenarios"

python init_scenarios.py

if [ $? -eq 0 ]; then
    print_success "Scenarios initialized"
else
    print_warning "Failed to initialize scenarios"
fi

# Step 7: Test system
print_header "Step 7: Testing System"

print_info "Running system tests..."
python test_system.py

if [ $? -eq 0 ]; then
    print_success "All tests passed!"
else
    print_warning "Some tests failed. Check configuration."
fi

# Final instructions
print_header "🎉 Setup Complete!"

echo ""
print_success "Your Call Center AI system is ready!"
echo ""
print_info "Next steps:"
echo ""
echo "1. Start the server:"
echo "   ${GREEN}python main.py${NC}"
echo ""
echo "2. In a new terminal, expose with ngrok:"
echo "   ${GREEN}ngrok http 8000${NC}"
echo ""
echo "3. Configure Twilio webhook with ngrok URL:"
echo "   Voice webhook: https://your-url.ngrok.io/voice/incoming"
echo ""
echo "4. Test by calling your Twilio number!"
echo ""
print_info "For detailed documentation, see:"
echo "  - README.md (full documentation)"
echo "  - QUICKSTART.md (quick start guide)"
echo "  - PROJECT_STRUCTURE.md (architecture details)"
echo ""
print_warning "Remember to activate virtual environment next time:"
echo "  ${YELLOW}source venv/bin/activate${NC}"
echo ""

# Deactivate is not needed as the script will exit
