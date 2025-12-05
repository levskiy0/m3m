import { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Save,
  GitBranch,
  Tag,
  Plus,
  Trash2,
  RotateCcw,
  Play,
} from 'lucide-react';
import { toast } from 'sonner';

import { pipelineApi, runtimeApi } from '@/api';
import { DEFAULT_SERVICE_CODE, RELEASE_TAGS } from '@/lib/constants';
import type { CreateBranchRequest, CreateReleaseRequest } from '@/types';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
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
import { Textarea } from '@/components/ui/textarea';
import { Badge } from '@/components/ui/badge';
import { Field, FieldGroup, FieldLabel, FieldDescription } from '@/components/ui/field';
import { CodeEditor } from '@/components/shared/code-editor';
import { ConfirmDialog } from '@/components/shared/confirm-dialog';
import { EmptyState } from '@/components/shared/empty-state';
import { Skeleton } from '@/components/ui/skeleton';
import { cn } from '@/lib/utils';

export function PipelinePage() {
  const { projectId } = useParams<{ projectId: string }>();
  const queryClient = useQueryClient();

  const [selectedBranch, setSelectedBranch] = useState<string>('develop');
  const [code, setCode] = useState('');
  const [hasChanges, setHasChanges] = useState(false);

  // Dialogs
  const [createBranchOpen, setCreateBranchOpen] = useState(false);
  const [createReleaseOpen, setCreateReleaseOpen] = useState(false);
  const [resetBranchOpen, setResetBranchOpen] = useState(false);
  const [deleteBranchOpen, setDeleteBranchOpen] = useState(false);
  const [deleteReleaseOpen, setDeleteReleaseOpen] = useState(false);
  const [releaseToDelete, setReleaseToDelete] = useState<string | null>(null);

  // New branch form
  const [newBranchName, setNewBranchName] = useState('');
  const [newBranchSource, setNewBranchSource] = useState<'branch' | 'release'>('branch');
  const [newBranchSourceName, setNewBranchSourceName] = useState('');

  // New release form
  const [releaseBranch, setReleaseBranch] = useState('');
  const [bumpType, setBumpType] = useState<'minor' | 'major'>('minor');
  const [releaseComment, setReleaseComment] = useState('');
  const [releaseTag, setReleaseTag] = useState<string>('develop');

  // Reset form
  const [resetTargetVersion, setResetTargetVersion] = useState('');

  const { data: branches = [], isLoading: branchesLoading } = useQuery({
    queryKey: ['branches', projectId],
    queryFn: () => pipelineApi.listBranches(projectId!),
    enabled: !!projectId,
  });

  const { data: releases = [], isLoading: releasesLoading } = useQuery({
    queryKey: ['releases', projectId],
    queryFn: () => pipelineApi.listReleases(projectId!),
    enabled: !!projectId,
  });

  const { data: runtimeTypes } = useQuery({
    queryKey: ['runtime-types'],
    queryFn: runtimeApi.getTypes,
  });

  const currentBranch = branches.find((b) => b.name === selectedBranch);

  // Load branch code
  useEffect(() => {
    if (currentBranch) {
      setCode(currentBranch.code);
      setHasChanges(false);
    } else if (branches.length === 0) {
      setCode(DEFAULT_SERVICE_CODE);
      setHasChanges(true);
    }
  }, [currentBranch, branches.length]);

  // Mutations
  const saveMutation = useMutation({
    mutationFn: () => pipelineApi.updateBranch(projectId!, selectedBranch, { code }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['branches', projectId] });
      setHasChanges(false);
      toast.success('Code saved');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to save');
    },
  });

  const createBranchMutation = useMutation({
    mutationFn: (data: CreateBranchRequest) =>
      pipelineApi.createBranch(projectId!, data),
    onSuccess: (branch) => {
      queryClient.invalidateQueries({ queryKey: ['branches', projectId] });
      setCreateBranchOpen(false);
      setSelectedBranch(branch.name);
      resetBranchForm();
      toast.success('Branch created');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to create branch');
    },
  });

  const deleteBranchMutation = useMutation({
    mutationFn: () => pipelineApi.deleteBranch(projectId!, selectedBranch),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['branches', projectId] });
      setDeleteBranchOpen(false);
      setSelectedBranch('develop');
      toast.success('Branch deleted');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to delete branch');
    },
  });

  const resetBranchMutation = useMutation({
    mutationFn: () =>
      pipelineApi.resetBranch(projectId!, selectedBranch, {
        target_version: resetTargetVersion,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['branches', projectId] });
      setResetBranchOpen(false);
      setResetTargetVersion('');
      toast.success('Branch reset');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to reset branch');
    },
  });

  const createReleaseMutation = useMutation({
    mutationFn: (data: CreateReleaseRequest) =>
      pipelineApi.createRelease(projectId!, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['releases', projectId] });
      setCreateReleaseOpen(false);
      resetReleaseForm();
      toast.success('Release created');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to create release');
    },
  });

  const deleteReleaseMutation = useMutation({
    mutationFn: (version: string) => pipelineApi.deleteRelease(projectId!, version),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['releases', projectId] });
      setDeleteReleaseOpen(false);
      setReleaseToDelete(null);
      toast.success('Release deleted');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to delete release');
    },
  });

  const activateReleaseMutation = useMutation({
    mutationFn: (version: string) => pipelineApi.activateRelease(projectId!, version),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['releases', projectId] });
      queryClient.invalidateQueries({ queryKey: ['project', projectId] });
      toast.success('Release activated');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to activate release');
    },
  });

  const resetBranchForm = () => {
    setNewBranchName('');
    setNewBranchSource('branch');
    setNewBranchSourceName('');
  };

  const resetReleaseForm = () => {
    setReleaseBranch('');
    setBumpType('minor');
    setReleaseComment('');
    setReleaseTag('develop');
  };

  const handleCreateBranch = () => {
    createBranchMutation.mutate({
      name: newBranchName,
      sourceType: newBranchSource,
      sourceName: newBranchSourceName,
    });
  };

  const handleCreateRelease = () => {
    createReleaseMutation.mutate({
      branch_name: releaseBranch,
      bump_type: bumpType,
      comment: releaseComment || undefined,
      tag: releaseTag as 'stable' | 'hot-fix' | 'night-build' | 'develop',
    });
  };

  const handleCodeChange = (value: string) => {
    setCode(value);
    setHasChanges(value !== (currentBranch?.code || DEFAULT_SERVICE_CODE));
  };

  const isLoading = branchesLoading || releasesLoading;

  if (isLoading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-10 w-full" />
        <Skeleton className="h-[600px] w-full" />
      </div>
    );
  }

  return (
    <div className="flex flex-col h-[calc(100vh-8rem)] gap-4">
      {/* Header */}
      <div className="flex items-center justify-between flex-shrink-0">
        <div className="flex items-center gap-4">
          <Select value={selectedBranch} onValueChange={setSelectedBranch}>
            <SelectTrigger className="w-48">
              <GitBranch className="mr-2 size-4" />
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {branches.map((branch) => (
                <SelectItem key={branch.name} value={branch.name}>
                  {branch.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Button
            variant="outline"
            size="sm"
            onClick={() => setCreateBranchOpen(true)}
          >
            <Plus className="mr-2 size-4" />
            New Branch
          </Button>
          {selectedBranch !== 'develop' && (
            <>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setResetBranchOpen(true)}
              >
                <RotateCcw className="mr-2 size-4" />
                Reset
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setDeleteBranchOpen(true)}
              >
                <Trash2 className="mr-2 size-4" />
                Delete
              </Button>
            </>
          )}
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            onClick={() => setCreateReleaseOpen(true)}
            disabled={branches.length === 0}
          >
            <Tag className="mr-2 size-4" />
            Create Release
          </Button>
          <Button onClick={() => saveMutation.mutate()} disabled={!hasChanges || saveMutation.isPending}>
            <Save className="mr-2 size-4" />
            {saveMutation.isPending ? 'Saving...' : 'Save'}
          </Button>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex gap-4 flex-1 min-h-0">
        {/* Code Editor */}
        <div className="flex-1 border rounded-lg overflow-hidden">
          <CodeEditor
            value={code}
            onChange={handleCodeChange}
            language="javascript"
            typeDefinitions={runtimeTypes}
          />
        </div>

        {/* Releases Panel */}
        <Card className="w-80 flex-shrink-0 flex flex-col">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium flex items-center gap-2">
              <Tag className="size-4" />
              Releases
            </CardTitle>
          </CardHeader>
          <CardContent className="flex-1 overflow-auto">
            {releases.length === 0 ? (
              <EmptyState
                title="No releases"
                description="Create a release to deploy your code"
                className="py-8"
              />
            ) : (
              <div className="space-y-2">
                {releases.map((release) => (
                  <div
                    key={release.version}
                    className={cn(
                      'flex items-center justify-between p-2 rounded-lg border',
                      release.isActive && 'border-primary bg-primary/5'
                    )}
                  >
                    <div className="min-w-0">
                      <div className="flex items-center gap-2">
                        <span className="font-mono font-medium">
                          {release.version}
                        </span>
                        {release.isActive && (
                          <Badge variant="default" className="text-xs">
                            Active
                          </Badge>
                        )}
                        {release.tag && (
                          <Badge variant="outline" className="text-xs capitalize">
                            {release.tag}
                          </Badge>
                        )}
                      </div>
                      {release.comment && (
                        <p className="text-xs text-muted-foreground truncate">
                          {release.comment}
                        </p>
                      )}
                    </div>
                    <div className="flex items-center gap-1">
                      {!release.isActive && (
                        <Button
                          variant="ghost"
                          size="icon"
                          className="size-8"
                          onClick={() => activateReleaseMutation.mutate(release.version)}
                          disabled={activateReleaseMutation.isPending}
                        >
                          <Play className="size-4" />
                        </Button>
                      )}
                      {!release.isActive && (
                        <Button
                          variant="ghost"
                          size="icon"
                          className="size-8"
                          onClick={() => {
                            setReleaseToDelete(release.version);
                            setDeleteReleaseOpen(true);
                          }}
                        >
                          <Trash2 className="size-4" />
                        </Button>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Create Branch Dialog */}
      <Dialog open={createBranchOpen} onOpenChange={setCreateBranchOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create Branch</DialogTitle>
            <DialogDescription>
              Create a new development branch
            </DialogDescription>
          </DialogHeader>
          <FieldGroup>
            <Field>
              <FieldLabel>Branch Name</FieldLabel>
              <Input
                value={newBranchName}
                onChange={(e) => setNewBranchName(e.target.value)}
                placeholder="feature/my-feature"
              />
            </Field>
            <Field>
              <FieldLabel>Source Type</FieldLabel>
              <Select
                value={newBranchSource}
                onValueChange={(v) => setNewBranchSource(v as 'branch' | 'release')}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="branch">Branch</SelectItem>
                  <SelectItem value="release">Release</SelectItem>
                </SelectContent>
              </Select>
            </Field>
            <Field>
              <FieldLabel>Source {newBranchSource === 'branch' ? 'Branch' : 'Release'}</FieldLabel>
              <Select value={newBranchSourceName} onValueChange={setNewBranchSourceName}>
                <SelectTrigger>
                  <SelectValue placeholder="Select source" />
                </SelectTrigger>
                <SelectContent>
                  {newBranchSource === 'branch'
                    ? branches.map((b) => (
                        <SelectItem key={b.name} value={b.name}>
                          {b.name}
                        </SelectItem>
                      ))
                    : releases.map((r) => (
                        <SelectItem key={r.version} value={r.version}>
                          {r.version}
                        </SelectItem>
                      ))}
                </SelectContent>
              </Select>
            </Field>
          </FieldGroup>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateBranchOpen(false)}>
              Cancel
            </Button>
            <Button
              onClick={handleCreateBranch}
              disabled={!newBranchName || !newBranchSourceName || createBranchMutation.isPending}
            >
              {createBranchMutation.isPending ? 'Creating...' : 'Create'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Create Release Dialog */}
      <Dialog open={createReleaseOpen} onOpenChange={setCreateReleaseOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create Release</DialogTitle>
            <DialogDescription>
              Publish a new release from a branch
            </DialogDescription>
          </DialogHeader>
          <FieldGroup>
            <Field>
              <FieldLabel>Source Branch</FieldLabel>
              <Select value={releaseBranch} onValueChange={setReleaseBranch}>
                <SelectTrigger>
                  <SelectValue placeholder="Select branch" />
                </SelectTrigger>
                <SelectContent>
                  {branches.map((b) => (
                    <SelectItem key={b.name} value={b.name}>
                      {b.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </Field>
            <Field>
              <FieldLabel>Version Bump</FieldLabel>
              <Select value={bumpType} onValueChange={(v) => setBumpType(v as 'minor' | 'major')}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="minor">Minor (X.+1)</SelectItem>
                  <SelectItem value="major">Major (+1.0)</SelectItem>
                </SelectContent>
              </Select>
              <FieldDescription>
                Current: {releases[0]?.version || '0.0'} → Next:{' '}
                {releases[0]?.version
                  ? bumpType === 'minor'
                    ? `${releases[0].version.split('.')[0]}.${
                        parseInt(releases[0].version.split('.')[1]) + 1
                      }`
                    : `${parseInt(releases[0].version.split('.')[0]) + 1}.0`
                  : bumpType === 'minor'
                  ? '0.1'
                  : '1.0'}
              </FieldDescription>
            </Field>
            <Field>
              <FieldLabel>Tag</FieldLabel>
              <Select value={releaseTag} onValueChange={setReleaseTag}>
                <SelectTrigger>
                  <SelectValue placeholder="Select tag" />
                </SelectTrigger>
                <SelectContent>
                  {RELEASE_TAGS.map((tag) => (
                    <SelectItem key={tag.value} value={tag.value}>
                      {tag.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </Field>
            <Field>
              <FieldLabel>Comment (optional)</FieldLabel>
              <Textarea
                value={releaseComment}
                onChange={(e) => setReleaseComment(e.target.value)}
                placeholder="Describe this release..."
                rows={3}
              />
            </Field>
          </FieldGroup>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateReleaseOpen(false)}>
              Cancel
            </Button>
            <Button
              onClick={handleCreateRelease}
              disabled={!releaseBranch || createReleaseMutation.isPending}
            >
              {createReleaseMutation.isPending ? 'Creating...' : 'Create Release'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Reset Branch Dialog */}
      <Dialog open={resetBranchOpen} onOpenChange={setResetBranchOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Reset Branch</DialogTitle>
            <DialogDescription>
              Reset "{selectedBranch}" to a specific release version
            </DialogDescription>
          </DialogHeader>
          <FieldGroup>
            <Field>
              <FieldLabel>Target Release</FieldLabel>
              <Select value={resetTargetVersion} onValueChange={setResetTargetVersion}>
                <SelectTrigger>
                  <SelectValue placeholder="Select release" />
                </SelectTrigger>
                <SelectContent>
                  {releases.map((r) => (
                    <SelectItem key={r.version} value={r.version}>
                      {r.version}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </Field>
          </FieldGroup>
          <DialogFooter>
            <Button variant="outline" onClick={() => setResetBranchOpen(false)}>
              Cancel
            </Button>
            <Button
              onClick={() => resetBranchMutation.mutate()}
              disabled={!resetTargetVersion || resetBranchMutation.isPending}
            >
              {resetBranchMutation.isPending ? 'Resetting...' : 'Reset'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Confirm Dialogs */}
      <ConfirmDialog
        open={deleteBranchOpen}
        onOpenChange={setDeleteBranchOpen}
        title="Delete Branch"
        description={`Are you sure you want to delete "${selectedBranch}"?`}
        confirmLabel="Delete"
        variant="destructive"
        onConfirm={() => deleteBranchMutation.mutate()}
        isLoading={deleteBranchMutation.isPending}
      />

      <ConfirmDialog
        open={deleteReleaseOpen}
        onOpenChange={setDeleteReleaseOpen}
        title="Delete Release"
        description={`Are you sure you want to delete release "${releaseToDelete}"?`}
        confirmLabel="Delete"
        variant="destructive"
        onConfirm={() => releaseToDelete && deleteReleaseMutation.mutate(releaseToDelete)}
        isLoading={deleteReleaseMutation.isPending}
      />
    </div>
  );
}
