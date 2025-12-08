import { useState, useEffect, useRef, useMemo } from 'react';
import { Check, ChevronsUpDown, Loader2, X } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@/components/ui/command';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';
import { modelsApi } from '@/api/models';
import type { ModelData, Model } from '@/types';

const INITIAL_LIMIT = 25;
const SEARCH_DEBOUNCE_MS = 300;

interface RefFieldInputProps {
  value: string | null;
  onChange: (value: string | null) => void;
  projectId: string;
  refModelSlug: string;
  models: Model[];
}

export function RefFieldInput({
  value,
  onChange,
  projectId,
  refModelSlug,
  models,
}: RefFieldInputProps) {
  const [open, setOpen] = useState(false);
  const [records, setRecords] = useState<ModelData[]>([]);
  const [selectedRecord, setSelectedRecord] = useState<ModelData | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [searchValue, setSearchValue] = useState('');
  const debounceRef = useRef<NodeJS.Timeout | null>(null);
  const initialLoadDone = useRef(false);

  // Find the referenced model
  const refModel = models.find((m) => m.slug === refModelSlug);
  const refModelId = refModel?.id;

  // Get display field (first string field) and searchable fields - memoized
  const displayField = useMemo(() =>
    refModel?.fields.find(f => f.type === 'string')?.key || '_id',
    [refModel?.fields]
  );

  const searchableFields = useMemo(() =>
    refModel?.fields
      .filter(f => f.type === 'string' || f.type === 'text')
      .map(f => f.key) || [],
    [refModel?.fields]
  );

  // Load selected record for display (if value exists)
  useEffect(() => {
    if (value && refModelId && !selectedRecord) {
      modelsApi
        .getData(projectId, refModelId, value)
        .then((record) => {
           
          setSelectedRecord(record);
        })
        .catch(() => {
           
          setSelectedRecord(null);
        });
    } else if (!value) {
      // eslint-disable-next-line react-hooks/set-state-in-effect -- intentional: clearing state when value becomes null
      setSelectedRecord(null);
    }
  }, [value, refModelId, projectId, selectedRecord]);

  // Load initial 25 records when popover opens
  useEffect(() => {
    if (open && refModelId && !initialLoadDone.current) {
      initialLoadDone.current = true;
      // eslint-disable-next-line react-hooks/set-state-in-effect -- intentional: setting loading state before API call
      setIsLoading(true);
      modelsApi
        .queryData(projectId, refModelId, { limit: INITIAL_LIMIT })
        .then((response) => {
          setRecords(response.data);
        })
        .catch((error) => {
          console.error('Failed to load ref records:', error);
          setRecords([]);
        })
        .finally(() => {
          setIsLoading(false);
        });
    }

    // Reset when popover closes
    if (!open) {
      initialLoadDone.current = false;
      setSearchValue('');
    }
  }, [open, refModelId, projectId]);

  // Handle search with debounce
  const handleSearchChange = (value: string) => {
    setSearchValue(value);

    if (debounceRef.current) {
      clearTimeout(debounceRef.current);
    }

    if (!refModelId) return;

    debounceRef.current = setTimeout(() => {
      setIsLoading(true);

      const query = value.trim();
      modelsApi
        .queryData(projectId, refModelId, {
          limit: INITIAL_LIMIT,
          search: query || undefined,
          searchIn: query && searchableFields.length > 0 ? searchableFields : undefined,
        })
        .then((response) => {
          setRecords(response.data);
        })
        .catch((error) => {
          console.error('Failed to search ref records:', error);
        })
        .finally(() => {
          setIsLoading(false);
        });
    }, SEARCH_DEBOUNCE_MS);
  };

  // Cleanup debounce on unmount
  useEffect(() => {
    return () => {
      if (debounceRef.current) {
        clearTimeout(debounceRef.current);
      }
    };
  }, []);

  // Get display value
  const displayValue = selectedRecord
    ? String(selectedRecord[displayField] || selectedRecord._id)
    : value || '';

  if (!refModel) {
    return (
      <div className="text-sm text-muted-foreground border rounded-md px-3 py-2">
        Model "{refModelSlug}" not found
      </div>
    );
  }

  return (
    <div className="flex gap-2">
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            variant="outline"
            role="combobox"
            aria-expanded={open}
            className="flex-1 justify-between font-normal"
          >
            {value ? (
              <span className="truncate">{displayValue}</span>
            ) : (
              <span className="text-muted-foreground">Select {refModel.name}...</span>
            )}
            <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-[var(--radix-popover-trigger-width)] p-0" align="start">
          <Command shouldFilter={false}>
            <CommandInput
              placeholder={`Search ${refModel.name}...`}
              value={searchValue}
              onValueChange={handleSearchChange}
            />
            <CommandList>
              {isLoading ? (
                <div className="flex items-center justify-center py-6">
                  <Loader2 className="h-4 w-4 animate-spin" />
                </div>
              ) : records.length === 0 ? (
                <CommandEmpty>
                  {searchValue ? 'No records found.' : 'No records available.'}
                </CommandEmpty>
              ) : (
                <CommandGroup>
                  {records.map((record) => (
                    <CommandItem
                      key={record._id}
                      value={record._id}
                      onSelect={() => {
                        onChange(record._id);
                        setSelectedRecord(record);
                        setOpen(false);
                      }}
                    >
                      <Check
                        className={cn(
                          'mr-2 h-4 w-4',
                          value === record._id ? 'opacity-100' : 'opacity-0'
                        )}
                      />
                      <span className="truncate">
                        {String(record[displayField] || record._id)}
                      </span>
                    </CommandItem>
                  ))}
                </CommandGroup>
              )}
            </CommandList>
          </Command>
        </PopoverContent>
      </Popover>
      {value && (
        <Button
          variant="outline"
          size="icon"
          onClick={() => {
            onChange(null);
            setSelectedRecord(null);
          }}
          className="shrink-0"
        >
          <X className="h-4 w-4" />
        </Button>
      )}
    </div>
  );
}
