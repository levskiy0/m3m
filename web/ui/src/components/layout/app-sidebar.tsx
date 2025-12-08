import { useMemo, useState } from 'react';
import { Link, useLocation, useNavigate, useParams } from 'react-router-dom';
import {
  FolderCode,
  Settings2,
  Target,
  Users,
  User,
  LogOut,
  Box,
  Code,
  HardDrive,
  Database,
  Variable,
  ChevronsUpDown,
  Plus,
  Globe,
  Package,
  Sun,
  Moon,
  Monitor,
  ChevronRight,
  Table2,
  BookOpen,
} from 'lucide-react';
import { useQuery } from '@tanstack/react-query';
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible';

import { useAuth } from '@/providers/auth-provider';
import { useTheme } from '@/providers/theme-provider';
import { projectsApi, modelsApi } from '@/api';
import { queryKeys } from '@/lib/query-keys';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenuSub,
  SidebarMenuSubItem,
  SidebarMenuSubButton,
  SidebarRail,
  useSidebar,
} from '@/components/ui/sidebar';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';

interface NavItem {
  title: string;
  url: string;
  icon: React.ElementType;
  isActive?: boolean;
  items?: { title: string; url: string }[];
}

function getInitials(name: string): string {
  return name
    .split(' ')
    .map((n) => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2);
}

const LAST_PROJECT_KEY = 'm3m_last_project';

