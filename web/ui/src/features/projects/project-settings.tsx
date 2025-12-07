import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Copy, RefreshCw, Trash2, UserPlus, X } from 'lucide-react';
import { toast } from 'sonner';

import { projectsApi, usersApi } from '@/api';
import { useAuth } from '@/providers/auth-provider';
import type { UpdateProjectRequest } from '@/types';
import { Button } from '@/components/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
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
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Input } from '@/components/ui/input';
import { Field, FieldGroup, FieldLabel, FieldDescription } from '@/components/ui/field';
import { ColorPicker } from '@/components/shared/color-picker';
import { ConfirmDialog } from '@/components/shared/confirm-dialog';
import { Skeleton } from '@/components/ui/skeleton';
import { slugify, copyToClipboard } from '@/lib/utils';

export function ProjectSettings() {
  const { projectId } = useParams<{ projectId: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { user } = useAuth();

  const [name, setName] = useState('');
  const [slug, setSlug] = useState('');
  const [color, setColor] = useState<string | undefined>();
  const [hasChanges, setHasChanges] = useState(false);

  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [regenerateDialogOpen, setRegenerateDialogOpen] = useState(false);
  const [addMemberDialogOpen, setAddMemberDialogOpen] = useState(false);
  const [selectedUserId, setSelectedUserId] = useState('');

  const { data: project, isLoading: projectLoading } = useQuery({
    queryKey: ['project', projectId],
    queryFn: () => projectsApi.get(projectId!),
    enabled: !!projectId,
  });

  const { data: users = [] } = useQuery({
    queryKey: ['users'],
    queryFn: usersApi.list,
    enabled: user?.permissions?.manageUsers || user?.isRoot,
  });

  // Initialize form when project loads
  useState(() => {
    if (project) {
      setName(project.name);
      setSlug(project.slug);
      setColor(project.color);
    }
  });

  // Update form when project changes
  if (project && !hasChanges) {
    if (name !== project.name) setName(project.name);
    if (slug !== project.slug) setSlug(project.slug);
    if (color !== project.color) setColor(project.color);
  }

  const updateMutation = useMutation({
    mutationFn: (data: UpdateProjectRequest) =>
      projectsApi.update(projectId!, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project', projectId] });
      queryClient.invalidateQueries({ queryKey: ['projects'] });
      setHasChanges(false);
      toast.success('Project updated');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Update failed');
    },
  });

  const deleteMutation = useMutation({
    mutationFn: () => projectsApi.delete(projectId!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects'] });
      toast.success('Project deleted');
      navigate('/projects');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Delete failed');
    },
  });

  const regenerateKeyMutation = useMutation({
    mutationFn: () => projectsApi.regenerateKey(projectId!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project', projectId] });
      setRegenerateDialogOpen(false);
      toast.success('API key regenerated');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to regenerate key');
    },
  });

  const addMemberMutation = useMutation({
    mutationFn: (userId: string) => projectsApi.addMember(projectId!, userId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project', projectId] });
      setAddMemberDialogOpen(false);
      setSelectedUserId('');
      toast.success('Member added');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to add member');
    },
  });

  const removeMemberMutation = useMutation({
    mutationFn: (userId: string) => projectsApi.removeMember(projectId!, userId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project', projectId] });
      toast.success('Member removed');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to remove member');
    },
  });

  const handleSave = () => {
    updateMutation.mutate({ name, slug, color });
  };

  const handleFieldChange = () => {
    setHasChanges(true);
  };

  const copyApiKey = async () => {
    if (project?.apiKey) {
      const success = await copyToClipboard(project.apiKey);
      if (success) {
        toast.success('API key copied to clipboard');
      }
    }
  };

  if (projectLoading) {
    return (
      <div className="space-y-6 max-w-2xl">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-64" />
        <Skeleton className="h-48" />
      </div>
    );
  }

  if (!project) {
    return <div>Project not found</div>;
  }

  const isOwner = project.ownerID === user?.id || user?.isRoot;
  const availableUsers = users.filter(
    (u) => u.id !== project.ownerID && !project.members.includes(u.id)
  );
  const memberUsers = users.filter((u) => project.members.includes(u.id));

  return (
    <div className="space-y-6 max-w-2xl">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Project Settings</h1>
        <p className="text-muted-foreground">
          Manage your project configuration
        </p>
      </div>

      {/* General Settings */}
      <Card>
        <CardHeader>
          <CardTitle>General</CardTitle>
          <CardDescription>Basic project information</CardDescription>
        </CardHeader>
        <CardContent>
          <FieldGroup>
            <Field>
              <FieldLabel htmlFor="name">Name</FieldLabel>
              <Input
                id="name"
                value={name}
                onChange={(e) => {
                  setName(e.target.value);
                  handleFieldChange();
                }}
              />
            </Field>
            <Field>
              <FieldLabel htmlFor="slug">Slug</FieldLabel>
              <Input
                id="slug"
                value={slug}
                onChange={(e) => {
                  setSlug(slugify(e.target.value));
                  handleFieldChange();
                }}
              />
              <FieldDescription>
                Used in URLs: /r/{slug}
              </FieldDescription>
            </Field>
            <Field>
              <FieldLabel>Color</FieldLabel>
              <ColorPicker
                value={color}
                onChange={(c) => {
                  setColor(c);
                  handleFieldChange();
                }}
              />
            </Field>
            {hasChanges && (
              <LoadingButton onClick={handleSave} loading={updateMutation.isPending}>
                Save Changes
              </LoadingButton>
            )}
          </FieldGroup>
        </CardContent>
      </Card>

      {/* Members */}
      {isOwner && (
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle>Members</CardTitle>
                <CardDescription>
                  Users who have access to this project
                </CardDescription>
              </div>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setAddMemberDialogOpen(true)}
                disabled={availableUsers.length === 0}
              >
                <UserPlus className="mr-2 size-4" />
                Add Member
              </Button>
            </div>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              {memberUsers.length === 0 ? (
                <p className="text-sm text-muted-foreground">
                  No additional members
                </p>
              ) : (
                memberUsers.map((member) => (
                  <div
                    key={member.id}
                    className="flex items-center justify-between py-2"
                  >
                    <div>
                      <p className="font-medium">{member.name}</p>
                      <p className="text-sm text-muted-foreground">
                        {member.email}
                      </p>
                    </div>
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => removeMemberMutation.mutate(member.id)}
                      disabled={removeMemberMutation.isPending}
                    >
                      <X className="size-4" />
                    </Button>
                  </div>
                ))
              )}
            </div>
          </CardContent>
        </Card>
      )}

      {/* API Key */}
      <Card>
        <CardHeader>
          <CardTitle>API Key</CardTitle>
          <CardDescription>
            Use this key to authenticate API requests
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-2">
            <code className="flex-1 bg-muted px-3 py-2 rounded-md text-sm font-mono truncate">
              {project.apiKey}
            </code>
            <Button variant="outline" size="icon" onClick={copyApiKey}>
              <Copy className="size-4" />
            </Button>
            {isOwner && (
              <Button
                variant="outline"
                size="icon"
                onClick={() => setRegenerateDialogOpen(true)}
              >
                <RefreshCw className="size-4" />
              </Button>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Danger Zone */}
      {isOwner && (
        <Card className="border-destructive/50">
          <CardHeader>
            <CardTitle className="text-destructive">Danger Zone</CardTitle>
            <CardDescription>
              Irreversible and destructive actions
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between">
              <div>
                <p className="font-medium">Delete Project</p>
                <p className="text-sm text-muted-foreground">
                  Permanently delete this project and all its data
                </p>
              </div>
              <Button
                variant="destructive"
                onClick={() => setDeleteDialogOpen(true)}
              >
                <Trash2 className="mr-2 size-4" />
                Delete
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Dialogs */}
      <ConfirmDialog
        open={deleteDialogOpen}
        onOpenChange={setDeleteDialogOpen}
        title="Delete Project"
        description={`Are you sure you want to delete "${project.name}"? This action cannot be undone.`}
        confirmLabel="Delete"
        variant="destructive"
        onConfirm={() => deleteMutation.mutate()}
        isLoading={deleteMutation.isPending}
      />

      <ConfirmDialog
        open={regenerateDialogOpen}
        onOpenChange={setRegenerateDialogOpen}
        title="Regenerate API Key"
        description="Are you sure? The current API key will be invalidated."
        confirmLabel="Regenerate"
        onConfirm={() => regenerateKeyMutation.mutate()}
        isLoading={regenerateKeyMutation.isPending}
      />

      <Dialog open={addMemberDialogOpen} onOpenChange={setAddMemberDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Add Member</DialogTitle>
            <DialogDescription>
              Grant a user access to this project
            </DialogDescription>
          </DialogHeader>
          <Field>
            <FieldLabel>User</FieldLabel>
            <Select value={selectedUserId} onValueChange={setSelectedUserId}>
              <SelectTrigger>
                <SelectValue placeholder="Select user" />
              </SelectTrigger>
              <SelectContent>
                {availableUsers.map((u) => (
                  <SelectItem key={u.id} value={u.id}>
                    {u.name} ({u.email})
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </Field>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setAddMemberDialogOpen(false)}
            >
              Cancel
            </Button>
            <LoadingButton
              onClick={() => addMemberMutation.mutate(selectedUserId)}
              disabled={!selectedUserId}
              loading={addMemberMutation.isPending}
            >
              Add
            </LoadingButton>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
