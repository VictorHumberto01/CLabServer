# CLab Server

A server for managing programming classrooms, tasks, and code submissions.

## Prerequisites

- Go 1.24 or higher
- PostgreSQL 14 or higher
- GCC compiler (for code execution)

## Setup

1. Clone the repository:
```bash
git clone https://github.com/yourusername/clab-server.git
cd clab-server
```

2. Install dependencies:
```bash
go mod download
```

3. Set up the database:
```bash
# Create PostgreSQL database
createdb clab

# Run migrations
./scripts/migrate.sh
```

4. Configure environment variables:
Create a `.env` file in the project root with the following content:
```env
DATABASE_URL=postgres://localhost:5432/clab?sslmode=disable
JWT_SECRET=your-secret-key
ENV=development
SERVER_PORT=8080
COMPILER_PATH=/usr/bin/gcc
MAX_CODE_SIZE=1048576
MAX_MEMORY_USAGE=268435456
```

## Running the Server

### Development Mode
```bash
./scripts/dev.sh
```

### Production Mode
```bash
./scripts/prod.sh
```

## API Documentation

### Authentication

#### Register User
- **POST** `/api/auth/register`
- **Body**:
```json
{
    "email": "user@example.com",
    "password": "password123",
    "name": "John Doe",
    "role": "student" // or "teacher" or "admin"
}
```

#### Login
- **POST** `/api/auth/login`
- **Body**:
```json
{
    "email": "user@example.com",
    "password": "password123"
}
```

### Admin Routes

#### Register Teacher
- **POST** `/api/admin/teachers`
- **Headers**: `Authorization: Bearer <token>`
- **Body**: Same as register user

#### List All Rooms
- **GET** `/api/admin/rooms`
- **Headers**: `Authorization: Bearer <token>`

### Teacher Routes

#### Create Room
- **POST** `/api/teacher/rooms`
- **Headers**: `Authorization: Bearer <token>`
- **Body**:
```json
{
    "name": "Introduction to Programming",
    "description": "Learn the basics of programming",
    "code": "INTRO101"
}
```

#### List Teacher's Rooms
- **GET** `/api/teacher/rooms`
- **Headers**: `Authorization: Bearer <token>`

#### Create Task
- **POST** `/api/teacher/rooms/{room_id}/tasks`
- **Headers**: `Authorization: Bearer <token>`
- **Body**:
```json
{
    "title": "Hello World",
    "description": "Write a program that prints 'Hello, World!'",
    "test_cases": [
        {
            "input": "",
            "expected_output": "Hello, World!"
        }
    ]
}
```

#### List Tasks in Room
- **GET** `/api/teacher/rooms/{room_id}/tasks`
- **Headers**: `Authorization: Bearer <token>`

#### List Submissions for Task
- **GET** `/api/teacher/tasks/{task_id}/submissions`
- **Headers**: `Authorization: Bearer <token>`

### Student Routes

#### Join Room
- **POST** `/api/student/rooms/join`
- **Headers**: `Authorization: Bearer <token>`
- **Body**:
```json
{
    "room_code": "INTRO101"
}
```

#### List Student's Rooms
- **GET** `/api/student/rooms`
- **Headers**: `Authorization: Bearer <token>`

#### List Tasks in Room
- **GET** `/api/student/rooms/{room_id}/tasks`
- **Headers**: `Authorization: Bearer <token>`

#### Submit Solution
- **POST** `/api/student/tasks/{task_id}/submit`
- **Headers**: `Authorization: Bearer <token>`
- **Body**:
```json
{
    "code": "#include <stdio.h>\n\nint main() {\n    printf(\"Hello, World!\\n\");\n    return 0;\n}"
}
```

#### List Student's Submissions
- **GET** `/api/student/tasks/{task_id}/submissions`
- **Headers**: `Authorization: Bearer <token>`

### Health Check
- **GET** `/health`
- Returns "OK" if the server is running

## Development

### Running Tests
```bash
./scripts/test.sh
```

### Linting
```bash
./scripts/lint.sh
```

### Building
```bash
./scripts/build.sh
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

