# OpenAI GPT-4o Realtime CLI

A Command Line Interface (CLI) tool in Go that interacts with OpenAI's GPT-4o-realtime-mini via WebSocket in real time. The CLI provides streaming responses and supports function calling for mathematical operations.

## Features

- ✅ **Real-time WebSocket communication** with OpenAI Realtime API
- ✅ **Streaming responses** - characters appear as they're generated (like ChatGPT)
- ✅ **Interactive chat interface** - continuous conversation with proper formatting
- ✅ **Function calling support** - multiply two numbers using AI function calls
- ✅ **Environment variable configuration** - secure API key handling

## Requirements

- Go 1.19+ 
- OpenAI API key with access to GPT-4o-realtime-preview-2024-10-01
- Internet connection

## Installation

1. **Clone the repository:**
   ```bash
   git clone https://github.com/avivheldman/wonderful.git
   cd wonderful
   ```

2. **Install dependencies:**
   ```bash
   go mod tidy
   ```

3. **Set up environment variables:**
   
   Create a `.env` file in the project root:
   ```bash
   OPENAI_KEY=your_openai_api_key_here
   OPENAI_URL=wss://api.openai.com/v1/realtime?model=gpt-4o-realtime-preview-2024-10-01
   ```
## Getting Your OpenAI API Key

1. Visit [OpenAI Platform](https://platform.openai.com/account/api-keys)
2. Create a new API key
3. Make sure your account has access to the GPT-4o Realtime API
4. Copy the key and add it to your `.env` file

## Usage

1. **Run the CLI:**
   ```bash
   go run main.go
   ```

2. **Start chatting:**
   ```
   Connected to OpenAI WebSocket
   
   You: Hello!
   AI: Hi there! How can I help you today? 😊
   
   You: What is 7 times 8?
   AI: [Calling function: multiply with args: {"a":7,"b":8}]
   Result: 56.00
   I'll calculate that for you! 7 times 8 equals 56.
   
   You: exit
   👋 Goodbye!
   ```

3. **Commands:**
   - Type any message to chat with the AI
   - `exit` or `quit` - Exit the program
   - `Ctrl+D` (Unix) / `Ctrl+Z` (Windows) - Exit the program

## Function Calling Examples

The CLI supports mathematical multiplication through function calling:

- "What is 5 times 3?"
- "Can you multiply 12 by 15?" 
- "Calculate 7 * 8"
- "Multiply 2.5 and 4"

When the AI detects a multiplication request, it automatically calls the `multiply` function and returns the result.

## Architecture

### Core Components

```
┌─────────────────┐    WebSocket     ┌─────────────────┐
│                 │◄────────────────►│                 │
│   Go CLI App    │                  │  OpenAI Realtime│
│                 │                  │      API        │
└─────────────────┘                  └─────────────────┘
```

### Key Functions

- **`connectToOpenAI()`** - Establishes WebSocket connection with authentication
- **`setSessionconfig()`** - Configures AI session with tools and instructions  
- **`sendMessage()`** - Sends user messages and requests AI responses
- **`listenForResponseMessages()`** - Handles streaming responses and function calls
- **`handleFunctionCall()`** - Executes function calls and sends results back
- **`multiply()`** - The mathematical function available to the AI

### Goroutine Management

```
Main Goroutine                    Listener Goroutine
│                                │
├─ Connect to OpenAI             │
├─ Start listener ──────────────►├─ Listen for messages
├─ User input loop               ├─ Stream text responses  
│  ├─ Read user input            ├─ Handle function calls
│  ├─ Send to OpenAI             │
│  └─ Wait for response ◄────────┤
│                                │
└─ Send shutdown signal ────────►└─ Clean exit
```

### WebSocket Events Handled

- `session.updated` - Session configuration confirmed
- `response.text.delta` - Streaming text chunks from AI
- `response.text.done` - Text response completed
- `response.function_call_arguments.done` - Function ready to execute
- `error` - Error messages from OpenAI

### Dependencies

- **[gorilla/websocket](https://github.com/gorilla/websocket)** - WebSocket client
- **[joho/godotenv](https://github.com/joho/godotenv)** - Environment variable loading
