import { Columns3, Filter, ArrowUpDown, Search } from 'lucide-react';

import type { ModelField, TableConfig } from '@/types';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Checkbox } from '@/components/ui/checkbox';

// System fields that can be displayed in tables
const SYSTEM_FIELDS = [
  { key: '_created_at', label: 'Created At', type: 'datetime' },
  { key: '_updated_at', label: 'Updated At', type: 'datetime' },
] as const;

interface TableConfigEditorProps {
  fields: ModelField[];
  config: TableConfig;
  onChange: (config: TableConfig) => void;
}

export function TableConfigEditor({ fields, config, onChange }: TableConfigEditorProps) {
  const toggleColumn = (key: string) => {
    const columns = config.columns.includes(key)
      ? config.columns.filter((c) => c !== key)
      : [...config.columns, key];
    onChange({ ...config, columns });
  };

  const toggleFilter = (key: string) => {
    const filters = config.filters.includes(key)
      ? config.filters.filter((f) => f !== key)
      : [...config.filters, key];
    onChange({ ...config, filters });
  };

  const toggleSortable = (key: string) => {
    const sort_columns = config.sort_columns.includes(key)
      ? config.sort_columns.filter((s) => s !== key)
      : [...config.sort_columns, key];
    onChange({ ...config, sort_columns });
  };

  const toggleSearchable = (key: string) => {
    const searchable = (config.searchable || []).includes(key)
      ? (config.searchable || []).filter((s) => s !== key)
      : [...(config.searchable || []), key];
    onChange({ ...config, searchable });
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Table Configuration</CardTitle>
        <CardDescription>
          Configure how data is displayed in the table view
        </CardDescription>
      </CardHeader>
      <CardContent>
        {fields.length === 0 ? (
          <div className="text-center py-8 text-muted-foreground">
            No fields defined. Add fields in the Schema tab first.
          </div>
        ) : (
          <div className="border rounded-lg overflow-hidden">
            <table className="w-full text-sm">
              <thead className="bg-muted/50">
                <tr>
                  <th className="text-left font-medium p-3">Field</th>
                  <th className="text-center font-medium p-3 w-28">
                    <div className="flex items-center justify-center gap-1.5">
                      <Columns3 className="size-4" />
                      <span>Column</span>
                    </div>
                  </th>
                  <th className="text-center font-medium p-3 w-28">
                    <div className="flex items-center justify-center gap-1.5">
                      <Filter className="size-4" />
                      <span>Filter</span>
                    </div>
                  </th>
                  <th className="text-center font-medium p-3 w-28">
                    <div className="flex items-center justify-center gap-1.5">
                      <ArrowUpDown className="size-4" />
                      <span>Sort</span>
                    </div>
                  </th>
                  <th className="text-center font-medium p-3 w-28">
                    <div className="flex items-center justify-center gap-1.5">
                      <Search className="size-4" />
                      <span>Search</span>
                    </div>
                  </th>
                </tr>
              </thead>
              <tbody>
                {fields.map((field) => (
                  <tr key={field.key} className="border-t">
                    <td className="p-3">
                      <div className="flex flex-col">
                        <span className="font-mono text-sm">{field.key}</span>
                        <span className="text-xs text-muted-foreground">{field.type}</span>
                      </div>
                    </td>
                    <td className="p-3 text-center">
                      <div className="flex items-center justify-center">
                        <Checkbox
                          id={`col-${field.key}`}
                          checked={config.columns.includes(field.key)}
                          onCheckedChange={() => toggleColumn(field.key)}
                        />
                      </div>
                    </td>
                    <td className="p-3 text-center">
                      <div className="flex items-center justify-center">
                        <Checkbox
                          id={`filter-${field.key}`}
                          checked={config.filters.includes(field.key)}
                          onCheckedChange={() => toggleFilter(field.key)}
                        />
                      </div>
                    </td>
                    <td className="p-3 text-center">
                      <div className="flex items-center justify-center">
                        <Checkbox
                          id={`sort-${field.key}`}
                          checked={config.sort_columns.includes(field.key)}
                          onCheckedChange={() => toggleSortable(field.key)}
                        />
                      </div>
                    </td>
                    <td className="p-3 text-center">
                      <div className="flex items-center justify-center">
                        <Checkbox
                          id={`search-${field.key}`}
                          checked={(config.searchable || []).includes(field.key)}
                          onCheckedChange={() => toggleSearchable(field.key)}
                          disabled={!['string', 'text'].includes(field.type)}
                        />
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        <div className="mt-4 text-xs text-muted-foreground space-y-1">
          <p><strong>Column:</strong> Show this field as a column in the table</p>
          <p><strong>Filter:</strong> Allow filtering by this field</p>
          <p><strong>Sort:</strong> Allow sorting by this field</p>
          <p><strong>Search:</strong> Include in full-text search (only string/text fields)</p>
        </div>
      </CardContent>
    </Card>
  );
}
