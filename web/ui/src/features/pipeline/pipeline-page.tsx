import { useState, useEffect, useMemo, useCallback } from 'react';
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
  Zap,
} from 'lucide-react';
import { toast } from 'sonner';

import { pipelineApi, runtimeApi, projectsApi, templatesApi, actionsApi } from '@/api';
import { queryKeys } from '@/lib/query-keys';
import { useTitle, useWebSocket } from '@/hooks';
import { RELEASE_TAGS } from '@/features/pipeline/constants';
import type { CreateBranchRequest, CreateReleaseRequest, LogEntry, ActionRuntimeState, ActionState } from '@/types';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { LogsViewer } from '@/components/shared/logs-viewer';
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
import { Textarea } from '@/components/ui/textarea';
import { Badge } from '@/components/ui/badge';
import { Field, FieldGroup, FieldLabel, FieldDescription } from '@/components/ui/field';
import { CodeEditor } from '@/components/shared/code-editor';
import { ConfirmDialog } from '@/components/shared/confirm-dialog';
import { Kbd } from '@/components/ui/kbd';
import { EditorTabs, EditorTab } from '@/components/ui/editor-tabs';
import { EmptyState } from '@/components/shared/empty-state';
import { Skeleton } from '@/components/ui/skeleton';
import { ActionsDropdown } from '@/components/shared/actions-dropdown';
import { ActionsList } from '@/features/pipeline/actions/actions-list';
import { cn, downloadBlob } from '@/lib/utils';
type PipelineTab = 'editor' | 'logs' | 'releases' | 'actions';

