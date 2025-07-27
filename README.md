# 🌐 Terraform Provider: userapi

This is a **custom Terraform provider** named `userapi` that allows Terraform to interact with a **local Flask-based REST API** for managing user data. It supports full CRUD operations and data source lookups via HTTP requests to endpoints like `/users` and `/users/{id}`.

---

## ⚙️ Flow Overview

The control flow follows this sequence:

1. `main.go`: Entrypoint for the provider binary.  
2. `provider.go`: Registers schema, resources, data sources, and configures the API client.  
3. `resource_user.go`: Defines full CRUD lifecycle: `Create`, `Read`, `Update`, `Delete`, and `Import`.  
4. `data_source.go`: Allows data lookups by ID (read-only).  
5. `app.py`: Local Flask-based REST API for user operations.

---

## ✅ Prerequisites

You need the following tools installed:

| Tool                                | Purpose                                  |
|------------------------------------|------------------------------------------|
| [Go](https://go.dev/dl/) 1.21+     | Build the Terraform provider             |
| [Python](https://www.python.org/) 3.8+ | Run the local REST API backend           |
| [Terraform](https://developer.hashicorp.com/terraform/downloads) 1.2+ | Use the provider in configurations       |
| Flask & SQLAlchemy (Python packages)| Backend API dependencies                  |

---

## 🚀 Setup Instructions (Local Development)

### Step 1: Install Python dependencies

pip install flask flask_sqlalchemy

### Step 2: Build the Terraform provider binary

go build -ldflags "-X main.version=0.1.0" -o terraform-provider-userapi.exe

### Step 3: Configure terraform.rc

Create or update your terraform.rc file at `%APPDATA%` (Windows) or `~/.terraformrc` (Linux/macOS) with the following content:

```hcl
provider_installation {
  dev_overrides {
    "local/userapi" = "PATH"
  }
  direct {}
}
Adjust the path to where your provider binary lives.


