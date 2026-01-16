# Future Roadmap & Project Improvements

This document outlines potential feature enhancements and project improvements planned for `Cloud Sentinel`.

## 1. Testing & Quality
- **`TESTING.md`**: Strategies for unit, integration, and E2E testing.
- **`e2e/`**: Directory for End-to-End tests (e.g., Playwright/Cypress).

## 2. Deployment & IaC
- **`k8s/`**: Manifests or Helm charts for deploying Cloud Sentinel.
- **`terraform/`**: IaC for provisioning required cloud resources.

## 3. Proposed Feature Enhancements

### Compute & Orchestration
- **Metrics Dashboard**: Integration with Prometheus/Grafana to show real-time CPU/Memory usage graphs for Pods and Nodes.
- **Resource Editor**: Advanced YAML editor with schema validation for safely editing live resources (beyond simple updates).

### Security & Governance
- **Vulnerability Scanning**: Integration with scanners (e.g., Trivy, Clair) to display vulnerability reports for running images.
- **Policy Enforcement**: Integration with OPA/Kyverno to visualize policy violations.

### Observability & Troubleshooting
- **Advanced Log Search**: Enhancing the existing Log Viewer with regex search, time-range filtering, and download capabilities.
- **Event Timeline**: A graphical timeline of cluster events to correlate cascading failures.
- **Network Policy Viewer**: Visualizer for allowed/blocked traffic flows between namespaces/pods.

### Cost & Operations
- **Cost Estimates**: Integration with OpenCost/Kubecost to display estimated run-rate for namespaces/workloads.
- **GitOps Status**: Native integration to view reconciliation status of Flux or ArgoCD resources.
- **Alerting Integration**: Webhooks for routing critical cluster alerts to Slack, Teams, or PagerDuty.
