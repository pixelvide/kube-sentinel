"use client";

import { useState } from "react";
import { Badge } from "@/components/ui/badge";
import { formatAge } from "@/lib/utils";
import { ChevronDown, ChevronRight, Tags, StickyNote } from "lucide-react";
import { cn } from "@/lib/utils";

interface KubePropertiesProps {
    resource: any;
}

export function KubeProperties({ resource }: KubePropertiesProps) {
    const [showLabels, setShowLabels] = useState(false);
    const [showAnnotations, setShowAnnotations] = useState(false);

    if (!resource || !resource.metadata) return null;

    const { metadata } = resource;
    const {
        name,
        namespace,
        creationTimestamp,
        labels,
        annotations
    } = metadata;

    const getResourceStatus = (res: any) => {
        if (!res.status) return null;

        const { status, kind } = res;

        // Specific handling for common kinds
        if (kind === "Pod") return status.phase;
        if (kind === "Node") {
            const readyCondition = status.conditions?.find((c: any) => c.type === "Ready");
            return readyCondition?.status === "True" ? "Ready" : "NotReady";
        }
        if (kind === "Job") {
            if (status.succeeded > 0) return "Succeeded";
            if (status.failed > 0) return "Failed";
            if (status.active > 0) return "Active";
        }
        if (kind === "CronJob") {
            return status.lastScheduleTime ? "Scheduled" : "Never Scheduled";
        }

        // Generic fallback for workloads with replicas
        if (status.replicas !== undefined && status.readyReplicas !== undefined) {
            return `${status.readyReplicas}/${status.replicas} Ready`;
        }

        // Check conditions for a generic "Ready" state
        const readyCond = status.conditions?.find((c: any) => c.type === "Ready" || c.type === "Available");
        if (readyCond) {
            return readyCond.status === "True" ? "Ready" : "NotReady";
        }

        // Phase is a common field for many resources
        if (status.phase) return status.phase;

        return null;
    };

    const statusValue = getResourceStatus(resource);

    const properties = [
        { label: "Name", value: name },
        { label: "Namespace", value: namespace },
        {
            label: "Created",
            value: creationTimestamp ? `${new Date(creationTimestamp).toLocaleString()} (${formatAge(creationTimestamp)} ago)` : null
        },
        { label: "Status", value: statusValue },
    ];

    const labelCount = labels ? Object.keys(labels).length : 0;
    const annotationCount = annotations ? Object.keys(annotations).length : 0;

    return (
        <div className="space-y-4">
            <h3 className="text-sm font-bold text-zinc-900 uppercase tracking-wider flex items-center gap-2">
                Properties
            </h3>
            <div className="rounded-xl border border-zinc-200 bg-white p-4 space-y-4 shadow-sm font-sans">
                <div className="space-y-3">
                    {properties.map((prop) => (
                        prop.value && (
                            <div key={prop.label} className="grid grid-cols-3 gap-2 items-start">
                                <span className="text-zinc-500 text-xs font-semibold uppercase tracking-tight">{prop.label}</span>
                                {prop.label === "Status" ? (
                                    <div className="col-span-2">
                                        <Badge
                                            variant="outline"
                                            className={cn(
                                                "text-[10px] font-bold px-2 py-0 h-auto",
                                                prop.value === "Running" || prop.value === "Ready" || prop.value === "Succeeded"
                                                    ? "bg-emerald-50 text-emerald-700 border-emerald-200"
                                                    : prop.value === "Pending" || prop.value === "Active"
                                                        ? "bg-amber-50 text-amber-700 border-amber-200"
                                                        : prop.value === "Failed" || prop.value === "NotReady"
                                                            ? "bg-red-50 text-red-700 border-red-200"
                                                            : "bg-zinc-50 text-zinc-700 border-zinc-200"
                                            )}
                                        >
                                            {prop.value}
                                        </Badge>
                                    </div>
                                ) : (
                                    <span className="text-zinc-950 col-span-2 break-all font-bold text-xs">{prop.value}</span>
                                )}
                            </div>
                        )
                    ))}
                </div>

                {labelCount > 0 && (
                    <div className="pt-3 border-t border-zinc-100">
                        <button
                            onClick={() => setShowLabels(!showLabels)}
                            className="flex items-center justify-between w-full text-left group"
                        >
                            <div className="flex items-center gap-2">
                                <Tags className="h-3.5 w-3.5 text-zinc-400" />
                                <span className="text-zinc-500 text-xs font-semibold uppercase tracking-tight">Labels</span>
                                <Badge variant="secondary" className="bg-zinc-100 text-zinc-600 border-none text-[10px] font-bold px-1.5 h-4">
                                    {labelCount}
                                </Badge>
                            </div>
                            {showLabels ? <ChevronDown className="h-3.5 w-3.5 text-zinc-400" /> : <ChevronRight className="h-3.5 w-3.5 text-zinc-400" />}
                        </button>

                        {showLabels && (
                            <div className="mt-3 flex flex-wrap gap-1.5 animate-in fade-in slide-in-from-top-1 duration-200">
                                {Object.entries(labels).map(([k, v]) => (
                                    <Badge key={k} variant="outline" className="text-[10px] font-bold border-zinc-200 bg-zinc-50 text-zinc-700 py-0.5 px-2 h-auto whitespace-normal break-all">
                                        {k}: {String(v)}
                                    </Badge>
                                ))}
                            </div>
                        )}
                    </div>
                )}

                {annotationCount > 0 && (
                    <div className="pt-3 border-t border-zinc-100">
                        <button
                            onClick={() => setShowAnnotations(!showAnnotations)}
                            className="flex items-center justify-between w-full text-left group"
                        >
                            <div className="flex items-center gap-2">
                                <StickyNote className="h-3.5 w-3.5 text-zinc-400" />
                                <span className="text-zinc-500 text-xs font-semibold uppercase tracking-tight">Annotations</span>
                                <Badge variant="secondary" className="bg-zinc-100 text-zinc-600 border-none text-[10px] font-bold px-1.5 h-4">
                                    {annotationCount}
                                </Badge>
                            </div>
                            {showAnnotations ? <ChevronDown className="h-3.5 w-3.5 text-zinc-400" /> : <ChevronRight className="h-3.5 w-3.5 text-zinc-400" />}
                        </button>

                        {showAnnotations && (
                            <div className="mt-3 flex flex-wrap gap-1.5 animate-in fade-in slide-in-from-top-1 duration-200">
                                {Object.entries(annotations).map(([k, v]) => (
                                    <Badge key={k} variant="outline" className="text-[10px] font-bold border-zinc-200 bg-zinc-50 text-zinc-700 py-0.5 px-2 max-w-full h-auto whitespace-normal break-all">
                                        {k}: {String(v)}
                                    </Badge>
                                ))}
                            </div>
                        )}
                    </div>
                )}
            </div>
        </div>
    );
}
