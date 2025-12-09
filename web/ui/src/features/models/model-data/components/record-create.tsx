import { useState, useCallback } from 'react';
import { X, Plus } from 'lucide-react';
import { formatFieldLabel } from '@/lib/format';
import type { ModelField, FormConfig, Model } from '@/types';
import { Button } from '@/components/ui/button';
import { LoadingButton } from '@/components/ui/loading-button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import {
  Table,
  TableBody,
  TableCell,
  TableRow,
} from '@/components/ui/table';
import { FieldInput } from './field-input';

interface RecordCreateProps {
  orderedFormFields: ModelField[];
  formConfig: FormConfig;
  onSave: (formData: Record<string, unknown>, closeAfterSave: boolean) => void;
  onCancel: () => void;
  isSaving: boolean;
  fieldErrors?: Record<string, string>;
  projectId?: string;
  models?: Model[];
  onFormChange?: (hasChanges: boolean) => void;
}

export function RecordCreate({
  orderedFormFields,
  formConfig,
  onSave,
  onCancel,
  isSaving,
  fieldErrors,
  projectId,
  models,
  onFormChange,
}: RecordCreateProps) {
  const [formData, setFormData] = useState<Record<string, unknown>>({});

  const handleSave = useCallback((closeAfterSave: boolean) => {
    onSave(formData, closeAfterSave);
  }, [formData, onSave]);

  const handleFieldChange = useCallback((key: string, value: unknown) => {
    setFormData(prev => ({ ...prev, [key]: value }));
    onFormChange?.(true);
  }, [onFormChange]);

  return (
    <Card className="rounded-t-none !mt-0 h-full max-w-4xl">
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>New Record</CardTitle>
            <CardDescription>
              Fill in the fields to create a new record
            </CardDescription>
          </div>
          <div className="flex items-center gap-2">
            <Button variant="outline" onClick={onCancel} disabled={isSaving}>
              <X className="mr-2 size-4" />
              Cancel
            </Button>
            <LoadingButton onClick={() => handleSave(true)} loading={isSaving}>
              <Plus className="mr-2 size-4" />
              Create
            </LoadingButton>
          </div>
        </div>
      </CardHeader>
      <CardContent className="flex-1 overflow-y-auto max-h-[calc(100vh-360px)]">
        <div>
          <div className="rounded-md border overflow-hidden">
            <Table>
              <TableBody>
                {orderedFormFields.map((field) => {
                  const hasError = !!fieldErrors?.[field.key];
                  return (
                    <TableRow key={field.key}>
                      <TableCell className="w-1/3 font-medium text-muted-foreground bg-muted/30 align-top py-3">
                        {formatFieldLabel(field.key)}
                        {field.required && <span className="text-destructive ml-1">*</span>}
                      </TableCell>
                      <TableCell className="py-2">
                        <div className="space-y-1">
                          <FieldInput
                            field={field}
                            value={formData[field.key]}
                            onChange={(value) => handleFieldChange(field.key, value)}
                            view={formConfig.field_views[field.key]}
                            projectId={projectId}
                            models={models}
                          />
                          {hasError && (
                            <p className="text-sm text-destructive">{fieldErrors![field.key]}</p>
                          )}
                        </div>
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
