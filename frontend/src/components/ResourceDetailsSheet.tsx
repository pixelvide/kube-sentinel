"use client";

import { useEffect, useState } from "react";
import {
    Sheet,
    SheetContent,
    SheetHeader,
    SheetTitle,
    SheetDescription,
} from "@/components/ui/sheet";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Badge } from "@/components/ui/badge";
import { KubeProperties } from "@/components/KubeProperties";

interface ResourceDetailsSheetProps {
    isOpen: boolean;
    onClose: () => void;
    context: string;
    namespace: string;
    name: string;
    kind: string;
}

interface EventSimple {
    type: string;
    reason: string;
    message: string;
    count: number;
    last_seen: string;
    age: string;
}

interface ResourceDetails {
    manifest: string;
    events: EventSimple[];
    raw: any;
}

export function ResourceDetailsSheet({
    isOpen,
    onClose,
    context,
    namespace,
    name,
    kind,
}: ResourceDetailsSheetProps) {
    const [details, setDetails] = useState<ResourceDetails | null>(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState("");

    const [scopes, setScopes] = useState<Record<string, string>>({});

    useEffect(() => {
        // Fetch scopes once
        fetch("/api/v1/kube/scopes")
            .then(res => res.json())
            .then(data => setScopes(data.scopes || {}))
            .catch(err => console.error("Failed to fetch scopes:", err));
    }, []);

    useEffect(() => {
        const kindLower = kind.toLowerCase();
        // Determine scope from fetched map, default to Namespaced if unknown (safe fallback)
        // Wait until scopes are loaded? Or just assume if not found, it might need namespace.
        // If scopes are empty, we might skip or error. Let's assume loaded.

        let shouldFetch = false;
        if (isOpen && context && name && kind) {
            // If scopes are not loaded yet, maybe wait?
            // But simpler: checking if scope knows about it.
            const scope = scopes[kindLower];
            if (scope) {
                const isClusterScoped = scope === "Cluster";
                if (isClusterScoped || namespace) {
                    shouldFetch = true;
                }
            } else {
                // Fallback if scopes not loaded or kind not found: try fetching if we have everything
                // Use defensive check: if kind matches known cluster scoped by default logic?
                // No, rely on API. If API fails, we might be stuck. 
                // Let's implement a simple optimize: if scopes empty, don't fetch yet?
                if (Object.keys(scopes).length > 0) {
                    // Scopes loaded but kind not found -> likely unsupported, but let's try if namespace is there
                    if (namespace) shouldFetch = true;
                }
                // If scopes not loaded, we can't decide perfectly. 
            }
        }

        if (Object.keys(scopes).length > 0 && isOpen && context && name && kind) {
            const scope = scopes[kindLower];
            const isClusterScoped = scope === "Cluster";
            // Valid if cluster scoped OR (namespaced and namespace provided)
            if (isClusterScoped || namespace) {
                shouldFetch = true;
            }
        }

        if (shouldFetch) {
            setLoading(true);
            setError("");
            setDetails(null);

            fetch(
                `/api/v1/kube/resource?context=${context}&namespace=${namespace}&name=${name}&kind=${kind}`
            )
                .then((res) => {
                    if (!res.ok) throw new Error("Failed to fetch details");
                    return res.json();
                })
                .then((data) => setDetails(data))
                .catch((err) => setError(err.message))
                .finally(() => setLoading(false));
        }
    }, [isOpen, context, namespace, name, kind, scopes]);

    return (
        <Sheet open={isOpen} onOpenChange={(open) => !open && onClose()}>
            <SheetContent className="w-full sm:max-w-2xl overflow-y-auto bg-zinc-50 border-l border-zinc-200 p-0 flex flex-col h-full">
                <SheetHeader className="p-6 border-b border-zinc-200 shrink-0 bg-white">
                    <SheetTitle className="text-xl font-bold font-mono text-zinc-900">
                        {kind}: {name}
                    </SheetTitle>
                    <SheetDescription className="text-zinc-500 font-mono text-xs">
                        {namespace} @ {context}
                    </SheetDescription>
                </SheetHeader>

                <div className="flex-1 overflow-y-auto">
                    {loading && (
                        <div className="p-6 text-zinc-500 font-mono text-sm animate-pulse">
                            Loading details...
                        </div>
                    )}

                    {error && (
                        <div className="p-6 text-red-400 font-mono text-sm">
                            Error: {error}
                        </div>
                    )}

                    {details && (
                        <div className="flex flex-col gap-6 p-6">
                            {/* Properties Section */}
                            <KubeProperties resource={details.raw} />

                            {/* Events Section */}
                            <div className="space-y-3">
                                <h3 className="text-sm font-semibold text-zinc-500 uppercase tracking-wider">
                                    Recent Events
                                </h3>
                                {details.events?.length > 0 ? (
                                    <div className="rounded-md border border-zinc-200 bg-white overflow-hidden shadow-sm">
                                        <table className="w-full text-xs text-left font-mono">
                                            <thead className="bg-zinc-50 text-zinc-500 border-b border-zinc-200">
                                                <tr>
                                                    <th className="p-2 font-medium">Type</th>
                                                    <th className="p-2 font-medium">Reason</th>
                                                    <th className="p-2 font-medium">Age</th>
                                                    <th className="p-2 font-medium">Message</th>
                                                </tr>
                                            </thead>
                                            <tbody className="divide-y divide-white/5">
                                                {details.events.map((e, i) => (
                                                    <tr key={i} className="hover:bg-white/5 transition-colors">
                                                        <td className="p-2">
                                                            <Badge
                                                                variant="outline"
                                                                className={
                                                                    e.type === "Warning"
                                                                        ? "text-red-600 border-red-200 bg-red-50"
                                                                        : "text-zinc-600 border-zinc-200 bg-zinc-50"
                                                                }
                                                            >
                                                                {e.type}
                                                            </Badge>
                                                        </td>
                                                        <td className="p-2 text-zinc-600">{e.reason}</td>
                                                        <td className="p-2 text-zinc-400 whitespace-nowrap text-[10px]">
                                                            {/* TODO: Format age better if needed, backend sends RFC3339 */}
                                                            {new Date(e.last_seen).toLocaleTimeString()}
                                                        </td>
                                                        <td className="p-2 text-zinc-500 break-words max-w-[200px]">
                                                            {e.message} ({e.count})
                                                        </td>
                                                    </tr>
                                                ))}
                                            </tbody>
                                        </table>
                                    </div>
                                ) : (
                                    <div className="text-zinc-400 text-sm italic py-4">
                                        No events found.
                                    </div>
                                )}
                            </div>

                            {/* Manifest Section */}
                            <div className="space-y-3">
                                <h3 className="text-sm font-semibold text-zinc-500 uppercase tracking-wider">
                                    YAML Manifest
                                </h3>
                                <div className="relative rounded-md border border-zinc-200 bg-white p-4 text-xs font-mono text-zinc-800 overflow-auto max-h-[600px] shadow-sm">
                                    <pre className="whitespace-pre-wrap break-all">{details.manifest}</pre>
                                </div>
                            </div>
                        </div>
                    )}
                </div>
            </SheetContent>
        </Sheet>
    );
}
