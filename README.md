# **Rune**

**A simple, secure, and performant open-source system for secrets management and service discovery, designed for modern cloud-native applications.**

<!-- Rune is being built live on YouTube\! You can follow the entire development process from scratch by watching the series [here](https://www.google.com/search?q=https://www.youtube.com/@thelamedev). -->

## **1\. Vision & Guiding Principles**

Rune is built on a few core tenets:

1. **Security is Not Optional:** The system is secure by default, utilizing concepts like Shamir's Secret Sharing and envelope encryption. The master key is never persisted.

2. **Operational Simplicity:** A single static Go binary with minimal external dependencies. Easy to deploy, manage, and monitor.

3. **Developer Experience is a Core Feature:** APIs are designed to be intuitive and integration should be trivial.

4. **Performance by Default:** Lightweight with minimal overhead.

5. **High Availability is Built-in:** Designed from the ground up to be a fault-tolerant, clustered system.

## **2\. Core Features**

* **Seal/Unseal Mechanism:** Rune starts in a sealed state and cannot decrypt any data. It requires a quorum of unseal keys to reconstruct the master key in memory.

* **Envelope Encryption:** Secrets are protected by a two-layer encryption strategy, ensuring the master key is used sparingly and data keys can be easily rotated.

* **Distributed & Highly Available:** Uses the Raft consensus algorithm to replicate data across a cluster for fault tolerance.

* **Service Discovery:** Provides a simple, TTL-based mechanism for services to register themselves and discover others.

* **API Driven:** All interactions happen over a secure gRPC API.

## **3\. Current Status**

**Phase 1: The Secure Core \- COMPLETE ✅**

Rune is currently a functional, single-node, secure vault. You can store and retrieve secrets from a running Rune server using the provided CLI. The next phase of development will focus on making the system distributed.

## **4\. Getting Started**

### **Prerequisites**

* Go (version 1.18+)

* Protocol Buffer Compiler (protoc)

### **Quick Start**

1. **Clone the repository:**  
   git clone [https://github.com/thelamedev/rune.git](https://github.com/thelamedev/rune.git)  
   cd rune

2. Generate API code:  
   If you modify the .proto files, you'll need to regenerate the Go code.  
   make api

3. Run the Rune Server:  
   In one terminal, start the server. It will generate the initial unseal keys and then listen for connections.  
   go run ./cmd/rune

4. Build and Use the CLI:  
   In a second terminal, build the CLI tool.  
   go build \-o rune-cli ./cmd/rune-cli

   Now use the CLI to interact with the server:  
   \# Store a secret  
   ./rune-cli put secrets/database/password "my-s3cr3t-p4ssw0rd\!"

   \# Retrieve the secret  
   ./rune-cli get secrets/database/password

## **5\. Roadmap**

The full product and development roadmap is detailed in [ROADMAP.md](ROADMAP.md).

## **6\. Contributing**

Contributions are welcome\! Please see the [CONTRIBUTING.md](CONTRIBUTING.md) file for guidelines on how to get started.

## **7\. License**

Rune is licensed under the [MIT License](LICENCE).
