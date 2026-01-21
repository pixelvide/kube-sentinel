import { useState } from 'react'
import { Check, ChevronsUpDown, Filter } from 'lucide-react'

import { useResources } from '@/lib/api'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@/components/ui/command'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'

export function NamespaceSelector({
  selectedNamespace,
  handleNamespaceChange,
  showAll = false,
  multiple = false,
}: {
  selectedNamespace?: string
  handleNamespaceChange: (namespace: string) => void
  showAll?: boolean
  multiple?: boolean
}) {
  const [open, setOpen] = useState(false)
  const { data } = useResources('namespaces')

  const selectedNamespaces =
    !selectedNamespace || selectedNamespace === '_all'
      ? []
      : selectedNamespace.split(',').filter(Boolean)

  const sortedNamespaces =
    data?.sort((a, b) => {
      const nameA = a.metadata?.name?.toLowerCase() || ''
      const nameB = b.metadata?.name?.toLowerCase() || ''
      return nameA.localeCompare(nameB)
    }) || []

  const isAllActive =
    !selectedNamespace ||
    selectedNamespace === '_all' ||
    (sortedNamespaces.length > 0 &&
      selectedNamespaces.length === sortedNamespaces.length)

  const toggleNamespace = (nsName: string) => {
    if (!multiple) {
      handleNamespaceChange(nsName)
      setOpen(false)
      return
    }

    let newSelected: string[]
    if (nsName === '_all') {
      newSelected = ['_all']
    } else {
      const current = selectedNamespaces.filter((n) => n !== '_all')
      if (current.includes(nsName)) {
        newSelected = current.filter((n) => n !== nsName)
      } else {
        newSelected = [...current, nsName]
      }

      if (
        newSelected.length === 0 ||
        (sortedNamespaces.length > 0 &&
          newSelected.length === sortedNamespaces.length)
      ) {
        newSelected = ['_all']
      }
    }
    handleNamespaceChange(newSelected.join(','))
  }

  const isSelected = (nsName: string) => {
    if (isAllActive) return true
    return selectedNamespaces.includes(nsName)
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          className="justify-between min-w-[150px] max-w-[300px] h-9"
        >
          <div className="flex items-center gap-2 overflow-hidden">
            <Filter className="h-4 w-4 shrink-0 opacity-50" />
            <span className="truncate">
              {isAllActive
                ? 'All Namespaces'
                : !multiple
                  ? selectedNamespace
                  : selectedNamespaces.length === 1
                    ? selectedNamespaces[0]
                    : `${selectedNamespaces.length} namespaces`}
            </span>
          </div>
          <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[250px] p-0" align="start">
        <Command>
          <CommandInput placeholder="Search namespace..." />
          <CommandList>
            <CommandEmpty>No namespace found.</CommandEmpty>
            <CommandGroup>
              {showAll && (
                <CommandItem
                  key="_all"
                  onSelect={() => {
                    toggleNamespace('_all')
                  }}
                  className="flex items-center gap-2"
                >
                  <Check
                    className={cn(
                      'h-4 w-4',
                      isAllActive ? 'opacity-100' : 'opacity-0'
                    )}
                  />
                  <span>All Namespaces</span>
                </CommandItem>
              )}
              {sortedNamespaces.map((ns) => (
                <CommandItem
                  key={ns.metadata!.name}
                  onSelect={() => {
                    toggleNamespace(ns.metadata!.name!)
                  }}
                  className="flex items-center gap-2"
                >
                  <Check
                    className={cn(
                      'h-4 w-4',
                      isSelected(ns.metadata!.name!)
                        ? 'opacity-100'
                        : 'opacity-0'
                    )}
                  />
                  <span>{ns.metadata!.name}</span>
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  )
}
