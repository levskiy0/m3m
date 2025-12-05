import { useState, useEffect, useRef } from 'react';
import { useParams, useLocation } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Save,
  GitBranch,
  Tag,
  Plus,
  Trash2,
  RotateCcw,
  Square,
  Bug,
  Code,
  ScrollText,
} from 'lucide-react';
import { toast } from 'sonner';

import { pipelineApi, runtimeApi, projectsApi } from '@/api';
import { DEFAULT_SERVICE_CODE, RELEASE_TAGS } from '@/lib/constants';
import type { CreateBranchRequest, CreateReleaseRequest, LogEntry } from '@/types';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { ScrollArea } from '@/components/ui/scroll-area';
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

type PipelineTab = 'editor' | 'logs' | 'releases';

export function PipelinePage() {
  const { projectId } = useParams<{ projectId: string }>();
  const location = useLocation();
  const queryClient = useQueryClient();

  // Get initial branch from location state (when coming from Overview)
  const initialBranchName = (location.state as { branch?: string } | null)?.branch;

  const [selectedBranchId, setSelectedBranchId] = useState<string>('');
  const [code, setCode] = useState('');
  const [hasChanges, setHasChanges] = useState(false);
  const [activeTab, setActiveTab] = useState<PipelineTab>('editor');

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

  // Logs ref for auto-scroll
  const logsRef = useRef<HTMLDivElement>(null);

  const { data: project } = useQuery({
    queryKey: ['project', projectId],
    queryFn: () => projectsApi.get(projectId!),
    enabled: !!projectId,
  });

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

  // Check if running debug mode
  const isRunning = project?.status === 'running';
  const isDebugMode = isRunning && project?.runningSource?.startsWith('debug:');
  const runningBranch = isDebugMode ? project?.runningSource?.replace('debug:', '') : null;
  const runningRelease = isRunning && project?.runningSource?.startsWith('release:')
    ? project?.runningSource?.replace('release:', '')
    : null;

  // Fetch logs when running debug mode
  const { data: logsData = [] } = useQuery({
    queryKey: ['logs', projectId],
    queryFn: () => runtimeApi.logs(projectId!),
    enabled: !!projectId && isDebugMode,
    refetchInterval: isDebugMode ? 2000 : false,
  });

  const logs: LogEntry[] = Array.isArray(logsData) ? logsData : [];

  // Auto-scroll logs
  useEffect(() => {
    if (logsRef.current && isDebugMode) {
      logsRef.current.scrollTop = logsRef.current.scrollHeight;
    }
  }, [logs, isDebugMode]);

  // Auto-select branch on load (prioritize initialBranchName from state)
  useEffect(() => {
    if (branches.length > 0 && !selectedBranchId) {
      // First try to select the branch from location state
      if (initialBranchName) {
        const targetBranch = branches.find((b) => b.name === initialBranchName);
        if (targetBranch) {
          setSelectedBranchId(targetBranch.id);
          return;
        }
      }
      // Otherwise select develop branch
      const developBranch = branches.find((b) => b.name === 'develop');
      if (developBranch) {
        setSelectedBranchId(developBranch.id);
      } else {
        setSelectedBranchId(branches[0].id);
      }
    }
  }, [branches, selectedBranchId, initialBranchName]);

  // Fetch full branch with code when selected
  const { data: currentBranch, isLoading: branchLoading } = useQuery({
    queryKey: ['branch', projectId, selectedBranchId],
    queryFn: () => pipelineApi.getBranch(projectId!, selectedBranchId),
    enabled: !!projectId && !!selectedBranchId,
  });

  const { data: runtimeTypes } = useQuery({
    queryKey: ['runtime-types'],
    queryFn: runtimeApi.getTypes,
  });

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
    mutationFn: () => pipelineApi.updateBranch(projectId!, selectedBranchId, { code }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['branch', projectId, selectedBranchId] });
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
      setSelectedBranchId(branch.id);
      resetBranchForm();
      toast.success('Branch created');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to create branch');
    },
  });

  const deleteBranchMutation = useMutation({
    mutationFn: () => pipelineApi.deleteBranch(projectId!, selectedBranchId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['branches', projectId] });
      setDeleteBranchOpen(false);
      // Select develop branch after deletion
      const developBranch = branches.find((b) => b.name === 'develop');
      if (developBranch) {
        setSelectedBranchId(developBranch.id);
      }
      toast.success('Branch deleted');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to delete branch');
    },
  });

  const resetBranchMutation = useMutation({
    mutationFn: () =>
      pipelineApi.resetBranch(projectId!, selectedBranchId, {
        target_version: resetTargetVersion,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['branch', projectId, selectedBranchId] });
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
    mutationFn: (releaseId: string) => pipelineApi.deleteRelease(projectId!, releaseId),
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

  const startDebugMutation = useMutation({
    mutationFn: (branchName: string) => runtimeApi.start(projectId!, { branch: branchName }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project', projectId] });
      queryClient.invalidateQueries({ queryKey: ['logs', projectId] });
      toast.success('Debug mode started');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to start debug');
    },
  });

  const stopDebugMutation = useMutation({
    mutationFn: () => runtimeApi.stop(projectId!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project', projectId] });
      toast.success('Service stopped');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to stop');
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

  const selectedBranchSummary = branches.find((b) => b.id === selectedBranchId);
  const isDevelopBranch = currentBranch?.name === 'develop';
  const isLoading = branchesLoading || releasesLoading || (selectedBranchId && branchLoading);

  if (isLoading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-10 w-full" />
        <Skeleton className="h-[600px] w-full" />
      </div>
    );
  }

  return (
    <>
      {/* Tabs */}
      <div className="w-full">
        <div className="flex items-end px-4">
          {/* Editor tab */}
          <button
            onClick={() => setActiveTab('editor')}
            className={cn(
              'flex items-center gap-2 px-4 py-2 text-sm border-t border-l border-r rounded-t-xl',
              activeTab === 'editor'
                ? 'border-border bg-card'
                : 'border-transparent text-muted-foreground hover:text-foreground'
            )}
            style={{
              marginBottom: activeTab === 'editor' ? -1 : 0,
            }}
          >
            <Code className="size-4" />
            Editor
            {hasChanges && <span className="text-orange-500">*</span>}
          </button>

          {/* Logs tab - only show when debug mode is active */}
          {isDebugMode && runningBranch === currentBranch?.name && (
            <button
              onClick={() => setActiveTab('logs')}
              className={cn(
                'flex items-center gap-2 px-4 py-2 text-sm border-t border-l border-r rounded-t-xl',
                activeTab === 'logs'
                  ? 'border-border bg-card'
                  : 'border-transparent text-muted-foreground hover:text-foreground'
              )}
              style={{
                marginBottom: activeTab === 'logs' ? -1 : 0,
              }}
            >
              <ScrollText className="size-4" />
              Logs
              <Badge variant="outline" className="ml-1 border-amber-500/50 text-amber-500 text-xs">
                {runningBranch}
              </Badge>
            </button>
          )}

          {/* Releases tab */}
          <button
            onClick={() => setActiveTab('releases')}
            className={cn(
              'flex items-center gap-2 px-4 py-2 text-sm border-t border-l border-r rounded-t-xl',
              activeTab === 'releases'
                ? 'border-border bg-card'
                : 'border-transparent text-muted-foreground hover:text-foreground'
            )}
            style={{
              marginBottom: activeTab === 'releases' ? -1 : 0,
            }}
          >
            <Tag className="size-4" />
            Releases
            {releases.length > 0 && (
              <Badge variant="secondary" className="ml-1 text-xs">
                {releases.length}
              </Badge>
            )}
          </button>
        </div>

        <Card className="flex flex-col gap-0 rounded-t-none py-0 overflow-hidden" style={{ height: 'calc(100vh - 120px)' }}>
          {/* Editor Content */}
          {activeTab === 'editor' && (
            <>
              {/* Editor Header */}
              <div className="flex items-center justify-between flex-shrink-0 px-4 py-3 border-b">
                <div className="flex items-center gap-4">
                  <Select value={selectedBranchId} onValueChange={setSelectedBranchId}>
                    <SelectTrigger className="w-48">
                      <GitBranch className="mr-2 size-4" />
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {branches.map((branch) => (
                        <SelectItem key={branch.id} value={branch.id}>
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
                  {!isDevelopBranch && (
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
                  {/* Debug Run/Stop Button */}
                  {currentBranch && (
                    runningBranch === currentBranch.name ? (
                      <Button
                        variant="destructive"
                        size="sm"
                        onClick={() => stopDebugMutation.mutate()}
                        disabled={stopDebugMutation.isPending}
                      >
                        <Square className="mr-2 size-4" />
                        Stop Debug
                      </Button>
                    ) : (
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => {
                          if (hasChanges) {
                            saveMutation.mutate();
                          }
                          startDebugMutation.mutate(currentBranch.name);
                        }}
                        disabled={startDebugMutation.isPending || (isRunning && !isDebugMode)}
                        className="border-amber-500/50 text-amber-600 hover:bg-amber-500/10"
                      >
                        <Bug className="mr-2 size-4" />
                        Run Debug
                      </Button>
                    )
                  )}
                  <Button onClick={() => saveMutation.mutate()} disabled={!hasChanges || saveMutation.isPending}>
                    <Save className="mr-2 size-4" />
                    {saveMutation.isPending ? 'Saving...' : 'Save'}
                  </Button>
                </div>
              </div>
              {/* Code Editor */}
              <div className="flex-1 min-h-0">
                <CodeEditor
                  value={code}
                  onChange={handleCodeChange}
                  language="javascript"
                  typeDefinitions={runtimeTypes}
                />
              </div>
            </>
          )}

          {/* Logs Content */}
          {activeTab === 'logs' && isDebugMode && runningBranch === currentBranch?.name && (
            <ScrollArea
              ref={logsRef}
              className="flex-1 bg-zinc-950 font-mono text-xs"
            >
              {logs.length === 0 ? (
                <div className="flex items-center justify-center h-full text-muted-foreground py-12">
                  Waiting for logs...
                </div>
              ) : (
                <div className="space-y-0.5 p-4">
                  {logs.slice(-500).map((log, index) => (
                    <div key={index} className="flex gap-2 text-gray-300">
                      <span className="text-gray-500 shrink-0">
                        {new Date(log.timestamp).toLocaleTimeString()}
                      </span>
                      <span
                        className={cn(
                          'shrink-0 uppercase w-12',
                          log.level === 'error' && 'text-red-400',
                          log.level === 'warn' && 'text-amber-400',
                          log.level === 'info' && 'text-blue-400',
                          log.level === 'debug' && 'text-gray-400'
                        )}
                      >
                        [{log.level}]
                      </span>
                      <span className="text-gray-200 break-all">{log.message}</span>
                    </div>
                  ))}
                </div>
              )}
            </ScrollArea>
          )}

          {/* Releases Content */}
          {activeTab === 'releases' && (
            <>
              {/* Releases Header */}
              <div className="flex items-center justify-between flex-shrink-0 px-4 py-3 border-b">
                <span className="text-sm text-muted-foreground">
                  {releases.length} release{releases.length !== 1 ? 's' : ''}
                </span>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setCreateReleaseOpen(true)}
                  disabled={branches.length === 0}
                >
                  <Tag className="mr-2 size-4" />
                  Create Release
                </Button>
              </div>
              <CardContent className="flex-1 overflow-auto p-6">
                {releases.length === 0 ? (
                  <EmptyState
                    title="No releases"
                    description="Create a release to deploy your code"
                    className="py-8"
                  />
                ) : (
                <div className="space-y-2">
                  {releases.map((release) => {
                    const isReleaseRunning = runningRelease === release.version;
                    return (
                      <div
                        key={release.id}
                        className={cn(
                          'flex items-center justify-between p-3 rounded-lg border',
                          isReleaseRunning && 'border-green-500/50 bg-green-500/5'
                        )}
                      >
                        <div className="min-w-0">
                          <div className="flex items-center gap-2">
                            <span className="font-mono font-medium">
                              {release.version}
                            </span>
                            {isReleaseRunning && (
                              <Badge variant="default" className="text-xs bg-green-600">
                                Running
                              </Badge>
                            )}
                            {release.tag && (
                              <Badge variant="outline" className="text-xs capitalize">
                                {release.tag}
                              </Badge>
                            )}
                          </div>
                          {release.comment && (
                            <p className="text-sm text-muted-foreground mt-1">
                              {release.comment}
                            </p>
                          )}
                          <p className="text-xs text-muted-foreground mt-1">
                            {new Date(release.createdAt).toLocaleString()}
                          </p>
                        </div>
                        <div className="flex items-center gap-1">
                          {!isReleaseRunning && (
                            <Button
                              variant="ghost"
                              size="icon"
                              className="size-8"
                              title="Delete release"
                              onClick={() => {
                                setReleaseToDelete(release.id);
                                setDeleteReleaseOpen(true);
                              }}
                            >
                              <Trash2 className="size-4" />
                            </Button>
                          )}
                        </div>
                      </div>
                    );
                  })}
                </div>
                )}
              </CardContent>
            </>
          )}
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
            <div className="grid grid-cols-2 gap-4">
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
                          <SelectItem key={b.id} value={b.name}>
                            {b.name}
                          </SelectItem>
                        ))
                      : releases.map((r) => (
                          <SelectItem key={r.id} value={r.version}>
                            {r.version}
                          </SelectItem>
                        ))}
                  </SelectContent>
                </Select>
              </Field>
            </div>
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
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>Create Release</DialogTitle>
            <DialogDescription>
              Publish a new release from a branch
            </DialogDescription>
          </DialogHeader>
          <FieldGroup>
            <div className="grid grid-cols-3 gap-4">
              <Field>
                <FieldLabel>Source Branch</FieldLabel>
                <Select value={releaseBranch} onValueChange={setReleaseBranch}>
                  <SelectTrigger>
                    <SelectValue placeholder="Select branch" />
                  </SelectTrigger>
                  <SelectContent>
                    {branches.map((b) => (
                      <SelectItem key={b.id} value={b.name}>
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
            </div>
            <FieldDescription className="text-center">
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
              Reset "{selectedBranchSummary?.name}" to a specific release version
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
                    <SelectItem key={r.id} value={r.version}>
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
        description={`Are you sure you want to delete "${selectedBranchSummary?.name}"?`}
        confirmLabel="Delete"
        variant="destructive"
        onConfirm={() => deleteBranchMutation.mutate()}
        isLoading={deleteBranchMutation.isPending}
      />

      <ConfirmDialog
        open={deleteReleaseOpen}
        onOpenChange={setDeleteReleaseOpen}
        title="Delete Release"
        description={`Are you sure you want to delete this release?`}
        confirmLabel="Delete"
        variant="destructive"
        onConfirm={() => releaseToDelete && deleteReleaseMutation.mutate(releaseToDelete)}
        isLoading={deleteReleaseMutation.isPending}
      />
    </>
  );
}
