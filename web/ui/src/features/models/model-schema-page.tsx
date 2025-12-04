import { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Plus, Trash2, GripVertical, Save, Table } from 'lucide-react';
import { toast } from 'sonner';

import { modelsApi } from '@/api';
import { FIELD_TYPES } from '@/lib/constants';
import type { ModelField, FieldType } from '@/types';
import { Button } from '@/components/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Input } from '@/components/ui/input';
import { Switch } from '@/components/ui/switch';
import { Field, FieldLabel } from '@/components/ui/field';
import { Skeleton } from '@/components/ui/skeleton';
import { Badge } from '@/components/ui/badge';

export function ModelSchemaPage() {
  const { projectId, modelId } = useParams<{ projectId: string; modelId: string }>();
  const queryClient = useQueryClient();

  const [fields, setFields] = useState<ModelField[]>([]);
  const [hasChanges, setHasChanges] = useState(false);

  const { data: model, isLoading } = useQuery({
    queryKey: ['model', projectId, modelId],
    queryFn: () => modelsApi.get(projectId!, modelId!),
    enabled: !!projectId && !!modelId,
  });

  useEffect(() => {
    if (model) {
      setFields(model.fields);
      setHasChanges(false);
    }
  }, [model]);

  const updateMutation = useMutation({
    mutationFn: () => modelsApi.update(projectId!, modelId!, { fields }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['model', projectId, modelId] });
      queryClient.invalidateQueries({ queryKey: ['models', projectId] });
      setHasChanges(false);
      toast.success('Schema saved');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to save schema');
    },
  });

  const addField = () => {
    const newField: ModelField = {
      key: `field_${fields.length + 1}`,
      type: 'string',
      required: false,
    };
    setFields([...fields, newField]);
    setHasChanges(true);
  };

  const updateField = (index: number, updates: Partial<ModelField>) => {
    const newFields = [...fields];
    newFields[index] = { ...newFields[index], ...updates };
    setFields(newFields);
    setHasChanges(true);
  };

  const removeField = (index: number) => {
    setFields(fields.filter((_, i) => i !== index));
    setHasChanges(true);
  };

  if (isLoading) {
    return (
      <div className="space-y-4 max-w-3xl">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-64" />
      </div>
    );
  }

  if (!model) {
    return <div>Model not found</div>;
  }

  return (
    <div className="space-y-4 max-w-3xl">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">{model.name}</h1>
          <p className="text-muted-foreground">
            Define the schema for this model
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button asChild variant="outline">
            <Link to={`/projects/${projectId}/models/${modelId}/data`}>
              <Table className="mr-2 size-4" />
              View Data
            </Link>
          </Button>
          <Button
            onClick={() => updateMutation.mutate()}
            disabled={!hasChanges || updateMutation.isPending}
          >
            <Save className="mr-2 size-4" />
            {updateMutation.isPending ? 'Saving...' : 'Save'}
          </Button>
        </div>
      </div>

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Fields</CardTitle>
              <CardDescription>
                Define the fields for your model schema
              </CardDescription>
            </div>
            <Button variant="outline" size="sm" onClick={addField}>
              <Plus className="mr-2 size-4" />
              Add Field
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {fields.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">
                No fields defined. Add a field to get started.
              </div>
            ) : (
              fields.map((field, index) => (
                <div
                  key={index}
                  className="flex items-start gap-4 p-4 border rounded-lg bg-muted/30"
                >
                  <div className="cursor-move text-muted-foreground mt-2">
                    <GripVertical className="size-5" />
                  </div>
                  <div className="flex-1 grid gap-4 md:grid-cols-4">
                    <Field>
                      <FieldLabel>Key</FieldLabel>
                      <Input
                        value={field.key}
                        onChange={(e) =>
                          updateField(index, {
                            key: e.target.value
                              .toLowerCase()
                              .replace(/[^a-z0-9_]/g, '_'),
                          })
                        }
                        placeholder="field_name"
                      />
                    </Field>
                    <Field>
                      <FieldLabel>Type</FieldLabel>
                      <Select
                        value={field.type}
                        onValueChange={(v) =>
                          updateField(index, { type: v as FieldType })
                        }
                      >
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          {FIELD_TYPES.map((type) => (
                            <SelectItem key={type.value} value={type.value}>
                              {type.label}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </Field>
                    <Field>
                      <FieldLabel>Default Value</FieldLabel>
                      <Input
                        value={field.defaultValue?.toString() || ''}
                        onChange={(e) =>
                          updateField(index, {
                            defaultValue: e.target.value || undefined,
                          })
                        }
                        placeholder="(optional)"
                      />
                    </Field>
                    <Field>
                      <FieldLabel>Required</FieldLabel>
                      <div className="flex items-center h-10">
                        <Switch
                          checked={field.required}
                          onCheckedChange={(checked) =>
                            updateField(index, { required: checked })
                          }
                        />
                      </div>
                    </Field>
                  </div>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="mt-6"
                    onClick={() => removeField(index)}
                  >
                    <Trash2 className="size-4 text-destructive" />
                  </Button>
                </div>
              ))
            )}
          </div>
        </CardContent>
      </Card>

      {/* Field Types Reference */}
      <Card>
        <CardHeader>
          <CardTitle className="text-sm">Field Types Reference</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-2 text-sm">
            {FIELD_TYPES.map((type) => (
              <div key={type.value} className="flex items-center gap-2">
                <Badge variant="outline" className="font-mono">
                  {type.value}
                </Badge>
                <span className="text-muted-foreground">{type.label}</span>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
