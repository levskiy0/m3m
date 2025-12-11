import { useState, useEffect } from 'react';
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
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { useUIDialogStore, type UIDialogRequest } from '@/stores/ui-dialog-store';
import type { UIFormField, UIFormAction } from '@/lib/websocket';

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
  const severity = dialog.options.severity || 'info';
  const Icon = severityIcons[severity];

  return (
    <AlertDialog open onOpenChange={() => closeDialog()}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle className="flex items-center gap-2">
            <Icon className={cn('h-5 w-5', severityColors[severity])} />
            {dialog.options.title || 'Alert'}
          </AlertDialogTitle>
          {dialog.options.text && (
            <AlertDialogDescription>{dialog.options.text}</AlertDialogDescription>
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

  return (
    <AlertDialog open onOpenChange={() => respondToDialog(false)}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{dialog.options.title || 'Confirm'}</AlertDialogTitle>
          {dialog.options.text && (
            <AlertDialogDescription>{dialog.options.text}</AlertDialogDescription>
          )}
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel onClick={() => respondToDialog(false)}>
            {dialog.options.no || 'No'}
          </AlertDialogCancel>
          <AlertDialogAction onClick={() => respondToDialog(true)}>
            {dialog.options.yes || 'Yes'}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}

function PromptDialogComponent({ dialog }: { dialog: UIDialogRequest }) {
  const { respondToDialog } = useUIDialogStore();
  const [value, setValue] = useState(dialog.options.defaultValue || '');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    respondToDialog(value);
  };

  return (
    <Dialog open onOpenChange={() => respondToDialog(null)}>
      <DialogContent>
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>{dialog.options.title || 'Enter Value'}</DialogTitle>
            {dialog.options.text && (
              <DialogDescription>{dialog.options.text}</DialogDescription>
            )}
          </DialogHeader>
          <div className="py-4">
            <Input
              value={value}
              onChange={(e) => setValue(e.target.value)}
              placeholder={dialog.options.placeholder}
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

function FormField({
  field,
  value,
  onChange,
  error,
}: {
  field: UIFormField;
  value: unknown;
  onChange: (value: unknown) => void;
  error?: string;
}) {
  const colSpan =
    field.colspan === 'full'
      ? 'col-span-6'
      : `col-span-${field.colspan || 6}`;

  const renderField = () => {
    switch (field.type) {
      case 'input':
        return (
          <Input
            value={(value as string) || ''}
            onChange={(e) => onChange(e.target.value)}
            placeholder={field.placeholder}
            required={field.required}
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
          />
        );

      case 'checkbox':
        return (
          <div className="flex items-center space-x-2">
            <Checkbox
              id={field.name}
              checked={!!value}
              onCheckedChange={(checked) => onChange(checked)}
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
          />
        );

      case 'datetime':
        return (
          <Input
            type="datetime-local"
            value={(value as string) || ''}
            onChange={(e) => onChange(e.target.value)}
            required={field.required}
          />
        );

      default:
        return (
          <Input
            value={(value as string) || ''}
            onChange={(e) => onChange(e.target.value)}
            placeholder={field.placeholder}
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
  const { respondToDialog } = useUIDialogStore();
  const schema = dialog.options.schema || [];
  const actions = dialog.options.actions || [
    { label: 'Cancel', variant: 'outline' as const, action: 'cancel' },
    { label: 'Submit', variant: 'default' as const, action: 'submit' },
  ];

  const [formData, setFormData] = useState<Record<string, unknown>>(() => {
    const initial: Record<string, unknown> = {};
    for (const field of schema) {
      initial[field.name] = field.defaultValue ?? '';
    }
    return initial;
  });

  const [errors, setErrors] = useState<Record<string, string>>({});

  const handleFieldChange = (name: string, value: unknown) => {
    setFormData((prev) => ({ ...prev, [name]: value }));
    // Clear error when field changes
    if (errors[name]) {
      setErrors((prev) => {
        const next = { ...prev };
        delete next[name];
        return next;
      });
    }
  };

  const handleAction = (action: UIFormAction) => {
    if (action.action === 'cancel') {
      respondToDialog(null);
      return;
    }

    // For submit or custom actions, send the form data
    respondToDialog({
      action: action.action,
      data: formData,
    });
  };

  // Handle validation errors from backend
  useEffect(() => {
    // This would be called if the backend returns validation errors
    // For now, validation is handled client-side
  }, []);

  return (
    <Dialog open onOpenChange={() => respondToDialog(null)}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>{dialog.options.title || 'Form'}</DialogTitle>
          {dialog.options.text && (
            <DialogDescription>{dialog.options.text}</DialogDescription>
          )}
        </DialogHeader>
        <div className="grid grid-cols-6 gap-4 py-4">
          {schema.map((field) => (
            <FormField
              key={field.name}
              field={field}
              value={formData[field.name]}
              onChange={(value) => handleFieldChange(field.name, value)}
              error={errors[field.name]}
            />
          ))}
        </div>
        <DialogFooter>
          {actions.map((action, idx) => (
            <Button
              key={idx}
              variant={action.variant || 'default'}
              onClick={() => handleAction(action)}
            >
              {action.label}
            </Button>
          ))}
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
