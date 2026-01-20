import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { Loader2, Save, FileUp } from 'lucide-react'
import { toast } from 'sonner'

import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { Card, CardContent, CardDescription, CardHeader, CardTitle, CardFooter } from '@/components/ui/card'
import { fetchUserAWSConfig, updateUserAWSConfig } from '@/lib/api'

export function AWSConfigManagement() {
    const { t } = useTranslation()
    const [content, setContent] = useState('')
    const [isLoading, setIsLoading] = useState(true)
    const [isSaving, setIsSaving] = useState(false)

    useEffect(() => {
        loadConfig()
    }, [])

    const loadConfig = async () => {
        setIsLoading(true)
        try {
            const config = await fetchUserAWSConfig()
            setContent(config.credentials_content || '')
        } catch (error) {
            console.error('Failed to load AWS config:', error)
        } finally {
            setIsLoading(false)
        }
    }

    const handleSave = async () => {
        setIsSaving(true)
        try {
            await updateUserAWSConfig({ credentials_content: content })
            toast.success(t('settings.aws.saved', 'AWS Credentials saved successfully'))
        } catch (_error) {
            toast.error(t('settings.aws.saveError', 'Failed to save AWS credentials'))
        } finally {
            setIsSaving(false)
        }
    }

    const handleFileUpload = (event: React.ChangeEvent<HTMLInputElement>) => {
        const file = event.target.files?.[0]
        if (!file) return

        const reader = new FileReader()
        reader.onload = (e) => {
            const text = e.target?.result
            if (typeof text === 'string') {
                setContent(text)
                toast.success(t('settings.aws.fileLoaded', 'File loaded successfully'))
            }
        }
        reader.readAsText(file)
    }

    if (isLoading) {
        return (
            <div className="flex items-center justify-center h-64">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
        )
    }

    return (
        <Card>
            <CardHeader>
                <CardTitle>{t('settings.aws.title', 'AWS Credentials')}</CardTitle>
                <CardDescription>
                    {t('settings.aws.description', 'Paste the content of your AWS credentials file (~/.aws/credentials) or upload it.')}
                </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
                <div className="flex items-center gap-4">
                    <Button variant="outline" onClick={() => document.getElementById('aws-file-upload')?.click()}>
                        <FileUp className="mr-2 h-4 w-4" />
                        {t('settings.aws.upload', 'Upload File')}
                    </Button>
                    <input
                        type="file"
                        id="aws-file-upload"
                        className="hidden"
                        onChange={handleFileUpload}
                    />
                </div>
                <Textarea
                    placeholder="[default]&#10;aws_access_key_id = ...&#10;aws_secret_access_key = ..."
                    value={content}
                    onChange={(e) => setContent(e.target.value)}
                    className="font-mono min-h-[300px]"
                />
            </CardContent>
            <CardFooter className="justify-end bg-muted/50 py-4 px-6">
                <Button onClick={handleSave} disabled={isSaving}>
                    {isSaving ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Save className="mr-2 h-4 w-4" />}
                    {t('common.save', 'Save')}
                </Button>
            </CardFooter>
        </Card>
    )
}
