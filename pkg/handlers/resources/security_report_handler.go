package resources

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/pixelvide/cloud-sentinel-k8s/pkg/cluster"
	"github.com/pixelvide/cloud-sentinel-k8s/pkg/model"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SecurityReportHandler struct{}

func NewSecurityReportHandler() *SecurityReportHandler {
	return &SecurityReportHandler{}
}

var vulnerabilityReportKind = schema.GroupVersionKind{
	Group:   "aquasecurity.github.io",
	Version: "v1alpha1",
	Kind:    "VulnerabilityReport",
}

var clusterVulnerabilityReportKind = schema.GroupVersionKind{
	Group:   "aquasecurity.github.io",
	Version: "v1alpha1",
	Kind:    "ClusterVulnerabilityReport",
}

var configAuditReportKind = schema.GroupVersionKind{
	Group:   "aquasecurity.github.io",
	Version: "v1alpha1",
	Kind:    "ConfigAuditReport",
}

var exposedSecretReportKind = schema.GroupVersionKind{
	Group:   "aquasecurity.github.io",
	Version: "v1alpha1",
	Kind:    "ExposedSecretReport",
}

var clusterComplianceReportKind = schema.GroupVersionKind{
	Group:   "aquasecurity.github.io",
	Version: "v1alpha1",
	Kind:    "ClusterComplianceReport",
}

var infraAssessmentReportKind = schema.GroupVersionKind{
	Group:   "aquasecurity.github.io",
	Version: "v1alpha1",
	Kind:    "InfraAssessmentReport",
}

var clusterInfraAssessmentReportKind = schema.GroupVersionKind{
	Group:   "aquasecurity.github.io",
	Version: "v1alpha1",
	Kind:    "ClusterInfraAssessmentReport",
}

var rbacAssessmentReportKind = schema.GroupVersionKind{
	Group:   "aquasecurity.github.io",
	Version: "v1alpha1",
	Kind:    "RbacAssessmentReport",
}

var clusterRbacAssessmentReportKind = schema.GroupVersionKind{
	Group:   "aquasecurity.github.io",
	Version: "v1alpha1",
	Kind:    "ClusterRbacAssessmentReport",
}

// CheckStatus checks if the Trivy Operator is installed by looking for the CRD
func (h *SecurityReportHandler) CheckStatus(c *gin.Context) {
	cs := c.MustGet("cluster").(*cluster.ClientSet)

	// Check if the CRD exists
	var crd apiextensionsv1.CustomResourceDefinition
	err := cs.K8sClient.Get(c.Request.Context(), client.ObjectKey{Name: "vulnerabilityreports.aquasecurity.github.io"}, &crd)

	installed := err == nil
	c.JSON(http.StatusOK, model.SecurityStatusResponse{TrivyInstalled: installed})
}

