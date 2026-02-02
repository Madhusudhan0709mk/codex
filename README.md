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
- [Scaling & Performance Strategy](#scaling--performance-strategy)
- [Fault Tolerance & Resiliency](#fault-tolerance--resiliency)
- [Caching Strategy](#caching-strategy)
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
- **Security**: OAuth2 / OIDC, JWT, Vault / KMS

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
| Identity | AuthN/AuthZ, RBAC, MFA, passwordless | REST/gRPC | `user.created` | `billing.subscription.updated` |
| Candidate Profile | Profile CRUD, consent, completeness scoring | REST/gRPC | `candidate.profile.updated` | `ai.profile.enriched` |
| Resume Parser | Parse & extract skills/experience | gRPC | `ai.profile.enriched` | `candidate.resume.uploaded` |
| Verification | Interview-ready workflow, human review | REST | `candidate.verification.status.changed` | `candidate.profile.updated` |
| Recruiter Search | Search + filters, ranking | REST | `recruiter.search.executed` | `candidate.profile.updated` |
| Decision Engine | Ranking + explainability, bias monitoring | gRPC | `ai.match.score.generated` | `recruiter.search.executed` |
| Recruiter Workflow | Shortlist + time-bound confirmation | REST | `recruiter.shortlist.created` | `candidate.consent.changed` |
| Chat | Consent-based chat, audit trail | REST/WebSocket | `chat.session.opened` | `recruiter.shortlist.created` |
| Placement Admin | College dashboard, student tracking | REST | `placement.status.updated` | `candidate.profile.updated` |
| Billing | Plans + payments, invoices | REST | `billing.subscription.updated` | `user.created` |
| Analytics | Reporting, KPI dashboards, funnel metrics | REST | - | all topics (stream) |
| Audit Log | Compliance logs, access tracking | gRPC | - | all topics (stream) |

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

### Event Flow Example
1. Candidate uploads resume → `candidate.resume.uploaded`
2. Resume Parsing Service consumes, extracts skills → `ai.profile.enriched`
3. Profile Service updates completeness score → `candidate.profile.updated`
4. Search Service re-indexes candidate in OpenSearch
5. AI Scoring Service updates ranking features in Feature Store

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
- `users` (id, role, auth_provider_id)
- `candidates` (id, user_id, readiness_status, completeness_score)
- `candidate_profiles` (education, skills, experience, resume_url)
- `recruiters` (id, company_id, plan_id)
- `roles`, `permissions`
- `shortlists` (id, recruiter_id, candidate_id, status, expires_at)
- `interview_requests`, `consents` (candidate_id, recruiter_id, status)
- `subscriptions` (user_id, plan_id, status, billing_cycle), `payments`, `invoices`
- `audit_logs` (actor_id, action, entity, timestamp)

### NoSQL (MongoDB)
- `candidate_profiles` (full profile JSON)
- `resume_raw_data` (parsed output)
- `ai_explanations`
- `candidate_activity_logs`

### Search / Vector
- OpenSearch index: `candidate_profiles`
- Vector DB: embeddings for semantic matching (pgvector / Pinecone)

### Indexing Strategy
- Postgres: composite indexes on `(candidate_id, status)` and `(recruiter_id, status)`
- MongoDB: text + hashed indexes on skills and college
- OpenSearch: multi-field analysis for skills, experience

---

## AI System Implementation

### Components
1. **Resume Parsing** → Extract structured data from PDF/DOC
2. **Skill Extraction** → Standardized taxonomy with confidence scoring
3. **Candidate Matching** → Match vs recruiter-defined ideal profile
4. **Ranking & Scoring** → Multi-factor score (hybrid approach)
5. **Decision Explanation** → "Why selected" justification
6. **Feedback Loop** → Recruiter outcomes feed retraining

### Resume Parsing Pipeline
1. Resume upload event (`candidate.resume.uploaded`)
2. Resume parser extracts skills/experience using NLP
3. Emit `ai.profile.enriched` with structured data
4. Candidate profile updated with auto-filled fields

### Skill Extraction Logic
- Normalization against a **skills taxonomy**
- Confidence scoring on each extracted skill
- Synonym mapping (e.g., "React.js" → "React")

### Matching & Ranking
- **Hybrid scoring** = semantic similarity + structured filters
- Weighted features: skills (40%), experience (30%), college (15%), verification status (15%)
- Personalized ranking based on recruiter history

### Feature Store
- Stores derived features (skills, experience, success rates)
- Decouples training from online inference
- Low-latency feature serving for real-time scoring

### Model Lifecycle
- Offline training (batch + scheduled)
- Online inference via API
- Continuous evaluation & drift detection
- A/B testing for new models

### Feedback Loop
- Recruiter outcomes (shortlist, interview, hire) feed training dataset
- Drift detection via metrics
- Model retraining triggers

### Explainability
- Store per-candidate explanation in `ai_explanations` collection
- Provide reasoning: "Matched on skills: Go, Kafka; 5+ years experience"

### Bias Monitoring
- Fairness checks on protected attributes
- Regular audits of ranking disparities
- Dashboard for bias metrics

---

## Search & Recommendation Engine
- **Keyword search** via OpenSearch with multi-field queries
- **Semantic search** via Vector DB (embeddings)
- **Hybrid scoring**: `0.6 * semantic_score + 0.4 * keyword_score`
- Personalized ranking based on recruiter search history
- Filters: skills, experience, location, verification status, college

---

## Auth & Security
- **OAuth2 / OIDC** with JWT tokens
- **RBAC** enforced at gateway + service level
- **MFA** and passwordless login options
- **TLS everywhere**, encrypted secrets (Vault / KMS)
- **Audit logs** for all critical actions
- **PII data minimization** & masking
- **Encryption at rest** for sensitive data

---

## Chat System
- Chat **only opens after mutual confirmation**
- Kafka event `chat.session.opened` triggers chat service
- WebSocket for real-time messaging
- Message audit trail for compliance
- Consent revocation immediately closes chat

---

## Scaling & Performance Strategy
- **Stateless services** scale horizontally with HPA
- Kafka partitions scale throughput (partition by `candidate_id`)
- Read-heavy endpoints cached via Redis (TTL-based)
- Async processing for AI/ETL workflows (decouple compute)
- Connection pooling for databases
- gRPC for low-latency inter-service communication

---

## Fault Tolerance & Resiliency
- **Retries + DLQ** for Kafka consumers (exponential backoff)
- **Circuit breakers** on external APIs (prevent cascading failures)
- **Multi-AZ database replication** for high availability
- **Graceful degradation** for AI services (fallback to rule-based scoring)
- Health checks and readiness probes in Kubernetes
- Rate limiting to prevent abuse

---

## Caching Strategy
- **Redis** for session tokens, profile caches
- Search results cached with TTL (5-15 minutes)
- AI inference responses cached short-term (1-5 minutes)
- Cache invalidation on `candidate.profile.updated` events
- Distributed caching for multi-region deployments

---

## Docker Setup
Each service has a Dockerfile. The Compose file under `infra/docker` runs Kafka, databases, search, and a subset of services for local development.

```bash
docker compose -f infra/docker/docker-compose.yml up --build
```

Example Dockerfile (Go service):
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
- **HPA on CPU + latency** metrics
- **Namespaces** per environment (dev, staging, prod)
- **Helm charts** for templated deployments
- **Ingress + API Gateway** for external traffic

### Auto-Scaling Methods
- **HPA**: Scale pods based on CPU/memory/custom metrics
- **VPA**: Optimize resource requests/limits
- **Cluster Autoscaler** for node pools

### Deployment Patterns
- **Blue-Green** for zero downtime
- **Canary** for gradual rollout (10% → 50% → 100%)

---

## CI/CD Pipeline
1. **Code push** triggers CI
2. **Lint + Unit tests**
3. **Build + Scan Docker images** (security scanning)
4. **Deploy to staging** (Helm upgrade)
5. **Integration tests** on staging
6. **Canary deploy to production** (monitor metrics)
7. **Full rollout** or rollback based on health

### CI/CD Tools
- GitHub Actions / GitLab CI / Jenkins
- ArgoCD for GitOps
- Helm for package management

---

## Observability
- **Logs**: Loki (centralized logging)
- **Metrics**: Prometheus + Grafana dashboards
- **Tracing**: OpenTelemetry + Jaeger (distributed tracing)
- **Alerting**: Prometheus Alertmanager (on-call alerts)
- **SLOs**: Track latency, error rate, availability

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
# Create candidate profile
curl -X POST http://localhost:8082/candidates \
  -H 'Content-Type: application/json' \
  -d '{"name":"Ada Lovelace","skills":["Go","Kafka"],"readiness_status":"verified"}'

# Index candidate in search
curl -X POST http://localhost:8084/index \
  -H 'Content-Type: application/json' \
  -d '{"id":"cand-1","name":"Ada Lovelace","skills":["Go","Kafka"],"readiness_status":"verified"}'

# Search candidates
curl -X POST http://localhost:8084/search \
  -H 'Content-Type: application/json' \
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
   - Kubernetes
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
- **Bias & fairness dashboards** with real-time monitoring
- **Multi-language resume parsing** (support 10+ languages)
- **Candidate career path recommendations** (AI-driven)
- **HRIS/ATS integrations** (Greenhouse, Lever, Workday)
- **Mobile apps** (native iOS/Android)
- **Advanced analytics** (predictive hiring success models)
- **Blockchain-based verification** for credentials

---

## Deliverables Checklist
- ✅ High-level architecture diagram
- ✅ Microservices list & responsibilities
- ✅ Kafka topics & event flow
- ✅ Database design (SQL + NoSQL)
- ✅ AI components architecture
- ✅ Search & recommendation engine design
- ✅ Scaling, fault tolerance, caching strategies
- ✅ Security best practices (OAuth2, RBAC, TLS)
- ✅ CI/CD pipeline design
- ✅ Docker setup & local development guide
- ✅ Kubernetes deployment & auto-scaling
- ✅ Observability stack (logs, metrics, traces)
- ✅ Future roadmap

---

**Platform is ready for implementation. Start with identity service, then candidate profile, then resume parser. Scale iteratively.**