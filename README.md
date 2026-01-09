# gRPC CLI Chat

A **command-line chat application** built to **learn and deeply understand gRPC**, especially **bidirectional streaming**, **event-driven communication**, and **concurrent client handling** in Go.

## 🎯 Purpose

This project exists to:

* Learn **gRPC bidirectional streaming**
* Understand **client ↔ server event flows**
* Practice **concurrency**, **fan-out broadcasting**, and **state tracking**
* Explore **Redis-backed message persistence**
* Build a real system that breaks in real ways (and fix it)

---

## 🎥 Demo (Video)

## Demo
[![Watch Demo](demo.png)](https://streamable.com/u55idz)



---

## Architecture Overview

```
┌────────────┐
│ CLI Client │
└─────┬──────┘
      │ gRPC (BiDi Stream)
┌─────▼──────┐
│  Chat      │
│  Server    │
└─────┬──────┘
      │
┌─────▼──────┐
│   Redis    │
└────────────┘
```

---

## Core Concepts Practiced

* gRPC **Bidirectional Streaming**
* Protobuf event modeling
* Stream lifecycle management
* Concurrent client fan-out
* Typing indicators via ephemeral events
* Message replay via Redis
* Rate limiting per client

---

## Tech Stack

* **Go**
* **gRPC**
* **Protocol Buffers**
* **Redis**
* **Docker**

---

## Requirements

* Go **1.21+**
* Docker
* `protoc`
* `protoc-gen-go`
* `protoc-gen-go-grpc`

---


## Running the Project

### 1. Generate protobuf code

```bash
make gen-go
```

### 2. Start Redis

```bash
make redis
```

### 3. Run the server

```bash
make server
```

### 4. Run one or more clients

```bash
make client
```

Or manually:

```bash
go run cmd/client/*.go --address 0.0.0.0:8080
```

---

## How It Works

### Client

* Opens a bidirectional gRPC stream
* Sends:

  * `USER_JOINED`
  * `CHAT_MESSAGE`
  * `TYPING_START / TYPING_STOP`
* Listens continuously for server events

### Server

* Accepts a stream per client
* Tracks active clients in memory
* Broadcasts:

  * Messages
  * Typing indicators
  * Join / leave events
* Persists chat messages in Redis

---

## Message Flow (Simplified)

```
Client ── CHAT_MESSAGE ──▶ Server
Server ── broadcast ────▶ All Clients
```

---

## Message Display Rules

* Sender sees:

  ```
  [you] hello
  ```
* Others see:

  ```
  [alice] hello
  ```

---

## Typing Indicators

* Clients send typing start/stop events
* Server broadcasts typing state to others
* Clients render:

  ```
  * alice typing...
  ```

---

## What This Project Is NOT

* ❌ Production-ready
* ❌ Secure
* ❌ Authenticated
* ❌ Feature-complete

It is:

* ✅ A **learning sandbox**
* ✅ A **gRPC stress test**
* ✅ A place to break things and understand why

---