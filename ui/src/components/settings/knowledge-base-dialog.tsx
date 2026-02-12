import { useState } from 'react'
import { IconBrain, IconPlus, IconTrash } from '@tabler/icons-react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import {
  addClusterKnowledge,
  ClusterKnowledgeBase,
  deleteClusterKnowledge,
  getClusterKnowledge,
} from '@/lib/api'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Textarea } from '@/components/ui/textarea'

interface KnowledgeBaseDialogProps {
  clusterId: number
  clusterName: string
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function KnowledgeBaseDialog({
  clusterId,
  clusterName,
  open,
  onOpenChange,
}: KnowledgeBaseDialogProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [newContent, setNewContent] = useState('')

  // Fetch Knowledge
  const { data: knowledgeList, isLoading } = useQuery({
    queryKey: ['cluster-knowledge', clusterId],
    queryFn: () => getClusterKnowledge(clusterId),
    enabled: open,
    select: (data) => data.data,
  })

  // Add Mutation
  const addMutation = useMutation({
    mutationFn: (content: string) => addClusterKnowledge(clusterId, content),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['cluster-knowledge', clusterId],
      })
      toast.success(t('knowledgeBase.added', 'Knowledge added'))
      setNewContent('')
    },
    onError: () => {
      toast.error(t('knowledgeBase.addError', 'Failed to add knowledge'))
    },
  })

  // Delete Mutation
  const deleteMutation = useMutation({
    mutationFn: (id: number) => deleteClusterKnowledge(clusterId, id),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['cluster-knowledge', clusterId],
      })
      toast.success(t('knowledgeBase.deleted', 'Knowledge deleted'))
    },
    onError: () => {
      toast.error(t('knowledgeBase.deleteError', 'Failed to delete knowledge'))
    },
  })

  const handleAdd = () => {
    if (!newContent.trim()) return
    addMutation.mutate(newContent)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[85vh] flex flex-col">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <IconBrain className="h-5 w-5 text-purple-500" />
            {t('knowledgeBase.title', 'AI Knowledge Base')} - {clusterName}
          </DialogTitle>
          <DialogDescription>
            {t(
              'knowledgeBase.description',
              'Manage persistent context and rules for the AI Assistant on this cluster.'
            )}
          </DialogDescription>
        </DialogHeader>

        <div className="flex-1 overflow-hidden flex flex-col gap-4">
          {/* Add New Section */}
          <div className="flex flex-col gap-2 p-4 border rounded-md bg-muted/30">
            <Label>
              {t('knowledgeBase.addNew', 'Add New Knowledge / Rule')}
            </Label>
            <div className="flex gap-2">
              <Textarea
                value={newContent}
                onChange={(e) => setNewContent(e.target.value)}
                placeholder={t(
                  'knowledgeBase.placeholder',
                  'e.g. "All prod deployments must have 2 replicas" or "Service foo maps to external-db"'
                )}
                className="resize-none"
                rows={2}
              />
              <Button
                onClick={handleAdd}
                disabled={!newContent.trim() || addMutation.isPending}
                className="self-end"
              >
                <IconPlus className="h-4 w-4 mr-1" />
                {t('common.add', 'Add')}
              </Button>
            </div>
          </div>

          {/* List Section */}
          <div className="flex-1 overflow-hidden">
            <h3 className="font-medium mb-2 text-sm text-muted-foreground">
              {t('knowledgeBase.existing', 'Existing Knowledge')}
            </h3>

            {isLoading ? (
              <div className="text-center py-8 text-muted-foreground">
                {t('common.loading', 'Loading...')}
              </div>
            ) : !knowledgeList || knowledgeList.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground border rounded-md border-dashed">
                {t('knowledgeBase.empty', 'No knowledge rules defined yet.')}
              </div>
            ) : (
              <ScrollArea className="h-[300px] border rounded-md">
                <div className="divide-y">
                  {knowledgeList.map((item: ClusterKnowledgeBase) => (
                    <div
                      key={item.id}
                      className="p-3 flex items-start justify-between gap-4 hover:bg-muted/50"
                    >
                      <div className="space-y-1">
                        <p className="text-sm whitespace-pre-wrap">
                          {item.content}
                        </p>
                        <div className="flex items-center gap-2 text-xs text-muted-foreground">
                          <span>{item.added_by}</span>
                          <span>•</span>
                          <span>
                            {new Date(item.created_at).toLocaleDateString()}
                          </span>
                        </div>
                      </div>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8 text-muted-foreground hover:text-destructive"
                        onClick={() => deleteMutation.mutate(item.id)}
                        disabled={deleteMutation.isPending}
                      >
                        <IconTrash className="h-4 w-4" />
                      </Button>
                    </div>
                  ))}
                </div>
              </ScrollArea>
            )}
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
