import { useQuery } from '@tanstack/react-query';
import { Package, Activity, ExternalLink } from 'lucide-react';

import { runtimeApi } from '@/api';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';

export function SystemInfoPage() {
  const { data: systemInfo, isLoading } = useQuery({
    queryKey: ['system-info'],
    queryFn: runtimeApi.getSystemInfo,
  });

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div>
          <Skeleton className="h-8 w-48" />
          <Skeleton className="mt-2 h-4 w-72" />
        </div>
        <div className="space-y-3">
          <Skeleton className="h-16" />
          <Skeleton className="h-16" />
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Loaded Modules</h1>
        <p className="text-muted-foreground">
          {systemInfo?.plugins?.length || 0} modules loaded
        </p>
      </div>

      {systemInfo?.plugins && systemInfo.plugins.length > 0 ? (
        <div className="space-y-3">
          {systemInfo.plugins.map((plugin) => (
            <div
              key={plugin.name}
              className="flex items-center justify-between rounded-lg border p-4"
            >
              <div className="flex items-center gap-3">
                <div className="flex h-10 w-10 items-center justify-center rounded-md bg-primary/10">
                  <Activity className="h-5 w-5 text-primary" />
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
        <div className="flex flex-col items-center justify-center rounded-lg border border-dashed py-12 text-center">
          <Package className="h-12 w-12 text-muted-foreground/50" />
          <p className="mt-2 text-sm text-muted-foreground">
            No modules loaded
          </p>
          <p className="text-xs text-muted-foreground">
            Place .so files in the plugins directory to load them
          </p>
        </div>
      )}
    </div>
  );
}
