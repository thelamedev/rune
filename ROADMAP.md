# **Rune: Product & Development Roadmap**

Version: 1.0  
Last Updated: September 13, 2025

## **1\. Vision & Core Tenets**

### **1.1. Mission Statement**

To provide a simple, secure, and performant open-source system for secrets management and service discovery, designed for modern cloud-native applications. Rune prioritizes operational simplicity and a best-in-class developer experience.

### **1.2. Guiding Principles (The Rune Creed)**

Every architectural decision and feature implementation must be weighed against these principles:

1. **Security is Not Optional:** The system must be secure by default. We will always choose the more secure path, even if it adds minor complexity. Zero-trust principles apply.

2. **Operational Simplicity:** Rune should be easy to deploy, manage, and monitor. A single static Go binary with minimal external dependencies is the goal.

3. **Developer Experience is a Core Feature:** APIs must be intuitive, documentation must be excellent, and client libraries should make integration trivial.

4. **Performance by Default:** The system must be lightweight and have minimal overhead, making it suitable for all environments from IoT devices to large-scale server clusters.

5. **High Availability is Built-in:** The system is designed from the ground up to be clustered and fault-tolerant. Running a single node is a special case of running a cluster of one.

## **2\. High-Level Architecture**

Rune operates on a client-server model, composed of two primary components:

* **Rune Server:** A Go application that runs in a cluster of 3 or 5 nodes. It manages state, replication, encryption, and serves the API.

* **Rune Agent / Client:** A lightweight process or library that co-locates with your applications. It authenticates with the Rune Server, fetches secrets, and manages service registration.

## **3\. Phased Development Roadmap**

This roadmap is broken into four distinct phases. Each phase represents a major milestone, a shippable version of the product, and a potential "season" for your YouTube series.

### **Phase 1: The Secure Core (v0.1 \- "Genesis")**

**Goal:** Build a functional, single-node, secure vault. This phase is all about getting the core cryptography and storage right.

| Feature ID | Feature Name | Description & Context | Key Technologies / Libraries | Cross-Reference |
| :---- | :---- | :---- | :---- | :---- |
| P1-F1 | Pluggable Storage Interface | Define a Go interface for the storage backend. Implement an initial version using an embedded key-value store. This ensures we aren't locked into one backend. | interface{}, **BoltDB** or **BadgerDB** | P4-F2 |
| P1-F2 | **The Seal / Unseal Mechanism** | This is the cornerstone of Rune's security. On startup, the server is "sealed" (cannot decrypt data). It requires a quorum of "unseal keys" to reconstruct the Master Key in memory. | **Shamir's Secret Sharing** (hashicorp/vault/shamir), crypto/rand | P1-F3 |
| P1-F3 | Envelope Encryption Layer | Build the cryptographic engine. When a secret is written, generate a Data Encryption Key (DEK), encrypt the secret with it, then encrypt the DEK with the Master Key. Store the encrypted secret and encrypted DEK. **The Master Key is never persisted.** | crypto/aes, crypto/cipher (GCM mode) | P1-F2, P1-F4 |
| P1-F4 | Core Secrets API (v1) | Expose basic gRPC and REST endpoints for Create, Read, Update, Delete (CRUD) operations on secrets at a given path (e.g., /v1/secrets/database/password). | **gRPC**, **Protocol Buffers**, **grpc-gateway** | P3-F3 |
| P1-F5 | Basic CLI for Operator | A simple command-line tool (rune) to initialize the vault (generating Shamir keys) and to perform the unseal operation. | cobra or urfave/cli | P4-F4 |

**YouTube Series for Phase 1:** "Building a Secure Vault in Go from Scratch\!"

### **Phase 2: The Distributed System (v0.2 \- "Fellowship")**

**Goal:** Transform the single node into a highly available, fault-tolerant cluster.

| Feature ID | Feature Name | Description & Context | Key Technologies / Libraries | Cross-Reference |
| :---- | :---- | :---- | :---- | :---- |
| P2-F1 | **Consensus & Replication** | Integrate a consensus algorithm to manage state across all nodes. All writes must go through a leader and be replicated to a quorum of followers before being committed. | **Raft Consensus Algorithm** (hashicorp/raft) | P2-F3 |
| P2-F2 | Node Discovery & Cluster Join | Implement a mechanism for nodes to discover each other and securely join the cluster. This could be via a static config list of peers or a cloud API. | \- |  |
| P2-F3 | Leader Forwarding | API requests that require a write (e.g., storing a secret) that arrive at a follower node must be automatically and transparently forwarded to the current leader. | HTTP/gRPC client logic | P2-F1 |
| P2-F4 | Health Check Endpoints | An internal /health endpoint that reports the node's status (e.g., Sealed, Unsealed, Leader, Follower) for monitoring. | net/http |  |

**YouTube Series for Phase 2:** "From Monolith to Distributed System: Adding High Availability"

### **Phase 3: The Ecosystem (v0.3 \- "Discovery")**

**Goal:** Make Rune truly useful by adding authentication, dynamic features, and service discovery.

