import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Plus,
  Users,
  Trash2,
  Edit,
  Ban,
  MoreHorizontal,
  ShieldOff,
} from 'lucide-react';
import { toast } from 'sonner';

import { usersApi, projectsApi } from '@/api';
import { queryKeys } from '@/lib/query-keys';
import { getInitials, toggleInArray } from '@/lib/utils';
import { useAuth } from '@/providers/auth-provider';
import { useFormDialog, useDeleteDialog, useTitle } from '@/hooks';
import type { User, CreateUserRequest, UpdateUserRequest } from '@/types';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { LoadingButton } from '@/components/ui/loading-button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Input } from '@/components/ui/input';
import { Switch } from '@/components/ui/switch';
import { Checkbox } from '@/components/ui/checkbox';
import { Field, FieldGroup, FieldLabel, FieldDescription } from '@/components/ui/field';
import { ConfirmDialog } from '@/components/shared/confirm-dialog';
import { EmptyState } from '@/components/shared/empty-state';
import { PageHeader } from '@/components/shared/page-header';
import { Skeleton } from '@/components/ui/skeleton';
import { Badge } from '@/components/ui/badge';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';

export function UsersPage() {
  useTitle('Users');
  const { user: currentUser } = useAuth();
  const queryClient = useQueryClient();

  const formDialog = useFormDialog<User>();
  const deleteDialog = useDeleteDialog<User>();

  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [name, setName] = useState('');
  const [createProjects, setCreateProjects] = useState(false);
  const [manageUsers, setManageUsers] = useState(false);
  const [projectAccess, setProjectAccess] = useState<string[]>([]);

  const { data: users = [], isLoading: usersLoading } = useQuery({
    queryKey: queryKeys.users.all,
    queryFn: usersApi.list,
  });

  const { data: projects = [] } = useQuery({
    queryKey: queryKeys.projects.all,
    queryFn: projectsApi.list,
  });

  const createMutation = useMutation({
    mutationFn: (data: CreateUserRequest) => usersApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.users.all });
      formDialog.close();
      resetForm();
      toast.success('User created');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to create user');
    },
  });

  const updateMutation = useMutation({
    mutationFn: (data: UpdateUserRequest) =>
      usersApi.update(formDialog.selectedItem!.id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.users.all });
      formDialog.close();
      resetForm();
      toast.success('User updated');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to update user');
    },
  });

  const deleteMutation = useMutation({
    mutationFn: () => usersApi.delete(deleteDialog.itemToDelete!.id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.users.all });
      deleteDialog.close();
      toast.success('User deleted');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to delete user');
    },
  });

  const blockMutation = useMutation({
    mutationFn: (userId: string) => usersApi.block(userId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.users.all });
      toast.success('User blocked');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to block user');
    },
  });

  const unblockMutation = useMutation({
    mutationFn: (userId: string) => usersApi.unblock(userId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.users.all });
      toast.success('User unblocked');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to unblock user');
    },
  });

  const resetForm = () => {
    setEmail('');
    setPassword('');
    setName('');
    setCreateProjects(false);
    setManageUsers(false);
    setProjectAccess([]);
  };

  const handleCreate = () => {
    createMutation.mutate({
      email,
      password,
      name,
      permissions: { create_projects: createProjects, manage_users: manageUsers, project_access: projectAccess },
    });
  };

  const handleEdit = (user: User) => {
    setName(user.name);
    setCreateProjects(user.permissions.create_projects);
    setManageUsers(user.permissions.manage_users);
    setProjectAccess(user.permissions.project_access || []);
    formDialog.openEdit(user);
  };

  const handleUpdate = () => {
    updateMutation.mutate({
      name,
      permissions: { create_projects: createProjects, manage_users: manageUsers, project_access: projectAccess },
    });
  };

  if (usersLoading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-64" />
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <PageHeader
        title="Users"
        description="Manage system users and permissions"
        action={
          <Button onClick={() => { resetForm(); formDialog.open(); }}>
            <Plus className="mr-2 size-4" />
            Add User
          </Button>
        }
      />

      {users.length === 0 ? (
        <EmptyState
          icon={<Users className="size-12" />}
          title="No users"
          description="Add users to grant access to the system"
          action={
            <Button onClick={() => { resetForm(); formDialog.open(); }}>
              <Plus className="mr-2 size-4" />
              Add User
            </Button>
          }
        />
      ) : (
        <Card>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>User</TableHead>
                  <TableHead>Email</TableHead>
                  <TableHead>Permissions</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead className="w-12"></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {users.map((user) => (
                  <TableRow key={user.id}>
                    <TableCell>
                      <div className="flex items-center gap-3">
                        <Avatar className="size-8">
                          <AvatarImage src={user.avatar} />
                          <AvatarFallback>{getInitials(user.name)}</AvatarFallback>
                        </Avatar>
                        <div>
                          <p className="font-medium">{user.name}</p>
                          {user.is_root && (
                            <Badge variant="secondary" className="text-xs">
                              Root
                            </Badge>
                          )}
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>{user.email}</TableCell>
                    <TableCell>
                      <div className="flex gap-1 flex-wrap">
                        {user.permissions.create_projects && (
                          <Badge variant="outline">Create Projects</Badge>
                        )}
                        {user.permissions.manage_users && (
                          <Badge variant="outline">Manage Users</Badge>
                        )}
                        {user.permissions.project_access?.length > 0 && (
                          <Badge variant="outline">
                            {user.permissions.project_access.length} projects
                          </Badge>
                        )}
                      </div>
                    </TableCell>
                    <TableCell>
                      {user.is_blocked ? (
                        <Badge variant="destructive">Blocked</Badge>
                      ) : (
                        <Badge variant="default">Active</Badge>
                      )}
                    </TableCell>
                    <TableCell>
                      {!user.is_root && user.id !== currentUser?.id && (
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="icon" className="size-8">
                              <MoreHorizontal className="size-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuItem onClick={() => handleEdit(user)}>
                              <Edit className="mr-2 size-4" />
                              Edit
                            </DropdownMenuItem>
                            {user.is_blocked ? (
                              <DropdownMenuItem onClick={() => unblockMutation.mutate(user.id)}>
                                <ShieldOff className="mr-2 size-4" />
                                Unblock
                              </DropdownMenuItem>
                            ) : (
                              <DropdownMenuItem onClick={() => blockMutation.mutate(user.id)}>
                                <Ban className="mr-2 size-4" />
                                Block
                              </DropdownMenuItem>
                            )}
                            <DropdownMenuSeparator />
                            <DropdownMenuItem
                              className="text-destructive"
                              onClick={() => deleteDialog.open(user)}
                            >
                              <Trash2 className="mr-2 size-4" />
                              Delete
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      )}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}

      <Dialog open={formDialog.isOpen} onOpenChange={(open) => !open && formDialog.close()}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>{formDialog.mode === 'create' ? 'Add User' : 'Edit User'}</DialogTitle>
            {formDialog.mode === 'create' && (
              <DialogDescription>Create a new user account</DialogDescription>
            )}
          </DialogHeader>
          <FieldGroup>
            {formDialog.mode === 'create' ? (
              <>
                <div className="grid grid-cols-2 gap-4">
                  <Field>
                    <FieldLabel>Name</FieldLabel>
                    <Input
                      value={name}
                      onChange={(e) => setName(e.target.value)}
                      placeholder="John Doe"
                    />
                  </Field>
                  <Field>
                    <FieldLabel>Email</FieldLabel>
                    <Input
                      type="email"
                      value={email}
                      onChange={(e) => setEmail(e.target.value)}
                      placeholder="john@example.com"
                    />
                  </Field>
                </div>
                <Field>
                  <FieldLabel>Password</FieldLabel>
                  <Input
                    type="password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    placeholder="••••••••"
                  />
                </Field>
              </>
            ) : (
              <>
                <Field>
                  <FieldLabel>Name</FieldLabel>
                  <Input value={name} onChange={(e) => setName(e.target.value)} />
                </Field>
                <Field>
                  <FieldLabel>Email</FieldLabel>
                  <Input value={formDialog.selectedItem?.email} disabled />
                </Field>
              </>
            )}
            <Field>
              <FieldLabel>Permissions</FieldLabel>
              <div className="space-y-3 mt-2">
                <div className="flex items-center gap-2">
                  <Switch checked={createProjects} onCheckedChange={setCreateProjects} />
                  <span className="text-sm">Can create projects</span>
                </div>
                <div className="flex items-center gap-2">
                  <Switch checked={manageUsers} onCheckedChange={setManageUsers} />
                  <span className="text-sm">Can manage users</span>
                </div>
              </div>
            </Field>
            <Field>
              <FieldLabel>Project Access</FieldLabel>
              <FieldDescription>Select which projects this user can access</FieldDescription>
              <div className="mt-2 space-y-2 max-h-48 overflow-y-auto border rounded-md p-3">
                {projects.map((project) => (
                  <div key={project.id} className="flex items-center gap-2">
                    <Checkbox
                      id={`${formDialog.mode}-${project.id}`}
                      checked={projectAccess.includes(project.id)}
                      onCheckedChange={() => setProjectAccess(toggleInArray(projectAccess, project.id))}
                    />
                    <label
                      htmlFor={`${formDialog.mode}-${project.id}`}
                      className="text-sm flex items-center gap-2 cursor-pointer"
                    >
                      <span
                        className="size-2 rounded-full"
                        style={{ backgroundColor: project.color || '#6b7280' }}
                      />
                      {project.name}
                    </label>
                  </div>
                ))}
              </div>
            </Field>
          </FieldGroup>
          <DialogFooter>
            <Button variant="outline" onClick={() => formDialog.close()}>
              Cancel
            </Button>
            <LoadingButton
              onClick={formDialog.mode === 'create' ? handleCreate : handleUpdate}
              disabled={
                formDialog.mode === 'create'
                  ? !name || !email || !password
                  : false
              }
              loading={createMutation.isPending || updateMutation.isPending}
            >
              {formDialog.mode === 'create' ? 'Create' : 'Save'}
            </LoadingButton>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <ConfirmDialog
        open={deleteDialog.isOpen}
        onOpenChange={(open) => !open && deleteDialog.close()}
        title="Delete User"
        description={`Are you sure you want to delete "${deleteDialog.itemToDelete?.name}"?`}
        confirmLabel="Delete"
        variant="destructive"
        onConfirm={() => deleteMutation.mutate()}
        isLoading={deleteMutation.isPending}
      />
    </div>
  );
}
