import { flexRender, getCoreRowModel, useReactTable, type ColumnDef } from "@tanstack/react-table";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { EmptyState } from "@/components/empty-state";

const pageSizeOptions = [10, 20, 50, 100];

export function DataTable<T>({
  columns,
  data,
  page,
  pageSize,
  totalPages,
  onPageChange,
  onPageSizeChange,
  emptyTitle = "No records",
  emptyDescription = "No data matches current filters.",
}: {
  columns: ColumnDef<T>[];
  data: T[];
  page?: number;
  pageSize?: number;
  totalPages?: number;
  onPageChange?: (page: number) => void;
  onPageSizeChange?: (pageSize: number) => void;
  emptyTitle?: string;
  emptyDescription?: string;
}) {
  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
    manualPagination: true,
  });
  const availablePageSizes = pageSize
    ? Array.from(new Set([...pageSizeOptions, pageSize])).sort((a, b) => a - b)
    : pageSizeOptions;

  return (
    <div>
      {!data.length ? (
        <EmptyState title={emptyTitle} description={emptyDescription} />
      ) : (
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((group) => (
              <TableRow key={group.id}>
                {group.headers.map((header) => (
                  <TableHead key={header.id}>
                    {header.isPlaceholder
                      ? null
                      : flexRender(header.column.columnDef.header, header.getContext())}
                  </TableHead>
                ))}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {table.getRowModel().rows.map((row) => (
              <TableRow key={row.id}>
                {row.getVisibleCells().map((cell) => (
                  <TableCell key={cell.id}>
                    {flexRender(cell.column.columnDef.cell, cell.getContext())}
                  </TableCell>
                ))}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}
      {page !== undefined && totalPages !== undefined && onPageChange && (
        <div className="flex flex-wrap items-center justify-between gap-3 border-t px-4 py-3">
          <p className="text-sm text-muted-foreground">
            Page {page} of {Math.max(totalPages, 1)}
          </p>
          <div className="flex flex-wrap items-center gap-3">
            {pageSize !== undefined && onPageSizeChange && (
              <div className="flex items-center gap-2">
                <span className="text-sm text-muted-foreground">Rows per page</span>
                <Select
                  value={String(pageSize)}
                  onValueChange={(value) => {
                    if (value !== null) onPageSizeChange(Number(value));
                  }}
                >
                  <SelectTrigger aria-label="Rows per page" size="sm">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent align="end">
                    {availablePageSizes.map((size) => (
                      <SelectItem key={size} value={String(size)}>
                        {size}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            )}
            <Button
              variant="outline"
              size="sm"
              disabled={page <= 1}
              onClick={() => onPageChange(page - 1)}
            >
              Previous
            </Button>
            <Button
              variant="outline"
              size="sm"
              disabled={page >= totalPages}
              onClick={() => onPageChange(page + 1)}
            >
              Next
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
