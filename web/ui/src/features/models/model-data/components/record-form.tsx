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
  FieldGroup,
  Field,
  FieldLabel,
  FieldDescription,
  FieldError,
} from '@/components/ui/field';
import { FieldInput } from './field-input';
import type { Tab } from '../types';

interface RecordFormProps {
  tab: Tab;
  orderedFormFields: ModelField[];
  formConfig: FormConfig;
  onFormDataChange: (tabId: string, formData: Record<string, unknown>) => void;
  onSave: (closeAfterSave: boolean) => void;
  onCancel: () => void;
  isSaving: boolean;
  projectId?: string;
  models?: Model[];
}

export function RecordForm({
  tab,
  orderedFormFields,
  formConfig,
  onFormDataChange,
  onSave,
  onCancel,
  isSaving,
  projectId,
  models,
}: RecordFormProps) {
  return (
    <Card className="rounded-t-none !mt-0 flex flex-col h-full">
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>{tab.type === 'create' ? 'New Record' : 'Edit Record'}</CardTitle>
            <CardDescription>
              {tab.type === 'create' ? 'Fill in the fields to create a new record' : 'Modify the fields and save changes'}
            </CardDescription>
          </div>
          <div className="flex items-center gap-2">
            <Button variant="outline" onClick={onCancel}>
              Cancel
            </Button>
            {tab.type === 'edit' && (
              <LoadingButton variant="outline" onClick={() => onSave(false)} loading={isSaving}>
                Save
              </LoadingButton>
            )}
            <LoadingButton onClick={() => onSave(true)} loading={isSaving}>
              {tab.type === 'edit' ? 'Save & Close' : 'Create'}
            </LoadingButton>
          </div>
        </div>
      </CardHeader>
      <CardContent className="flex-1 overflow-y-auto max-h-[calc(100vh-360px)]">
        <FieldGroup className="max-w-2xl space-y-2">
          {orderedFormFields.map((field) => {
            const hasError = !!tab.fieldErrors?.[field.key];
            return (
              <Field key={field.key} data-invalid={hasError || undefined}>
                <FieldLabel htmlFor={field.key}>
                  {formatFieldLabel(field.key)}
                  {field.required && <span className="text-destructive ml-1">*</span>}
                </FieldLabel>
                <FieldInput
                  field={field}
                  value={tab.formData?.[field.key]}
                  onChange={(value) => onFormDataChange(tab.id, { ...tab.formData, [field.key]: value })}
                  view={formConfig.field_views[field.key]}
                  projectId={projectId}
                  models={models}
                />
                {field.description && (
                  <FieldDescription>{field.description}</FieldDescription>
                )}
                {hasError && (
                  <FieldError>{tab.fieldErrors![field.key]}</FieldError>
                )}
              </Field>
            );
          })}
        </FieldGroup>
      </CardContent>
    </Card>
  );
}
