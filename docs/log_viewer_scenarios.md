# Log Viewer Behavior Scenarios

This document outlines the behavior of the Cloud Sentinel Log Viewer across different container configurations and user interactions.

## 1. Single-Container Pods

**Scenario:** A user opens logs for a pod that contains only one container.

*   **Default View:**
    *   Logs are streamed immediately.
    *   **Prefixes:** Hidden by default. The user sees raw log output.
    *   **Timestamps:** OFF by default. User can click the clock icon to enable.
*   **Container Selector:**
    *   Shows: "All Containers" (Active) and the single container name.
    *   Since there is only one container, switching selection effectively does nothing to the stream content, but is available for consistency.
*   **Pod Selector:**
    *   **Visibility:** Hidden by default when viewing a single pod (e.g., from Pods page).
    *   Can be explicitly shown via `showPodSelector={true}` prop.
*   **Prefix Toggle:**
    *   **Visible:** Yes.
    *   **State:** Defaults to `OFF` for single pod views, `ON` for multi-pod views.
    *   **Interaction:** User can click the tag icon to manually show/hide prefixes in any view. Preference is respected and passed to the backend.

## 2. Multi-Container Pods

**Scenario:** A user opens logs for a pod with multiple containers (e.g., `app-container` and `sidecar-proxy`).

*   **Default View:**
    *   Logs are streamed from **ALL** containers concurrently.
    *   **Prefixes:** Visible by default (e.g., `[app-container] Log message...`).
*   **Container Selector:**
    *   **Default:** "All Containers" is selected.
    *   **Dropdown:** Lists "All Containers" (Active) and individual checkboxes for each container.
*   **Interaction Scenarios:**
    *   **Click a specific container (Focus Mode):**
*   **Default View**: "All Containers" selected. Logs from all containers stream in parallel.
*   **Prefixes**: **On** by default.
    *   **Behavior**: Prefix state is set initially. Changing selection **does not** automatically toggle prefixes (User preference persists).
    *   **Format**: `[container-name] log line...`
*   **Selection Logic**:
    *   **Mixed Selection**: User can select any subset of containers.
    *   **Focus Mode**: Clicking a specific container when "All" is active instantly selects ONLY that container.
    *   **Deselection**: User can deselect all containers (showing "Select at least one container" empty state).
    *   **Auto-Revert**: Selecting all individual containers automatically reverts to the "All Containers" sentinel state.

## 3. Multi-Pod Views (Deployments, DaemonSets, StatefulSets, Jobs)

**Scenario:** A user views logs from multiple pods for a workload resource.

*   **Default View:**
    *   **Pod Selector:** Visible and defaults to `__all__` (all pods selected).
    *   **Container Selector:** Shows "All Containers" by default.
    *   **Prefixes:** Enabled by default to distinguish between pods/containers.
*   **Pod Selection:**
    *   **Default:** `__all__` is selected (streaming logs from all pods matching the selector).
    *   **Dropdown:** MultiSelect component showing all available pods.
    *   **Behavior:** User can select specific pods or multiple pods. Prefix is automatically enabled when multiple pods are selected.

## 4. Init Containers
*   **Visibility**: Listed in a separate "INIT CONTAINERS" section in the dropdown.
*   **Behavior**: Fully selectable, included in "All Containers" stream.
*   **Styling**: Distinct headers in dropdown to separate from standard containers.

## 5. Technical Implementation Details
*   **Concurrency**: Backend uses goroutines to stream multiple containers.
*   **Stability**: WebSocket heartbeat (Ping every 15s) prevents connection timeouts.
*   **Streaming**: Unified logic (`streamContainers`) handles both single and multi-container requests efficiently.
*   **Validation:**
    *   **Pod Selection:** WebSocket connection prevented if no pods are selected. Shows empty state overlay.
    *   **Container Selection:** WebSocket connection prevented if no containers are selected. Shows empty state overlay.
    *   **Priority:** Pod selection validation takes precedence over container validation.
*   **API Changes:**
    *   **Removed:** `pod` prop (replaced with `pods` array).
    *   **Added:** `showPodSelector?: boolean` prop to control pod selector visibility.
    *   **Default:** Pod selector shows when `pods` array has items, unless explicitly disabled.
*   **UI**:
    *   **Dropdown**: Rounded-lg container, rounded-sm checkboxes.
    *   **Empty States**:
        *   No pods selected: Terminal icon with "Select at least one pod"
        *   No containers selected: Tag icon with "Select at least one container"
    *   **Pod Selector**: 300px wide MultiSelect component with search disabled.
