/**
 * RefFilterInput - compact ref field selector for filters
 */

import { useState, useEffect, useRef, useMemo } from 'react';
import { Check, ChevronsUpDown, Loader2 } from 'lucide-react';
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

interface RefFilterInputProps {
  value: string;
  onChange: (value: string) => void;
  projectId: string;
  refModelSlug: string;
  models: Model[];
  onCommit: () => void;
}

export function RefFilterInput({
  value,
  onChange,
  projectId,
  refModelSlug,
  models,
  onCommit,
}: RefFilterInputProps) {
  const [open, setOpen] = useState(false);
  const [records, setRecords] = useState<ModelData[]>([]);
  const [selectedRecord, setSelectedRecord] = useState<ModelData | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [searchValue, setSearchValue] = useState('');
  const debounceRef = useRef<NodeJS.Timeout | null>(null);
  const initialLoadDone = useRef(false);

  const refModel = models.find((m) => m.slug === refModelSlug);
  const refModelId = refModel?.id;

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

  // Load selected record
  useEffect(() => {
    if (value && refModelId && !selectedRecord) {
      modelsApi
        .getData(projectId, refModelId, value)
         
        .then(setSelectedRecord)
         
        .catch(() => setSelectedRecord(null));
    } else if (!value) {
      // eslint-disable-next-line react-hooks/set-state-in-effect -- intentional: clearing state when value becomes null
      setSelectedRecord(null);
    }
  }, [value, refModelId, projectId, selectedRecord]);

  // Load initial records
  useEffect(() => {
    if (open && refModelId && !initialLoadDone.current) {
      initialLoadDone.current = true;
      // eslint-disable-next-line react-hooks/set-state-in-effect -- intentional: setting loading state before API call
      setIsLoading(true);
      modelsApi
        .queryData(projectId, refModelId, { limit: 25 })
        .then((response) => setRecords(response.data))
        .catch(() => setRecords([]))
        .finally(() => setIsLoading(false));
    }
    if (!open) {
      setSearchValue('');
      initialLoadDone.current = false;
    }
  }, [open, refModelId, projectId]);

  // Search with debounce
  useEffect(() => {
    if (!open || !refModelId) return;
    if (debounceRef.current) clearTimeout(debounceRef.current);

    if (!searchValue.trim()) {
      modelsApi
        .queryData(projectId, refModelId, { limit: 25 })
        .then((response) => setRecords(response.data))
        .catch(() => setRecords([]));
      return;
    }

    debounceRef.current = setTimeout(() => {
      setIsLoading(true);
      modelsApi
        .queryData(projectId, refModelId, {
          limit: 25,
          search: searchValue,
          searchIn: searchableFields.length > 0 ? searchableFields : undefined,
        })
        .then((response) => setRecords(response.data))
        .catch(() => setRecords([]))
        .finally(() => setIsLoading(false));
    }, 300);

    return () => {
      if (debounceRef.current) clearTimeout(debounceRef.current);
    };
  }, [searchValue, open, refModelId, projectId, searchableFields]);

  const getRecordLabel = (record: ModelData) => {
    const displayValue = record[displayField];
    if (displayValue) return String(displayValue);
    return record._id.slice(-8);
  };

  const handleSelect = (recordId: string) => {
    const record = records.find(r => r._id === recordId);
    if (record) {
      setSelectedRecord(record);
      onChange(recordId);
      setOpen(false);
      onCommit();
    }
  };

  if (!refModel) {
    return <span className="text-xs text-muted-foreground">Model not found</span>;
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          className="h-8 w-[140px] justify-between text-xs font-normal"
        >
          <span className="truncate">
            {selectedRecord ? getRecordLabel(selectedRecord) : 'Select...'}
          </span>
          <ChevronsUpDown className="ml-1 h-3 w-3 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[200px] p-0" align="start">
        <Command shouldFilter={false}>
          <CommandInput
            placeholder="Search..."
            value={searchValue}
            onValueChange={setSearchValue}
            className="h-8 text-xs"
          />
          <CommandList>
            {isLoading ? (
              <div className="flex items-center justify-center py-4">
                <Loader2 className="h-4 w-4 animate-spin" />
              </div>
            ) : records.length === 0 ? (
              <CommandEmpty>No results</CommandEmpty>
            ) : (
              <CommandGroup>
                {records.map((record) => (
                  <CommandItem
                    key={record._id}
                    value={record._id}
                    onSelect={handleSelect}
                    className="text-xs"
                  >
                    <Check
                      className={cn(
                        'mr-2 h-3 w-3',
                        value === record._id ? 'opacity-100' : 'opacity-0'
                      )}
                    />
                    <span className="truncate">{getRecordLabel(record)}</span>
                  </CommandItem>
                ))}
              </CommandGroup>
            )}
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
