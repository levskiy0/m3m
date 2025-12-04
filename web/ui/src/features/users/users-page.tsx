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
import { useAuth } from '@/providers/auth-provider';
import type { User, CreateUserRequest, UpdateUserRequest } from '@/types';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
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
import { Skeleton } from '@/components/ui/skeleton';
import { Badge } from '@/components/ui/badge';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';

function getInitials(name: string): string {
  return name
    .split(' ')
    .map((n) => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2);
}

export function UsersPage() {
  const { user: currentUser } = useAuth();
  const queryClient = useQueryClient();

  const [createOpen, setCreateOpen] = useState(false);
  const [editOpen, setEditOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [selectedUser, setSelectedUser] = useState<User | null>(null);

  // Form state
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [name, setName] = useState('');
  const [createProjects, setCreateProjects] = useState(false);
  const [manageUsers, setManageUsers] = useState(false);
  const [projectAccess, setProjectAccess] = useState<string[]>([]);

  const { data: users = [], isLoading: usersLoading } = useQuery({
    queryKey: ['users'],
    queryFn: usersApi.list,
  });

  const { data: projects = [] } = useQuery({
    queryKey: ['projects'],
    queryFn: projectsApi.list,
  });

  const createMutation = useMutation({
    mutationFn: (data: CreateUserRequest) => usersApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      setCreateOpen(false);
      resetForm();
      toast.success('User created');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to create user');
    },
  });

  const updateMutation = useMutation({
    mutationFn: (data: UpdateUserRequest) =>
      usersApi.update(selectedUser!.id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      setEditOpen(false);
      setSelectedUser(null);
      resetForm();
      toast.success('User updated');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to update user');
    },
  });

  const deleteMutation = useMutation({
    mutationFn: () => usersApi.delete(selectedUser!.id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      setDeleteOpen(false);
      setSelectedUser(null);
      toast.success('User deleted');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to delete user');
    },
  });

  const blockMutation = useMutation({
    mutationFn: (userId: string) => usersApi.block(userId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      toast.success('User blocked');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to block user');
    },
  });

  const unblockMutation = useMutation({
    mutationFn: (userId: string) => usersApi.unblock(userId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
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
      permissions: {
        createProjects,
        manageUsers,
        projectAccess,
      },
    });
  };

  const handleEdit = (user: User) => {
    setSelectedUser(user);
    setName(user.name);
    setCreateProjects(user.permissions.createProjects);
    setManageUsers(user.permissions.manageUsers);
    setProjectAccess(user.permissions.projectAccess || []);
    setEditOpen(true);
  };

  const handleUpdate = () => {
    updateMutation.mutate({
      name,
      permissions: {
        createProjects,
        manageUsers,
        projectAccess,
      },
    });
  };

  const toggleProjectAccess = (projectId: string) => {
    setProjectAccess((prev) =>
      prev.includes(projectId)
        ? prev.filter((id) => id !== projectId)
        : [...prev, projectId]
    );
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
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Users</h1>
          <p className="text-muted-foreground">
            Manage system users and permissions
          </p>
        </div>
        <Button onClick={() => setCreateOpen(true)}>
          <Plus className="mr-2 size-4" />
          Add User
        </Button>
      </div>

      {users.length === 0 ? (
        <EmptyState
          icon={<Users className="size-12" />}
          title="No users"
          description="Add users to grant access to the system"
          action={
            <Button onClick={() => setCreateOpen(true)}>
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
                          {user.isRoot && (
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
                        {user.permissions.createProjects && (
                          <Badge variant="outline">Create Projects</Badge>
                        )}
                        {user.permissions.manageUsers && (
                          <Badge variant="outline">Manage Users</Badge>
                        )}
                        {user.permissions.projectAccess?.length > 0 && (
                          <Badge variant="outline">
                            {user.permissions.projectAccess.length} projects
                          </Badge>
                        )}
                      </div>
                    </TableCell>
                    <TableCell>
                      {user.isBlocked ? (
                        <Badge variant="destructive">Blocked</Badge>
                      ) : (
                        <Badge variant="default">Active</Badge>
                      )}
                    </TableCell>
                    <TableCell>
                      {!user.isRoot && user.id !== currentUser?.id && (
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
                            {user.isBlocked ? (
                              <DropdownMenuItem
                                onClick={() => unblockMutation.mutate(user.id)}
                              >
                                <ShieldOff className="mr-2 size-4" />
                                Unblock
                              </DropdownMenuItem>
                            ) : (
                              <DropdownMenuItem
                                onClick={() => blockMutation.mutate(user.id)}
                              >
                                <Ban className="mr-2 size-4" />
                                Block
                              </DropdownMenuItem>
                            )}
                            <DropdownMenuSeparator />
                            <DropdownMenuItem
                              className="text-destructive"
                              onClick={() => {
                                setSelectedUser(user);
                                setDeleteOpen(true);
                              }}
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

      {/* Create Dialog */}
      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Add User</DialogTitle>
            <DialogDescription>Create a new user account</DialogDescription>
          </DialogHeader>
          <FieldGroup>
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
            <Field>
              <FieldLabel>Permissions</FieldLabel>
              <div className="space-y-3 mt-2">
                <div className="flex items-center gap-2">
                  <Switch
                    checked={createProjects}
                    onCheckedChange={setCreateProjects}
                  />
                  <span className="text-sm">Can create projects</span>
                </div>
                <div className="flex items-center gap-2">
                  <Switch
                    checked={manageUsers}
                    onCheckedChange={setManageUsers}
                  />
                  <span className="text-sm">Can manage users</span>
                </div>
              </div>
            </Field>
            <Field>
              <FieldLabel>Project Access</FieldLabel>
              <FieldDescription>
                Select which projects this user can access
              </FieldDescription>
              <div className="mt-2 space-y-2 max-h-48 overflow-y-auto border rounded-md p-3">
                {projects.map((project) => (
                  <div key={project.id} className="flex items-center gap-2">
                    <Checkbox
                      id={`create-${project.id}`}
                      checked={projectAccess.includes(project.id)}
                      onCheckedChange={() => toggleProjectAccess(project.id)}
                    />
                    <label
                      htmlFor={`create-${project.id}`}
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
            <Button variant="outline" onClick={() => setCreateOpen(false)}>
              Cancel
            </Button>
            <Button
              onClick={handleCreate}
              disabled={
                !name || !email || !password || createMutation.isPending
              }
            >
              {createMutation.isPending ? 'Creating...' : 'Create'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Edit Dialog */}
      <Dialog open={editOpen} onOpenChange={setEditOpen}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Edit User</DialogTitle>
          </DialogHeader>
          <FieldGroup>
            <Field>
              <FieldLabel>Name</FieldLabel>
              <Input
                value={name}
                onChange={(e) => setName(e.target.value)}
              />
            </Field>
            <Field>
              <FieldLabel>Email</FieldLabel>
              <Input value={selectedUser?.email} disabled />
            </Field>
            <Field>
              <FieldLabel>Permissions</FieldLabel>
              <div className="space-y-3 mt-2">
                <div className="flex items-center gap-2">
                  <Switch
                    checked={createProjects}
                    onCheckedChange={setCreateProjects}
                  />
                  <span className="text-sm">Can create projects</span>
                </div>
                <div className="flex items-center gap-2">
                  <Switch
                    checked={manageUsers}
                    onCheckedChange={setManageUsers}
                  />
                  <span className="text-sm">Can manage users</span>
                </div>
              </div>
            </Field>
            <Field>
              <FieldLabel>Project Access</FieldLabel>
              <div className="mt-2 space-y-2 max-h-48 overflow-y-auto border rounded-md p-3">
                {projects.map((project) => (
                  <div key={project.id} className="flex items-center gap-2">
                    <Checkbox
                      id={`edit-${project.id}`}
                      checked={projectAccess.includes(project.id)}
                      onCheckedChange={() => toggleProjectAccess(project.id)}
                    />
                    <label
                      htmlFor={`edit-${project.id}`}
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
            <Button variant="outline" onClick={() => setEditOpen(false)}>
              Cancel
            </Button>
            <Button
              onClick={handleUpdate}
              disabled={updateMutation.isPending}
            >
              {updateMutation.isPending ? 'Saving...' : 'Save'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Confirm */}
      <ConfirmDialog
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
        title="Delete User"
        description={`Are you sure you want to delete "${selectedUser?.name}"?`}
        confirmLabel="Delete"
        variant="destructive"
        onConfirm={() => deleteMutation.mutate()}
        isLoading={deleteMutation.isPending}
      />
    </div>
  );
}
