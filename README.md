# 🖥️ CLab Server - Backend Engine

Servidor backend do CLab responsável pela compilação segura de código C, execução em sandbox, integração com IA para feedback educacional e gerenciamento de dados locais.

![CLab Server](https://img.shields.io/badge/CLab-Server%20Backend-green?style=for-the-badge)
![Go](https://img.shields.io/badge/Go-1.21+-blue?style=flat-square&logo=go)


## 🎯 Visão Geral

O CLab Server é o núcleo do sistema de ensino de programação C, fornecendo:
- **Compilação segura** de código C em ambiente isolado
- **Feedback inteligente** via IA local (LLaMA/Ollama)
- **API REST** para comunicação com o frontend Electron
- **Gerenciamento de dados** com SQLite local

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
│  💾 Database Service (Go + SQLite)          │
│  • Gerenciamento de usuários               │
│  • Armazenamento de tarefas                │
│  • Histórico de submissões                 │
└─────────────────────────────────────────────┘
```

## 📁 Estrutura Atual do Projeto

```
clab-server/
├── main.go                     # ✅ Servidor principal com API de compilação
├── go.mod                      # ✅ Dependências do projeto
├── go.sum                      # ✅ Checksums das dependências
└── README.md                   # ✅ Este arquivo

# Estrutura planejada para expansão:
├── cmd/
│   └── server/              
│       └── main.go             # Ponto de entrada da aplicação
├── internal/
│   ├── api/                    # Handlers e rotas da API
│   │   ├── handlers/           # Controllers REST
│   │   ├── middleware/         # Middlewares HTTP
│   │   └── routes/             # Definição de rotas
│   ├── compiler/               # 🔄 Serviço de compilação (refatorar)
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
├── ai-service/                 # Serviço Python de IA (planejado)
├── database/                   # Schema e migrações
├── docker/                     # Configurações Docker (planejado)
└── scripts/                    # Scripts utilitários
```

## 🚀 Tecnologias Utilizadas

### Backend Core (Go)
- **Gin** - Framework web para API REST ✅
- **Firejail** - Sandbox para execução segura de código
- **GCC** - Compilador C integrado ✅
- **SQLite** - Banco de dados local (planejado)
- **GORM** - ORM para gerenciamento do banco de dados (planejado)

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

### Executar o Servidor Atual

```bash
# Clone o repositório
git clone https://github.com/VictorHumberto01/clab-server.git
cd clab-server

# Instale dependências
go mod tidy

# Execute o servidor
go run main.go
```

O servidor estará disponível em `http://localhost:8080`

### Testar a API

```bash
# Teste básico de compilação
curl -X POST http://localhost:8080/compile \
  -H "Content-Type: application/json" \
  -d '{
    "code": "#include <stdio.h>\nint main() {\n    printf(\"Hello, CLab!\\n\");\n    return 0;\n}",
    "input": ""
  }'
```

### Recursos Implementados ✅

- **Compilação segura** de código C com GCC
- **Execução em sandbox** usando Firejail (quando disponível)
- **Modo inseguro** com confirmação do usuário (fallback)
- **Timeout de execução** (3 segundos)
- **Suporte a input** do usuário para programas interativos
- **Logs detalhados** para debugging
- **Limpeza automática** de arquivos temporários


## 🔒 Segurança - Implementação Atual

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