// listReportsGeneric is a helper to list and filter reports based on common logic
func (h *SecurityReportHandler) listReportsGeneric(
	c *gin.Context,
	reportGVK schema.GroupVersionKind,
	crdName string,
	workloadKind string,
	workloadName string,
	namespace string,
) ([]unstructured.Unstructured, error) {
	cs := c.MustGet("cluster").(*cluster.ClientSet)

	// Check if CRD exists
	var crd apiextensionsv1.CustomResourceDefinition
	if err := cs.K8sClient.Get(c.Request.Context(), client.ObjectKey{Name: crdName}, &crd); err != nil {
		return []unstructured.Unstructured{}, nil //nolint:nilerr // Intentional: if CRD missing, return empty list, no error
	}

	var list unstructured.UnstructuredList
	list.SetGroupVersionKind(reportGVK)
	opts := []client.ListOption{}

	if namespace != "" {
		opts = append(opts, client.InNamespace(namespace))
	}

	switch {
	case workloadKind == "Deployment":
		var rsList appsv1.ReplicaSetList
		if err := cs.K8sClient.List(c.Request.Context(), &rsList, client.InNamespace(namespace)); err != nil {
			return nil, fmt.Errorf("failed to list recyclasets: %w", err)
		}

		var targetRSNames []string
		for _, rs := range rsList.Items {
			for _, owner := range rs.OwnerReferences {
				if owner.Kind == "Deployment" && owner.Name == workloadName {
					targetRSNames = append(targetRSNames, rs.Name)
					break
				}
			}
		}

		if len(targetRSNames) == 0 {
			return []unstructured.Unstructured{}, nil
		}

		labels := client.MatchingLabels{
			"trivy-operator.resource.kind": "ReplicaSet",
		}
		opts = append(opts, labels)

		if err := cs.K8sClient.List(c.Request.Context(), &list, opts...); err != nil {
			return nil, fmt.Errorf("failed to list reports: %w", err)
		}

		filteredItems := []unstructured.Unstructured{}
		for _, item := range list.Items {
			lbls := item.GetLabels()
			reportResourceName := lbls["trivy-operator.resource.name"]
			for _, target := range targetRSNames {
				if reportResourceName == target {
					filteredItems = append(filteredItems, item)
					break
				}
			}
		}
		return filteredItems, nil

	case workloadKind != "" && workloadName != "":
		labels := client.MatchingLabels{
			"trivy-operator.resource.kind": workloadKind,
			"trivy-operator.resource.name": workloadName,
		}
		opts = append(opts, labels)
		if err := cs.K8sClient.List(c.Request.Context(), &list, opts...); err != nil {
			return nil, fmt.Errorf("failed to list reports: %w", err)
		}
		return list.Items, nil

	default:
		if err := cs.K8sClient.List(c.Request.Context(), &list, opts...); err != nil {
			return nil, fmt.Errorf("failed to list reports: %w", err)
		}
		return list.Items, nil
	}
}

// ListReports fetches vulnerability reports, optionally filtered by workload
func (h *SecurityReportHandler) ListReports(c *gin.Context) {
	cs := c.MustGet("cluster").(*cluster.ClientSet)
	namespace := c.Query("namespace")
	workloadKind := c.Query("workloadKind") // e.g. Pod, Deployment
	workloadName := c.Query("workloadName")

	if namespace == "" && workloadKind != "Node" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace is required for namespaced resources"})
		return
	}

	// 1. Special handling for Node (ClusterVulnerabilityReport) - separate path
	if workloadKind == "Node" {
		// ClusterVulnerabilityReport is cluster-scoped
		var crd apiextensionsv1.CustomResourceDefinition
		if err := cs.K8sClient.Get(c.Request.Context(), client.ObjectKey{Name: "clustervulnerabilityreports.aquasecurity.github.io"}, &crd); err != nil {
			c.JSON(http.StatusOK, model.VulnerabilityReportList{Items: []model.VulnerabilityReport{}})
			return
		}
		var list unstructured.UnstructuredList
		list.SetGroupVersionKind(clusterVulnerabilityReportKind)

		if err := cs.K8sClient.List(c.Request.Context(), &list); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to list cluster vulnerability reports: %v", err)})
			return
		}

		reports := make([]model.VulnerabilityReport, 0, len(list.Items))
		for _, u := range list.Items {
			var report model.VulnerabilityReport
			if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &report); err != nil {
				continue
			}
			reports = append(reports, report)
		}
		c.JSON(http.StatusOK, model.VulnerabilityReportList{Items: reports})
		return
	}

	// 2. Resolve Pod owner if needed
	if workloadKind == "Pod" {
		var pod corev1.Pod
		if err := cs.K8sClient.Get(c.Request.Context(), client.ObjectKey{Namespace: namespace, Name: workloadName}, &pod); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "pod not found"})
			return
		}

		ownerKind := ""
		ownerName := ""

		for _, owner := range pod.OwnerReferences {
			if owner.Controller != nil && *owner.Controller {
				ownerKind = owner.Kind
				ownerName = owner.Name
				break
			}
		}

		switch ownerKind {
		case "":
			ownerKind = "Pod"
			ownerName = workloadName
		case "ReplicaSet":
			// ReplicaSet will be handled by listReportsGeneric if it recurses.
			// Currently, we just look up reports for the RS.
		}

		workloadKind = ownerKind
		workloadName = ownerName
	}

	// 3. Use generic helper
	items, err := h.listReportsGeneric(c, vulnerabilityReportKind, "vulnerabilityreports.aquasecurity.github.io", workloadKind, workloadName, namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	reports := make([]model.VulnerabilityReport, 0, len(items))
	for _, u := range items {
		var report model.VulnerabilityReport
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &report); err != nil {
			continue
		}
		reports = append(reports, report)
	}

	c.JSON(http.StatusOK, model.VulnerabilityReportList{Items: reports})
}

