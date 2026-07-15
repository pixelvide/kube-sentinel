import { useRouteError, isRouteErrorResponse } from 'react-router-dom'
import { AlertCircle } from 'lucide-react'
import { Alert, AlertDescription, AlertTitle } from './alert'
import { Button } from './button'

export function ErrorBoundary() {
  const error = useRouteError()
  console.error("ErrorBoundary caught:", error)

  let errorMessage = "An unexpected error occurred."

  if (isRouteErrorResponse(error)) {
    errorMessage = error.data?.message || error.statusText
  } else if (error instanceof Error) {
    errorMessage = error.message
  } else if (typeof error === 'string') {
    errorMessage = error
  }

  return (
    <div className="flex h-full w-full flex-col items-center justify-center p-6 bg-background">
      <div className="mx-auto max-w-md w-full space-y-6">
        <div className="flex flex-col items-center space-y-4 text-center">
          <div className="rounded-full bg-destructive/10 p-4">
            <AlertCircle className="h-10 w-10 text-destructive" />
          </div>
          <h1 className="text-2xl font-bold tracking-tight">Something went wrong</h1>
          <p className="text-muted-foreground">
            We encountered an unexpected error while rendering this page.
          </p>
        </div>

        <Alert variant="destructive">
          <AlertTitle>Error Details</AlertTitle>
          <AlertDescription className="mt-2 text-xs font-mono break-words overflow-auto max-h-[200px]">
            {errorMessage}
          </AlertDescription>
        </Alert>

        <div className="flex justify-center gap-4 pt-4">
          <Button variant="outline" onClick={() => window.location.reload()}>
            Reload Page
          </Button>
          <Button variant="default" onClick={() => window.location.href = '/'}>
            Go to Home
          </Button>
        </div>
      </div>
    </div>
  )
}
