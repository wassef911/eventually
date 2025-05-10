# Docs in progress, still an ongoing repo!

## Prerequisites

- Docker
- Docker Compose
- go 1.24

## Local Setup

1. Copy `.env.example` to `.env`:
   ```sh
   cp .env.example .env
   go mod download
   ```

2. Start the application:
   ```sh
   make start
   ```

3. To run tests:
   ```sh
   make test
   ```

## Swagger

The REST API documentation is available at:  http://localhost:5007/swagger/index.html

## Project Structure

### Internal Structure (DDD Approach)

The internal structure follows Domain-Driven Design (DDD) principles:

```sh
/internal
    ├── api
    │   ├── constants
    │   ├── dto
    │   ├── handlers
    │   ├── middlewares
    │   ├── server.go
    │   └── utils
    ├── delivery
    │   ├── aggregate
    │   ├── commands
    │   ├── events
    │   ├── models
    │   ├── projections
    │   ├── queries
    │   ├── repository
    │   └── services
    └── infrastructure
        ├── elasticsearch
        ├── es
        ├── eventstore
        ├── mongodb
        └── tracing
```

The structure separates concerns according to DDD layers:
- `api`: Presentation layer handling HTTP requests/responses
- `delivery`: Core domain logic and business rules
- `infrastructure`: Technical implementation details and external integrations

#### Plant UML:

![Puml](http://www.plantuml.com/plantuml/proxy?cache=no&src=https://raw.githubusercontent.com/wassef911/eventually/refs/heads/main/internal.puml)

---

### Cluster Setup Notes

1. **Grafana Agent Operator**
   Followed the [official documentation](https://grafana.com/docs/agent/latest/operator/getting-started/) for setup.
   - Manually created Grafana CRDs as required by their installation process
   - Modified Loki deployment to remove strict anti-affinity rules (could alternatively use `soft` anti-affinity)

2. **Key Configuration**
   - Kustomize overlays for environment-specific configurations, although only worked on a "prod" setup.
   - Separate RBAC restrictions for **developer-base** access in production VS **sre**.
   - A base cluster policy using **kyverno** to ensure all namespaced resources are in **[dev-prod-monitoring]**.

3. **Observability Stack**
   - Jaeger for distributed tracing
   - Loki for log aggregation
   - Grafana Agent for metrics collection

#### Cluster Diagram
![current cluster](./diagram_cluster.png)

#### Cluster Structure
```sh
    ├────── base
    │   ├────── api
    │   ├────── configs
    │   │   ├────────── clusterroles.yaml
    │   │   └────────── policy.yaml
    │   │   ├────────── configmaps.yaml
    │   │   ├────────── kustomization.yaml
                        # git ignored and needs to be created manually...
    │   │   ├────────── github-registry-secret.yaml
    │   │   ├────────── mongodb-secret.yaml

    │   ├────── elasticsearch
    │   ├────── eventstore
    │   ├────── ingress.yaml
    │   ├────── jaeger
    │   ├────── kustomization.yaml
    │   └────── mongodb
    ├────── overlays
    │   └────── prod
    │       ├────── charts
    │       ├────────── kustomization.yaml
    │       ├────────── kustomizeconfig.yaml
    │       ├────── loki
    │       │   ├────────── generator.yaml
    │       │   └────────── values.yaml
    │       └────── patches
    │           ├────────── api-svc.yaml
    │           └────────── restrict-developer-permissions.yaml
```