// GetClusterSummary aggregates vulnerabilities across the entire cluster (or filtered namespace)
func (h *SecurityReportHandler) GetClusterSummary(c *gin.Context) {
	cs := c.MustGet("cluster").(*cluster.ClientSet)

	summary := model.ClusterSecuritySummary{}

	// 1. Aggregate VulnerabilityReports
	var vulnCRD apiextensionsv1.CustomResourceDefinition
	if err := cs.K8sClient.Get(c.Request.Context(), client.ObjectKey{Name: "vulnerabilityreports.aquasecurity.github.io"}, &vulnCRD); err == nil {
		var vulnList unstructured.UnstructuredList
		vulnList.SetGroupVersionKind(vulnerabilityReportKind)

		if err := cs.K8sClient.List(c.Request.Context(), &vulnList); err == nil {
			summary.ScannedImages = len(vulnList.Items)

			for _, u := range vulnList.Items {
				var report model.VulnerabilityReport
				if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &report); err != nil {
					continue
				}

				s := report.Report.Summary
				summary.TotalVulnerabilities.CriticalCount += s.CriticalCount
				summary.TotalVulnerabilities.HighCount += s.HighCount
				summary.TotalVulnerabilities.MediumCount += s.MediumCount
				summary.TotalVulnerabilities.LowCount += s.LowCount
				summary.TotalVulnerabilities.UnknownCount += s.UnknownCount

				if s.CriticalCount > 0 || s.HighCount > 0 || s.MediumCount > 0 || s.LowCount > 0 {
					summary.VulnerableImages++
				}
			}
		}
	}

	// 2. Aggregate ConfigAuditReports
	var configCRD apiextensionsv1.CustomResourceDefinition
	if err := cs.K8sClient.Get(c.Request.Context(), client.ObjectKey{Name: "configauditreports.aquasecurity.github.io"}, &configCRD); err == nil {
		var configList unstructured.UnstructuredList
		configList.SetGroupVersionKind(configAuditReportKind)

		if err := cs.K8sClient.List(c.Request.Context(), &configList); err == nil {
			for _, u := range configList.Items {
				var report model.ConfigAuditReport
				if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &report); err != nil {
					continue
				}

				s := report.Report.Summary
				summary.TotalConfigAuditIssues.CriticalCount += s.CriticalCount
				summary.TotalConfigAuditIssues.HighCount += s.HighCount
				summary.TotalConfigAuditIssues.MediumCount += s.MediumCount
				summary.TotalConfigAuditIssues.LowCount += s.LowCount
			}
		}
	}

	// 3. Aggregate ExposedSecretReports
	var secretCRD apiextensionsv1.CustomResourceDefinition
	if err := cs.K8sClient.Get(c.Request.Context(), client.ObjectKey{Name: "exposedsecretreports.aquasecurity.github.io"}, &secretCRD); err == nil {
		var secretList unstructured.UnstructuredList
		secretList.SetGroupVersionKind(exposedSecretReportKind)

		if err := cs.K8sClient.List(c.Request.Context(), &secretList); err == nil {
			for _, u := range secretList.Items {
				var report model.ExposedSecretReport
				if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &report); err != nil {
					continue
				}

				s := report.Report.Summary
				summary.TotalExposedSecrets.CriticalCount += s.CriticalCount
				summary.TotalExposedSecrets.HighCount += s.HighCount
				summary.TotalExposedSecrets.MediumCount += s.MediumCount
				summary.TotalExposedSecrets.LowCount += s.LowCount
			}
		}
	}

	c.JSON(http.StatusOK, summary)
}

