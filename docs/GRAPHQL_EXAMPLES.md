# GraphQL API Examples

Comprehensive examples for using the GraphQL API with various operations, patterns, and best practices.

## Table of Contents

- [Authentication Flow](#authentication-flow)
- [User Operations](#user-operations)
- [Using Fragments](#using-fragments)
- [Query Variables](#query-variables)
- [Error Handling](#error-handling)
- [Pagination](#pagination)
- [Dataloaders in Action](#dataloaders-in-action)
- [Testing with cURL](#testing-with-curl)
- [Client Libraries](#client-libraries)

## Authentication Flow

### Complete Registration and Login

#### 1. Register a New User

```graphql
mutation RegisterUser {
  createUser(input: {
    name: "Alice Johnson"
    email: "alice@example.com"
    password: "SecurePass123!"
  }) {
    id
    name
    email
    createdAt
  }
}
```

**Response**:

```json
{
  "data": {
    "createUser": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Alice Johnson",
      "email": "alice@example.com",
      "createdAt": "2026-02-26T18:00:00Z"
    }
  }
}
```

#### 2. Login to Get JWT Token

```graphql
mutation Login {
  login(input: {
    email: "alice@example.com"
    password: "SecurePass123!"
  }) {
    token
    user {
      id
      name
      email
    }
  }
}
```

**Response**:

```json
{
  "data": {
    "login": {
      "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI1NTBlODQwMC1lMjliLTQxZDQtYTcxNi00NDY2NTU0NDAwMDAiLCJleHAiOjE3MDkwNjg4MDB9.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
      "user": {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "name": "Alice Johnson",
        "email": "alice@example.com"
      }
    }
  }
}
```

#### 3. Use Token for Protected Queries

In GraphQL Playground, add the token to HTTP Headers:

```json
{
  "Authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

Then query protected data:

```graphql
query GetCurrentUser {
  me {
    id
    name
    email
    createdAt
    updatedAt
  }
}
```

## User Operations

### Query Single User

```graphql
query GetUser($id: ID!) {
  user(id: $id) {
    id
    name
    email
    createdAt
    updatedAt
  }
}
```

**Variables**:

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### Query Multiple Users

```graphql
query ListUsers {
  users {
    id
    name
    email
  }
}
```

### Update User

```graphql
mutation UpdateUser($id: ID!, $input: UpdateUserInput!) {
  updateUser(id: $id, input: $input) {
    id
    name
    email
    updatedAt
  }
}
```

**Variables**:

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "input": {
    "name": "Alice Johnson Updated"
  }
}
```

### Delete User

```graphql
mutation DeleteUser($id: ID!) {
  deleteUser(id: $id)
}
```

**Variables**:

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Response**:

```json
{
  "data": {
    "deleteUser": true
  }
}
```

## Using Fragments

Fragments help you reuse common fields across multiple queries.

### Define a Fragment

```graphql
fragment UserFields on User {
  id
  name
  email
  createdAt
}

query GetUser($id: ID!) {
  user(id: $id) {
    ...UserFields
  }
}

query ListUsers {
  users {
    ...UserFields
  }
}
```

### Nested Fragments

```graphql
fragment BasicUserInfo on User {
  id
  name
}

fragment DetailedUserInfo on User {
  ...BasicUserInfo
  email
  createdAt
  updatedAt
}

query GetUserDetails($id: ID!) {
  user(id: $id) {
    ...DetailedUserInfo
  }
}
```

## Query Variables

### Optional Variables

```graphql
query SearchUsers(
  $email: String
  $name: String
) {
  users(email: $email, name: $name) {
    id
    name
    email
  }
}
```

**Variables (partial)**:

```json
{
  "email": "alice@example.com"
}
```

### Default Values

```graphql
query ListUsers(
  $limit: Int = 10
  $offset: Int = 0
) {
  users(limit: $limit, offset: $offset) {
    id
    name
  }
}
```

### Input Objects

```graphql
mutation CreateUserWithInput($input: CreateUserInput!) {
  createUser(input: $input) {
    id
    name
    email
  }
}
```

**Variables**:

```json
{
  "input": {
    "name": "Bob Smith",
    "email": "bob@example.com",
    "password": "Password123!"
  }
}
```

## Error Handling

### Validation Errors

```graphql
mutation CreateInvalidUser {
  createUser(input: {
    name: "Alice"
    email: "invalid-email"
    password: "123"
  }) {
    id
  }
}
```

**Response**:

```json
{
  "errors": [
    {
      "message": "invalid email format",
      "path": ["createUser"],
      "extensions": {
        "code": "VALIDATION_ERROR",
        "field": "email"
      }
    },
    {
      "message": "password must be at least 8 characters",
      "path": ["createUser"],
      "extensions": {
        "code": "VALIDATION_ERROR",
        "field": "password"
      }
    }
  ],
  "data": {
    "createUser": null
  }
}
```

### Authentication Errors

Query without token:

```graphql
query GetMe {
  me {
    id
    email
  }
}
```

**Response**:

```json
{
  "errors": [
    {
      "message": "Unauthorized: missing or invalid token",
      "path": ["me"],
      "extensions": {
        "code": "UNAUTHORIZED"
      }
    }
  ],
  "data": {
    "me": null
  }
}
```

### Not Found Errors

```graphql
query GetNonExistentUser {
  user(id: "00000000-0000-0000-0000-000000000000") {
    id
  }
}
```

**Response**:

```json
{
  "errors": [
    {
      "message": "user not found",
      "path": ["user"],
      "extensions": {
        "code": "NOT_FOUND"
      }
    }
  ],
  "data": {
    "user": null
  }
}
```

### Handling Errors in Client

```javascript
const result = await client.query({ query: GET_USER });

if (result.errors) {
  result.errors.forEach(error => {
    if (error.extensions.code === 'UNAUTHORIZED') {
      // Redirect to login
      router.push('/login');
    } else if (error.extensions.code === 'NOT_FOUND') {
      // Show 404 page
      showNotFound();
    } else {
      // Show generic error
      showError(error.message);
    }
  });
}
```

## Pagination

### Offset-based Pagination

```graphql
query ListUsersPaginated(
  $limit: Int!
  $offset: Int!
) {
  users(limit: $limit, offset: $offset) {
    id
    name
    email
  }
  usersCount
}
```

**Variables**:

```json
{
  "limit": 10,
  "offset": 0
}
```

### Cursor-based Pagination (Relay-style)

```graphql
query ListUsersWithCursor(
  $first: Int!
  $after: String
) {
  usersConnection(first: $first, after: $after) {
    edges {
      node {
        id
        name
        email
      }
      cursor
    }
    pageInfo {
      hasNextPage
      endCursor
    }
  }
}
```

**First Page**:

```json
{
  "first": 10
}
```

**Next Page** (use `endCursor` from previous response):

```json
{
  "first": 10,
  "after": "Y3Vyc29yOjEw"
}
```

## Dataloaders in Action

Dataloaders batch and cache requests to prevent N+1 queries.

### Without Dataloaders (N+1 Problem)

Imagine querying posts with their authors:

```graphql
query GetPosts {
  posts {
    id
    title
    author {
      id
      name
    }
  }
}
```

**Without dataloaders**: 
- 1 query for posts
- N queries for authors (one per post)
- Total: N+1 queries

### With Dataloaders (Optimized)

**With dataloaders**:
- 1 query for posts
- 1 batched query for all authors
- Total: 2 queries

### Observing Dataloader Batching

Enable debug logging:

```env
APP_DEBUG=true
```

You'll see logs like:

```
DEBUG dataloader: batching 10 user IDs
DEBUG dataloader: executing batch query
DEBUG SQL: SELECT * FROM users WHERE id = ANY($1)
```

## Testing with cURL

### Health Check

```bash
curl -X POST http://localhost:8080/query \
  -H 'Content-Type: application/json' \
  -d '{"query": "{ health }" }'
```

### Register User

```bash
curl -X POST http://localhost:8080/query \
  -H 'Content-Type: application/json' \
  -d '{
    "query": "mutation CreateUser($input: CreateUserInput!) { createUser(input: $input) { id name email } }",
    "variables": {
      "input": {
        "name": "Alice",
        "email": "alice@example.com",
        "password": "SecurePass123!"
      }
    }
  }'
```

### Login and Save Token

```bash
# Login and extract token
TOKEN=$(curl -s -X POST http://localhost:8080/query \
  -H 'Content-Type: application/json' \
  -d '{
    "query": "mutation Login($input: LoginInput!) { login(input: $input) { token } }",
    "variables": {
      "input": {
        "email": "alice@example.com",
        "password": "SecurePass123!"
      }
    }
  }' | jq -r '.data.login.token')

echo "Token: $TOKEN"
```

### Query with Authentication

```bash
curl -X POST http://localhost:8080/query \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"query": "{ me { id name email } }" }'
```

### Pretty Print Response

```bash
curl -s -X POST http://localhost:8080/query \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"query": "{ users { id name email } }" }' | jq
```

## Client Libraries

### Apollo Client (React)

```javascript
import { ApolloClient, InMemoryCache, createHttpLink } from '@apollo/client';
import { setContext } from '@apollo/client/link/context';

const httpLink = createHttpLink({
  uri: 'http://localhost:8080/query',
});

const authLink = setContext((_, { headers }) => {
  const token = localStorage.getItem('token');
  return {
    headers: {
      ...headers,
      authorization: token ? `Bearer ${token}` : "",
    }
  };
});

const client = new ApolloClient({
  link: authLink.concat(httpLink),
  cache: new InMemoryCache(),
});

// Usage in component
import { useQuery, gql } from '@apollo/client';

const GET_USERS = gql`
  query GetUsers {
    users {
      id
      name
      email
    }
  }
`;

function UsersList() {
  const { loading, error, data } = useQuery(GET_USERS);

  if (loading) return <p>Loading...</p>;
  if (error) return <p>Error: {error.message}</p>;

  return data.users.map(user => (
    <div key={user.id}>
      {user.name} - {user.email}
    </div>
  ));
}
```

### urql (React/Vue/Svelte)

```javascript
import { createClient, cacheExchange, fetchExchange } from 'urql';
import { authExchange } from '@urql/exchange-auth';

const client = createClient({
  url: 'http://localhost:8080/query',
  exchanges: [
    cacheExchange,
    authExchange(async utils => ({
      addAuthToOperation(operation) {
        const token = localStorage.getItem('token');
        if (!token) return operation;
        return utils.appendHeaders(operation, {
          Authorization: `Bearer ${token}`,
        });
      },
      didAuthError(error) {
        return error.graphQLErrors.some(
          e => e.extensions?.code === 'UNAUTHORIZED',
        );
      },
    })),
    fetchExchange,
  ],
});
```

### graphql-request (Simple)

```javascript
import { GraphQLClient, gql } from 'graphql-request';

const endpoint = 'http://localhost:8080/query';

const client = new GraphQLClient(endpoint, {
  headers: {
    authorization: `Bearer ${token}`,
  },
});

const query = gql`
  query GetUsers {
    users {
      id
      name
      email
    }
  }
`;

const data = await client.request(query);
console.log(data.users);
```

### Go Client

```go
package main

import (
	"context"
	"github.com/machinebox/graphql"
)

func main() {
	client := graphql.NewClient("http://localhost:8080/query")

	req := graphql.NewRequest(`
		query GetUsers {
			users {
				id
				name
				email
			}
		}
	`)

	req.Header.Set("Authorization", "Bearer "+token)

	var response struct {
		Users []struct {
			ID    string
			Name  string
			Email string
		}
	}

	if err := client.Run(context.Background(), req, &response); err != nil {
		panic(err)
	}

	for _, user := range response.Users {
		fmt.Printf("%s: %s\n", user.Name, user.Email)
	}
}
```

### Python Client (gql)

```python
from gql import gql, Client
from gql.transport.requests import RequestsHTTPTransport

transport = RequestsHTTPTransport(
    url='http://localhost:8080/query',
    headers={'Authorization': f'Bearer {token}'}
)

client = Client(transport=transport, fetch_schema_from_transport=True)

query = gql("""
    query GetUsers {
        users {
            id
            name
            email
        }
    }
""")

result = client.execute(query)
for user in result['users']:
    print(f"{user['name']}: {user['email']}")
```

## Performance Tips

### Query Only What You Need

❌ **Bad** - Fetching unnecessary fields:

```graphql
query GetUserNames {
  users {
    id
    name
    email
    password
    createdAt
    updatedAt
    lastLoginAt
    profilePicture
  }
}
```

✅ **Good** - Fetch only required fields:

```graphql
query GetUserNames {
  users {
    id
    name
  }
}
```

### Use Fragments for Reusability

✅ **Good**:

```graphql
fragment UserCard on User {
  id
  name
  email
}

query GetUsers {
  users {
    ...UserCard
  }
}
```

### Leverage Dataloaders

Always fetch related entities in the same query to benefit from dataloaders:

```graphql
query GetPostsWithAuthors {
  posts {
    id
    title
    author {
      id
      name
    }
  }
}
```

## Next Steps

- **Deployment**: Check [Deployment Guide](./DEPLOYMENT.md)
- **Dataloaders**: Read [Dataloader Guide](./DATALOADERS.md)
- **Development**: See [Development Guide](./DEVELOPMENT.md)
- **Architecture**: Review [README](../README.md)
