import { useVersionInfo } from '@/lib/api'

export function VersionInfo() {
  const { data: versionInfo } = useVersionInfo()

  if (!versionInfo) return null

  const handleCommitClick = (e: React.MouseEvent) => {
    e.preventDefault()
    e.stopPropagation()
    // GitHub repository URL - you can modify this to match your repository
    const repoUrl = 'https://github.com/pixelvide/kube-sentinel'
    const commitUrl = `${repoUrl}/commit/${versionInfo.commitId}`
    window.open(commitUrl, '_blank')
  }

  // Safely access properties
  const version = versionInfo.version || ''
  const commitId = versionInfo.commitId || ''

  return (
    <div className="text-[10px] text-muted-foreground/60 font-mono leading-none">
      v{version.replace(/^v/, '')} •{' '}
      <button
        onClick={handleCommitClick}
        className="hover:text-primary/80 hover:underline transition-colors cursor-pointer"
        title={`View commit ${commitId} on GitHub`}
      >
        {commitId.slice(0, 7)}
      </button>
    </div>
  )
}