// GetTopVulnerableWorkloads fetches workloads with most vulnerabilities
func (h *SecurityReportHandler) GetTopVulnerableWorkloads(c *gin.Context) {
	cs := c.MustGet("cluster").(*cluster.ClientSet)

	var vulnCRD apiextensionsv1.CustomResourceDefinition
	if err := cs.K8sClient.Get(c.Request.Context(), client.ObjectKey{Name: "vulnerabilityreports.aquasecurity.github.io"}, &vulnCRD); err != nil {
		c.JSON(http.StatusOK, model.WorkloadSummaryList{Items: []model.WorkloadSummary{}})
		return
	}

	var vulnList unstructured.UnstructuredList
	vulnList.SetGroupVersionKind(vulnerabilityReportKind)

	if err := cs.K8sClient.List(c.Request.Context(), &vulnList); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to list vulnerability reports: %v", err)})
		return
	}

	workloadMap := make(map[string]*model.WorkloadSummary)

	for _, u := range vulnList.Items {
		var report model.VulnerabilityReport
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &report); err != nil {
			continue
		}

		s := report.Report.Summary

		// Aggregate by workload
		lbls := u.GetLabels()
		kind := lbls["trivy-operator.resource.kind"]
		name := lbls["trivy-operator.resource.name"]
		namespace := u.GetNamespace()

		if kind != "" && name != "" {
			key := fmt.Sprintf("%s/%s/%s", namespace, kind, name)
			if _, exists := workloadMap[key]; !exists {
				workloadMap[key] = &model.WorkloadSummary{
					Namespace: namespace,
					Kind:      kind,
					Name:      name,
				}
			}
			w := workloadMap[key]
			w.Vulnerabilities.CriticalCount += s.CriticalCount
			w.Vulnerabilities.HighCount += s.HighCount
			w.Vulnerabilities.MediumCount += s.MediumCount
			w.Vulnerabilities.LowCount += s.LowCount
			w.Vulnerabilities.UnknownCount += s.UnknownCount
		}
	}

	// Convert map to slice and sort
	var workloads []model.WorkloadSummary
	for _, w := range workloadMap {
		workloads = append(workloads, *w)
	}

	sort.Slice(workloads, func(i, j int) bool {
		if workloads[i].Vulnerabilities.CriticalCount != workloads[j].Vulnerabilities.CriticalCount {
			return workloads[i].Vulnerabilities.CriticalCount > workloads[j].Vulnerabilities.CriticalCount
		}
		if workloads[i].Vulnerabilities.HighCount != workloads[j].Vulnerabilities.HighCount {
			return workloads[i].Vulnerabilities.HighCount > workloads[j].Vulnerabilities.HighCount
		}
		if workloads[i].Vulnerabilities.MediumCount != workloads[j].Vulnerabilities.MediumCount {
			return workloads[i].Vulnerabilities.MediumCount > workloads[j].Vulnerabilities.MediumCount
		}
		return workloads[i].Vulnerabilities.LowCount > workloads[j].Vulnerabilities.LowCount
	})

	limit := 10
	if len(workloads) < limit {
		limit = len(workloads)
	}

	c.JSON(http.StatusOK, model.WorkloadSummaryList{Items: workloads[:limit]})
}

