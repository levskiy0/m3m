import { useQuery } from '@tanstack/react-query';
import { Cpu, HardDrive, Package, Server, Activity, ExternalLink } from 'lucide-react';

import { runtimeApi } from '@/api';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

export function SystemInfoPage() {
  const { data: systemInfo, isLoading } = useQuery({
    queryKey: ['system-info'],
    queryFn: runtimeApi.getSystemInfo,
    refetchInterval: 5000,
  });

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div>
          <Skeleton className="h-8 w-48" />
          <Skeleton className="mt-2 h-4 w-72" />
        </div>
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {[1, 2, 3].map((i) => (
            <Skeleton key={i} className="h-40" />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">System Info</h1>
        <p className="text-muted-foreground">
          Runtime information and loaded modules
        </p>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {/* Runtime Info */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Runtime</CardTitle>
            <Server className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <div className="flex justify-between">
                <span className="text-muted-foreground text-sm">Version</span>
                <Badge variant="secondary">{systemInfo?.version}</Badge>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground text-sm">Go</span>
                <span className="text-sm font-medium">{systemInfo?.go_version}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground text-sm">Platform</span>
                <span className="text-sm font-medium">
                  {systemInfo?.go_os}/{systemInfo?.go_arch}
                </span>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* CPU Info */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">CPU</CardTitle>
            <Cpu className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <div className="flex justify-between">
                <span className="text-muted-foreground text-sm">Cores</span>
                <span className="text-sm font-medium">{systemInfo?.num_cpu}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground text-sm">Goroutines</span>
                <span className="text-sm font-medium">{systemInfo?.num_goroutine}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground text-sm">Running Projects</span>
                <Badge variant="outline">{systemInfo?.running_projects_count}</Badge>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Memory Info */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Memory</CardTitle>
            <HardDrive className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <div className="flex justify-between">
                <span className="text-muted-foreground text-sm">Allocated</span>
                <span className="text-sm font-medium">
                  {formatBytes(systemInfo?.memory?.alloc || 0)}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground text-sm">System</span>
                <span className="text-sm font-medium">
                  {formatBytes(systemInfo?.memory?.sys || 0)}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground text-sm">GC Cycles</span>
                <span className="text-sm font-medium">{systemInfo?.memory?.num_gc}</span>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Loaded Plugins */}
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Package className="h-5 w-5 text-muted-foreground" />
            <div>
              <CardTitle>Loaded Modules</CardTitle>
              <CardDescription>SO plugins loaded at runtime</CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {systemInfo?.plugins && systemInfo.plugins.length > 0 ? (
            <div className="space-y-3">
              {systemInfo.plugins.map((plugin) => (
                <div
                  key={plugin.name}
                  className="flex items-center justify-between rounded-lg border p-3"
                >
                  <div className="flex items-center gap-3">
                    <div className="flex h-9 w-9 items-center justify-center rounded-md bg-primary/10">
                      <Activity className="h-4 w-4 text-primary" />
                    </div>
                    <div>
                      <p className="font-medium">{plugin.name}</p>
                      {plugin.description && (
                        <p className="text-sm text-muted-foreground">
                          {plugin.description}
                        </p>
                      )}
                      {plugin.author && (
                        <p className="text-xs text-muted-foreground">
                          by {plugin.author}
                        </p>
                      )}
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    {plugin.url && (
                      <a
                        href={plugin.url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-muted-foreground hover:text-foreground transition-colors"
                      >
                        <ExternalLink className="h-4 w-4" />
                      </a>
                    )}
                    {plugin.version && (
                      <Badge variant="outline">{plugin.version}</Badge>
                    )}
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center py-8 text-center">
              <Package className="h-12 w-12 text-muted-foreground/50" />
              <p className="mt-2 text-sm text-muted-foreground">
                No plugins loaded
              </p>
              <p className="text-xs text-muted-foreground">
                Place .so files in the plugins directory to load them
              </p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
