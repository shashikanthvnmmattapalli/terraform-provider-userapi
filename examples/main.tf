terraform {
  required_providers {
    userapi = {
      source = "local/userapi"
      version = "0.1.0"
    }
  }
}

provider "userapi" {
  # Configuration options if any
  # Example: endpoint = "http://localhost:5000"
}

resource "userapi_user" "example" {
  name     = "John"
  email    = "john@example.com"
  username = "johndoe"
}

output "user_id" {
  value = userapi_user.example.id
}

resource "userapi_user" "example1" {
  name     = "Alice"
  email    = "alice@example.com"
  username = "alice123"
}

resource "userapi_user" "example2" {
  name     = "Alice1"
  email    = "alice1@example.com"
  username = "alice1231"
}