// GetTopMisconfiguredWorkloads fetches workloads with most misconfigurations
//
//nolint:dupl // aggregation logic is similar
func (h *SecurityReportHandler) GetTopMisconfiguredWorkloads(c *gin.Context) {
	cs := c.MustGet("cluster").(*cluster.ClientSet)

	var configCRD apiextensionsv1.CustomResourceDefinition
	if err := cs.K8sClient.Get(c.Request.Context(), client.ObjectKey{Name: "configauditreports.aquasecurity.github.io"}, &configCRD); err != nil {
		c.JSON(http.StatusOK, model.WorkloadSummaryList{Items: []model.WorkloadSummary{}})
		return
	}

	var configList unstructured.UnstructuredList
	configList.SetGroupVersionKind(configAuditReportKind)

	if err := cs.K8sClient.List(c.Request.Context(), &configList); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to list config audit reports: %v", err)})
		return
	}

	misconfiguredMap := make(map[string]*model.WorkloadSummary)

	for _, u := range configList.Items {
		var report model.ConfigAuditReport
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &report); err != nil {
			continue
		}

		s := report.Report.Summary

		// Aggregate by workload
		lbls := u.GetLabels()
		kind := lbls["trivy-operator.resource.kind"]
		name := lbls["trivy-operator.resource.name"]
		namespace := u.GetNamespace()

		if kind != "" && name != "" {
			key := fmt.Sprintf("%s/%s/%s", namespace, kind, name)
			if _, exists := misconfiguredMap[key]; !exists {
				misconfiguredMap[key] = &model.WorkloadSummary{
					Namespace: namespace,
					Kind:      kind,
					Name:      name,
				}
			}
			w := misconfiguredMap[key]
			w.Vulnerabilities.CriticalCount += s.CriticalCount
			w.Vulnerabilities.HighCount += s.HighCount
			w.Vulnerabilities.MediumCount += s.MediumCount
			w.Vulnerabilities.LowCount += s.LowCount
		}
	}

	// Convert map to slice and sort
	var misconfigured []model.WorkloadSummary
	for _, w := range misconfiguredMap {
		// Only include workloads with issues
		if w.Vulnerabilities.CriticalCount > 0 || w.Vulnerabilities.HighCount > 0 ||
			w.Vulnerabilities.MediumCount > 0 || w.Vulnerabilities.LowCount > 0 {
			misconfigured = append(misconfigured, *w)
		}
	}

	sort.Slice(misconfigured, func(i, j int) bool {
		if misconfigured[i].Vulnerabilities.CriticalCount != misconfigured[j].Vulnerabilities.CriticalCount {
			return misconfigured[i].Vulnerabilities.CriticalCount > misconfigured[j].Vulnerabilities.CriticalCount
		}
		if misconfigured[i].Vulnerabilities.HighCount != misconfigured[j].Vulnerabilities.HighCount {
			return misconfigured[i].Vulnerabilities.HighCount > misconfigured[j].Vulnerabilities.HighCount
		}
		return misconfigured[i].Vulnerabilities.MediumCount > misconfigured[j].Vulnerabilities.MediumCount
	})

	limit := 10
	if len(misconfigured) < limit {
		limit = len(misconfigured)
	}

	c.JSON(http.StatusOK, model.WorkloadSummaryList{Items: misconfigured[:limit]})
}