| Feature ID | Feature Name | Description & Context | Key Technologies / Libraries | Cross-Reference |
| :---- | :---- | :---- | :---- | :---- |
| P3-F1 | Authentication & Authorization | Implement a token-based auth system. Each token is associated with policies. Policies are path-based rules defining capabilities (e.g., read, write) on secret paths. | golang-jwt, Casbin (optional for policies) | P4-F1 |
| P3-F2 | **Leasing and Renewal** | Secrets and tokens are not permanent. When read, they are granted a lease with a TTL. Clients must renew the lease before it expires, otherwise it is revoked. | Time-based logic, background cleanup goroutines |  |
| P3-F3 | Secure Audit Logging | Every request (authenticated or not) and every response must be logged to a secure, append-only audit log. This provides a critical trail of all activity. | \- | P1-F4 |
| P3-F4 | Service Discovery API (v1) | Create endpoints for registering a service (/v1/discovery/register), de-registering, and querying for healthy instances (/v1/discovery/query/my-app). | \- | P3-F5 |
| P3-F5 | Passive Health Checks (TTL) | When a service registers, it is given a TTL. The service (or its agent) is responsible for periodically sending a heartbeat to renew the TTL. If it expires, Rune automatically removes the service instance. | Time-based logic | P3-F4 |

**YouTube Series for Phase 3:** "Adding Auth, Auditing, and Service Discovery"

### **Phase 4: Production Hardening (v1.0 \- "Maturity")**

**Goal:** Add features that make Rune ready for production use, focusing on observability, usability, and extensibility.

| Feature ID | Feature Name | Description & Context | Key Technologies / Libraries | Cross-Reference |
| :---- | :---- | :---- | :---- | :---- |
| P4-F1 | Pluggable Auth Backends | Refactor auth into an interface to allow for different methods (e.g., OIDC, Kubernetes Service Account, Cloud IAM). | Go interface{} | P3-F1 |
| P4-F2 | Pluggable Storage Backends | Implement additional storage backends like **etcd**, **Consul**, or cloud object storage like **S3**. | Go interface{} | P1-F1 |
| P4-F3 | Built-in DNS Interface | For seamless service discovery, run an optional DNS server on port 53 that can resolve queries like my-app.service.rune to the IP addresses of healthy instances. | miekg/dns | P3-F4 |
| P4-F4 | Advanced CLI | Enhance the rune CLI to be the primary operator tool for managing policies, auth, and inspecting cluster state. | cobra, urfave/cli | P1-F5 |
| P4-F5 | Official Rune Agent | Develop a polished, standalone agent that can be run as a sidecar. It will handle token renewal, secret fetching (and caching to disk), and service registration for an application. | \- | P3-F5 |
| P4-F6 | Observability | Expose detailed metrics in a standard format (e.g., Prometheus) for monitoring cluster health, API latency, and more. | prometheus/client\_golang |  |
| P4-F7 | Web UI (Stretch Goal) | A simple, read-only web interface for visualizing cluster status, services, and managing unsealing. | Go embed, a simple JS framework like Svelte or Vue. |  |

## **4\. Technical Deep Dive & Reference**

### **4.1. Core Data Models (Go Structs)**

// Represents a secret stored in the backend.  
type StoredSecret struct {  
    Path           string    // e.g., "database/mysql/password"  
    EncryptedValue \[\]byte    // The secret value, encrypted with the DEK.  
    EncryptedDEK   \[\]byte    // The DEK, encrypted with the Master Key.  
    Metadata       map\[string\]string  
    Version        int  
    CreatedAt      time.Time  
}

// Represents a live service instance for discovery.  
type ServiceInstance struct {  
    ServiceName string    // e.g., "api-service"  
    InstanceID  string    // a unique ID for this instance  
    Address     string    // "10.0.1.123:8080"  
    Tags        \[\]string  // \["v1.2", "region:us-east-1"\]  
    LeaseID     string    // ID for the TTL lease  
}

// Represents an authorization policy.  
type Policy struct {  
    Name  string  
    Rules \[\]PolicyRule  
}

type PolicyRule struct {  
    Path         string   // Path glob, e.g., "secrets/prod/\*"  
    Capabilities \[\]string // "read", "write", "delete", "list"  
}

### **4.2. Security Considerations Checklist**

* \[ \] **TLS Everywhere:** All communication between nodes and between clients/servers MUST use mTLS.

* \[ \] **Memory Protection:** Secrets should have a minimal lifetime in memory. Use memguard or similar libraries to securely zero out memory after use.

* \[ \] **Input Validation:** Rigorously validate all API inputs to prevent injection or path traversal attacks.

* \[ \] **Rate Limiting:** Implement rate limiting on sensitive endpoints (e.g., auth, unseal) to prevent brute-force attacks.

* \[ \] **Dependencies:** Regularly scan dependencies for known vulnerabilities (govulncheck).

### **4.3. OSS & Community Strategy**

* **Source Code:** Host on GitHub.

* **License:** Choose a permissive license like **MIT** or **Apache 2.0**.

* **Contribution Guide:** Create a CONTRIBUTING.md file explaining how to set up the dev environment, run tests, and submit pull requests.

* **Communication:** Use GitHub Issues for bug tracking and feature requests. Consider a Discord or Slack for community discussion.

* **YouTube:** Use the video series as the primary driver for community growth. Link to the GitHub repo in every video.