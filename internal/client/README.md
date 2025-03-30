# Masthead Data Client (Go)

A Go client package used by the Masthead Data Terraform provider [terraform-provider-masthead](https://github.com/masthead-data/terraform-provider-masthead) and product API. It establishes a new client and sends HTTP(s) requests to perform CRUD operations.

## API Reference

### Authentication

All API requests require authentication using API keys passed in the headers.

```http
X-API-TOKEN: <token-value>
```

### User Management APIs

#### List Users

```http
GET /clientApi/user/list
```

Returns a list of all users.

Example Response:

```json
{
    "values": [
        {
            "email": "user@example.com",
            "role": "OWNER"
        }
    ],
    "extra": null,
    "error": null
}
```

#### Create User

```http
POST /clientApi/user
```

Creates a new user in the system.

Request Body:

```json
{
    "email": "user@example.com",
    "role": "USER"
}
```

#### Update User Role

```http
PUT /clientApi/user/role
```

Updates an existing user's role.

Request Body:

```json
{
    "email": "user@example.com",
    "role": "OWNER"
}
```

#### Delete User

```http
DELETE /clientApi/user/{email}
```

Removes a user from the system by their email address.
