# item-modify

A CRUD service for **Items**, backed by **IBM Cloud Object Storage (COS)** with credentials sourced from **IBM Secrets Manager**. Includes a single-page web UI served directly by the backend.

## Layout

```
item-modify/
├── api/                  ← Go backend (Gin REST API)
│   ├── main.go           ← server entry-point, config boot, route wiring
│   ├── config.go         ← COSConfig + SecretsManagerConfig from env vars
│   ├── secrets.go        ← IBM Secrets Manager → IAM API key
│   ├── store.go          ← IBM COS CRUD (PutObject / GetObject / ListObjectsV2 / DeleteObject)
│   ├── model.go          ← Item struct
│   ├── handler.go        ← Gin HTTP handlers
│   ├── go.mod
│   └── go.sum
├── ui/
│   └── index.html        ← Single-page CRUD UI
├── arch/
│   └── architecture.drawio  ← Component and code-flow architecture diagram
├── Dockerfile            ← Multi-stage build (golang:1.25-alpine → scratch)
├── .dockerignore
├── deploy.sh             ← Build → push to ICR → deploy to Code Engine
└── README.md
```

## Prerequisites

- Go 1.25+
- Docker
- IBM Cloud CLI with plugins: `code-engine`, `container-registry`
- IBM Cloud account with:
  - A **Cloud Object Storage** instance and bucket (`resilience-app`)
  - A **Secrets Manager** instance with an `iam_credentials` secret

## Configuration

### Option A — IBM Secrets Manager ✅ recommended

The COS IAM API key is fetched at startup from Secrets Manager.
`COS_API_KEY` is **not** required in this mode.

```bash
# Secrets Manager
export SM_INSTANCE_URL="https://d53e26db-b7c1-461d-94e6-8de136174d04.eu-gb.secrets-manager.appdomain.cloud"
export SM_API_KEY="<iam-api-key-with-secretsreader-role>"
export SM_SECRET_ID="2c54b838-360b-fc07-cff9-3eb9a13075e1"   # resilience-forge-iam-api-key

# COS infrastructure (not secrets)
export COS_INSTANCE_CRN="crn:v1:bluemix:public:cloud-object-storage:global:a/1f7277194bb748cdb1d35fd8fb85a7cb:4c1fe46b-a89d-4b23-8729-d9d8221ecf5f::"
export COS_ENDPOINT="s3.us-south.cloud-object-storage.appdomain.cloud"
export COS_BUCKET="resilience-app"
```

### Option B — direct environment variables

```bash
export COS_API_KEY="<iam-api-key>"
export COS_INSTANCE_CRN="crn:v1:bluemix:public:cloud-object-storage:global:a/1f7277194bb748cdb1d35fd8fb85a7cb:4c1fe46b-a89d-4b23-8729-d9d8221ecf5f::"
export COS_ENDPOINT="s3.us-south.cloud-object-storage.appdomain.cloud"
export COS_BUCKET="resilience-app"
```

## Run locally

```bash
cd item-modify/api
go run .
```

Open **http://localhost:8080** — the web UI loads automatically.

---

## Deploy to IBM Code Engine

The app is deployed in the **`event-notification-server`** project (`us-south`).

**Live URL:** https://item-modify.27wc2juyqulv.us-south.codeengine.appdomain.cloud

### Step 1 — Login and target the project

```bash
ibmcloud login --sso
ibmcloud target -r us-south -g Default
ibmcloud ce project select --id 5b578840-3506-444d-a21a-d5a768c6e324
```

### Step 2 — Login to IBM Container Registry

```bash
ibmcloud cr region-set us-south
ibmcloud cr login
```

### Step 3 — Build and push the image

```bash
cd item-modify

docker build --platform linux/amd64 --push \
  -t us.icr.io/resilience-forge/item-modify:latest .
```

### Step 4 — Deploy (first time)

```bash
ibmcloud ce app create \
  --name item-modify \
  --image us.icr.io/resilience-forge/item-modify:latest \
  --registry-secret icr-resilience-forge \
  --port 8080 \
  --min-scale 1 \
  --max-scale 5 \
  --cpu 0.5 \
  --memory 1G \
  --env UI_DIR=/ui \
  --env COS_INSTANCE_CRN="crn:v1:bluemix:public:cloud-object-storage:global:a/1f7277194bb748cdb1d35fd8fb85a7cb:4c1fe46b-a89d-4b23-8729-d9d8221ecf5f::" \
  --env COS_ENDPOINT="s3.us-south.cloud-object-storage.appdomain.cloud" \
  --env COS_BUCKET="resilience-app" \
  --env SM_INSTANCE_URL="https://d53e26db-b7c1-461d-94e6-8de136174d04.eu-gb.secrets-manager.appdomain.cloud" \
  --env SM_SECRET_ID="2c54b838-360b-fc07-cff9-3eb9a13075e1" \
  --env SM_API_KEY="<your-iam-api-key>" \
  --no-wait
```

