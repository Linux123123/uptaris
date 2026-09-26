import { DeleteResourceDialog } from "@/components/delete-resource-dialog";
import { createFileRoute, redirect } from "@tanstack/react-router";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import type { ColumnDef } from "@tanstack/react-table";
import { Trash2 } from "lucide-react";
import { toast } from "sonner";
import { z } from "zod";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { DataTable } from "@/components/data-table";
import { PageHeader } from "@/components/page-header";
import { ErrorState } from "@/components/error-state";
import { LoadingState } from "@/components/loading-state";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { adminApi, type Role, type User } from "@/lib/api";
import { authStore } from "@/lib/auth-store";
import { usersQuery } from "@/lib/queries";

const searchSchema = z.object({
  page: z.coerce.number().int().positive().catch(1),
  pageSize: z.coerce.number().int().min(10).max(100).catch(20),
});
export const Route = createFileRoute("/app/admin/users")({
  validateSearch: searchSchema,
  beforeLoad: ({ context }) => {
    if (context.auth.user?.role !== "admin") throw redirect({ to: "/app/dashboard" });
  },
  loaderDeps: ({ search }) => search,
  loader: ({ context, deps }) => context.queryClient.ensureQueryData(usersQuery(deps)),
  component: UsersPage,
});

function UsersPage() {
  const search = Route.useSearch();
  const navigate = Route.useNavigate();
  const queryClient = useQueryClient();
  const query = useQuery(usersQuery(search));
  const currentUserId = authStore.state.user?.id;
  const remove = useMutation({
    mutationFn: (id: number) => adminApi.removeUser(id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["users"] });
      toast.success("User deleted and sessions revoked");
    },
    onError: (error) => toast.error(error.message),
  });
  const updateRole = useMutation({
    mutationFn: ({ id, role }: { id: number; role: Role }) => adminApi.updateRole(id, role),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["users"] });
      toast.success("User role updated; active sessions revoked");
    },
    onError: (error) => toast.error(error.message),
  });
  const columns: ColumnDef<User>[] = [
    {
      accessorKey: "email",
      header: "User",
      cell: ({ row }) => (
        <div>
          <p className="font-medium">{row.original.email}</p>
          <p className="text-xs text-muted-foreground">ID #{row.original.id}</p>
        </div>
      ),
    },
    {
      accessorKey: "role",
      header: "Role",
      cell: ({ row }) =>
        row.original.id === currentUserId ? (
          <Badge variant="outline" className="capitalize">
            {row.original.role}
          </Badge>
        ) : (
          <Select
            value={row.original.role}
            disabled={updateRole.isPending}
            onValueChange={(role) =>
              role && updateRole.mutate({ id: row.original.id, role: role as Role })
            }
          >
            <SelectTrigger
              className="w-32 capitalize"
              aria-label={`Role for ${row.original.email}`}
            >
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="viewer">Viewer</SelectItem>
              <SelectItem value="operator">Operator</SelectItem>
              <SelectItem value="admin">Admin</SelectItem>
            </SelectContent>
          </Select>
        ),
    },
    {
      accessorKey: "createdAt",
      header: "Created",
      cell: ({ row }) => new Date(row.original.createdAt).toLocaleDateString(),
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) =>
        row.original.id === currentUserId ? (
          <span className="text-xs text-muted-foreground">Current user</span>
        ) : (
          <DeleteResourceDialog
            trigger={
              <Button variant="ghost" size="icon-sm" aria-label={`Delete ${row.original.email}`}>
                <Trash2 />
              </Button>
            }
            title={<>Delete {row.original.email}?</>}
            description={
              <>
                User access and active sessions will be revoked. Resource ownership remains subject
                to API rules.
              </>
            }
            actionLabel="Delete user"
            onConfirm={() => remove.mutateAsync(row.original.id)}
          />
        ),
    },
  ];

  return (
    <div className="space-y-6">
      <PageHeader
        title="Users"
        description="Manage console access. Deleting user immediately revokes active sessions."
      />
      <Card>
        <CardContent className="p-1">
          {query.isPending ? (
            <LoadingState label="Loading users" />
          ) : query.isError ? (
            <ErrorState message={query.error.message} retry={() => void query.refetch()} />
          ) : (
            <DataTable
              columns={columns}
              data={query.data.data}
              page={query.data.pagination.page}
              totalPages={query.data.pagination.totalPages}
              onPageChange={(page) => void navigate({ search: { ...search, page } })}
              emptyTitle="No users"
              emptyDescription="No active user accounts found."
            />
          )}
        </CardContent>
      </Card>
    </div>
  );
}
