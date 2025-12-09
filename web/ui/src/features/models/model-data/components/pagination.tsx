import { Trash2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';

interface PaginationProps {
  page: number;
  totalPages: number;
  total: number;
  limit: number;
  selectedCount: number;
  onPageChange: (page: number) => void;
  onLimitChange: (limit: number) => void;
  onBulkDelete?: () => void;
}

export function Pagination({
  page,
  totalPages,
  total,
  limit,
  selectedCount,
  onPageChange,
  onLimitChange,
  onBulkDelete,
}: PaginationProps) {
  return (
    <div className="flex items-center justify-between py-4">
      <div className="flex items-center gap-2">
        <span className="text-sm text-muted-foreground">
          {selectedCount > 0 ? `${selectedCount} of ${total} selected` : `${total} rows`}
        </span>
        {selectedCount > 0 && onBulkDelete && (
          <Button
            variant="destructive"
            size="sm"
            onClick={onBulkDelete}
          >
            <Trash2 className="mr-2 size-4" />
            Delete Selected
          </Button>
        )}
      </div>
      <div className="flex items-center gap-4">
        <Select
          value={limit.toString()}
          onValueChange={(v) => {
            onLimitChange(parseInt(v));
            onPageChange(1);
          }}
        >
          <SelectTrigger className="w-[100px]">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="25">25</SelectItem>
            <SelectItem value="50">50</SelectItem>
            <SelectItem value="100">100</SelectItem>
          </SelectContent>
        </Select>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => onPageChange(page - 1)}
            disabled={page === 1}
          >
            Previous
          </Button>
          <span className="text-sm">{page} / {totalPages || 1}</span>
          <Button
            variant="outline"
            size="sm"
            onClick={() => onPageChange(page + 1)}
            disabled={page >= totalPages}
          >
            Next
          </Button>
        </div>
      </div>
    </div>
  );
}
