import { useState, useEffect } from 'react'
import {
    Check,
    Loader2,
    Shield,
    User,
} from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { toast } from 'sonner'

import { api, InitCheckResponse } from '@/lib/api'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

// Internal hook since we don't have swr/react-query
function useInitCheck() {
    const [data, setData] = useState<InitCheckResponse | null>(null)
    const [isLoading, setIsLoading] = useState(true)
    const [error, setError] = useState<Error | null>(null)

    const fetchCheck = async () => {
        setIsLoading(true)
        try {
            const res = await api.checkInit()
            setData(res)
            setError(null)
        } catch (err) {
            setError(err as Error)
        } finally {
            setIsLoading(false)
        }
    }

    useEffect(() => {
        fetchCheck()
    }, [])

    return { data, isLoading, error, refetch: fetchCheck }
}

interface InitStepProps {
    step: number
    currentStep: number
    title: string
    description: string
    icon: React.ElementType
    completed: boolean
    children: React.ReactNode
}

function InitStep({
    step,
    currentStep,
    title,
    description,
    icon: Icon,
    completed,
    children,
}: InitStepProps) {
    const isActive = step === currentStep
    const isPending = step > currentStep

    return (
        <div className={`space-y-4 ${isPending ? 'opacity-50' : ''}`}>
            <div className="flex items-center space-x-3">
                <div
                    className={`flex aspect-square h-10 w-10 items-center justify-center rounded-full border-2 flex-shrink-0 ${completed
                        ? 'border-green-500 bg-green-500 text-white'
                        : isActive
                            ? 'border-blue-500 bg-blue-50 text-blue-600'
                            : 'border-gray-300 bg-gray-50 text-gray-400'
                        }`}
                >
                    {completed ? (
                        <Check className="h-5 w-5" />
                    ) : (
                        <Icon className="h-5 w-5" />
                    )}
                </div>
                <div>
                    <h3
                        className={`text-lg font-medium ${completed
                            ? 'text-green-600'
                            : isActive
                                ? 'text-gray-900'
                                : 'text-gray-400'
                            }`}
                    >
                        {title}
                    </h3>
                    <p
                        className={`text-xs text-muted-foreground ${completed
                            ? 'text-green-600'
                            : isActive
                                ? 'text-gray-600'
                                : 'text-gray-400'
                            }`}
                    >
                        {description}
                    </p>
                </div>
            </div>
            {isActive && <div className="ml-14">{children}</div>}
        </div>
    )
}

