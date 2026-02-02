# AI-Powered Recruitment Platform (Candidate-Only Marketplace)

> **Production-ready, event-driven, AI-first architecture for enterprise-scale candidate discovery**

## Platform Overview
This platform lists **only candidate profiles**. Recruiters search and select candidates; **no job postings** exist. Candidates are visible with a clear status:

- **Interview-Ready (Verified)**
- **Not Interview-Ready (Unverified)**

The system is designed for **millions of users**, **enterprise-grade performance**, and **AI-first decisioning** with **Kafka**-based event-driven microservices on **Kubernetes**.

## High-Level Architecture (ASCII)
```
                              ┌───────────────────────────┐
                              │        API Gateway        │
                              │  AuthN/AuthZ, Rate Limit  │
                              └─────────────┬─────────────┘
                                            │
        ┌───────────────────────────────────┼───────────────────────────────────┐
        │                                   │                                   │
┌───────▼────────┐                 ┌────────▼────────┐                 ┌────────▼────────┐
│  Web / Mobile  │                 │  Recruiter UI   │                 │ College Admin  │
│  Candidate UI  │                 │  Dashboard UI   │                 │  Dashboard UI  │
└───────┬────────┘                 └────────┬────────┘                 └────────┬────────┘
        │                                   │                                   │
        └───────────────┬───────────────────┴───────────────┬───────────────────┘
                        │                                   │
                 ┌──────▼──────┐                     ┌──────▼──────┐
                 │  Core APIs  │                     │  Admin APIs │
                 └──────┬──────┘                     └──────┬──────┘
                        │                                   │
        ┌───────────────┴───────────────┐      ┌────────────┴─────────────┐
        │        Microservices          │      │      Observability       │
        │ (K8s + Docker Containers)     │      │  Logs, Metrics, Traces   │
        └───────────────┬───────────────┘      └────────────┬─────────────┘
                        │                                   │
                 ┌──────▼──────────────────────────────────▼──────┐
                 │               Kafka Event Bus                  │
                 └──────┬──────────────────────────────────┬──────┘
                        │                                  │
         ┌──────────────▼──────────────┐        ┌──────────▼───────────┐
         │  AI/ML Pipeline & Feature   │        │ Search/Rank/Recommend │
         │  Store (async + GPU-ready)  │        │ (Vector + Keyword)    │
         └──────────────┬──────────────┘        └──────────┬───────────┘
                        │                                  │
                 ┌──────▼──────┐                     ┌─────▼─────┐
                 │ Data Layer  │                     │  Caching  │
                 │ SQL + NoSQL │                     │  Redis    │
                 └─────────────┘                     └───────────┘
```

## Tech Stack (Suggested)
- **Frontend**: React / Next.js, TypeScript, Tailwind
- **Backend**: Go / Java / Node.js (NestJS)
- **Data**: PostgreSQL, MongoDB, Elasticsearch/OpenSearch, Redis
- **Event Bus**: Apache Kafka
- **AI/ML**: Python (FastAPI), MLflow, Feature Store (Feast), Vector DB (pgvector / Pinecone)
- **Infrastructure**: Docker, Kubernetes (EKS/GKE/AKS), Terraform
- **Observability**: Prometheus, Grafana, Loki, OpenTelemetry
- **Security**: OAuth2 / OIDC, JWT, Vault / KMS

---

## Microservices Breakdown

### 1. **Identity & Access Service**
- OAuth2/OIDC, SSO, RBAC
- MFA, passwordless login
- Token lifecycle, refresh, revocation

### 2. **Candidate Profile Service**
- Profile CRUD, resume upload
- Profile completeness scoring
- Availability and consent management

### 3. **Resume Parsing & AI Extraction Service**
- PDF/DOC ingestion
- Skill extraction, experience parsing
- Auto-fill profile data

### 4. **Verification Workflow Service**
- Interview-ready verification steps
- Human review + AI validation
- Status transitions

### 5. **Recruiter Search & Discovery Service**
- Search API + filters (skills, readiness)
- Ranking & match scoring

### 6. **Decision Engine / AI Scoring Service**
- Best-fit scoring
- Explainability (why selected)
- Bias monitoring + fairness checks

### 7. **Recruiter Workflow Service**
- Shortlist & candidate selection
- Time-bound confirmation flow
- Status updates (Confirmed / Not Interested)

### 8. **Chat & Consent Service**
- Chat enabled only on mutual consent
- Audit of consent and message activity

### 9. **College Placement Service**
- Student tracking, placement status
- Employer interactions
- Analytics dashboard

### 10. **Subscription & Billing Service**
- Membership plans
- Payment + invoices

