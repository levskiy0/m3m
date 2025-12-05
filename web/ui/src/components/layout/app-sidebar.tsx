import { useEffect, useState } from 'react';
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
  Server,
} from 'lucide-react';
import { useQuery } from '@tanstack/react-query';

import { useAuth } from '@/providers/auth-provider';
import { projectsApi } from '@/api';
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

  const { data: projects = [] } = useQuery({
    queryKey: ['projects'],
    queryFn: projectsApi.list,
  });

  // Use route projectId or fallback to last selected project from localStorage
  const [selectedProjectId, setSelectedProjectId] = useState<string | null>(() => {
    return routeProjectId || localStorage.getItem(LAST_PROJECT_KEY);
  });

  // Update selected project when route changes
  useEffect(() => {
    if (routeProjectId) {
      setSelectedProjectId(routeProjectId);
      localStorage.setItem(LAST_PROJECT_KEY, routeProjectId);
    }
  }, [routeProjectId]);

  // Also update from localStorage when projects load (in case the stored project still exists)
  useEffect(() => {
    if (!routeProjectId && projects.length > 0) {
      const storedId = localStorage.getItem(LAST_PROJECT_KEY);
      if (storedId && projects.some(p => p.id === storedId)) {
        setSelectedProjectId(storedId);
      }
    }
  }, [projects, routeProjectId]);

  const currentProject = projects?.find((p) => p.id === selectedProjectId);

  // Navigation items for current project
  const projectNavItems: NavItem[] = currentProject
    ? [
        { title: 'Overview', url: `/projects/${selectedProjectId}`, icon: Box },
        { title: 'Pipeline', url: `/projects/${selectedProjectId}/pipeline`, icon: Code },
        { title: 'Storage', url: `/projects/${selectedProjectId}/storage`, icon: HardDrive },
        { title: 'Models', url: `/projects/${selectedProjectId}/models`, icon: Database },
        { title: 'Goals', url: `/projects/${selectedProjectId}/goals`, icon: Target },
        { title: 'Environment', url: `/projects/${selectedProjectId}/environment`, icon: Variable },
        { title: 'Settings', url: `/projects/${selectedProjectId}/settings`, icon: Settings2 },
      ]
    : [];

  const isAdmin = user?.permissions?.manageUsers || user?.isRoot;

  const handleProjectSelect = (id: string) => {
    setSelectedProjectId(id);
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
        {/* Current Project Navigation */}
        {currentProject && projectNavItems.length > 0 && (
          <SidebarGroup>
            <SidebarGroupLabel>Project</SidebarGroupLabel>
            <SidebarMenu>
              {projectNavItems.map((item) => (
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
