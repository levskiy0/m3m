/**
 * SortableEnvRow component
 * Draggable environment variable row for inline table editing
 */

import { GripVertical, Trash2, Eye, EyeOff, Copy } from 'lucide-react';
import { useSortable } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { useState } from 'react';
import { toast } from 'sonner';

import type { EnvType } from '@/types';
import { ENV_TYPES } from '@/lib/constants';
import { copyToClipboard } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Switch } from '@/components/ui/switch';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';

export interface EnvRowData {
  key: string;
  type: EnvType;
  value: string;
  isNew?: boolean;
}

interface SortableEnvRowProps {
  id: string;
  env: EnvRowData;
  onUpdate: (updates: Partial<EnvRowData>) => void;
  onRemove: () => void;
}

const SENSITIVE_PATTERNS = ['PASS', 'PWD', 'TOKEN', 'SECRET'];

function isSensitiveKey(key: string): boolean {
  const upper = key.toUpperCase();
  return SENSITIVE_PATTERNS.some((p) => upper.includes(p));
}

function ValueInput({
  type,
  value,
  onChange,
  showValue,
}: {
  type: EnvType;
  value: string;
  onChange: (value: string) => void;
  showValue: boolean;
}) {
  if (!showValue) {
    return (
      <Input
        type="password"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="h-8 font-mono text-sm"
      />
    );
  }

  switch (type) {
    case 'text':
    case 'json':
      return (
        <Textarea
          value={value}
          onChange={(e) => onChange(e.target.value)}
          rows={2}
          className={`font-mono text-sm resize-none ${type === 'json' ? '' : ''}`}
          placeholder={type === 'json' ? '{}' : ''}
        />
      );
    case 'integer':
      return (
        <Input
          type="number"
          value={value}
          onChange={(e) => onChange(e.target.value)}
          step="1"
          className="h-8 font-mono text-sm"
        />
      );
    case 'float':
      return (
        <Input
          type="number"
          value={value}
          onChange={(e) => onChange(e.target.value)}
          step="0.01"
          className="h-8 font-mono text-sm"
        />
      );
    case 'boolean':
      return (
        <div className="flex items-center gap-2 h-8">
          <Switch
            checked={value === 'true'}
            onCheckedChange={(checked) => onChange(checked ? 'true' : 'false')}
          />
          <span className="text-sm text-muted-foreground">
            {value === 'true' ? 'True' : 'False'}
          </span>
        </div>
      );
    default:
      return (
        <Input
          value={value}
          onChange={(e) => onChange(e.target.value)}
          className="h-8 font-mono text-sm"
        />
      );
  }
}

export function SortableEnvRow({
  id,
  env,
  onUpdate,
  onRemove,
}: SortableEnvRowProps) {
  const isSensitive = isSensitiveKey(env.key);
  const [showValue, setShowValue] = useState(!isSensitive);

  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  };

  const handleKeyChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    // Sanitize: uppercase, replace non-alphanumeric with underscore
    const sanitized = e.target.value.toUpperCase().replace(/[^A-Z0-9_]/g, '_');
    onUpdate({ key: sanitized });
  };

  const handleTypeChange = (newType: EnvType) => {
    // Reset value when switching types for better UX
    let newValue = env.value;
    if (newType === 'boolean' && env.value !== 'true' && env.value !== 'false') {
      newValue = 'false';
    } else if (newType === 'integer' && env.type !== 'integer' && env.type !== 'float') {
      newValue = '0';
    } else if (newType === 'float' && env.type !== 'integer' && env.type !== 'float') {
      newValue = '0';
    } else if (newType === 'json' && env.type !== 'json') {
      newValue = '{}';
    }
    onUpdate({ type: newType, value: newValue });
  };

  const handleCopy = async () => {
    const success = await copyToClipboard(env.value);
    if (success) {
      toast.success('Copied to clipboard');
    }
  };

  const keyError = !env.key || env.key.length === 0;

  return (
    <tr
      ref={setNodeRef}
      style={style}
      className={`border-t bg-background ${isDragging ? 'z-50' : ''}`}
    >
      <td className="p-3 w-10">
        <div
          {...attributes}
          {...listeners}
          className="cursor-grab active:cursor-grabbing text-muted-foreground hover:text-foreground transition-colors"
        >
          <GripVertical className="size-4" />
        </div>
      </td>
      <td className="p-3 w-[200px]">
        <Input
          value={env.key}
          onChange={handleKeyChange}
          placeholder="API_KEY"
          disabled={!env.isNew}
          className={`h-8 font-mono text-sm ${keyError ? 'border-destructive focus-visible:ring-destructive' : ''} ${!env.isNew ? 'bg-muted' : ''}`}
        />
      </td>
      <td className="p-3 w-[140px]">
        <Select value={env.type} onValueChange={handleTypeChange}>
          <SelectTrigger className="h-8">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {ENV_TYPES.map((t) => (
              <SelectItem key={t.value} value={t.value}>
                {t.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </td>
      <td className="p-3">
        <ValueInput
          type={env.type}
          value={env.value}
          onChange={(v) => onUpdate({ value: v })}
          showValue={isSensitive ? showValue : true}
        />
      </td>
      <td className="p-3 w-32">
        <div className="flex items-center gap-1">
          {isSensitive && (
            <Button
              variant="ghost"
              size="icon"
              className="size-8"
              onClick={() => setShowValue(!showValue)}
            >
              {showValue ? <EyeOff className="size-4" /> : <Eye className="size-4" />}
            </Button>
          )}
          <Button
            variant="ghost"
            size="icon"
            className="size-8"
            onClick={handleCopy}
          >
            <Copy className="size-4" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="size-8 text-muted-foreground hover:text-destructive"
            onClick={onRemove}
          >
            <Trash2 className="size-4" />
          </Button>
        </div>
      </td>
    </tr>
  );
}