### Step 5 — Update an existing deployment

```bash
# After rebuilding and pushing a new image:
ibmcloud ce app update \
  --name item-modify \
  --image us.icr.io/resilience-forge/item-modify:latest \
  --no-wait
```

### Step 6 — Update a single env var (e.g. rotate SM_API_KEY)

```bash
ibmcloud ce app update \
  --name item-modify \
  --env SM_API_KEY="<new-key>" \
  --no-wait
```

### Step 7 — Check status

```bash
ibmcloud ce app get --name item-modify
```

Wait until `Status Summary: Application is ready`.

### Step 8 — View logs

```bash
ibmcloud ce app logs -f --name item-modify
```

### One-command deploy script

```bash
cd item-modify
export SM_API_KEY="<your-iam-api-key>"
./deploy.sh
```

---

## Web UI

The UI is a zero-dependency single-page application served at **`/`**.
File: [`ui/index.html`](ui/index.html)

### URLs

| Environment | URL |
|---|---|
| Local | `http://localhost:8080` |
| IBM Code Engine | `https://item-modify.27wc2juyqulv.us-south.codeengine.appdomain.cloud` |

### Features

| Feature | Detail |
|---------|--------|
| **Create item** | Name + Value form at the top; `name` is required, `value` is optional |
| **List items** | Table showing ID, Name, Value, and last-updated timestamp |
| **Edit item** | Click **Edit** to open a modal pre-filled with current values; saved via `PUT` |
| **Delete item** | Click **Delete** → confirmation prompt → removes the item |
| **Refresh** | Manual **Refresh** button re-fetches the list from COS |
| **Toast notifications** | Green (success) / red (error) pop-up for every operation |

### How the UI calls the API

| Action | Method | Endpoint |
|--------|--------|----------|
| Page load | `GET` | `/api/items` |
| Submit Create form | `POST` | `/api/items` |
| Click Edit → Save | `PUT` | `/api/items/:id` |
| Click Delete → Confirm | `DELETE` | `/api/items/:id` |
| Click Refresh | `GET` | `/api/items` |

---

## API reference

Base path: `/api/items`

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/items` | Create an item |
| `GET` | `/api/items` | List all items |
| `GET` | `/api/items/:id` | Get a single item |
| `PUT` | `/api/items/:id` | Update an item |
| `DELETE` | `/api/items/:id` | Delete an item |

### Item object

```json
{
  "id":         "3f8a2c1d...",
  "name":       "widget",
  "value":      "blue",
  "created_at": "2026-07-17T10:00:00Z",
  "updated_at": "2026-07-17T10:05:00Z"
}
```

### Create / Update body

```json
{ "name": "widget", "value": "blue" }
```

`name` is required. `value` is optional.

### curl examples

```bash
BASE=https://item-modify.27wc2juyqulv.us-south.codeengine.appdomain.cloud

# Create
curl -s -X POST $BASE/api/items \
  -H "Content-Type: application/json" \
  -d '{"name":"widget","value":"blue"}' | jq

# List
curl -s $BASE/api/items | jq

# Get  (replace <id>)
curl -s $BASE/api/items/<id> | jq

# Update
curl -s -X PUT $BASE/api/items/<id> \
  -H "Content-Type: application/json" \
  -d '{"name":"widget","value":"red"}' | jq

# Delete
curl -s -X DELETE $BASE/api/items/<id>
```

---

## IBM Cloud resources

| Resource | Name | ID |
|---|---|---|
| Code Engine project | `event-notification-server` | `5b578840-3506-444d-a21a-d5a768c6e324` |
| Code Engine app | `item-modify` | `6a65c9cc-4c76-4b59-8831-8d7cd8fc63dc` |
| Container Registry | `us.icr.io/resilience-forge/item-modify` | namespace: `resilience-forge` |
| Cloud Object Storage | `resilience-forge` | `4c1fe46b-a89d-4b23-8729-d9d8221ecf5f` |
| COS Bucket | `resilience-app` | region: `us-south-smart` |
| Secrets Manager | `eu-gb` instance | `d53e26db-b7c1-461d-94e6-8de136174d04` |
| IAM Credentials Secret | `resilience-forge-iam-api-key` | `2c54b838-360b-fc07-cff9-3eb9a13075e1` |
