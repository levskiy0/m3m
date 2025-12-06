import { formatFieldLabel } from '@/lib/format';
import type { ModelField, FormConfig, Model } from '@/types';
import { Button } from '@/components/ui/button';
import { FieldInput } from './field-input';
import type { Tab } from '../types';

interface RecordFormProps {
  tab: Tab;
  orderedFormFields: ModelField[];
  formConfig: FormConfig;
  onFormDataChange: (tabId: string, formData: Record<string, unknown>) => void;
  onSave: () => void;
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
    <>
      <div className="overflow-y-auto max-h-[calc(100vh-360px)]">
        <div className="max-w-2xl space-y-4">
          {orderedFormFields.map((field) => (
            <div key={field.key} className="space-y-2">
              <label className="text-sm font-medium">
                {formatFieldLabel(field.key)}
                {field.required && <span className="text-destructive ml-1">*</span>}
              </label>
              <FieldInput
                field={field}
                value={tab.formData?.[field.key]}
                onChange={(value) => onFormDataChange(tab.id, { ...tab.formData, [field.key]: value })}
                view={formConfig.field_views[field.key]}
                projectId={projectId}
                models={models}
              />
              {tab.fieldErrors?.[field.key] && (
                <p className="text-sm text-destructive">{tab.fieldErrors[field.key]}</p>
              )}
            </div>
          ))}
        </div>
      </div>
      {/* Actions */}
      <div className="flex items-center gap-2 mt-4">
        <Button variant="outline" onClick={onCancel}>Cancel</Button>
        <Button onClick={onSave} disabled={isSaving}>
          {isSaving ? 'Saving...' : 'Save'}
        </Button>
      </div>
    </>
  );
}
