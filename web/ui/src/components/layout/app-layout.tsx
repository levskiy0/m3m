import React from 'react';
import { Outlet, useLocation, useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { Star, Github } from 'lucide-react';

import { projectsApi, versionApi } from '@/api';
import { AppSidebar } from './app-sidebar';
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb';
import { Separator } from '@/components/ui/separator';
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from '@/components/ui/sidebar';
import { Toaster } from '@/components/ui/sonner';

interface BreadcrumbItem {
  label: string;
  href?: string;
}

function useBreadcrumbs(): BreadcrumbItem[] {
  const location = useLocation();
  const { projectId, modelId } = useParams();
  const { data: projects } = useQuery({
    queryKey: ['projects'],
    queryFn: projectsApi.list,
    enabled: !!projectId,
  });

  const project = projects?.find((p) => p.id === projectId);
  const pathSegments = location.pathname.split('/').filter(Boolean);

  const breadcrumbs: BreadcrumbItem[] = [];

  // Build breadcrumbs based on path
  if (pathSegments[0] === 'projects') {
    if (projectId && project) {
      breadcrumbs.push({ label: 'Projects', href: '/projects' });

      if (pathSegments.length === 2) {
        // /projects/:id
        breadcrumbs.push({ label: project.name });
      } else {
        breadcrumbs.push({ label: project.name, href: `/projects/${projectId}` });

        // Handle sub-pages
        const subPage = pathSegments[2];
        switch (subPage) {
          case 'pipeline':
            breadcrumbs.push({ label: 'Pipeline' });
            break;
          case 'storage':
            breadcrumbs.push({ label: 'Storage' });
            break;
          case 'models':
            if (modelId) {
              breadcrumbs.push({ label: 'Models', href: `/projects/${projectId}/models` });
              const subSubPage = pathSegments[4];
              if (subSubPage === 'schema') {
                breadcrumbs.push({ label: 'Schema' });
              } else if (subSubPage === 'data') {
                breadcrumbs.push({ label: 'Data' });
              }
            } else {
              breadcrumbs.push({ label: 'Models' });
            }
            break;
          case 'goals':
            breadcrumbs.push({ label: 'Goals' });
            break;
          case 'environment':
            breadcrumbs.push({ label: 'Environment' });
            break;
          case 'logs':
            breadcrumbs.push({ label: 'Logs' });
            break;
          case 'settings':
            breadcrumbs.push({ label: 'Settings' });
            break;
        }
      }
    } else {
      breadcrumbs.push({ label: 'Projects' });
    }
  } else if (pathSegments[0] === 'docs') {
    breadcrumbs.push({ label: 'Documentation' });
  } else if (pathSegments[0] === 'system') {
    breadcrumbs.push({ label: 'System Info' });
  } else if (pathSegments[0] === 'settings') {
    breadcrumbs.push({ label: 'Settings', href: '/settings/profile' });
    if (pathSegments[1] === 'users') {
      breadcrumbs.push({ label: 'Users' });
    } else if (pathSegments[1] === 'profile') {
      breadcrumbs.push({ label: 'Profile' });
    }
  }

  return breadcrumbs;
}

export function AppLayout() {
  const breadcrumbs = useBreadcrumbs();

  const { data: versionInfo } = useQuery({
    queryKey: ['version'],
    queryFn: versionApi.get,
    staleTime: Infinity,
  });

  return (
    <SidebarProvider>
      <AppSidebar />
      <SidebarInset>
        <header className="sticky top-0 z-10 bg-background flex h-16 shrink-0 items-center justify-between gap-2 transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12">
          <div className="flex items-center gap-2 px-4">
            <SidebarTrigger className="-ml-1" />
            <Separator
              orientation="vertical"
              className="mr-2 data-[orientation=vertical]:h-4"
            />
            <Breadcrumb>
              <BreadcrumbList>
                {breadcrumbs.map((item, index) => (
                  <React.Fragment key={index}>
                    {index > 0 && <BreadcrumbSeparator />}
                    <BreadcrumbItem>
                      {item.href ? (
                        <BreadcrumbLink href={item.href}>{item.label}</BreadcrumbLink>
                      ) : (
                        <BreadcrumbPage>{item.label}</BreadcrumbPage>
                      )}
                    </BreadcrumbItem>
                  </React.Fragment>
                ))}
              </BreadcrumbList>
            </Breadcrumb>
          </div>

          {/* Right section: Server Time + Version + GitHub */}
          <div className="flex items-center gap-3 px-4">
            <div className="flex items-center gap-2 text-sm">
              <span className="font-semibold">{versionInfo?.name || 'M3M'}</span>
              {versionInfo?.version && (
                <span className="text-muted-foreground bg-muted px-1.5 py-0.5 rounded text-xs">
                  {versionInfo.version}
                </span>
              )}
            </div>
            <Separator orientation="vertical" className="h-4" />
            <a
              href="https://github.com/levskiy0/m3m"
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-2 px-2.5 py-1 border rounded-md text-sm hover:bg-muted transition-colors"
            >
              <Github className="h-4 w-4" />
              <Star className="h-3.5 w-3.5 text-yellow-500" />
            </a>
          </div>
        </header>
        <main className="relative flex flex-1 flex-col gap-4 p-4 pt-0">
          <Outlet/>
        </main>
      </SidebarInset>
      <Toaster/>
    </SidebarProvider>
  );
}