// GetTopRbacRiskyWorkloads fetches workloads with most RBAC risks
//
//nolint:dupl // aggregation logic is similar
func (h *SecurityReportHandler) GetTopRbacRiskyWorkloads(c *gin.Context) {
	cs := c.MustGet("cluster").(*cluster.ClientSet)

	// Check if CRD exists
	var crd apiextensionsv1.CustomResourceDefinition
	if err := cs.K8sClient.Get(c.Request.Context(), client.ObjectKey{Name: "rbacassessmentreports.aquasecurity.github.io"}, &crd); err != nil {
		c.JSON(http.StatusOK, model.WorkloadSummaryList{Items: []model.WorkloadSummary{}})
		return
	}

	var list unstructured.UnstructuredList
	list.SetGroupVersionKind(rbacAssessmentReportKind)

	if err := cs.K8sClient.List(c.Request.Context(), &list); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to list rbac reports: %v", err)})
		return
	}

	rbacMap := make(map[string]*model.WorkloadSummary)

	for _, u := range list.Items {
		var report model.RbacAssessmentReport
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &report); err != nil {
			continue
		}

		s := report.Report.Summary

		// Aggregate by workload
		lbls := u.GetLabels()
		kind := lbls["trivy-operator.resource.kind"]
		name := lbls["trivy-operator.resource.name"]
		namespace := u.GetNamespace()

		if kind != "" && name != "" {
			key := fmt.Sprintf("%s/%s/%s", namespace, kind, name)
			if _, exists := rbacMap[key]; !exists {
				rbacMap[key] = &model.WorkloadSummary{
					Namespace: namespace,
					Kind:      kind,
					Name:      name,
				}
			}
			w := rbacMap[key]
			w.Vulnerabilities.CriticalCount += s.CriticalCount
			w.Vulnerabilities.HighCount += s.HighCount
			w.Vulnerabilities.MediumCount += s.MediumCount
			w.Vulnerabilities.LowCount += s.LowCount
		}
	}

	// Also check ClusterRbacAssessmentReports (for ClusterRoles etc)
	var clusterCrd apiextensionsv1.CustomResourceDefinition
	if err := cs.K8sClient.Get(c.Request.Context(), client.ObjectKey{Name: "clusterrbacassessmentreports.aquasecurity.github.io"}, &clusterCrd); err == nil {
		var clusterList unstructured.UnstructuredList
		clusterList.SetGroupVersionKind(clusterRbacAssessmentReportKind)
		if err := cs.K8sClient.List(c.Request.Context(), &clusterList); err == nil {
			for _, u := range clusterList.Items {
				var report model.ClusterRbacAssessmentReport
				if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &report); err != nil {
					continue
				}

				s := report.Report.Summary
				lbls := u.GetLabels()
				kind := lbls["trivy-operator.resource.kind"]
				name := lbls["trivy-operator.resource.name"]
				// Cluster scoped, no namespace

				if kind != "" && name != "" {
					key := fmt.Sprintf("cluster/%s/%s", kind, name)
					if _, exists := rbacMap[key]; !exists {
						rbacMap[key] = &model.WorkloadSummary{
							Namespace: "", // Cluster Scoped
							Kind:      kind,
							Name:      name,
						}
					}
					w := rbacMap[key]
					w.Vulnerabilities.CriticalCount += s.CriticalCount
					w.Vulnerabilities.HighCount += s.HighCount
					w.Vulnerabilities.MediumCount += s.MediumCount
					w.Vulnerabilities.LowCount += s.LowCount
				}
			}
		}
	}

	// Convert map to slice and sort
	var risky []model.WorkloadSummary
	for _, w := range rbacMap {
		// Only include workloads with issues
		if w.Vulnerabilities.CriticalCount > 0 || w.Vulnerabilities.HighCount > 0 ||
			w.Vulnerabilities.MediumCount > 0 || w.Vulnerabilities.LowCount > 0 {
			risky = append(risky, *w)
		}
	}

	sort.Slice(risky, func(i, j int) bool {
		if risky[i].Vulnerabilities.CriticalCount != risky[j].Vulnerabilities.CriticalCount {
			return risky[i].Vulnerabilities.CriticalCount > risky[j].Vulnerabilities.CriticalCount
		}
		if risky[i].Vulnerabilities.HighCount != risky[j].Vulnerabilities.HighCount {
			return risky[i].Vulnerabilities.HighCount > risky[j].Vulnerabilities.HighCount
		}
		return risky[i].Vulnerabilities.MediumCount > risky[j].Vulnerabilities.MediumCount
	})

	limit := 10
	if len(risky) < limit {
		limit = len(risky)
	}

	c.JSON(http.StatusOK, model.WorkloadSummaryList{Items: risky[:limit]})
}

// ListConfigAuditReports fetches config audit reports for a workload
//
//nolint:dupl // boilerplate code for different report types
func (h *SecurityReportHandler) ListConfigAuditReports(c *gin.Context) {
	namespace := c.Query("namespace")
	workloadKind := c.Query("workloadKind")
	workloadName := c.Query("workloadName")

	if namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace is required"})
		return
	}

	items, err := h.listReportsGeneric(c, configAuditReportKind, "configauditreports.aquasecurity.github.io", workloadKind, workloadName, namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	reports := make([]model.ConfigAuditReport, 0, len(items))
	for _, u := range items {
		var report model.ConfigAuditReport
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &report); err != nil {
			continue
		}
		reports = append(reports, report)
	}

	c.JSON(http.StatusOK, model.ConfigAuditReportList{Items: reports})
}

// ListInfraAssessmentReports fetches infra assessment reports for a workload
//
//nolint:dupl // boilerplate code for different report types
func (h *SecurityReportHandler) ListInfraAssessmentReports(c *gin.Context) {
	namespace := c.Query("namespace")
	workloadKind := c.Query("workloadKind")
	workloadName := c.Query("workloadName")

	if namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace is required"})
		return
	}

	items, err := h.listReportsGeneric(c, infraAssessmentReportKind, "infraassessmentreports.aquasecurity.github.io", workloadKind, workloadName, namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	reports := make([]model.InfraAssessmentReport, 0, len(items))
	for _, u := range items {
		var report model.InfraAssessmentReport
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &report); err != nil {
			continue
		}
		reports = append(reports, report)
	}

	c.JSON(http.StatusOK, model.InfraAssessmentReportList{Items: reports})
}