export function SetupPage() {
    const { data: initCheck, isLoading, refetch } = useInitCheck()
    const [isSubmitting, setIsSubmitting] = useState(false)
    const [error, setError] = useState<string | null>(null)
    const navigate = useNavigate()

    // User form state
    const [email, setEmail] = useState('')
    const [password, setPassword] = useState('')
    const [confirmPassword, setConfirmPassword] = useState('')
    const [name, setName] = useState('')

    // If loading, show spinner
    if (isLoading) {
        return (
            <div className="flex items-center justify-center min-h-screen">
                <Loader2 className="h-8 w-8 animate-spin text-primary" />
            </div>
        )
    }

    // If already initialized, redirect to home
    if (initCheck?.initialized) {
        setTimeout(() => navigate('/', { replace: true }), 0);
        return null;
    }

    const step = initCheck?.step || 0
    const currentStep = step + 1

    const handleCreateUser = async (e: React.FormEvent) => {
        e.preventDefault()
        setError(null)

        if (password !== confirmPassword) {
            setError("Passwords do not match")
            return
        }

        setIsSubmitting(true)
        try {
            await api.createSuperUser({
                email,
                password,
                name: name || undefined,
            })
            toast.success("Super user created successfully")
            await refetch()
        } catch (err) {
            setError(
                err instanceof Error
                    ? err.message
                    : "Failed to create user"
            )
        } finally {
            setIsSubmitting(false)
        }
    }

    const handleSkipOIDC = async () => {
        setError(null)
        setIsSubmitting(true)
        try {
            await api.skipOIDC()
            toast.success("OIDC configuration skipped")
            await refetch()
        } catch (err) {
            setError(
                err instanceof Error
                    ? err.message
                    : "Failed to skip OIDC"
            )
        } finally {
            setIsSubmitting(false)
        }
    }

    return (
        <div className="min-h-screen flex flex-col bg-gray-50/50">
            <div className="flex-1 flex items-center justify-center py-8 px-4">
                <div className="w-full max-w-2xl">
                    <div className="text-center mb-8">
                        <h1 className="text-3xl font-bold tracking-tight text-gray-900">Cloud Sentinel</h1>
                        <p className="mt-2 text-lg text-gray-600">Initial Setup</p>
                    </div>

                    <Card className="shadow-lg border-t-4 border-t-primary">
                        <CardHeader className="pb-6">
                            <CardTitle className="text-xl">
                                Welcome to Cloud Sentinel
                            </CardTitle>
                            <CardDescription>
                                Let's get your instance configured. This will only take a moment.
                            </CardDescription>
                        </CardHeader>

                        <CardContent className="space-y-6">
                            {error && (
                                <Alert variant="destructive">
                                    <AlertDescription>{error}</AlertDescription>
                                </Alert>
                            )}

                            {/* Step 1: Create Super Admin User */}
                            <InitStep
                                step={1}
                                currentStep={currentStep}
                                title="Create Administrator"
                                description="Set up your super admin account"
                                icon={User}
                                completed={step >= 1}
                            >
                                <form onSubmit={handleCreateUser} className="space-y-4 pt-4">
                                    <div className="space-y-2">
                                        <Label htmlFor="email">Email Address</Label>
                                        <Input
                                            id="email"
                                            type="email"
                                            placeholder="admin@example.com"
                                            value={email}
                                            onChange={(e) => setEmail(e.target.value)}
                                            required
                                        />
                                    </div>
                                    <div className="space-y-2">
                                        <Label htmlFor="name">Display Name (Optional)</Label>
                                        <Input
                                            id="name"
                                            type="text"
                                            placeholder="Admin User"
                                            value={name}
                                            onChange={(e) => setName(e.target.value)}
                                        />
                                    </div>
                                    <div className="grid grid-cols-2 gap-4">
                                        <div className="space-y-2">
                                            <Label htmlFor="password">Password</Label>
                                            <Input
                                                id="password"
                                                type="password"
                                                value={password}
                                                onChange={(e) => setPassword(e.target.value)}
                                                required
                                                minLength={8}
                                            />
                                        </div>
                                        <div className="space-y-2">
                                            <Label htmlFor="confirmPassword">Confirm Password</Label>
                                            <Input
                                                id="confirmPassword"
                                                type="password"
                                                value={confirmPassword}
                                                onChange={(e) => setConfirmPassword(e.target.value)}
                                                required
                                                minLength={8}
                                            />
                                        </div>
                                    </div>
                                    <Button
                                        type="submit"
                                        disabled={isSubmitting}
                                        className="w-full"
                                    >
                                        {isSubmitting ? (
                                            <div className="flex items-center space-x-2">
                                                <Loader2 className="h-4 w-4 animate-spin" />
                                                <span>Creating Account...</span>
                                            </div>
                                        ) : (
                                            "Create Account"
                                        )}
                                    </Button>
                                </form>
                            </InitStep>

                            {/* Step 2: Configure OIDC */}
                            <InitStep
                                step={2}
                                currentStep={currentStep}
                                title="Configure Authentication"
                                description="Set up OpenID Connect (OIDC)"
                                icon={Shield}
                                completed={step >= 2}
                            >
                                <div className="space-y-4 pt-4">
                                    <div className="bg-blue-50 p-4 rounded-md text-sm text-blue-700">
                                        <p className="font-semibold mb-1">OIDC Configuration</p>
                                        <p>
                                            You can configure OIDC (Okta, Google, etc.) to allow users to log in with their existing credentials.
                                            Currently, this is best configured via environment variables.
                                        </p>
                                    </div>

                                    <div className="flex flex-col gap-3">
                                        <Button
                                            variant="outline"
                                            onClick={handleSkipOIDC}
                                            disabled={isSubmitting}
                                            className="w-full"
                                        >
                                            {isSubmitting ? (
                                                <Loader2 className="h-4 w-4 animate-spin mr-2" />
                                            ) : null}
                                            Skip for now
                                        </Button>
                                    </div>
                                    <p className="text-xs text-gray-500 text-center">
                                        You can always configure this later via Environment Variables.
                                    </p>
                                </div>
                            </InitStep>

                        </CardContent>
                    </Card>
                </div>
            </div>
        </div>
    )
}
