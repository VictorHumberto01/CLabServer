# 🖥️ CLab Server - Backend Engine

Servidor backend do CLab responsável pela compilação segura de código C, execução em sandbox, integração com IA para feedback educacional e gerenciamento de dados.

![CLab Server](https://img.shields.io/badge/CLab-Server%20Backend-green?style=for-the-badge)
![Go](https://img.shields.io/badge/Go-1.24+-blue?style=flat-square&logo=go)

## 🎯 Visão Geral

O CLab Server é o núcleo do sistema de ensino de programação C, fornecendo:
- **Compilação segura** de código C em ambiente isolado
- **Feedback inteligente** via IA local (LLaMA/Ollama)
- **API REST** para comunicação com o frontend Electron
- **Gerenciamento de dados** com PostgreSQL

## 🏗️ Arquitetura

```
┌─────────────────────────────────────────────┐
│              CLab Server                    │
├─────────────────────────────────────────────┤
│  🌐 API Gateway (Go)                        │
│  • Roteamento de requisições               │
│  • Autenticação e middleware               │
│  • Rate limiting e validação               │
├─────────────────────────────────────────────┤
│  ⚙️ Compiler Service (Go)                   │
│  • Compilação de código C                  │
│  • Execução em sandbox                     │
│  • Captura de stdout/stderr                │
├─────────────────────────────────────────────┤
│  🧠 AI Service (Python)                     │
│  • Integração LLaMA via Ollama             │
│  • Análise de erros de compilação          │
│  • Geração de feedback educativo           │
├─────────────────────────────────────────────┤
│  💾 Database Service (Go + PostgreSQL)      │
│  • Gerenciamento de usuários               │
│  • Armazenamento de tarefas                │
│  • Histórico de submissões                 │
└─────────────────────────────────────────────┘
```

## 📁 Estrutura do Projeto

```
clab-server/
├── cmd/
│   └── server/              
│       └── main.go             # Ponto de entrada da aplicação
├── internal/
│   ├── api/                    # Handlers e rotas da API
│   │   ├── handlers/           # Controllers REST
│   │   ├── middleware/         # Middlewares HTTP
│   │   └── routes/             # Definição de rotas
│   ├── compiler/               # Serviço de compilação
│   │   ├── sandbox/            # Sistema de sandbox
│   │   ├── executor/           # Executor de código C
│   │   └── validator/          # Validação de código
│   ├── database/               # Camada de dados
│   │   ├── migrations/         # Scripts de migração
│   │   ├── models/             # Modelos de dados
│   │   └── repositories/       # Repositórios de acesso
│   ├── ai/                     # Interface com serviço Python
│   │   ├── client/             # Cliente HTTP para AI service
│   │   └── types/              # Tipos para comunicação
│   └── config/                 # Configurações da aplicação
├── scripts/                    # Scripts utilitários
└── docker/                     # Configurações Docker
```

## 🚀 Tecnologias Utilizadas

### Backend Core (Go)
- **Gin** - Framework web para API REST ✅
- **Firejail** - Sandbox para execução segura de código ✅
- **GCC** - Compilador C integrado ✅
- **PostgreSQL** - Banco de dados ✅
- **GORM** - ORM para gerenciamento do banco de dados ✅

### AI Service (Python)
- **FastAPI/Flask** - Framework web para API de IA
- **Ollama** - Interface para modelos LLaMA
- **Langchain** - Framework para aplicações com LLM
- **Pydantic** - Validação de dados
- **aiohttp** - Cliente HTTP assíncrono

### Segurança & Isolamento
- **Docker** - Containerização para sandbox
- **Firejail** - Isolamento adicional de processos
- **chroot** - Isolamento de filesystem
- **ulimit** - Limitação de recursos

## ⚡ Quick Start

### Configuração do Ambiente

```bash
# Clone o repositório
git clone https://github.com/VictorHumberto01/CLabServer.git
cd clab-server

# Instale dependências
go mod download

# Configure o banco de dados
createdb clab
./scripts/migrate.sh
```

### Configuração do Ambiente
Crie um arquivo `.env` na raiz do projeto:
```env
DATABASE_URL=postgres://localhost:5432/clab?sslmode=disable
JWT_SECRET=sua-chave-secreta
ENV=development
SERVER_PORT=8080
COMPILER_PATH=/usr/bin/gcc
MAX_CODE_SIZE=1048576
MAX_MEMORY_USAGE=268435456
```

### Executando o Servidor

#### Modo Desenvolvimento
```bash
./scripts/dev.sh
```

#### Modo Produção
```bash
./scripts/prod.sh
```

## 🔒 Segurança

### Sistema de Sandbox ✅
- **Firejail Integration**: Execução isolada quando disponível
  - `--quiet`: Execução silenciosa
  - `--net=none`: Sem acesso à rede
  - `--private=tmpdir`: Filesystem isolado
- **Modo Inseguro Controlado**: Fallback com confirmação dupla do usuário
- **Timeout de Execução**: Limite de 3 segundos para prevenir loops infinitos
- **Diretório Temporário**: Cada execução usa um diretório isolado
- **Limpeza Automática**: Remoção de arquivos temporários após execução

### Validação de Entrada ✅
- **JSON Binding**: Validação automática de requests
- **Timeout Protection**: Processo killado após limite de tempo
- **Concurrent Safe**: Goroutines para execução não-bloqueante

### Próximas Implementações 🔄
- **Rate limiting** para prevenir abuse
- **Validação** de tamanho de código
- **Filtragem** de comandos perigosos
- **Logs de auditoria** estruturados

## 🤝 Contribuição

### Estrutura de Commits
```
feat: adiciona nova funcionalidade
fix: corrige bug existente
docs: atualiza documentação
test: adiciona ou corrige testes
refactor: refatora código sem mudar funcionalidade
perf: melhora performance
chore: tarefas de manutenção
```

### Pull Request Guidelines
1. Fork o repositório
2. Crie uma branch descritiva
3. Implemente a funcionalidade com testes
4. Atualize a documentação se necessário
5. Submeta o PR com descrição clara

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

