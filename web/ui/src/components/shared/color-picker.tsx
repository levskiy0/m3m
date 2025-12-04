import { PRESET_COLORS } from '@/lib/constants';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';

interface ColorPickerProps {
  value?: string;
  onChange: (color: string | undefined) => void;
  allowNone?: boolean;
}

export function ColorPicker({ value, onChange, allowNone = true }: ColorPickerProps) {
  const selectedColor = PRESET_COLORS.find((c) => c.value === value);

  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="outline" className="w-full justify-start gap-2">
          {value ? (
            <>
              <span
                className="size-4 rounded-full border"
                style={{ backgroundColor: value }}
              />
              <span className="capitalize">{selectedColor?.name || 'Custom'}</span>
            </>
          ) : (
            <span className="text-muted-foreground">Select color...</span>
          )}
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-64">
        <div className="grid grid-cols-5 gap-2">
          {allowNone && (
            <button
              type="button"
              className={cn(
                'size-8 rounded-full border-2 flex items-center justify-center',
                !value && 'border-primary'
              )}
              onClick={() => onChange(undefined)}
              title="No color"
            >
              <span className="text-xs text-muted-foreground">-</span>
            </button>
          )}
          {PRESET_COLORS.map((color) => (
            <button
              key={color.name}
              type="button"
              className={cn(
                'size-8 rounded-full border-2 transition-transform hover:scale-110',
                value === color.value ? 'border-primary scale-110' : 'border-transparent'
              )}
              style={{ backgroundColor: color.value }}
              onClick={() => onChange(color.value)}
              title={color.name}
            />
          ))}
        </div>
      </PopoverContent>
    </Popover>
  );
}
