import { useState, useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Search, Book, Code2, Braces, ChevronRight, Copy, Check } from 'lucide-react';

import { runtimeApi } from '@/api';
import type { ModuleSchema, MethodSchema, TypeSchema } from '@/api/runtime';
import { useTitle } from '@/hooks';
import { Input } from '@/components/ui/input';
import { Skeleton } from '@/components/ui/skeleton';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';

function TypeBadge({ type, optional }: { type: string; optional?: boolean }) {
  const colors: Record<string, string> = {
    string: 'bg-green-500/10 text-green-600 dark:text-green-400',
    number: 'bg-blue-500/10 text-blue-600 dark:text-blue-400',
    boolean: 'bg-purple-500/10 text-purple-600 dark:text-purple-400',
    void: 'bg-gray-500/10 text-gray-600 dark:text-gray-400',
    any: 'bg-orange-500/10 text-orange-600 dark:text-orange-400',
    object: 'bg-yellow-500/10 text-yellow-600 dark:text-yellow-400',
  };

  const baseType = type.replace(/[[\]?]/g, '').split(' ')[0];
  const colorClass = colors[baseType] || 'bg-cyan-500/10 text-cyan-600 dark:text-cyan-400';

  return (
    <code className={cn('px-1.5 py-0.5 rounded text-xs font-mono', colorClass)}>
      {type}
      {optional && '?'}
    </code>
  );
}

function MethodSignature({ method, moduleName }: { method: MethodSchema; moduleName: string }) {
  const [copied, setCopied] = useState(false);

  const signature = useMemo(() => {
    const params = method.params?.map(p => `${p.name}${p.optional ? '?' : ''}`).join(', ') || '';
    return `${moduleName}.${method.name}(${params})`;
  }, [method, moduleName]);

  const handleCopy = () => {
    navigator.clipboard.writeText(signature);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="group relative">
      <pre className="bg-muted/50 rounded-lg p-3 text-sm font-mono overflow-x-auto">
        <code>{signature}</code>
      </pre>
      <Button
        variant="ghost"
        size="icon"
        className="absolute top-2 right-2 h-6 w-6 opacity-0 group-hover:opacity-100 transition-opacity"
        onClick={handleCopy}
      >
        {copied ? <Check className="h-3 w-3" /> : <Copy className="h-3 w-3" />}
      </Button>
    </div>
  );
}

