import { useEffect, useState } from 'react';
import { Link, useLocation, useParams } from 'react-router-dom';
import {
  ChevronRight,
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
  ScrollText,
  ChevronsUpDown,
  Plus,
} from 'lucide-react';
import { useQuery } from '@tanstack/react-query';

import { useAuth } from '@/providers/auth-provider';
import { projectsApi } from '@/api';
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible';
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
  SidebarMenuSubButton,
  SidebarMenuSubItem,
  SidebarRail,
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

export function AppSidebar() {
  const location = useLocation();
  const { projectId } = useParams();
  const { user, logout } = useAuth();

  const { data: projects = [] } = useQuery({
    queryKey: ['projects'],
    queryFn: projectsApi.list,
  });

  const [openProjects, setOpenProjects] = useState(true);
  const [openSettings, setOpenSettings] = useState(false);

  const currentProject = projects.find((p) => p.id === projectId);

  // Navigation items for current project
  const projectNavItems: NavItem[] = currentProject
    ? [
        { title: 'Overview', url: `/projects/${projectId}`, icon: Box },
        { title: 'Pipeline', url: `/projects/${projectId}/pipeline`, icon: Code },
        { title: 'Storage', url: `/projects/${projectId}/storage`, icon: HardDrive },
        { title: 'Models', url: `/projects/${projectId}/models`, icon: Database },
        { title: 'Goals', url: `/projects/${projectId}/goals`, icon: Target },
        { title: 'Environment', url: `/projects/${projectId}/environment`, icon: Variable },
        { title: 'Logs', url: `/projects/${projectId}/logs`, icon: ScrollText },
        { title: 'Settings', url: `/projects/${projectId}/settings`, icon: Settings2 },
      ]
    : [];

  const isAdmin = user?.permissions?.manageUsers || user?.isRoot;

  useEffect(() => {
    if (location.pathname.includes('/settings')) {
      setOpenSettings(true);
    }
    if (location.pathname.includes('/projects')) {
      setOpenProjects(true);
    }
  }, [location.pathname]);

  return (
    <Sidebar collapsible="icon">
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton size="lg" asChild>
              <Link to="/projects">
                <div className="bg-primary text-primary-foreground flex aspect-square size-8 items-center justify-center rounded-lg">
                  <FolderCode className="size-4" />
                </div>
                <div className="grid flex-1 text-left text-sm leading-tight">
                  <span className="truncate font-semibold">M3M</span>
                  <span className="truncate text-xs">Mini Services</span>
                </div>
              </Link>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>

      <SidebarContent>
        {/* Current Project Navigation */}
        {currentProject && projectNavItems.length > 0 && (
          <SidebarGroup>
            <SidebarGroupLabel>
              <span
                className="mr-2 inline-block size-2 rounded-full"
                style={{ backgroundColor: currentProject.color || '#6b7280' }}
              />
              {currentProject.name}
            </SidebarGroupLabel>
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

        {/* Projects */}
        <SidebarGroup>
          <Collapsible open={openProjects} onOpenChange={setOpenProjects}>
            <SidebarMenuItem>
              <CollapsibleTrigger asChild>
                <SidebarMenuButton tooltip="Projects">
                  <FolderCode />
                  <span>Projects</span>
                  <ChevronRight className="ml-auto transition-transform duration-200 group-data-[state=open]/collapsible:rotate-90" />
                </SidebarMenuButton>
              </CollapsibleTrigger>
              <CollapsibleContent>
                <SidebarMenuSub>
                  {projects.map((project) => (
                    <SidebarMenuSubItem key={project.id}>
                      <SidebarMenuSubButton
                        asChild
                        isActive={projectId === project.id}
                      >
                        <Link to={`/projects/${project.id}`}>
                          <span
                            className="mr-2 inline-block size-2 rounded-full"
                            style={{ backgroundColor: project.color || '#6b7280' }}
                          />
                          <span>{project.name}</span>
                          {project.status === 'running' && (
                            <span className="ml-auto size-2 rounded-full bg-green-500" />
                          )}
                        </Link>
                      </SidebarMenuSubButton>
                    </SidebarMenuSubItem>
                  ))}
                  <SidebarMenuSubItem>
                    <SidebarMenuSubButton asChild>
                      <Link
                        to="/projects"
                        className="text-muted-foreground"
                        state={{ openCreate: true }}
                      >
                        <Plus className="size-4" />
                        <span>New Project</span>
                      </Link>
                    </SidebarMenuSubButton>
                  </SidebarMenuSubItem>
                </SidebarMenuSub>
              </CollapsibleContent>
            </SidebarMenuItem>
          </Collapsible>
        </SidebarGroup>

        {/* Global Goals */}
        <SidebarGroup>
          <SidebarMenu>
            <SidebarMenuItem>
              <SidebarMenuButton
                asChild
                isActive={location.pathname === '/goals'}
                tooltip="Goals"
              >
                <Link to="/goals">
                  <Target />
                  <span>Global Goals</span>
                </Link>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarGroup>

        {/* Settings */}
        <SidebarGroup>
          <Collapsible open={openSettings} onOpenChange={setOpenSettings}>
            <SidebarMenuItem>
              <CollapsibleTrigger asChild>
                <SidebarMenuButton tooltip="Settings">
                  <Settings2 />
                  <span>Settings</span>
                  <ChevronRight className="ml-auto transition-transform duration-200 group-data-[state=open]/collapsible:rotate-90" />
                </SidebarMenuButton>
              </CollapsibleTrigger>
              <CollapsibleContent>
                <SidebarMenuSub>
                  {isAdmin && (
                    <SidebarMenuSubItem>
                      <SidebarMenuSubButton
                        asChild
                        isActive={location.pathname === '/settings/users'}
                      >
                        <Link to="/settings/users">
                          <Users className="size-4" />
                          <span>Users</span>
                        </Link>
                      </SidebarMenuSubButton>
                    </SidebarMenuSubItem>
                  )}
                  <SidebarMenuSubItem>
                    <SidebarMenuSubButton
                      asChild
                      isActive={location.pathname === '/settings/profile'}
                    >
                      <Link to="/settings/profile">
                        <User className="size-4" />
                        <span>Profile</span>
                      </Link>
                    </SidebarMenuSubButton>
                  </SidebarMenuSubItem>
                </SidebarMenuSub>
              </CollapsibleContent>
            </SidebarMenuItem>
          </Collapsible>
        </SidebarGroup>
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
