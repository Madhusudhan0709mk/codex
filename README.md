# AI-Powered Recruitment Platform (Candidate Marketplace)

> **Production-ready, AI-first, Kafka-driven microservices platform for candidate discovery**

This repository provides a **build-ready implementation skeleton** for a scalable, event-driven recruitment platform that lists **only candidate profiles** (no job postings). It includes a **complete repository structure**, **service boundaries**, **Kafka design**, **database ownership**, **AI systems**, **search + ranking**, and **deployment instructions** for Docker and Kubernetes, plus runnable service templates and local Docker Compose orchestration.

---

## Table of Contents
- [Platform Overview](#platform-overview)
- [Architecture Diagram (ASCII)](#architecture-diagram-ascii)
- [Tech Stack (Mandatory + Justified)](#tech-stack-mandatory--justified)
- [Repository Structure](#repository-structure)
- [Microservices Implementation Plan](#microservices-implementation-plan)
- [Kafka Design](#kafka-design)
- [Database Design](#database-design)
- [AI System Implementation](#ai-system-implementation)
- [Search & Recommendation Engine](#search--recommendation-engine)
- [Auth & Security](#auth--security)
- [Chat System](#chat-system)
- [Docker Setup](#docker-setup)
- [Kubernetes Setup](#kubernetes-setup)
- [CI/CD Pipeline](#cicd-pipeline)
- [Observability](#observability)
- [How to Run Locally](#how-to-run-locally)
- [How to Deploy to Kubernetes](#how-to-deploy-to-kubernetes)
- [Future Roadmap](#future-roadmap)

---

## Platform Overview
This is **not** a job-posting platform. It is a **candidate profile marketplace**.

- Candidates are visible as:
  - **Interview-Ready (Verified)**
  - **Not Interview-Ready (Unverified)**
- Recruiters search, evaluate, and initiate interview requests.
- Candidates must confirm interview requests within **X days**.
- Chat opens **only after mutual confirmation**.
- The full hiring lifecycle is tracked.

### User Roles
1. Candidate
2. Recruiter
3. College Placement Admin
4. Platform Admin

---

## Architecture Diagram (ASCII)
```
                                ┌──────────────────────────────┐
                                │          API Gateway          │
                                │ AuthN/AuthZ, Rate Limit, WAF  │
                                └───────────────┬───────────────┘
                                                │
        ┌───────────────────────────────────────┼───────────────────────────────────────┐
        │                                       │                                       │
┌───────▼────────┐                     ┌────────▼────────┐                     ┌────────▼────────┐
│ Candidate UI   │                     │ Recruiter UI    │                     │ College Admin   │
│ (Web/Mobile)   │                     │ Dashboard       │                     │ Dashboard       │
└───────┬────────┘                     └────────┬────────┘                     └────────┬────────┘
        │                                       │                                       │
        └──────────────────────┬────────────────┴────────────────┬──────────────────────┘
                               │                                 │
                      ┌────────▼────────┐               ┌────────▼────────┐
                      │   Core APIs     │               │   Admin APIs    │
                      └────────┬────────┘               └────────┬────────┘
                               │                                 │
                 ┌─────────────▼─────────────────────────────────▼─────────────┐
                 │                   Microservices (K8s)                        │
                 └─────────────┬─────────────────────────────────┬─────────────┘
                               │                                 │
                     ┌─────────▼──────────┐           ┌─────────▼──────────┐
                     │     Kafka Bus      │           │   Observability    │
                     └─────────┬──────────┘           └─────────┬──────────┘
                               │                                 │
         ┌─────────────────────▼───────────────────┐   ┌─────────▼─────────┐
         │       AI/ML Services + Feature Store     │   │ Search & Ranking  │
         │   Resume Parsing, Matching, Explain      │   │ OpenSearch/Vector │
         └─────────────────────┬───────────────────┘   └─────────┬─────────┘
                               │                                 │
                          ┌────▼────┐                      ┌─────▼─────┐
                          │ SQL DB  │                      │  NoSQL    │
                          │Postgres │                      │ Profiles  │
                          └─────────┘                      └───────────┘
```

---

## Tech Stack (Mandatory + Justified)
**Backend Language Choice: Go**
- **Why Go?** High concurrency, low latency, simple deployment, strong ecosystem for microservices, ideal for Kafka consumers and high-throughput APIs.

**Core Stack**
- **Backend**: Go (Gin/Fiber), gRPC for internal service-to-service, REST for external APIs
- **API Gateway**: Kong / Envoy
- **Event Streaming**: Apache Kafka
- **Databases**:
  - **SQL**: PostgreSQL (transactions, payments, auth)
  - **NoSQL**: MongoDB (profile documents, AI outputs)
- **Search**: OpenSearch + Vector DB (pgvector or Pinecone)
- **AI Services**: Python (FastAPI), MLflow, Feast (Feature Store)
- **Infra**: Docker, Kubernetes, Terraform
- **Observability**: Prometheus, Grafana, OpenTelemetry, Loki

---

## Repository Structure
We use a **monorepo** to simplify shared libraries, consistent CI/CD, and cross-service refactors.

```
/ (repo)
├── apps/
│   └── web/                 # Next.js recruiter & admin portal
├── services/
│   ├── api-gateway/
│   ├── identity/
│   ├── candidate-profile/
│   ├── resume-parser/
│   ├── verification/
│   ├── recruiter-search/
│   ├── decision-engine/
│   ├── recruiter-workflow/
│   ├── chat/
│   ├── placement-admin/
│   ├── billing/
│   ├── analytics/
│   └── audit-log/
├── libs/
│   ├── proto/               # gRPC contracts
│   ├── kafka/               # shared Kafka client wrappers
│   ├── auth/                # shared auth middleware
│   └── observability/       # tracing, metrics, logging
├── infra/
│   ├── docker/
│   ├── k8s/
│   └── terraform/
├── docs/
│   ├── architecture/
│   └── api/
└── README.md
```

---

## Microservices Implementation Plan
Each service owns its **data store**, exposes APIs (REST external, gRPC internal), and publishes domain events to Kafka.

| Service | Responsibility | API | Kafka Produce | Kafka Consume |
|---------|----------------|-----|--------------|---------------|
| Identity | AuthN/AuthZ, RBAC | REST/gRPC | `user.created` | `billing.subscription.updated` |
| Candidate Profile | Profile CRUD, consent | REST/gRPC | `candidate.profile.updated` | `ai.profile.enriched` |
| Resume Parser | Parse & extract | gRPC | `ai.profile.enriched` | `candidate.resume.uploaded` |
| Verification | Interview-ready workflow | REST | `candidate.verification.status.changed` | `candidate.profile.updated` |
| Recruiter Search | Search + filters | REST | `recruiter.search.executed` | `candidate.profile.updated` |
| Decision Engine | Ranking + explainability | gRPC | `ai.match.score.generated` | `recruiter.search.executed` |
| Recruiter Workflow | Shortlist + confirmation | REST | `recruiter.shortlist.created` | `candidate.consent.changed` |
| Chat | Consent-based chat | REST/WebSocket | `chat.session.opened` | `recruiter.shortlist.created` |
| Placement Admin | College dashboard | REST | `placement.status.updated` | `candidate.profile.updated` |
| Billing | Plans + payments | REST | `billing.subscription.updated` | `user.created` |
| Analytics | Reporting | REST | - | all topics (stream) |
| Audit Log | Compliance logs | gRPC | - | all topics (stream) |

---

## Kafka Design

### Topics
- `user.created`
- `candidate.resume.uploaded`
- `ai.profile.enriched`
- `candidate.profile.updated`
- `candidate.verification.status.changed`
- `recruiter.search.executed`
- `ai.match.score.generated`
- `recruiter.shortlist.created`
- `candidate.consent.changed`
- `chat.session.opened`
- `billing.subscription.updated`
- `placement.status.updated`

### Event Schema (Example: candidate.profile.updated)
```json
{
  "event_id": "uuid",
  "event_type": "candidate.profile.updated",
  "timestamp": "2025-01-01T12:00:00Z",
  "candidate_id": "uuid",
  "fields_changed": ["skills", "experience"],
  "version": 3
}
```

### Retry & DLQ Strategy
- **Retry topics**: `topic.retry.5s`, `topic.retry.1m`
- **DLQ**: `topic.dlq`
- Idempotent consumers, max attempts = 5

---

## Database Design

### Data Ownership
- Each service owns its **own schema** and writes only to its database.
- Cross-service data flow via **Kafka events**.

### SQL (Postgres)
- `users`, `roles`, `permissions`
- `candidates`, `recruiters`, `colleges`
- `shortlists`, `interview_requests`, `consents`
- `subscriptions`, `payments`, `invoices`

### NoSQL (MongoDB)
- `candidate_profiles` (full profile JSON)
- `resume_raw_data` (parsed output)
- `ai_explanations`

### Indexing Strategy
- Postgres: composite indexes on `(candidate_id, status)` and `(recruiter_id, status)`
- MongoDB: text + hashed indexes on skills and college

---

## AI System Implementation

### Resume Parsing Pipeline
1. Resume upload event (`candidate.resume.uploaded`)
2. Resume parser extracts skills/experience
3. Emit `ai.profile.enriched`
4. Candidate profile updated

### Skill Extraction Logic
- Normalization against a **skills taxonomy**
- Confidence scoring on each extracted skill

### Matching & Ranking
- **Hybrid scoring** = semantic similarity + structured filters
- Weighted features: skills, experience, college, verification status

### Feedback Loop
- Recruiter outcomes feed training dataset
- Drift detection via metrics

### Explainability
- Store per-candidate explanation in `ai_explanations`

---

## Search & Recommendation Engine
- OpenSearch for keyword filters
- Vector DB for embeddings
- Hybrid score = `0.6 * semantic + 0.4 * keyword`

---

## Auth & Security
- OAuth2 + JWT
- RBAC enforced at gateway + services
- TLS everywhere, encrypted secrets
- Audit logs for all critical actions

---

## Chat System
- Chat **only opens after mutual confirmation**
- Kafka event `chat.session.opened` triggers chat service

---

## Docker Setup
Each service has a Dockerfile. The Compose file under `infra/docker` runs Kafka, databases, search, and a subset of services for local development.

```bash
docker compose -f infra/docker/docker-compose.yml up --build
```

Example Dockerfile:
```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o service ./cmd/service

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/service .
CMD ["./service"]
```

---

## Kubernetes Setup
- One Deployment + Service per microservice (see `infra/k8s/*-deployment.yaml`)
- ConfigMaps for non-secret config
- Secrets in K8s Secrets or Vault
- HPA on CPU + latency

---

## CI/CD Pipeline
1. Lint + Unit tests
2. Build + Scan Docker images
3. Deploy to staging
4. Canary deploy to production

---

## Observability
- Logs: Loki
- Metrics: Prometheus + Grafana
- Tracing: OpenTelemetry + Jaeger

---

## How to Run Locally
```bash
make bootstrap
make up
```

Frontend:
- `http://localhost:3000`

Local endpoints:
- `GET http://localhost:8081/healthz` (identity)
- `GET http://localhost:8082/healthz` (candidate-profile)
- `POST http://localhost:8083/parse` (resume-parser)
- `GET http://localhost:8084/healthz` (recruiter-search)
- `POST http://localhost:8085/requests` (recruiter-workflow)
- `POST http://localhost:8087/sessions` (chat)

Sample API calls:
```bash
curl -X POST http://localhost:8082/candidates \\
  -H 'Content-Type: application/json' \\
  -d '{"name":"Ada Lovelace","skills":["Go","Kafka"],"readiness_status":"verified"}'

curl -X POST http://localhost:8084/index \\
  -H 'Content-Type: application/json' \\
  -d '{"id":"cand-1","name":"Ada Lovelace","skills":["Go","Kafka"],"readiness_status":"verified"}'

curl -X POST http://localhost:8084/search \\
  -H 'Content-Type: application/json' \\
  -d '{"skills":["Kafka"],"readiness_status":"verified","minimum_score":1}'
```

Integration wiring:
- `candidate-profile` auto-indexes to recruiter-search via `SEARCH_URL`.
- `recruiter-workflow` opens chat sessions on confirmation via `CHAT_URL`.

---

## VS Code Setup & Run
1. Install **Docker Desktop** (for local services) and **Go**/**Node.js** toolchains.
2. Install VS Code extensions:
   - Go
   - Docker
   - ESLint
3. Clone the repo and open `/workspace/codex` in VS Code.
4. Run the local stack from the integrated terminal:
   ```bash
   make up
   ```
5. Open the frontend at `http://localhost:3000`.
6. Use the sample API calls above to seed data.

---

## How to Deploy to Kubernetes
```bash
kubectl apply -f infra/k8s/namespace.yaml
kubectl apply -f infra/k8s/
```

---

## Future Roadmap
- Bias & fairness dashboards
- Multi-language resume parsing
- HRIS/ATS integrations