// ListClusterInfraAssessmentReports fetches cluster infra assessment reports (for nodes)
func (h *SecurityReportHandler) ListClusterInfraAssessmentReports(c *gin.Context) {
	cs := c.MustGet("cluster").(*cluster.ClientSet)
	// Typically filtered by a node name which might be passed as workloadName
	workloadName := c.Query("workloadName")

	// Check if CRD exists
	var crd apiextensionsv1.CustomResourceDefinition
	if err := cs.K8sClient.Get(c.Request.Context(), client.ObjectKey{Name: "clusterinfraassessmentreports.aquasecurity.github.io"}, &crd); err != nil {
		c.JSON(http.StatusOK, model.ClusterInfraAssessmentReportList{Items: []model.ClusterInfraAssessmentReport{}})
		return
	}

	var list unstructured.UnstructuredList
	list.SetGroupVersionKind(clusterInfraAssessmentReportKind)

	// Cluster scope, no namespace
	opts := []client.ListOption{}

	// Filter by labels if a node name is provided.
	// Typically Trivy Operator labels node reports with the node name.
	if workloadName != "" {
		// Assuming Trivy labels them with resource.name or similar.
		// For nodes, it's often labeled as 'trivy-operator.resource.name' or exists with the name 'node-<nodename>'.
		// Let's try label matching first as it's more robust than guessing specific name pattern.
		labels := client.MatchingLabels{
			"trivy-operator.resource.name": workloadName,
			// "trivy-operator.resource.kind": "Node", // Optional, likely redundant if looking in ClusterInfraAssessment
		}
		opts = append(opts, labels)
	}

	if err := cs.K8sClient.List(c.Request.Context(), &list, opts...); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to list cluster infra assessment reports: %v", err)})
		return
	}

	reports := make([]model.ClusterInfraAssessmentReport, 0, len(list.Items))
	for _, u := range list.Items {
		var report model.ClusterInfraAssessmentReport
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &report); err != nil {
			continue
		}
		reports = append(reports, report)
	}

	c.JSON(http.StatusOK, model.ClusterInfraAssessmentReportList{Items: reports})
}

// ListExposedSecretReports fetches exposed secret reports for a workload
//
//nolint:dupl // boilerplate code for different report types
func (h *SecurityReportHandler) ListExposedSecretReports(c *gin.Context) {
	namespace := c.Query("namespace")
	workloadKind := c.Query("workloadKind")
	workloadName := c.Query("workloadName")

	if namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace is required"})
		return
	}

	items, err := h.listReportsGeneric(c, exposedSecretReportKind, "exposedsecretreports.aquasecurity.github.io", workloadKind, workloadName, namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	reports := make([]model.ExposedSecretReport, 0, len(items))
	for _, u := range items {
		var report model.ExposedSecretReport
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &report); err != nil {
			continue
		}
		reports = append(reports, report)
	}

	c.JSON(http.StatusOK, model.ExposedSecretReportList{Items: reports})
}

// ListComplianceReports fetches ClusterComplianceReports (cluster-scoped)
func (h *SecurityReportHandler) ListComplianceReports(c *gin.Context) {
	cs := c.MustGet("cluster").(*cluster.ClientSet)

	// Check if CRD exists
	var crd apiextensionsv1.CustomResourceDefinition
	if err := cs.K8sClient.Get(c.Request.Context(), client.ObjectKey{Name: "clustercompliancereports.aquasecurity.github.io"}, &crd); err != nil {
		c.JSON(http.StatusOK, model.ClusterComplianceReportList{Items: []model.ClusterComplianceReport{}})
		return
	}

	var list unstructured.UnstructuredList
	list.SetGroupVersionKind(clusterComplianceReportKind)

	// ClusterComplianceReport is cluster-scoped, no namespace
	if err := cs.K8sClient.List(c.Request.Context(), &list); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to list compliance reports: %v", err)})
		return
	}

	reports := make([]model.ClusterComplianceReport, 0, len(list.Items))
	for _, u := range list.Items {
		var report model.ClusterComplianceReport
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &report); err != nil {
			continue
		}
		reports = append(reports, report)
	}

	c.JSON(http.StatusOK, model.ClusterComplianceReportList{Items: reports})
}

