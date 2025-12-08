import { Outlet, Link, useLocation } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { Search, Rocket, Code2, RefreshCw } from 'lucide-react';
import { useState, useMemo } from 'react';

import { runtimeApi } from '@/api';
import { Input } from '@/components/ui/input';
import { ScrollArea } from '@/components/ui/scroll-area';
import { cn } from '@/lib/utils';

const staticPages = [
  { id: 'getting-started', title: 'Getting Started', icon: Rocket },
  { id: 'lifecycle', title: 'Development Guide', icon: RefreshCw },
];

export function DocsLayout() {
  const location = useLocation();
  const [search, setSearch] = useState('');

  const { data: schemas = [] } = useQuery({
    queryKey: ['runtime-schemas'],
    queryFn: runtimeApi.getSchemas,
  });

  const filteredModules = useMemo(() => {
    if (!search) return schemas;
    const q = search.toLowerCase();
    return schemas.filter(
      (m) =>
        m.name.toLowerCase().includes(q) ||
        m.description.toLowerCase().includes(q)
    );
  }, [schemas, search]);

  const currentPath = location.pathname;

  return (
    <div className="flex gap-6">
      {/* Sidebar */}
      <div className="w-56 shrink-0">
        <div className="fixed w-56 top-20">
          <div className="relative mb-4">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              placeholder="Search..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="pl-9 h-9"
            />
          </div>

          <ScrollArea className="h-[calc(100vh-9rem)]">
            <div className="space-y-4 pr-2">
              {/* Static pages */}
              <div>
                <div className="text-xs font-medium text-muted-foreground mb-2 px-2">
                  Guides
                </div>
                <div className="space-y-0.5">
                  {staticPages.map((page) => (
                    <Link
                      key={page.id}
                      to={`/docs/${page.id}`}
                      className={cn(
                        'flex items-center gap-2 px-2 py-1.5 rounded-md text-sm transition-colors',
                        currentPath === `/docs/${page.id}`
                          ? 'bg-primary/10 text-primary font-medium'
                          : 'text-muted-foreground hover:bg-muted hover:text-foreground'
                      )}
                    >
                      <page.icon className="h-4 w-4" />
                      {page.title}
                    </Link>
                  ))}
                </div>
              </div>

              {/* Modules */}
              <div>
                <div className="text-xs font-medium text-muted-foreground mb-2 px-2">
                  API Reference
                </div>
                <div className="space-y-0.5">
                  {filteredModules.map((module) => (
                    <Link
                      key={module.name}
                      to={`/docs/api/${encodeURIComponent(module.name)}`}
                      className={cn(
                        'flex items-center gap-2 px-2 py-1.5 rounded-md text-sm font-mono transition-colors',
                        currentPath === `/docs/api/${encodeURIComponent(module.name)}`
                          ? 'bg-primary/10 text-primary font-medium'
                          : 'text-muted-foreground hover:bg-muted hover:text-foreground'
                      )}
                    >
                      <Code2 className="h-3.5 w-3.5" />
                      {module.name}
                    </Link>
                  ))}
                </div>
              </div>
            </div>
          </ScrollArea>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 min-w-0 max-w-3xl">
        <Outlet />
      </div>
    </div>
  );
}
