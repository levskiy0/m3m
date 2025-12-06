import * as React from 'react';
import { format, parse } from 'date-fns';
import { ChevronDownIcon } from 'lucide-react';

import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Calendar } from '@/components/ui/calendar';
import { Input } from '@/components/ui/input';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';

interface DateTimePickerProps {
  value?: string;
  onChange: (value: string | undefined) => void;
  placeholder?: string;
  disabled?: boolean;
}

export function DateTimePicker({
  value,
  onChange,
  placeholder = 'Select date',
  disabled,
}: DateTimePickerProps) {
  const [open, setOpen] = React.useState(false);
  const [date, setDate] = React.useState<Date | undefined>(() => {
    if (!value) return undefined;
    try {
      const d = new Date(value);
      return isNaN(d.getTime()) ? undefined : d;
    } catch {
      return undefined;
    }
  });
  const [time, setTime] = React.useState(() => {
    if (!value) return '';
    try {
      const d = new Date(value);
      if (isNaN(d.getTime())) return '';
      return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`;
    } catch {
      return '';
    }
  });

  // Update internal state when value prop changes
  React.useEffect(() => {
    if (!value) {
      setDate(undefined);
      setTime('');
      return;
    }
    try {
      const d = new Date(value);
      if (!isNaN(d.getTime())) {
        setDate(d);
        setTime(`${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`);
      }
    } catch {
      // Invalid date
    }
  }, [value]);

  const updateValue = (newDate: Date | undefined, newTime: string) => {
    if (!newDate) {
      onChange(undefined);
      return;
    }
    const [hours, minutes] = (newTime || '00:00').split(':').map(Number);
    const result = new Date(newDate);
    result.setHours(hours || 0, minutes || 0, 0, 0);
    onChange(format(result, "yyyy-MM-dd'T'HH:mm"));
  };

  const handleDateSelect = (selectedDate: Date | undefined) => {
    setDate(selectedDate);
    updateValue(selectedDate, time);
    setOpen(false);
  };

  const handleTimeChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newTime = e.target.value;
    setTime(newTime);
    updateValue(date, newTime);
  };

  return (
    <div className="flex gap-2">
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            variant="outline"
            disabled={disabled}
            className={cn(
              'flex-1 justify-between font-normal',
              !date && 'text-muted-foreground'
            )}
          >
            {date ? date.toLocaleDateString() : placeholder}
            <ChevronDownIcon className="h-4 w-4 opacity-50" />
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-auto overflow-hidden p-0" align="start">
          <Calendar
            mode="single"
            selected={date}
            captionLayout="dropdown"
            onSelect={handleDateSelect}
          />
        </PopoverContent>
      </Popover>
      <Input
        type="time"
        value={time}
        onChange={handleTimeChange}
        disabled={disabled}
        className="w-28 bg-background appearance-none [&::-webkit-calendar-picker-indicator]:hidden [&::-webkit-calendar-picker-indicator]:appearance-none"
      />
    </div>
  );
}

interface DatePickerProps {
  value?: string;
  onChange: (value: string | undefined) => void;
  placeholder?: string;
  disabled?: boolean;
}

export function DatePicker({
  value,
  onChange,
  placeholder = 'Select date',
  disabled,
}: DatePickerProps) {
  const [open, setOpen] = React.useState(false);
  const [date, setDate] = React.useState<Date | undefined>(() => {
    if (!value) return undefined;
    try {
      const d = parse(value, 'yyyy-MM-dd', new Date());
      return isNaN(d.getTime()) ? undefined : d;
    } catch {
      return undefined;
    }
  });

  // Update internal state when value prop changes
  React.useEffect(() => {
    if (!value) {
      setDate(undefined);
      return;
    }
    try {
      const d = parse(value, 'yyyy-MM-dd', new Date());
      if (!isNaN(d.getTime())) {
        setDate(d);
      }
    } catch {
      // Invalid date
    }
  }, [value]);

  const handleSelect = (selectedDate: Date | undefined) => {
    setDate(selectedDate);
    if (selectedDate) {
      onChange(format(selectedDate, 'yyyy-MM-dd'));
    } else {
      onChange(undefined);
    }
    setOpen(false);
  };

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          disabled={disabled}
          className={cn(
            'w-full justify-between font-normal',
            !date && 'text-muted-foreground'
          )}
        >
          {date ? date.toLocaleDateString() : placeholder}
          <ChevronDownIcon className="h-4 w-4 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-auto overflow-hidden p-0" align="start">
        <Calendar
          mode="single"
          selected={date}
          captionLayout="dropdown"
          onSelect={handleSelect}
        />
      </PopoverContent>
    </Popover>
  );
}
