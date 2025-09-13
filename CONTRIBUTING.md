# **Contributing to Rune**

First off, thank you for considering contributing\! Rune is an open-source project made possible by the community, and we welcome any contributions, from bug reports to new features.

This document provides guidelines for contributing to the project.

## **Code of Conduct**

This project and everyone participating in it is governed by our [Code of Conduct](https://www.google.com/search?q=CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## **How Can I Contribute?**

### **Reporting Bugs**

If you find a bug, please ensure the bug was not already reported by searching on GitHub under [Issues](https://www.google.com/search?q=https://github.com/thelamedev/rune/issues).

When you are creating a bug report, please include as many details as possible:

* A clear and descriptive title.

* Steps to reproduce the behavior.

* The expected behavior and what happened instead.

* Your Go version and operating system.

### **Suggesting Enhancements**

If you have an idea for a new feature or an improvement to an existing one, please open an issue to discuss it. This allows us to coordinate efforts and ensure the feature aligns with the project's goals.

### **Your First Code Contribution**

Ready to contribute code? Here’s how to set up your environment and submit a pull request.

1. **Fork the repository** on GitHub.

2. **Clone your fork** locally:  
   git clone \[https://github.com/YOUR\_USERNAME/rune.git\](https://github.com/YOUR\_USERNAME/rune.git)

3. **Create a new branch** for your feature or bug fix:  
   git checkout \-b feature/your-amazing-feature

4. **Set up your development environment:**

   * Ensure you have Go (1.18+) and protoc installed.

   * Install the necessary Go plugins for protoc:  
     go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28  
     go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

5. **Make your changes\!**

6. **Run the tests:** Before submitting your changes, ensure that all tests pass.  
   go test \-v ./...

7. **Commit your changes** with a clear and descriptive commit message.  
   git commit \-m "feat: Add some amazing feature"

8. **Push your branch** to your fork on GitHub:  
   git push origin feature/your-amazing-feature

9. **Open a Pull Request** to the main branch of the original repository. Provide a clear description of your changes and reference any related issues.

## **Styleguides**

### **Go Code**

Please run gofmt on your code before committing to ensure it is formatted according to Go's standard style. Most IDEs do this automatically.

### **Git Commit Messages**

* Use the present tense ("Add feature" not "Added feature").

* Use the imperative mood ("Move cursor to..." not "Moves cursor to...").

* Limit the first line to 72 characters or less.

* Reference issues and pull requests liberally after the first line.

We look forward to your contributions\!