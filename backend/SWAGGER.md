# Swagger API Documentation

MyKanban Backend includes auto-generated Swagger/OpenAPI documentation powered by [Swaggo](https://github.com/swaggo/swag).

## Accessing Swagger UI

Once the server is running, open your browser and navigate to:

```
http://localhost:8080/swagger/index.html
```

> Replace `localhost:8080` with your actual host and port if different.

## Authentication in Swagger UI

Most API endpoints require JWT authentication. Follow these steps:

### 1. Get a JWT Token

Use the **POST `/api/v1/auth/login`** endpoint in Swagger UI (under the **Auth** tag):

```json
{
  "email": "admin@mykanban.local",
  "password": "admin123"
}
```

Copy the `token` value from the response.

### 2. Authorize

1. Click the **🔓 Authorize** button at the top-right of Swagger UI.
2. In the **BearerAuth** field, enter:
   ```
   Bearer <your-jwt-token>
   ```
   For example:
   ```
   Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
   ```
3. Click **Authorize**, then **Close**.

All subsequent requests in Swagger UI will include the JWT token.

## API Tags

The API is organized into the following groups:

| Tag          | Description                                |
|--------------|--------------------------------------------|
| **System**   | Health check                               |
| **Auth**     | Login, password change, Google OAuth       |
| **Projects** | CRUD operations for projects               |
| **Boards**   | CRUD operations for boards                 |
| **Tasks**    | CRUD + PATCH operations for tasks          |
| **Schedulers** | CRUD for recurring task schedulers       |
| **Resources** | CRUD for resources (people/services)      |

## Example Workflow in Swagger UI

1. **Health check** — `GET /api/health` (no auth needed)
2. **Login** — `POST /api/v1/auth/login` → copy token
3. **Authorize** — click 🔓 and paste `Bearer <token>`
4. **Create a project** — `POST /api/v1/projects`
5. **Create a board** — `POST /api/v1/boards` (provide `project_id`)
6. **Create a task** — `POST /api/v1/tasks` (provide `board_id`, valid `swimlane`, `task_type`)
7. **Move task** — `PATCH /api/v1/tasks/{id}` with `{"swimlane": "Done"}`

## Regenerating Swagger Docs

If you modify handler annotations, regenerate the docs:

```bash
# Install swag CLI (one-time)
go install github.com/swaggo/swag/cmd/swag@latest

# Generate docs
swag init

# Or use the build script (runs swag init automatically)
bash scripts/build.sh
```

Generated files are in the `docs/` folder:
- `docs/docs.go` — Go bindings for embedding
- `docs/swagger.json` — OpenAPI 2.0 spec (JSON)
- `docs/swagger.yaml` — OpenAPI 2.0 spec (YAML)

## Downloading the OpenAPI Spec

The raw spec files are available at:

- **JSON:** `http://localhost:8080/swagger/doc.json`
- **YAML:** Download from the `docs/swagger.yaml` file

These can be imported into tools like Postman, Insomnia, or used for client code generation.

## CORS Support

The API supports Cross-Origin Resource Sharing (CORS), which allows Swagger UI and frontend applications running on different origins to interact with the backend. Allowed origins are configured in `config.json` via the `allowed_origins` field. See [Backend README – CORS Configuration](README.md#cors-configuration) for details.