// ListRbacAssessmentReports fetches RBAC assessment reports for a workload
//
//nolint:dupl // boilerplate code for different report types
func (h *SecurityReportHandler) ListRbacAssessmentReports(c *gin.Context) {
	namespace := c.Query("namespace")
	workloadKind := c.Query("workloadKind")
	workloadName := c.Query("workloadName")

	if namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace is required"})
		return
	}

	items, err := h.listReportsGeneric(c, rbacAssessmentReportKind, "rbacassessmentreports.aquasecurity.github.io", workloadKind, workloadName, namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	reports := make([]model.RbacAssessmentReport, 0, len(items))
	for _, u := range items {
		var report model.RbacAssessmentReport
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &report); err != nil {
			continue
		}
		reports = append(reports, report)
	}

	c.JSON(http.StatusOK, model.RbacAssessmentReportList{Items: reports})
}

// ListClusterRbacAssessmentReports fetches cluster RBAC assessment reports (for cluster roles/bindings)
func (h *SecurityReportHandler) ListClusterRbacAssessmentReports(c *gin.Context) {
	cs := c.MustGet("cluster").(*cluster.ClientSet)
	workloadName := c.Query("workloadName")
	workloadKind := c.Query("workloadKind")

	// Check if CRD exists
	var crd apiextensionsv1.CustomResourceDefinition
	if err := cs.K8sClient.Get(c.Request.Context(), client.ObjectKey{Name: "clusterrbacassessmentreports.aquasecurity.github.io"}, &crd); err != nil {
		c.JSON(http.StatusOK, model.ClusterRbacAssessmentReportList{Items: []model.ClusterRbacAssessmentReport{}})
		return
	}

	var list unstructured.UnstructuredList
	list.SetGroupVersionKind(clusterRbacAssessmentReportKind)

	// Cluster scope
	opts := []client.ListOption{}

	if workloadKind != "" && workloadName != "" {
		labels := client.MatchingLabels{
			"trivy-operator.resource.kind": workloadKind,
			"trivy-operator.resource.name": workloadName,
		}
		opts = append(opts, labels)
	}

	if err := cs.K8sClient.List(c.Request.Context(), &list, opts...); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to list cluster rbac assessment reports: %v", err)})
		return
	}

	reports := make([]model.ClusterRbacAssessmentReport, 0, len(list.Items))
	for _, u := range list.Items {
		var report model.ClusterRbacAssessmentReport
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &report); err != nil {
			continue
		}
		reports = append(reports, report)
	}

	c.JSON(http.StatusOK, model.ClusterRbacAssessmentReportList{Items: reports})
}

func (h *SecurityReportHandler) RegisterRoutes(group *gin.RouterGroup) {
	securityParams := group.Group("/security")
	securityParams.GET("/status", h.CheckStatus)
	securityParams.GET("/reports", h.ListReports)
	securityParams.GET("/config-audit/reports", h.ListConfigAuditReports)
	securityParams.GET("/infra-assessment/reports", h.ListInfraAssessmentReports)
	securityParams.GET("/cluster-infra-assessment/reports", h.ListClusterInfraAssessmentReports)
	securityParams.GET("/rbac-assessment/reports", h.ListRbacAssessmentReports)
	securityParams.GET("/cluster-rbac-assessment/reports", h.ListClusterRbacAssessmentReports)
	securityParams.GET("/secrets/reports", h.ListExposedSecretReports)
	securityParams.GET("/compliance/reports", h.ListComplianceReports)
	securityParams.GET("/summary", h.GetClusterSummary)
	securityParams.GET("/reports/top-vulnerable", h.GetTopVulnerableWorkloads)
	securityParams.GET("/reports/top-misconfigured", h.GetTopMisconfiguredWorkloads)
	securityParams.GET("/reports/top-rbac-risky", h.GetTopRbacRiskyWorkloads)
}
