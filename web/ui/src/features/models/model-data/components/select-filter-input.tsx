/**
 * SelectFilterInput component
 * Multi-select filter for select/multiselect field types
 */

import { X, Check, ChevronDown } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@/components/ui/command';
import { cn } from '@/lib/utils';

interface SelectFilterInputProps {
  value: string[];
  onChange: (value: string[]) => void;
  options: string[];
  onCommit: () => void;
}

export function SelectFilterInput({
  value,
  onChange,
  options,
  onCommit,
}: SelectFilterInputProps) {
  const selectedValues = Array.isArray(value) ? value : [];

  const toggleOption = (option: string) => {
    const newValue = selectedValues.includes(option)
      ? selectedValues.filter(v => v !== option)
      : [...selectedValues, option];
    onChange(newValue);
  };

  const handleOpenChange = (open: boolean) => {
    if (!open && selectedValues.length > 0) {
      onCommit();
    }
  };

  return (
    <Popover onOpenChange={handleOpenChange}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          className="h-8 w-[140px] justify-between px-2 text-xs"
        >
          {selectedValues.length > 0 ? (
            <div className="flex items-center gap-1 overflow-hidden">
              {selectedValues.length === 1 ? (
                <span className="truncate">{selectedValues[0]}</span>
              ) : (
                <Badge variant="secondary" className="rounded-sm px-1 font-normal text-xs">
                  {selectedValues.length} selected
                </Badge>
              )}
            </div>
          ) : (
            <span className="text-muted-foreground">Select...</span>
          )}
          <ChevronDown className="size-3.5 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-52 p-0" align="start">
        <Command>
          <CommandInput placeholder="Search..." className="h-8" />
          <CommandList>
            <CommandEmpty>No options found</CommandEmpty>
            <CommandGroup>
              {options.map((option) => {
                const isSelected = selectedValues.includes(option);
                return (
                  <CommandItem
                    key={option}
                    onSelect={() => toggleOption(option)}
                    className="cursor-pointer"
                  >
                    <div
                      className={cn(
                        'mr-2 flex h-4 w-4 items-center justify-center rounded-sm border border-primary',
                        isSelected
                          ? 'bg-primary text-primary-foreground'
                          : 'opacity-50 [&_svg]:invisible'
                      )}
                    >
                      <Check className="size-3" />
                    </div>
                    <span className="truncate">{option}</span>
                  </CommandItem>
                );
              })}
            </CommandGroup>
          </CommandList>
        </Command>
        {selectedValues.length > 0 && (
          <div className="border-t p-2">
            <Button
              variant="ghost"
              size="sm"
              className="h-7 w-full text-xs"
              onClick={() => {
                onChange([]);
              }}
            >
              <X className="mr-1 size-3" />
              Clear selection
            </Button>
          </div>
        )}
      </PopoverContent>
    </Popover>
  );
}
