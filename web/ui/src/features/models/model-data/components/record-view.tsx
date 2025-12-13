import { useState, useCallback, useEffect } from 'react';
import { Edit, Trash2, X, Save } from 'lucide-react';
import { formatFieldLabel } from '@/lib/format';
import type { ModelField, ModelData, FormConfig, Model } from '@/types';
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
import { SYSTEM_FIELDS, type SystemField } from '../constants';
import { formatCellValue, formatSystemFieldValue } from '../utils';
import { FieldInput } from './field-input';

interface RecordViewProps {
  data: ModelData;
  orderedFormFields: ModelField[];
  formConfig: FormConfig;
  onSave: (formData: Record<string, unknown>, closeAfterSave: boolean) => void;
  onDelete: (data: ModelData) => void;
  isSaving: boolean;
  fieldErrors?: Record<string, string>;
  projectId?: string;
  models?: Model[];
  onFormChange?: (hasChanges: boolean) => void;
}

export function RecordView({
  data,
  orderedFormFields,
  formConfig,
  onSave,
  onDelete,
  isSaving,
  fieldErrors,
  projectId,
  models,
  onFormChange,
}: RecordViewProps) {
  const [isEditing, setIsEditing] = useState(false);
  const [formData, setFormData] = useState<Record<string, unknown>>({});

  // Initialize form data from record data
  useEffect(() => {
    const initialData: Record<string, unknown> = {};
    for (const field of orderedFormFields) {
      initialData[field.key] = data[field.key];
    }
    // eslint-disable-next-line react-hooks/set-state-in-effect -- intentional: sync form data with external data prop
    setFormData(initialData);
  }, [data, orderedFormFields]);

  const handleStartEdit = useCallback(() => {
    setIsEditing(true);
  }, []);

  const handleCancel = useCallback(() => {
    // Reset form data to original
    const initialData: Record<string, unknown> = {};
    for (const field of orderedFormFields) {
      initialData[field.key] = data[field.key];
    }
    setFormData(initialData);
    setIsEditing(false);
    onFormChange?.(false);
  }, [data, orderedFormFields, onFormChange]);

  const handleSave = useCallback((closeAfterSave: boolean) => {
    onSave(formData, closeAfterSave);
    if (closeAfterSave) {
      setIsEditing(false);
    }
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
            <CardTitle>{isEditing ? 'Edit Record' : 'View Record'}</CardTitle>
            <CardDescription>
              {isEditing ? 'Modify the fields and save changes' : 'Record details and field values'}
            </CardDescription>
          </div>
          <div className="flex items-center gap-2">
            {isEditing ? (
              <>
                <Button variant="outline" onClick={handleCancel} disabled={isSaving}>
                  <X className="mr-2 size-4" />
                  Cancel
                </Button>
                <LoadingButton variant="outline" onClick={() => handleSave(false)} loading={isSaving}>
                  <Save className="mr-2 size-4" />
                  Save
                </LoadingButton>
                <LoadingButton onClick={() => handleSave(true)} loading={isSaving}>
                  Save & Close
                </LoadingButton>
              </>
            ) : (
              <>
                <Button variant="outline" onClick={handleStartEdit}>
                  <Edit className="mr-2 size-4" />
                  Edit
                </Button>
                <Button
                  variant="outline"
                  onClick={() => onDelete(data)}
                  className="text-destructive hover:text-destructive"
                >
                  <Trash2 className="mr-2 size-4" />
                  Delete
                </Button>
              </>
            )}
          </div>
        </div>
      </CardHeader>
      <CardContent className="flex-1 overflow-y-auto max-h-[calc(100vh-360px)]">
        <div>
          <div className="rounded-md border overflow-hidden">
            <Table className="table-fixed">
              <TableBody>
                {/* Regular fields */}
                {orderedFormFields.map((field) => {
                  const hasError = !!fieldErrors?.[field.key];
                  return (
                    <TableRow key={field.key}>
                      <TableCell className="w-1/3 font-medium text-muted-foreground bg-muted/30 align-top py-3">
                        {formatFieldLabel(field.key)}
                        {field.required && <span className="text-destructive ml-1">*</span>}
                      </TableCell>
                      <TableCell className="py-2 w-2/3 whitespace-normal">
                        {isEditing ? (
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
                        ) : (
                          <>
                            {field.type === 'document' ? (
                              <pre className="text-md bg-muted p-2 rounded overflow-auto max-h-48 max-w-full whitespace-pre-wrap break-all">
                                {JSON.stringify(data?.[field.key], null, 2)}
                              </pre>
                            ) : (
                              <span className="font-mono text-md">
                                {formatCellValue(data?.[field.key], field.type) || '—'}
                              </span>
                            )}
                          </>
                        )}
                      </TableCell>
                    </TableRow>
                  );
                })}
                {/* System fields - always read-only */}
                {SYSTEM_FIELDS.map((sf) => (
                  <TableRow key={sf.key}>
                    <TableCell className="w-1/3 font-medium text-md text-muted-foreground bg-muted/30">
                      {sf.label}
                    </TableCell>
                    <TableCell className="font-mono text-md text-muted-foreground">
                      {formatSystemFieldValue(sf.key as SystemField, data?.[sf.key])}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
