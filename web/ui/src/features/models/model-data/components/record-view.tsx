import { Edit, Trash2 } from 'lucide-react';
import { formatFieldLabel } from '@/lib/format';
import type { ModelField, ModelData } from '@/types';
import { Button } from '@/components/ui/button';
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

interface RecordViewProps {
  data: ModelData;
  orderedFormFields: ModelField[];
  onEdit: (data: ModelData) => void;
  onDelete: (data: ModelData) => void;
}

export function RecordView({
  data,
  orderedFormFields,
  onEdit,
  onDelete,
}: RecordViewProps) {
  return (
    <Card className="rounded-t-none !mt-0 h-full max-w-4xl">
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>View Record</CardTitle>
            <CardDescription>
              Record details and field values
            </CardDescription>
          </div>
          <div className="flex items-center gap-2">
            <Button variant="outline" onClick={() => onEdit(data)}>
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
          </div>
        </div>
      </CardHeader>
      <CardContent className="flex-1 overflow-y-auto max-h-[calc(100vh-360px)]">
        <div>
          <div className="rounded-md border overflow-hidden">
            <Table>
              <TableBody>
                {/* Regular fields */}
                {orderedFormFields.map((field) => (
                  <TableRow key={field.key}>
                    <TableCell className="w-1/3 font-medium text-muted-foreground bg-muted/30">
                      {formatFieldLabel(field.key)}
                    </TableCell>
                    <TableCell>
                      {field.type === 'document' ? (
                        <pre className="text-md bg-muted p-2 rounded overflow-auto max-h-32">
                          {JSON.stringify(data?.[field.key], null, 2)}
                        </pre>
                      ) : (
                        <span className="font-mono text-md">
                          {formatCellValue(data?.[field.key], field.type) || '—'}
                        </span>
                      )}
                    </TableCell>
                  </TableRow>
                ))}
                {/* System fields */}
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