export function PipelinePage() {
  useTitle('Pipeline');
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

  // Fetch logs (always load, refresh when running)
  const { data: logsData = [] } = useQuery({
    queryKey: ['logs', projectId],
    queryFn: () => runtimeApi.logs(projectId!),
    enabled: !!projectId,
    refetchInterval: isRunning ? 2000 : false,
  });

  const logs: LogEntry[] = Array.isArray(logsData) ? logsData : [];

  // Actions
  const { data: actions = [] } = useQuery({
    queryKey: queryKeys.actions.all(projectId!),
    queryFn: () => actionsApi.list(projectId!),
    enabled: !!projectId,
  });

  const [actionStates, setActionStates] = useState<Map<string, ActionState>>(new Map());

  // WebSocket for logs, status and action states
  // Always enabled to receive running status changes
  useWebSocket({
    projectId,
    enabled: !!projectId,
    onLog: () => {
      queryClient.refetchQueries({ queryKey: ['logs', projectId] });
    },
    onRunning: () => {
      // Refetch project data when running status changes
      queryClient.invalidateQueries({ queryKey: queryKeys.projects.detail(projectId!) });
    },
    onActions: (data: ActionRuntimeState[]) => {
      const newStates = new Map<string, ActionState>();
      data.forEach((item) => {
        newStates.set(item.slug, item.state);
      });
      setActionStates(newStates);
    },
  });

  // Auto-select branch on load (prioritize initialBranchName from state)
  useEffect(() => {
    if (branches.length > 0 && !selectedBranchId) {
      // First try to select the branch from location state
      if (initialBranchName) {
        const targetBranch = branches.find((b) => b.name === initialBranchName);
        if (targetBranch) {
          // eslint-disable-next-line react-hooks/set-state-in-effect -- intentional: initial selection based on external data
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

  // Fetch default service template
  const { data: defaultTemplate } = useQuery({
    queryKey: ['service-template'],
    queryFn: templatesApi.getServiceTemplate,
    staleTime: Infinity, // Template doesn't change during session
  });

  // Load branch code
  useEffect(() => {
    if (currentBranch) {
      // eslint-disable-next-line react-hooks/set-state-in-effect -- intentional: sync code state with fetched branch data
      setCode(currentBranch.code);

      setHasChanges(false);
    } else if (branches.length === 0 && defaultTemplate) {

      setCode(defaultTemplate);

      setHasChanges(true);
    }
  }, [currentBranch, branches.length, defaultTemplate]);

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

  const restartDebugMutation = useMutation({
    mutationFn: (branchName: string) => runtimeApi.restart(projectId!, { branch: branchName }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project', projectId] });
      queryClient.invalidateQueries({ queryKey: ['logs', projectId] });
      toast.success('Service restarted');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to restart');
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
    setHasChanges(value !== (currentBranch?.code || defaultTemplate || ''));
  };

  const handleDownloadLogs = async () => {
    try {
      const blob = await runtimeApi.downloadLogs(projectId!);
      await downloadBlob(blob, `${project?.slug || projectId}-logs.txt`);
    } catch (err) {
      console.error('Failed to download logs:', err);
    }
  };

  const selectedBranchSummary = branches.find((b) => b.id === selectedBranchId);
  const isDevelopBranch = currentBranch?.name === 'develop';
  const isLoading = branchesLoading || releasesLoading || (selectedBranchId && branchLoading);

  // Keyboard shortcuts handlers
  const handleSave = useCallback(() => {
    if (hasChanges && !saveMutation.isPending) {
      saveMutation.mutate();
    }
  }, [hasChanges, saveMutation]);

  const handleRunOrRestart = useCallback(async () => {
    if (!currentBranch || saveMutation.isPending || startDebugMutation.isPending || restartDebugMutation.isPending) return;
    // Can't start if another release is running
    if (isRunning && !isDebugMode) return;

    if (hasChanges) {
      await saveMutation.mutateAsync();
    }
    if (runningBranch === currentBranch.name) {
      // Restart with the same branch
      restartDebugMutation.mutate(currentBranch.name);
    } else {
      // Start new debug session
      startDebugMutation.mutate(currentBranch.name);
    }
  }, [currentBranch, hasChanges, runningBranch, isRunning, isDebugMode, saveMutation, startDebugMutation, restartDebugMutation]);

  const handleStop = useCallback(() => {
    if (isDebugMode && !stopDebugMutation.isPending) {
      stopDebugMutation.mutate();
    }
  }, [isDebugMode, stopDebugMutation]);

  // Key bindings for Monaco editor
  const keyBindings = useMemo(() => [
    { key: 'ctrl+s', label: 'Save', action: handleSave },
    { key: 'ctrl+,', label: 'Run / Restart', action: handleRunOrRestart },
    { key: 'ctrl+.', label: 'Stop', action: handleStop },
  ], [handleSave, handleRunOrRestart, handleStop]);

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
        <EditorTabs>
          <EditorTab
            active={activeTab === 'editor'}
            onClick={() => setActiveTab('editor')}
            icon={<Code className="size-4" />}
            dirty={hasChanges}
          >
            Editor
          </EditorTab>

          <EditorTab
            active={activeTab === 'logs'}
            onClick={() => setActiveTab('logs')}
            icon={<ScrollText className="size-4" />}
            badge={
              logs.filter(l => l.level === 'error').length > 0 ? (
                <Badge variant="destructive" className="ml-1 h-5 px-1.5 text-xs">
                  {logs.filter(l => l.level === 'error').length}
                </Badge>
              ) : isDebugMode && runningBranch === currentBranch?.name ? (
                <Badge variant="outline" className="ml-1 border-amber-500/50 text-amber-500 text-[10px] h-[20px] py-0">
                  {runningBranch}
                </Badge>
              ) : undefined
            }
          >
            Logs
          </EditorTab>

          <EditorTab
            active={activeTab === 'releases'}
            onClick={() => setActiveTab('releases')}
            icon={<Tag className="size-4" />}
            badge={
              releases.length > 0 ? (
                <Badge variant="secondary" className="ml-1 text-xs">
                  {releases.length}
                </Badge>
              ) : undefined
            }
          >
            Releases
          </EditorTab>

          <EditorTab
            active={activeTab === 'actions'}
            onClick={() => setActiveTab('actions')}
            icon={<Zap className="size-4" />}
            badge={
              actions.length > 0 ? (
                <Badge variant="secondary" className="ml-1 text-xs">
                  {actions.length}
                </Badge>
              ) : undefined
            }
          >
            Actions
          </EditorTab>
        </EditorTabs>

        <Card
            className={cn("flex flex-col gap-0 rounded-t-none py-0 overflow-hidden", ['releases', 'actions'].includes(activeTab) ? 'max-w-4xl h-auto': '')}
            style={{ height: ['editor', 'logs'].includes(activeTab) ? 'calc(100vh - 120px)' : 'auto', maxHeight: ['releases', 'actions'].includes(activeTab) ? 'calc(100vh - 120px)' : 'auto' }}
        >
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
                  {/* Debug Run/Stop/Restart Buttons */}
                  {currentBranch && (
                    runningBranch === currentBranch.name ? (
                      <div className="flex items-center gap-2">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={handleRunOrRestart}
                          disabled={restartDebugMutation.isPending || saveMutation.isPending}
                          className="border-amber-500/50 text-amber-600 hover:bg-amber-500/10"
                        >
                          <RotateCcw className={cn("mr-2 size-4", restartDebugMutation.isPending && "animate-spin")} />
                          Restart
                          <Kbd className="ml-2">^,</Kbd>
                        </Button>
                        <Button
                          variant="destructive"
                          size="sm"
                          onClick={handleStop}
                          disabled={stopDebugMutation.isPending}
                        >
                          <Square className="mr-2 size-4" />
                          Stop
                          <Kbd className="ml-2">^.</Kbd>
                        </Button>
                        {actions.length > 0 && project?.slug && (
                          <ActionsDropdown
                            projectSlug={project.slug}
                            actions={actions}
                            actionStates={actionStates}
                          />
                        )}
                      </div>
                    ) : (
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={handleRunOrRestart}
                        disabled={saveMutation.isPending || startDebugMutation.isPending || (isRunning && !isDebugMode)}
                        className="border-amber-500/50 text-amber-600 hover:bg-amber-500/10"
                      >
                        <Bug className="mr-2 size-4" />
                        Run Debug
                        <Kbd className="ml-2">^,</Kbd>
                      </Button>
                    )
                  )}
                  <LoadingButton onClick={() => saveMutation.mutate()} disabled={!hasChanges} loading={saveMutation.isPending}>
                    <Save className="mr-2 size-4" />
                    Save
                    <Kbd className="ml-2">^S</Kbd>
                  </LoadingButton>
                </div>
              </div>
              {/* Code Editor */}
              <div className="flex-1 min-h-0">
                <CodeEditor
                  value={code}
                  onChange={handleCodeChange}
                  language="javascript"
                  typeDefinitions={runtimeTypes}
                  keyBindings={keyBindings}
                />
              </div>
            </>
          )}

          {/* Logs Content */}
          {activeTab === 'logs' && (
            <LogsViewer
              logs={logs}
              limit={500}
              emptyMessage={isRunning ? "Waiting for logs..." : "No logs available. Start the service to see logs."}
              onDownload={handleDownloadLogs}
              height="100%"
              className="flex-1"
            />
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
                          'flex items-center justify-between p-3 rounded-lg bg-background border',
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
                            {new Date(release.created_at).toLocaleString()}
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

          {/* Actions Content */}
          {activeTab === 'actions' && (
            <ActionsList projectId={projectId!} />
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
            <LoadingButton
              onClick={handleCreateBranch}
              disabled={!newBranchName || !newBranchSourceName}
              loading={createBranchMutation.isPending}
            >
              Create
            </LoadingButton>
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
