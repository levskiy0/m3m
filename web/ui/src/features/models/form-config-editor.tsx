import { GripVertical, Eye, EyeOff, ChevronUp, ChevronDown } from 'lucide-react';

import type { ModelField, FormConfig, FieldView } from '@/types';
import { FIELD_VIEWS } from '@/lib/constants';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';

interface FormConfigEditorProps {
  fields: ModelField[];
  config: FormConfig;
  onChange: (config: FormConfig) => void;
}

export function FormConfigEditor({ fields, config, onChange }: FormConfigEditorProps) {
  // Get ordered fields list based on config.field_order
  const getOrderedFields = () => {
    const fieldMap = new Map(fields.map((f) => [f.key, f]));
    const ordered: ModelField[] = [];

    // First add fields in specified order
    for (const key of config.field_order) {
      const field = fieldMap.get(key);
      if (field) {
        ordered.push(field);
        fieldMap.delete(key);
      }
    }

    // Then add remaining fields not in the order
    for (const field of fieldMap.values()) {
      ordered.push(field);
    }

    return ordered;
  };

  const orderedFields = getOrderedFields();

  const toggleHidden = (key: string) => {
    const hidden_fields = config.hidden_fields.includes(key)
      ? config.hidden_fields.filter((k) => k !== key)
      : [...config.hidden_fields, key];
    onChange({ ...config, hidden_fields });
  };

  const setFieldView = (key: string, view: FieldView | '') => {
    const field_views = { ...config.field_views };
    if (view === '') {
      delete field_views[key];
    } else {
      field_views[key] = view;
    }
    onChange({ ...config, field_views });
  };

  const moveField = (key: string, direction: 'up' | 'down') => {
    const currentOrder = orderedFields.map((f) => f.key);
    const index = currentOrder.indexOf(key);
    if (index === -1) return;

    const newIndex = direction === 'up' ? index - 1 : index + 1;
    if (newIndex < 0 || newIndex >= currentOrder.length) return;

    // Swap
    [currentOrder[index], currentOrder[newIndex]] = [currentOrder[newIndex], currentOrder[index]];
    onChange({ ...config, field_order: currentOrder });
  };

  const getDefaultView = (field: ModelField): string => {
    const views = FIELD_VIEWS[field.type];
    return views?.[0]?.value || 'input';
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Form Configuration</CardTitle>
        <CardDescription>
          Configure the order, visibility, and widget types for form fields
        </CardDescription>
      </CardHeader>
      <CardContent>
        {fields.length === 0 ? (
          <div className="text-center py-8 text-muted-foreground">
            No fields defined. Add fields in the Schema tab first.
          </div>
        ) : (
          <div className="space-y-2">
            {orderedFields.map((field, index) => {
              const isHidden = config.hidden_fields.includes(field.key);
              const currentView = config.field_views[field.key] || getDefaultView(field);
              const availableViews = FIELD_VIEWS[field.type] || [];

              return (
                <div
                  key={field.key}
                  className={`flex items-center gap-3 p-3 border rounded-lg transition-colors ${
                    isHidden ? 'bg-muted/30 opacity-60' : 'bg-background'
                  }`}
                >
                  <div className="flex items-center gap-1">
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-7 w-7"
                      disabled={index === 0}
                      onClick={() => moveField(field.key, 'up')}
                    >
                      <ChevronUp className="size-4" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-7 w-7"
                      disabled={index === orderedFields.length - 1}
                      onClick={() => moveField(field.key, 'down')}
                    >
                      <ChevronDown className="size-4" />
                    </Button>
                  </div>

                  <div className="text-muted-foreground">
                    <GripVertical className="size-5" />
                  </div>

                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="font-mono text-sm font-medium">{field.key}</span>
                      <Badge variant="outline" className="text-xs">
                        {field.type}
                      </Badge>
                      {field.required && (
                        <Badge variant="secondary" className="text-xs">
                          required
                        </Badge>
                      )}
                    </div>
                  </div>

                  <div className="flex items-center gap-3">
                    <Select
                      value={currentView}
                      onValueChange={(v) => setFieldView(field.key, v as FieldView)}
                      disabled={availableViews.length <= 1}
                    >
                      <SelectTrigger className="w-40 h-8">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {availableViews.map((view) => (
                          <SelectItem key={view.value} value={view.value}>
                            {view.label}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>

                    <Button
                      variant={isHidden ? 'secondary' : 'ghost'}
                      size="icon"
                      className="h-8 w-8"
                      onClick={() => toggleHidden(field.key)}
                      title={isHidden ? 'Show in form' : 'Hide from form'}
                    >
                      {isHidden ? (
                        <EyeOff className="size-4" />
                      ) : (
                        <Eye className="size-4" />
                      )}
                    </Button>
                  </div>
                </div>
              );
            })}
          </div>
        )}

        <div className="mt-4 text-xs text-muted-foreground space-y-1">
          <p><strong>Order:</strong> Use arrows to reorder fields in the form</p>
          <p><strong>Widget:</strong> Choose how the field is displayed in forms</p>
          <p><strong>Visibility:</strong> Click the eye icon to show/hide fields from forms</p>
        </div>
      </CardContent>
    </Card>
  );
}
