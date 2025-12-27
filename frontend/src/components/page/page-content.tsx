import { cn } from '@/lib/utils'

interface PageContentProps {
  children: React.ReactNode
  className?: string
}

export function PageContent({ children, className }: PageContentProps) {
  return (
    <div className={cn('mt-6 space-y-6', className)}>
      {children}
    </div>
  )
}
