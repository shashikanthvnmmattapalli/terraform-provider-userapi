terraform {
  required_providers {
    userapi = {
      source  = "local/userapi"
      version = "0.1.0"
    }
  }
}

provider "userapi" {
  # Optionally configure endpoint here if needed
  # endpoint = "http://localhost:5000"
}

resource "userapi_user" "example" {
  name     = "John"
  email    = "john@example.com"
  username = "johndoe"
}

# Data source fetching user by ID from resource created above
data "userapi_user" "fetched_user" {
  id = userapi_user.example.id
}

output "fetched_user_name" {
  value = data.userapi_user.fetched_user.name
}

output "fetched_user_email" {
  value = data.userapi_user.fetched_user.email
}