export function AppSidebar() {
  const location = useLocation();
  const navigate = useNavigate();
  const { projectId: routeProjectId } = useParams();
  const { user, logout } = useAuth();
  const { isMobile } = useSidebar();
  const { theme, setTheme } = useTheme();

  const { data: projects = [] } = useQuery({
    queryKey: ['projects'],
    queryFn: projectsApi.list,
  });

  // Compute selected project ID from route or localStorage
  const selectedProjectId = useMemo(() => {
    if (routeProjectId) {
      localStorage.setItem(LAST_PROJECT_KEY, routeProjectId);
      return routeProjectId;
    }
    const storedId = localStorage.getItem(LAST_PROJECT_KEY);
    if (storedId && projects.some(p => p.id === storedId)) {
      return storedId;
    }
    return null;
  }, [routeProjectId, projects]);

  // Load models for current project
  const { data: models = [] } = useQuery({
    queryKey: queryKeys.models.all(selectedProjectId || ''),
    queryFn: () => modelsApi.list(selectedProjectId!),
    enabled: !!selectedProjectId,
  });

  // Track Data Storage collapsible state
  const [dataStorageOpen, setDataStorageOpen] = useState(true);

  const currentProject = projects?.find((p) => p.id === selectedProjectId);

  // Navigation items grouped by category
  const coreNavItems: NavItem[] = currentProject
    ? [
        { title: 'Overview', url: `/projects/${selectedProjectId}`, icon: Box },
        { title: 'Pipeline', url: `/projects/${selectedProjectId}/pipeline`, icon: Code },
      ]
    : [];

  const storageNavItems: NavItem[] = currentProject
    ? [
        { title: 'File Storage', url: `/projects/${selectedProjectId}/storage`, icon: HardDrive },
      ]
    : [];

  const metricsNavItems: NavItem[] = currentProject
    ? [
        { title: 'Goals', url: `/projects/${selectedProjectId}/goals`, icon: Target },
      ]
    : [];

  const settingsNavItems: NavItem[] = currentProject
    ? [
        { title: 'Environment', url: `/projects/${selectedProjectId}/environment`, icon: Variable },
        { title: 'Settings', url: `/projects/${selectedProjectId}/settings`, icon: Settings2 },
      ]
    : [];

  // Check if Data Storage section is active
  const isDataStorageActive = location.pathname.startsWith(`/projects/${selectedProjectId}/models`);

  const isAdmin = user?.permissions?.manageUsers || user?.isRoot;

  const handleProjectSelect = (id: string) => {
    localStorage.setItem(LAST_PROJECT_KEY, id);
    navigate(`/projects/${id}`);
  };

  const handleNewProject = () => {
    navigate('/projects', { state: { openCreate: true } });
  };

  return (
    <Sidebar collapsible="icon">
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <SidebarMenuButton
                  size="lg"
                  className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
                >
                  {currentProject ? (
                    <>
                      <div
                        className="flex aspect-square size-8 items-center justify-center rounded-lg"
                        style={{ backgroundColor: currentProject.color || '#6b7280' }}
                      >
                        <FolderCode className="size-4 text-white" />
                      </div>
                      <div className="grid flex-1 text-left text-sm leading-tight">
                        <span className="truncate font-semibold">{currentProject.name}</span>
                        <span className="truncate text-xs text-muted-foreground">
                          {currentProject.status === 'running' ? 'Running' : 'Stopped'}
                        </span>
                      </div>
                    </>
                  ) : (
                    <>
                      <div className="bg-primary text-primary-foreground flex aspect-square size-8 items-center justify-center rounded-lg">
                        <FolderCode className="size-4" />
                      </div>
                      <div className="grid flex-1 text-left text-sm leading-tight">
                        <span className="truncate font-semibold">M3M</span>
                        <span className="truncate text-xs">Select project</span>
                      </div>
                    </>
                  )}
                  <ChevronsUpDown className="ml-auto" />
                </SidebarMenuButton>
              </DropdownMenuTrigger>
              <DropdownMenuContent
                className="w-(--radix-dropdown-menu-trigger-width) min-w-56 rounded-lg"
                align="start"
                side={isMobile ? 'bottom' : 'right'}
                sideOffset={4}
              >
                <DropdownMenuLabel className="text-muted-foreground text-xs">
                  Projects
                </DropdownMenuLabel>
                {projects?.map((project) => (
                  <DropdownMenuItem
                    key={project.id}
                    onClick={() => handleProjectSelect(project.id)}
                    className="gap-2 p-2"
                  >
                    <div
                      className="flex size-6 items-center justify-center rounded-md"
                      style={{ backgroundColor: project.color || '#6b7280' }}
                    >
                      <FolderCode className="size-3.5 shrink-0 text-white" />
                    </div>
                    <span className="flex-1">{project.name}</span>
                    {project.status === 'running' && (
                      <span className="size-2 rounded-full bg-green-500" />
                    )}
                  </DropdownMenuItem>
                ))}
                <DropdownMenuSeparator />
                <DropdownMenuItem onClick={handleNewProject} className="gap-2 p-2">
                  <div className="flex size-6 items-center justify-center rounded-md border bg-transparent">
                    <Plus className="size-4" />
                  </div>
                  <span className="text-muted-foreground font-medium">New Project</span>
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>

      <SidebarContent>
        {/* Core Navigation */}
        {currentProject && coreNavItems.length > 0 && (
          <SidebarGroup>
            <SidebarGroupLabel>Project</SidebarGroupLabel>
            <SidebarMenu>
              {coreNavItems.map((item) => (
                <SidebarMenuItem key={item.title}>
                  <SidebarMenuButton
                    asChild
                    isActive={location.pathname === item.url}
                    tooltip={item.title}
                  >
                    <Link to={item.url}>
                      <item.icon />
                      <span>{item.title}</span>
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroup>
        )}

        {/* Storage Section */}
        {currentProject && (
          <SidebarGroup>
            <SidebarGroupLabel>Storage</SidebarGroupLabel>
            <SidebarMenu>
              {/* Data Storage with collapsible models */}
              <Collapsible
                asChild
                open={dataStorageOpen}
                onOpenChange={setDataStorageOpen}
                className="group/collapsible"
              >
                <SidebarMenuItem>
                  <CollapsibleTrigger asChild>
                    <SidebarMenuButton
                      tooltip="Data Storage"
                      isActive={isDataStorageActive}
                    >
                      <Database />
                      <span>Data Storage</span>
                      <ChevronRight className="ml-auto transition-transform duration-200 group-data-[state=open]/collapsible:rotate-90" />
                    </SidebarMenuButton>
                  </CollapsibleTrigger>
                  <CollapsibleContent>
                    <SidebarMenuSub>
                      <SidebarMenuSubItem>
                        <SidebarMenuSubButton
                          asChild
                          isActive={location.pathname === `/projects/${selectedProjectId}/models`}
                        >
                          <Link to={`/projects/${selectedProjectId}/models`}>
                            <span>Manage Storage</span>
                          </Link>
                        </SidebarMenuSubButton>
                      </SidebarMenuSubItem>
                      {models.map((model) => (
                        <SidebarMenuSubItem key={model.id}>
                          <SidebarMenuSubButton
                            asChild
                            isActive={
                              location.pathname === `/projects/${selectedProjectId}/models/${model.id}/data` ||
                              location.pathname === `/projects/${selectedProjectId}/models/${model.id}/schema`
                            }
                          >
                            <Link to={`/projects/${selectedProjectId}/models/${model.id}/data`}>
                              <Table2 className="size-3" />
                              <span>{model.name}</span>
                            </Link>
                          </SidebarMenuSubButton>
                        </SidebarMenuSubItem>
                      ))}
                    </SidebarMenuSub>
                  </CollapsibleContent>
                </SidebarMenuItem>
              </Collapsible>

              {/* File Storage */}
              {storageNavItems.map((item) => (
                <SidebarMenuItem key={item.title}>
                  <SidebarMenuButton
                    asChild
                    isActive={location.pathname === item.url}
                    tooltip={item.title}
                  >
                    <Link to={item.url}>
                      <item.icon />
                      <span>{item.title}</span>
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroup>
        )}

        {/* Metrics Section */}
        {currentProject && metricsNavItems.length > 0 && (
          <SidebarGroup>
            <SidebarGroupLabel>Metrics</SidebarGroupLabel>
            <SidebarMenu>
              {metricsNavItems.map((item) => (
                <SidebarMenuItem key={item.title}>
                  <SidebarMenuButton
                    asChild
                    isActive={location.pathname === item.url}
                    tooltip={item.title}
                  >
                    <Link to={item.url}>
                      <item.icon />
                      <span>{item.title}</span>
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroup>
        )}

        {/* Settings Section */}
        {currentProject && settingsNavItems.length > 0 && (
          <SidebarGroup>
            <SidebarGroupLabel>Configuration</SidebarGroupLabel>
            <SidebarMenu>
              {settingsNavItems.map((item) => (
                <SidebarMenuItem key={item.title}>
                  <SidebarMenuButton
                    asChild
                    isActive={location.pathname === item.url}
                    tooltip={item.title}
                  >
                    <Link to={item.url}>
                      <item.icon />
                      <span>{item.title}</span>
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroup>
        )}

        {/* Global Section */}
        <SidebarGroup>
          <SidebarGroupLabel>Global</SidebarGroupLabel>
          <SidebarMenu>
            <SidebarMenuItem>
              <SidebarMenuButton
                asChild
                isActive={location.pathname === '/goals'}
                tooltip="Global Goals"
              >
                <Link to="/goals">
                  <Globe />
                  <span>Global Goals</span>
                </Link>
              </SidebarMenuButton>
            </SidebarMenuItem>
            <SidebarMenuItem>
              <SidebarMenuButton
                asChild
                isActive={location.pathname === '/modules'}
                tooltip="Modules"
              >
                <Link to="/modules">
                  <Package />
                  <span>Modules</span>
                </Link>
              </SidebarMenuButton>
            </SidebarMenuItem>
            <SidebarMenuItem>
              <SidebarMenuButton
                asChild
                isActive={location.pathname === '/docs'}
                tooltip="Documentation"
              >
                <Link to="/docs">
                  <BookOpen />
                  <span>Documentation</span>
                </Link>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarGroup>

        {/* Admin Settings - only show if user is admin */}
        {isAdmin && (
          <SidebarGroup>
            <SidebarGroupLabel>Admin</SidebarGroupLabel>
            <SidebarMenu>
              <SidebarMenuItem>
                <SidebarMenuButton
                  asChild
                  isActive={location.pathname === '/settings/users'}
                  tooltip="Users"
                >
                  <Link to="/settings/users">
                    <Users />
                    <span>Users</span>
                  </Link>
                </SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarGroup>
        )}
      </SidebarContent>

      <SidebarFooter>
        <SidebarMenu>
          <SidebarMenuItem>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <SidebarMenuButton
                  size="lg"
                  className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
                >
                  <Avatar className="h-8 w-8 rounded-lg">
                    <AvatarImage src={user?.avatar} alt={user?.name} />
                    <AvatarFallback className="rounded-lg">
                      {user?.name ? getInitials(user.name) : 'U'}
                    </AvatarFallback>
                  </Avatar>
                  <div className="grid flex-1 text-left text-sm leading-tight">
                    <span className="truncate font-medium">{user?.name}</span>
                    <span className="truncate text-xs">{user?.email}</span>
                  </div>
                  <ChevronsUpDown className="ml-auto size-4" />
                </SidebarMenuButton>
              </DropdownMenuTrigger>
              <DropdownMenuContent
                className="w-(--radix-dropdown-menu-trigger-width) min-w-56 rounded-lg"
                side="right"
                align="end"
                sideOffset={4}
              >
                <DropdownMenuLabel className="p-0 font-normal">
                  <div className="flex items-center gap-2 px-1 py-1.5 text-left text-sm">
                    <Avatar className="h-8 w-8 rounded-lg">
                      <AvatarImage src={user?.avatar} alt={user?.name} />
                      <AvatarFallback className="rounded-lg">
                        {user?.name ? getInitials(user.name) : 'U'}
                      </AvatarFallback>
                    </Avatar>
                    <div className="grid flex-1 text-left text-sm leading-tight">
                      <span className="truncate font-medium">{user?.name}</span>
                      <span className="truncate text-xs">{user?.email}</span>
                    </div>
                  </div>
                </DropdownMenuLabel>
                <DropdownMenuSeparator />
                <DropdownMenuGroup>
                  <DropdownMenuItem asChild>
                    <Link to="/settings/profile">
                      <User />
                      Profile
                    </Link>
                  </DropdownMenuItem>
                </DropdownMenuGroup>
                <DropdownMenuSeparator />
                <DropdownMenuLabel className="text-muted-foreground text-xs">
                  Theme
                </DropdownMenuLabel>
                <DropdownMenuGroup>
                  <DropdownMenuItem onClick={() => setTheme('light')}>
                    <Sun />
                    Light
                    {theme === 'light' && <span className="ml-auto text-xs">*</span>}
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={() => setTheme('dark')}>
                    <Moon />
                    Dark
                    {theme === 'dark' && <span className="ml-auto text-xs">*</span>}
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={() => setTheme('system')}>
                    <Monitor />
                    System
                    {theme === 'system' && <span className="ml-auto text-xs">*</span>}
                  </DropdownMenuItem>
                </DropdownMenuGroup>
                <DropdownMenuSeparator />
                <DropdownMenuItem onClick={logout}>
                  <LogOut />
                  Log out
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  );
}