function MethodCard({ method, moduleName }: { method: MethodSchema; moduleName: string }) {
  return (
    <div id={`${moduleName}-${method.name}`} className="border rounded-lg p-4 scroll-mt-20">
      <div className="flex items-center gap-2 mb-2">
        <Code2 className="h-4 w-4 text-primary" />
        <h4 className="font-semibold font-mono">{method.name}</h4>
        {method.returns && (
          <>
            <ChevronRight className="h-3 w-3 text-muted-foreground" />
            <TypeBadge type={method.returns.type} />
          </>
        )}
      </div>
      <p className="text-sm text-muted-foreground mb-3">{method.description}</p>

      <MethodSignature method={method} moduleName={moduleName} />

      {method.params && method.params.length > 0 && (
        <div className="mt-3">
          <h5 className="text-xs font-medium text-muted-foreground mb-2">Parameters</h5>
          <div className="space-y-2">
            {method.params.map((param) => (
              <div key={param.name} className="flex items-start gap-2 text-sm">
                <code className="font-mono text-foreground">{param.name}</code>
                <TypeBadge type={param.type} optional={param.optional} />
                <span className="text-muted-foreground">{param.description}</span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

function TypeCard({ type }: { type: TypeSchema }) {
  return (
    <div id={`type-${type.name}`} className="border rounded-lg p-4 scroll-mt-20">
      <div className="flex items-center gap-2 mb-2">
        <Braces className="h-4 w-4 text-primary" />
        <h4 className="font-semibold font-mono">{type.name}</h4>
      </div>
      {type.description && (
        <p className="text-sm text-muted-foreground mb-3">{type.description}</p>
      )}

      <div className="bg-muted/50 rounded-lg p-3 font-mono text-sm">
        <div className="text-muted-foreground">{'interface'} <span className="text-foreground">{type.name}</span> {'{'}</div>
        {type.fields.map((field) => (
          <div key={field.name} className="pl-4 flex items-center gap-2">
            <span className="text-foreground">{field.name}</span>
            <span className="text-muted-foreground">{field.optional ? '?:' : ':'}</span>
            <TypeBadge type={field.type} />
            {field.description && (
              <span className="text-muted-foreground text-xs">// {field.description}</span>
            )}
          </div>
        ))}
        <div className="text-muted-foreground">{'}'}</div>
      </div>
    </div>
  );
}

function ModuleSection({ module }: { module: ModuleSchema }) {
  return (
    <div id={module.name} className="scroll-mt-20">
      <div className="flex items-center gap-3 mb-4">
        <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10">
          <Book className="h-5 w-5 text-primary" />
        </div>
        <div>
          <h2 className="text-xl font-bold font-mono">{module.name}</h2>
          <p className="text-sm text-muted-foreground">{module.description}</p>
        </div>
      </div>

      {module.types && module.types.length > 0 && (
        <div className="mb-6">
          <h3 className="text-sm font-medium text-muted-foreground mb-3">Types</h3>
          <div className="space-y-3">
            {module.types.map((type) => (
              <TypeCard key={type.name} type={type} />
            ))}
          </div>
        </div>
      )}

      {module.methods.length > 0 && (
        <div className="mb-6">
          <h3 className="text-sm font-medium text-muted-foreground mb-3">Methods</h3>
          <div className="space-y-3">
            {module.methods.map((method) => (
              <MethodCard key={method.name} method={method} moduleName={module.name} />
            ))}
          </div>
        </div>
      )}

      {module.nested && module.nested.length > 0 && (
        <div className="mb-6">
          <h3 className="text-sm font-medium text-muted-foreground mb-3">Nested Modules</h3>
          {module.nested.map((nested) => (
            <div key={nested.name} className="border rounded-lg p-4 mb-3">
              <h4 className="font-semibold font-mono mb-2">{module.name}.{nested.name}</h4>
              <p className="text-sm text-muted-foreground mb-3">{nested.description}</p>
              <div className="space-y-3">
                {nested.methods.map((method) => (
                  <MethodCard
                    key={method.name}
                    method={method}
                    moduleName={`${module.name}.${nested.name}`}
                  />
                ))}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

function NavItem({
  module,
  isActive,
  onClick,
}: {
  module: ModuleSchema;
  isActive: boolean;
  onClick: () => void;
}) {
  return (
    <button
      onClick={onClick}
      className={cn(
        'w-full text-left px-3 py-2 rounded-md text-sm font-mono transition-colors',
        isActive
          ? 'bg-primary/10 text-primary'
          : 'text-muted-foreground hover:bg-muted hover:text-foreground'
      )}
    >
      {module.name}
    </button>
  );
}

export function DocsPage() {
  useTitle('Documentation');
  const [search, setSearch] = useState('');
  const [activeModule, setActiveModule] = useState<string | null>(null);

  const { data: schemas, isLoading } = useQuery({
    queryKey: ['runtime-schemas'],
    queryFn: runtimeApi.getSchemas,
  });

  const filteredSchemas = useMemo(() => {
    if (!schemas) return [];
    if (!search) return schemas;

    const searchLower = search.toLowerCase();
    return schemas.filter((module) => {
      if (module.name.toLowerCase().includes(searchLower)) return true;
      if (module.description.toLowerCase().includes(searchLower)) return true;
      if (module.methods.some((m) => m.name.toLowerCase().includes(searchLower))) return true;
      if (module.types?.some((t) => t.name.toLowerCase().includes(searchLower))) return true;
      return false;
    });
  }, [schemas, search]);

  const handleModuleClick = (moduleName: string) => {
    setActiveModule(moduleName);
    const element = document.getElementById(moduleName);
    if (element) {
      element.scrollIntoView({ behavior: 'smooth' });
    }
  };

  if (isLoading) {
    return (
      <div className="flex gap-6">
        <div className="w-64 shrink-0">
          <Skeleton className="h-10 w-full mb-4" />
          <div className="space-y-2">
            {[1, 2, 3, 4, 5].map((i) => (
              <Skeleton key={i} className="h-8 w-full" />
            ))}
          </div>
        </div>
        <div className="flex-1 space-y-6">
          <Skeleton className="h-8 w-64" />
          <Skeleton className="h-32 w-full" />
          <Skeleton className="h-32 w-full" />
        </div>
      </div>
    );
  }

  return (
    <div className="flex gap-6">
      {/* Sidebar Navigation - Fixed */}
      <div className="w-64 shrink-0">
        <div className="fixed w-64 top-20">
          <div className="relative mb-4">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              placeholder="Search modules..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="pl-9"
            />
          </div>
          <ScrollArea className="h-[calc(100vh-8rem)]">
            <div className="space-y-1 pr-4">
              {filteredSchemas.map((module) => (
                <NavItem
                  key={module.name}
                  module={module}
                  isActive={activeModule === module.name}
                  onClick={() => handleModuleClick(module.name)}
                />
              ))}
            </div>
          </ScrollArea>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 min-w-0">
        <div className="mb-6">
          <h1 className="text-2xl font-bold tracking-tight">API Documentation</h1>
          <p className="text-muted-foreground">
            Reference documentation for M3M runtime modules. These modules are available in your
            JavaScript services with the <code className="font-mono text-sm">$</code> prefix.
          </p>
        </div>

        {filteredSchemas.length === 0 ? (
          <div className="flex flex-col items-center justify-center rounded-lg border border-dashed py-12 text-center">
            <Book className="h-12 w-12 text-muted-foreground/50" />
            <p className="mt-2 text-sm text-muted-foreground">No modules found</p>
            {search && (
              <p className="text-xs text-muted-foreground">
                Try a different search term
              </p>
            )}
          </div>
        ) : (
          <div className="space-y-12">
            {filteredSchemas.map((module) => (
              <ModuleSection key={module.name} module={module} />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
