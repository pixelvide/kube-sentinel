import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

interface APIKeyDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSubmit: (data: { name: string; expiresAt?: string }) => void
  isLoading?: boolean
}

export function APIKeyDialog({
  open,
  onOpenChange,
  onSubmit,
  isLoading,
}: APIKeyDialogProps) {
  const { t } = useTranslation()
  const [name, setName] = useState('')
  const [expiresAt, setExpiresAt] = useState('')
  const [error, setError] = useState('')

  useEffect(() => {
    if (open) {
      const defaultDate = new Date()
      defaultDate.setDate(defaultDate.getDate() + 14)
      setExpiresAt(defaultDate.toISOString().split('T')[0])
    } else {
      setName('')
      setExpiresAt('')
      setError('')
    }
  }, [open])

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()

    if (!name.trim()) {
      setError(t('common.required', 'This field is required'))
      return
    }

    onSubmit({ name: name.trim(), expiresAt: expiresAt || undefined })
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>
            {t('apikeyManagement.dialog.title', 'Create API Key')}
          </DialogTitle>
          <DialogDescription>
            {t(
              'apikeyManagement.dialog.description',
              'Create a new API key for programmatic access.'
            )}
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit}>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="name">
                {t('apikeyManagement.dialog.name', 'Name')}
              </Label>
              <Input
                id="name"
                placeholder={t(
                  'apikeyManagement.dialog.namePlaceholder',
                  'e.g., CI API Key'
                )}
                value={name}
                onChange={(e) => {
                  setName(e.target.value)
                  setError('')
                }}
              />
              {error && <p className="text-sm text-destructive">{error}</p>}
            </div>

            <div className="space-y-2">
              <Label htmlFor="expiresAt">
                {t(
                  'apikeyManagement.dialog.expiresAt',
                  'Expiry Date (Optional)'
                )}
              </Label>
              <Input
                id="expiresAt"
                type="date"
                value={expiresAt}
                min={new Date().toISOString().split('T')[0]}
                max={
                  new Date(new Date().setDate(new Date().getDate() + 365))
                    .toISOString()
                    .split('T')[0]
                }
                onChange={(e) => setExpiresAt(e.target.value)}
              />
              <p className="text-xs text-muted-foreground">
                {t(
                  'apikeyManagement.dialog.expiryHint',
                  'Maximum 365 days from creation'
                )}
              </p>
            </div>
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
            >
              {t('common.cancel', 'Cancel')}
            </Button>
            <Button type="submit" disabled={isLoading}>
              {isLoading
                ? t('common.creating', 'Creating...')
                : t('common.create', 'Create')}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