### 11. **Analytics & Reporting Service**
- KPI dashboards
- Funnel & placement metrics

### 12. **Audit Log Service**
- Immutable audit trail
- Access log for compliance

---

## Kafka Topics & Event Flow

### Core Topics
- `candidate.profile.updated`
- `candidate.resume.uploaded`
- `candidate.verification.status.changed`
- `recruiter.search.executed`
- `recruiter.shortlist.created`
- `candidate.consent.changed`
- `ai.profile.enriched`
- `ai.match.score.generated`
- `chat.session.opened`
- `billing.subscription.updated`

### Event Flow Example
1. Candidate uploads resume → `candidate.resume.uploaded`
2. Resume Parsing Service consumes, extracts skills → `ai.profile.enriched`
3. Profile Service updates completeness score → `candidate.profile.updated`
4. Search Service re-indexes candidate in OpenSearch
5. AI Scoring Service updates ranking features in Feature Store

---

## Database Design (High-Level)

### SQL (PostgreSQL)
- `users` (id, role, auth_provider_id)
- `candidates` (id, user_id, readiness_status, completeness_score)
- `candidate_profiles` (education, skills, experience, resume_url)
- `recruiters` (id, company_id, plan_id)
- `shortlists` (id, recruiter_id, candidate_id, status, expires_at)
- `consents` (candidate_id, recruiter_id, status)
- `subscriptions` (user_id, plan_id, status, billing_cycle)
- `audit_logs` (actor_id, action, entity, timestamp)

### NoSQL (MongoDB)
- `resume_raw_data`
- `candidate_activity_logs`
- `ai_explanations`

### Search / Vector
- OpenSearch index: candidate_profiles
- Vector DB: embeddings for semantic matching

---

## AI System Design

### Components
1. **Resume Parsing** → Extract structured data
2. **Skill Extraction** → Standardized taxonomy
3. **Candidate-Job Matching** → Match vs recruiter-defined ideal profile
4. **Ranking & Scoring** → Multi-factor score
5. **Decision Explanation** → “Why selected” justification
6. **Feedback Loop** → Recruiter outcomes feed retraining

### Feature Store
- Stores derived features (skills, experience, success rates)
- Decouples training from online inference

### Model Lifecycle
- Offline training (batch + scheduled)
- Online inference via API
- Continuous evaluation & drift detection

---

## Search & Recommendation Engine
- **Keyword search** via OpenSearch
- **Semantic search** via Vector DB
- Hybrid scoring: relevance + AI rank score
- Personalized ranking based on recruiter history

---

## Scaling & Performance Strategy
- **Stateless services** scale horizontally
- Kafka partitions scale throughput
- Read-heavy endpoints cached via Redis
- Async processing for AI/ETL workflows

---

## Fault Tolerance & Resiliency
- Retries + DLQ for Kafka consumers
- Circuit breakers on external APIs
- Multi-AZ database replication
- Graceful degradation for AI services

---

## Caching Strategy
- Redis for session tokens, profile caches
- Search results cached with TTL
- AI inference responses cached short-term

---

## API Gateway Pattern
- Centralized authentication & authorization
- Request validation
- Rate limiting
- API versioning

---

## Security Best Practices
- OAuth2 / OIDC with JWT
- RBAC enforced at gateway + service level
- Encryption at rest + in transit (TLS)
- Audit logs for all access events
- PII data minimization & masking

---

## CI/CD Pipeline
1. Code push triggers CI
2. Unit + integration tests
3. Container build + scan
4. Deploy to staging (Helm)
5. Canary / blue-green deploy to prod

---

## Docker Architecture
- Each microservice packaged as Docker image
- Images stored in registry (ECR/GCR/ACR)
- Immutable version tagging

---

## Kubernetes Deployment Strategy
- Namespaces per environment
- Helm charts for services
- Ingress + API Gateway
- HPA/VPA for auto scaling

---

## Auto-Scaling Methods
- **HPA**: Scale pods based on CPU/latency
- **VPA**: Optimize resource requests
- **Cluster Autoscaler** for node pools

---

## Deployment Patterns
- **Blue-Green** for zero downtime
- **Canary** for gradual rollout

---

## Future Roadmap
- Multi-language AI resume parsing
- Advanced fairness + bias dashboards
- Candidate career path recommendations
- Integration with HRIS/ATS systems

---

## Deliverables Checklist (Mapped)
- ✅ High-level architecture diagram
- ✅ Microservices list & responsibilities
- ✅ Kafka topics & event flow
- ✅ AI components architecture
- ✅ Search & recommendation engine design
- ✅ Scaling, fault tolerance, caching
- ✅ Security, CI/CD, Docker, K8s, auto-scaling
