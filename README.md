# Docs in progress, still an ongoing repo

#### platUML

![Puml](http://www.plantuml.com/plantuml/proxy?cache=no&src=https://raw.githubusercontent.com/wassef911/eventually/refs/heads/main/internal.puml)

will provide something more detailed later...

under /internal
```sh
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


#### current cluster
![current cluster](./diagram_cluster.png)


#### trying to achieve:
under /deployment

```sh
├── base
│   ├── api
│   │   ├── deployment.yaml
│   │   ├── kustomization.yaml
│   │   └── service.yaml
│   ├── configs
│   │   ├── configmaps.yaml
│   │   ├── github-registry-secret.yaml
│   │   ├── kustomization.yaml
│   │   ├── mongodb-secret.yaml
│   │   └── rbac
│   │       ├── clusterroles.yaml
│   │       ├── clusterrolebindings.yaml
│   │       ├── roles.yaml
│   │       ├── rolebindings.yaml
│   │       ├── serviceaccounts.yaml
│   │       └── kustomization.yaml
│   ├── hpc
│   │   ├── batch-operator
│   │   ├── mpi-operator
│   │   ├── gpu-operator
│   │   ├── slurm
│   │   └── kustomization.yaml
│   ├── elasticsearch
│   │   ├── kustomization.yaml
│   │   ├── service.yaml
│   │   └── statefulset.yaml
│   ├── eventstore
│   │   ├── kustomization.yaml
│   │   ├── service.yaml
│   │   └── statefulset.yaml
│   ├── jaeger
│   │   ├── deployment.yaml
│   │   ├── kustomization.yaml
│   │   └── service.yaml
│   ├── kustomization.yaml
│   └── mongodb                 <---- might swap it for a realistic setup using a known chart
│       ├── kustomization.yaml
│       ├── service.yaml
│       └── statefulset.yaml
└── overlays
    └── prod
        ├── charts
        ├── grafana
        ├── helm-loki
        ├── kustomization.yaml
        ├── kustomizeconfig.yaml
        ├── patch-api-svc.yaml
        ├── network-policies
        ├── pod-security
        └── resource-quotas
```
---
### to mention in docs
https://grafana.com/docs/agent/latest/operator/getting-started/

* had to create Grafana CRDs by hand... (mentioned in their docs)

* removed anti-affinity rules from Loki, but could've been turned soft


---
### TODO:
- add loki?
- add rbac?
- add svc mesh?
- add argo?
