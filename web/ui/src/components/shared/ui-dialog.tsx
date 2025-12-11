import { useState } from 'react';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Checkbox } from '@/components/ui/checkbox';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';
import { Label } from '@/components/ui/label';
import {
  AlertCircle,
  CheckCircle2,
  AlertTriangle,
  Info,
  Loader2,
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { useUIDialogStore, type UIDialogRequest } from '@/stores/ui-dialog-store';
import type { UIFormField, UIFormAction, UIRequestOptions } from '@/lib/websocket';

const severityIcons = {
  info: Info,
  success: CheckCircle2,
  warning: AlertTriangle,
  error: AlertCircle,
};

const severityColors = {
  info: 'text-blue-500',
  success: 'text-green-500',
  warning: 'text-yellow-500',
  error: 'text-red-500',
};

function AlertDialogComponent({ dialog }: { dialog: UIDialogRequest }) {
  const { closeDialog } = useUIDialogStore();
  const options = dialog.options as UIRequestOptions;
  const severity = options.severity || 'info';
  const Icon = severityIcons[severity];

  return (
    <AlertDialog open onOpenChange={() => closeDialog()}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle className="flex items-center gap-2">
            <Icon className={cn('h-5 w-5', severityColors[severity])} />
            {options.title || 'Alert'}
          </AlertDialogTitle>
          {options.text && (
            <AlertDialogDescription>{options.text}</AlertDialogDescription>
          )}
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogAction onClick={() => closeDialog()}>OK</AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}

function ConfirmDialogComponent({ dialog }: { dialog: UIDialogRequest }) {
  const { respondToDialog } = useUIDialogStore();
  const options = dialog.options as UIRequestOptions;

  return (
    <AlertDialog open onOpenChange={() => respondToDialog(false)}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{options.title || 'Confirm'}</AlertDialogTitle>
          {options.text && (
            <AlertDialogDescription>{options.text}</AlertDialogDescription>
          )}
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel onClick={() => respondToDialog(false)}>
            {options.no || 'No'}
          </AlertDialogCancel>
          <AlertDialogAction onClick={() => respondToDialog(true)}>
            {options.yes || 'Yes'}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}

function PromptDialogComponent({ dialog }: { dialog: UIDialogRequest }) {
  const { respondToDialog } = useUIDialogStore();
  const options = dialog.options as UIRequestOptions;
  const [value, setValue] = useState(options.defaultValue || '');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    respondToDialog(value);
  };

  return (
    <Dialog open onOpenChange={() => respondToDialog(null)}>
      <DialogContent>
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>{options.title || 'Enter Value'}</DialogTitle>
            {options.text && (
              <DialogDescription>{options.text}</DialogDescription>
            )}
          </DialogHeader>
          <div className="py-4">
            <Input
              value={value}
              onChange={(e) => setValue(e.target.value)}
              placeholder={options.placeholder}
              autoFocus
            />
          </div>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => respondToDialog(null)}
            >
              Cancel
            </Button>
            <Button type="submit">OK</Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

// Static mapping for Tailwind col-span classes (dynamic classes don't work)
const colSpanClasses: Record<number | 'full', string> = {
  1: 'col-span-1',
  2: 'col-span-2',
  3: 'col-span-3',
  4: 'col-span-4',
  5: 'col-span-5',
  6: 'col-span-6',
  'full': 'col-span-6',
};

function FormField({
  field,
  value,
  onChange,
  error,
  disabled,
}: {
  field: UIFormField;
  value: unknown;
  onChange: (value: unknown) => void;
  error?: string;
  disabled?: boolean;
}) {
  const colSpan = colSpanClasses[field.colspan || 6] || 'col-span-6';

  const renderField = () => {
    switch (field.type) {
      case 'input':
        return (
          <Input
            value={(value as string) || ''}
            onChange={(e) => onChange(e.target.value)}
            placeholder={field.placeholder}
            required={field.required}
            disabled={disabled}
          />
        );

      case 'textarea':
        return (
          <Textarea
            value={(value as string) || ''}
            onChange={(e) => onChange(e.target.value)}
            placeholder={field.placeholder}
            required={field.required}
            rows={3}
            disabled={disabled}
          />
        );

      case 'checkbox':
        return (
          <div className="flex items-center space-x-2">
            <Checkbox
              id={field.name}
              checked={!!value}
              onCheckedChange={(checked) => onChange(checked)}
              disabled={disabled}
            />
            <Label htmlFor={field.name} className="cursor-pointer">
              {field.label}
            </Label>
          </div>
        );

      case 'select':
      case 'combobox': {
        const options = field.options || [];
        return (
          <Select
            value={(value as string) || ''}
            onValueChange={onChange}
            disabled={disabled}
          >
            <SelectTrigger>
              <SelectValue placeholder={field.placeholder || 'Select...'} />
            </SelectTrigger>
            <SelectContent>
              {options.map((opt) => {
                const optValue = typeof opt === 'string' ? opt : opt.value;
                const optLabel = typeof opt === 'string' ? opt : opt.label;
                return (
                  <SelectItem key={optValue} value={optValue}>
                    {optLabel}
                  </SelectItem>
                );
              })}
            </SelectContent>
          </Select>
        );
      }

      case 'radiogroup': {
        const options = field.options || [];
        return (
          <RadioGroup
            value={(value as string) || ''}
            onValueChange={onChange}
            disabled={disabled}
          >
            {options.map((opt) => {
              const optValue = typeof opt === 'string' ? opt : opt.value;
              const optLabel = typeof opt === 'string' ? opt : opt.label;
              return (
                <div key={optValue} className="flex items-center space-x-2">
                  <RadioGroupItem value={optValue} id={`${field.name}-${optValue}`} />
                  <Label htmlFor={`${field.name}-${optValue}`}>{optLabel}</Label>
                </div>
              );
            })}
          </RadioGroup>
        );
      }

      case 'date':
        return (
          <Input
            type="date"
            value={(value as string) || ''}
            onChange={(e) => onChange(e.target.value)}
            required={field.required}
            disabled={disabled}
          />
        );

      case 'datetime':
        return (
          <Input
            type="datetime-local"
            value={(value as string) || ''}
            onChange={(e) => onChange(e.target.value)}
            required={field.required}
            disabled={disabled}
          />
        );

      default:
        return (
          <Input
            value={(value as string) || ''}
            onChange={(e) => onChange(e.target.value)}
            placeholder={field.placeholder}
            disabled={disabled}
          />
        );
    }
  };

  // Checkbox has its own label
  if (field.type === 'checkbox') {
    return (
      <div className={cn(colSpan, 'space-y-1')}>
        {renderField()}
        {field.hint && (
          <p className="text-xs text-muted-foreground">{field.hint}</p>
        )}
        {error && <p className="text-xs text-destructive">{error}</p>}
      </div>
    );
  }

  return (
    <div className={cn(colSpan, 'space-y-1')}>
      {field.label && (
        <Label className="text-sm font-medium">
          {field.label}
          {field.required && <span className="text-destructive ml-1">*</span>}
        </Label>
      )}
      {renderField()}
      {field.hint && (
        <p className="text-xs text-muted-foreground">{field.hint}</p>
      )}
      {error && <p className="text-xs text-destructive">{error}</p>}
    </div>
  );
}

function FormDialogComponent({ dialog }: { dialog: UIDialogRequest }) {
  const { respondToDialog, closeDialog, getFormState } = useUIDialogStore();
  const options = dialog.options as UIRequestOptions;
  const schema = options.schema || [];
  const actions = options.actions || [
    { label: 'Cancel', variant: 'outline' as const, action: 'cancel' },
    { label: 'Submit', variant: 'default' as const, action: 'submit' },
  ];

  // Get form state from store (loading, errors)
  const formState = getFormState(dialog.requestId);

  const [formData, setFormData] = useState<Record<string, unknown>>(() => {
    const initial: Record<string, unknown> = {};
    for (const field of schema) {
      initial[field.name] = field.defaultValue ?? '';
    }
    return initial;
  });

  const handleFieldChange = (name: string, value: unknown) => {
    setFormData((prev) => ({ ...prev, [name]: value }));
  };

  const handleAction = (action: UIFormAction) => {
    if (action.action === 'cancel') {
      // Cancel closes the form without waiting for backend
      closeDialog();
      return;
    }

    // For submit or custom actions, send the form data
    // Form stays open - backend will send close or errors via form_update
    respondToDialog({
      action: action.action,
      data: formData,
    });
  };

  return (
    <Dialog open onOpenChange={() => closeDialog()}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>{options.title || 'Form'}</DialogTitle>
          {options.text && (
            <DialogDescription>{options.text}</DialogDescription>
          )}
        </DialogHeader>
        <div className="grid grid-cols-6 gap-4 py-4">
          {schema.map((field) => (
            <FormField
              key={field.name}
              field={field}
              value={formData[field.name]}
              onChange={(value) => handleFieldChange(field.name, value)}
              error={formState.errors[field.name]}
              disabled={formState.loading}
            />
          ))}
        </div>
        <DialogFooter>
          {actions.map((action, idx) => {
            const isSubmit = action.action === 'submit';
            const isCancel = action.action === 'cancel';
            return (
              <Button
                key={idx}
                variant={action.variant || 'default'}
                onClick={() => handleAction(action)}
                disabled={formState.loading && !isCancel}
              >
                {formState.loading && isSubmit && (
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                )}
                {action.label}
              </Button>
            );
          })}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

export function UIDialog() {
  const { currentDialog } = useUIDialogStore();

  if (!currentDialog) return null;

  switch (currentDialog.dialogType) {
    case 'alert':
      return <AlertDialogComponent dialog={currentDialog} />;
    case 'confirm':
      return <ConfirmDialogComponent dialog={currentDialog} />;
    case 'prompt':
      return <PromptDialogComponent dialog={currentDialog} />;
    case 'form':
      return <FormDialogComponent dialog={currentDialog} />;
    default:
      return null;
  }
}
