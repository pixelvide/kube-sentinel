# LogViewerModal Component

The `LogViewerModal` component provides a full-featured terminal interface for viewing Kubernetes pod logs with support for multi-pod and multi-container streaming.

## Features

- **Real-time log streaming** via WebSocket
- **Multi-pod support** with pod selector dropdown
- **Multi-container support** with container checkboxes
- **Init container support** with separate section
- **Configurable display options:**
  - Timestamps toggle
  - Container/Pod prefix toggle
  - Line wrapping toggle
- **Connection status indicators**
- **Terminal-based UI** using xterm.js

## Props

| Prop | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `isOpen` | `boolean` | - | ✅ | Controls modal visibility |
| `onClose` | `() => void` | - | ✅ | Callback when modal is closed |
| `context` | `string` | - | ✅ | Kubernetes context name |
| `namespace` | `string` | - | ✅ | Kubernetes namespace |
| `containers` | `string[]` | - | ✅ | List of container names |
| `initContainers` | `string[]` | `[]` | ❌ | List of init container names |
| `selector` | `string` | - | ❌ | Label selector for fetching pods |
| `pods` | `Array<{ name: string, status: string }>` | `[]` | ❌ | List of pods to stream from |
| `showPodSelector` | `boolean` | `pods.length > 0` | ❌ | Whether to show pod selector dropdown |
| `title` | `string` | - | ❌ | Resource name to display in modal title |

## Usage Examples

### Single Pod View (from Pods page)

```tsx
<LogViewerModal
    isOpen={!!logPod}
    onClose={() => setLogPod(null)}
    context={selectedContext}
    namespace={logPod.namespace}
    containers={logPod.containers}
    initContainers={logPod.init_containers || []}
    pods={[{ name: logPod.name, status: logPod.status }]}
    showPodSelector={false}  // Hide pod selector for single pod
    title={logPod.name}  // Pod name as title
/>
```

### Multi-Pod View (from Deployments/DaemonSets/etc)

```tsx
<LogViewerModal
    isOpen={!!logResource}
    onClose={() => setLogResource(null)}
    context={selectedContext}
    namespace={logResource.namespace}
    selector={logResource.selector}
    containers={["__all__"]}
    initContainers={[]}
    pods={logResource.pods}
    showPodSelector={true}  // Show pod selector
    title={logResource.name}  // Resource name (deployment/daemonset/etc)
/>
```

## Behavior

### Pod Selection
- **Default:** 
  - When `showPodSelector={false}` and single pod provided: Uses that specific pod name
  - When `showPodSelector={true}` or multiple pods: Defaults to `["__all__"]`
- **Smart Initialization:** Automatically detects single pod context to prevent sending `__all__` when viewing specific pod logs
- **MultiSelect Integration:** Uses the MultiSelect component with `allOption={{ label: "All Pods", value: "__all__" }}`
- **Auto-prefix:** Automatically enables prefix when multiple pods or `__all__` is selected

### Container Selection
- **Default:** Defaults to `["__all__"]` (all containers selected)
- **Focus Mode:** Clicking a specific container when "All" is active switches to that container only
- **Auto-revert:** Selecting all individual containers reverts back to "__all__"
- **Init Containers:** Listed in separate section with "INIT CONTAINERS" header

### Prefix Display
- **Single container:** Defaults to OFF
- **Multiple containers:** Defaults to ON
- **Format:** `[container-name] log line` or `[pod-name][container-name] log line`
- **User Control:** Can be toggled manually via tag icon button

### WebSocket Connection
- **Protocol:** `ws://` or `wss://` based on page protocol
- **Endpoint:** `/api/v1/kube/logs`
- **Parameters:** 
  - `context`, `namespace`
  - `pod` (comma-separated or `__all__`)
  - `container` (comma-separated or `__all__`)
  - `timestamps` (boolean)
  - `prefix` (boolean)
  - `selector` (optional label selector)
- **Heartbeat:** Implemented on backend to prevent connection timeouts

### UI Controls

| Button | Icon | Function |
|--------|------|----------|
| Pod Selector | MultiSelect Dropdown | Select which pods to stream from |
| Container Selector | Dropdown Checklist | Select which containers to stream from |
| Prefix Toggle | Tag Icon | Show/hide pod/container prefixes |
| Timestamps Toggle | Clock Icon | Show/hide log timestamps |
| Wrap Toggle | WrapText Icon | Enable/disable line wrapping |

## Terminal Features

- **Library:** xterm.js with FitAddon
- **Font:** Menlo, Monaco, "Courier New", monospace at 12px
- **Theme:** Dark background (#09090b) with light foreground (#f4f4f5)
- **Line Wrapping:** 
  - When enabled: Terminal fits to container width
  - When disabled: 1000 columns width for horizontal scrolling
- **Auto-resize:** Responds to container size changes via ResizeObserver

## Empty States

The component shows different empty state overlays based on user selection:

- **No Pod Selected:** Shows overlay with Terminal icon and message "Select at least one pod"
  - Displayed when `selectedPods.length === 0`
  - Takes priority over container empty state
- **No Container Selected:** Shows overlay with Tag icon and message "Select at least one container"
  - Displayed when `selectedContainers.length === 0` and at least one pod is selected
- **Connection States:**
  - **Connecting:** Yellow spinner
  - **Connected:** Green pulsing dot
  - **Disconnected:** Gray dot
  - **Error:** Red dot

## API Changes (v2.0)

### Removed
- ❌ `pod: string` prop - Replaced with `pods` array for consistency

### Added
- ✅ `showPodSelector?: boolean` - Explicit control over pod selector visibility
- ✅ Default pod selection to `["__all__"]`

### Migration Guide

**Before:**
```tsx
<LogViewerModal
    pod={podName}
    pods={[]}  // Usually empty
    ...
/>
```

**After:**
```tsx
<LogViewerModal
    pods={[{ name: podName, status: podStatus }]}
    showPodSelector={false}  // For single pod view
    ...
/>
```
