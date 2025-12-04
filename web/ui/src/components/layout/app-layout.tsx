import { Outlet, useLocation, useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';

import { projectsApi } from '@/api';
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
  } else if (pathSegments[0] === 'goals') {
    breadcrumbs.push({ label: 'Global Goals' });
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

  return (
    <SidebarProvider>
      <AppSidebar />
      <SidebarInset>
        <header className="flex h-16 shrink-0 items-center gap-2 transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12">
          <div className="flex items-center gap-2 px-4">
            <SidebarTrigger className="-ml-1" />
            <Separator
              orientation="vertical"
              className="mr-2 data-[orientation=vertical]:h-4"
            />
            <Breadcrumb>
              <BreadcrumbList>
                {breadcrumbs.map((item, index) => (
                  <BreadcrumbItem key={index}>
                    {index > 0 && <BreadcrumbSeparator />}
                    {item.href ? (
                      <BreadcrumbLink href={item.href}>{item.label}</BreadcrumbLink>
                    ) : (
                      <BreadcrumbPage>{item.label}</BreadcrumbPage>
                    )}
                  </BreadcrumbItem>
                ))}
              </BreadcrumbList>
            </Breadcrumb>
          </div>
        </header>
        <main className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <Outlet />
        </main>
      </SidebarInset>
      <Toaster />
    </SidebarProvider>
  );
}
