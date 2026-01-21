export function HelmChartListPage() {
  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold">Helm Charts</h1>
        </div>
      </div>

      <div className="p-8 text-center text-muted-foreground border rounded-lg bg-muted/20">
        <h3 className="text-lg font-medium mb-2">Coming Soon</h3>
        <p>
          Helm chart repository management and installation will be available in
          a future update.
        </p>
      </div>
    </div>
  )
}
