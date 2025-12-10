import { useParams, Navigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { Code2, Braces, ChevronRight, Copy, Check, List } from 'lucide-react';
import { useState, useMemo } from 'react';

import { runtimeApi } from '@/api';
import type { MethodSchema, TypeSchema, ParamSchema } from '@/api/runtime';
import { useTitle } from '@/hooks';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
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
    const params = method.params?.map((p) => `${p.name}${p.optional ? '?' : ''}`).join(', ') || '';
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
    <div id={method.name} className="border rounded-lg p-4 scroll-mt-20">
      <div className="flex items-center gap-2 mb-2">
        <Code2 className="h-4 w-4 text-primary" />
        <h3 className="font-semibold font-mono text-lg">{method.name}</h3>
        {method.returns && (
          <>
            <ChevronRight className="h-3 w-3 text-muted-foreground" />
            <TypeBadge type={method.returns.type} />
          </>
        )}
      </div>
      <p className="text-muted-foreground mb-3">{method.description}</p>

      <MethodSignature method={method} moduleName={moduleName} />

      {method.params && method.params.length > 0 && (
        <div className="mt-4">
          <h4 className="text-sm font-medium mb-2">Parameters</h4>
          <div className="space-y-2">
            {method.params.map((param) => (
              <div key={param.name} className="flex items-start gap-3 text-sm">
                <code className="font-mono text-foreground shrink-0">{param.name}</code>
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

// Determine if a field looks like a method (has function type signature)
function isMethodField(field: ParamSchema): boolean {
  return field.type.includes('=>') || field.type.includes('(');
}

// Parse rawTypes string into separate interface blocks
interface ParsedInterface {
  name: string;
  body: string;
}

function parseRawTypes(rawTypes: string): ParsedInterface[] {
  const interfaces: ParsedInterface[] = [];
  const regex = /interface\s+(\w+)\s*\{/g;
  let match;
  const positions: { name: string; start: number }[] = [];

  while ((match = regex.exec(rawTypes)) !== null) {
    positions.push({ name: match[1], start: match.index });
  }

  for (let i = 0; i < positions.length; i++) {
    const start = positions[i].start;
    const end = i < positions.length - 1 ? positions[i + 1].start : rawTypes.length;
    const body = rawTypes.slice(start, end).trim();
    interfaces.push({ name: positions[i].name, body });
  }

  return interfaces;
}

function RawTypeCard({ iface }: { iface: ParsedInterface }) {
  const [copied, setCopied] = useState(false);

  const handleCopy = () => {
    navigator.clipboard.writeText(iface.body);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div id={`rawtype-${iface.name}`} className="border rounded-lg p-4 scroll-mt-20">
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-2">
          <Braces className="h-4 w-4 text-primary" />
          <h3 className="font-semibold font-mono text-lg">{iface.name}</h3>
        </div>
        <Button
          variant="ghost"
          size="icon"
          className="h-7 w-7"
          onClick={handleCopy}
        >
          {copied ? <Check className="h-3.5 w-3.5" /> : <Copy className="h-3.5 w-3.5" />}
        </Button>
      </div>
      <pre className="bg-muted/50 rounded-lg p-3 text-sm font-mono overflow-x-auto whitespace-pre">
        <code>{iface.body}</code>
      </pre>
    </div>
  );
}

function TypeCard({ type }: { type: TypeSchema }) {
  const methodFields = type.fields.filter(isMethodField);
  const dataFields = type.fields.filter((f) => !isMethodField(f));

  return (
    <div id={`type-${type.name}`} className="border rounded-lg p-4 scroll-mt-20">
      <div className="flex items-center gap-2 mb-2">
        <Braces className="h-4 w-4 text-primary" />
        <h3 className="font-semibold font-mono text-lg">{type.name}</h3>
      </div>
      {type.description && <p className="text-muted-foreground mb-4">{type.description}</p>}

      {/* Data fields as compact interface */}
      {dataFields.length > 0 && (
        <div className="bg-muted/50 rounded-lg p-3 font-mono text-sm overflow-x-auto mb-4">
          <div className="text-muted-foreground">
            {'interface'} <span className="text-foreground">{type.name}</span> {'{'}
          </div>
          {dataFields.map((field) => (
            <div key={field.name} className="pl-4 flex items-center gap-2 flex-wrap">
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
      )}

      {/* Method fields as detailed cards */}
      {methodFields.length > 0 && (
        <div className="space-y-3">
          <h4 className="text-sm font-medium text-muted-foreground">Methods</h4>
          {methodFields.map((field) => (
            <div key={field.name} className="border rounded-md p-3 bg-background">
              <div className="flex items-start gap-2 mb-1">
                <code className="font-mono font-medium text-foreground">{field.name}</code>
              </div>
              <div className="text-xs font-mono text-muted-foreground mb-2 break-all">
                {field.type}
              </div>
              {field.description && (
                <p className="text-sm text-muted-foreground">{field.description}</p>
              )}
            </div>
          ))}
        </div>
      )}

      {/* Fallback for types with no categorized fields */}
      {dataFields.length === 0 && methodFields.length === 0 && (
        <div className="bg-muted/50 rounded-lg p-3 font-mono text-sm overflow-x-auto">
          <div className="text-muted-foreground">
            {'interface'} <span className="text-foreground">{type.name}</span> {'{'}
          </div>
          {type.fields.map((field) => (
            <div key={field.name} className="pl-4 flex items-center gap-2 flex-wrap">
              <span className="text-foreground">{field.name}</span>
              <span className="text-muted-foreground">{field.optional ? '?:' : ':'}</span>
              <TypeBadge type={field.type} />
            </div>
          ))}
          <div className="text-muted-foreground">{'}'}</div>
        </div>
      )}
    </div>
  );
}

function TableOfContents({
  types,
  methods,
  nested,
  rawTypes,
  moduleName,
}: {
  types?: TypeSchema[];
  methods: MethodSchema[];
  nested?: { name: string; methods: MethodSchema[] }[];
  rawTypes?: string;
  moduleName: string;
}) {
  const parsedRawTypes = rawTypes ? parseRawTypes(rawTypes) : [];

  return (
    <div className="border rounded-lg p-4 bg-muted/30">
      <div className="flex items-center gap-2 mb-3">
        <List className="h-4 w-4" />
        <h3 className="font-medium">Contents</h3>
      </div>
      <div className="space-y-3 text-sm">
        {types && types.length > 0 && (
          <div>
            <div className="text-muted-foreground text-xs uppercase mb-1">Types</div>
            <div className="space-y-0.5">
              {types.map((t) => (
                <a
                  key={t.name}
                  href={`#type-${t.name}`}
                  className="block text-foreground hover:text-primary transition-colors font-mono"
                >
                  {t.name}
                </a>
              ))}
            </div>
          </div>
        )}
        {parsedRawTypes.length > 0 && (
          <div>
            <div className="text-muted-foreground text-xs uppercase mb-1">Type Definitions</div>
            <div className="space-y-0.5">
              {parsedRawTypes.map((t) => (
                <a
                  key={t.name}
                  href={`#rawtype-${t.name}`}
                  className="block text-foreground hover:text-primary transition-colors font-mono"
                >
                  {t.name}
                </a>
              ))}
            </div>
          </div>
        )}
        {methods.length > 0 && (
          <div>
            <div className="text-muted-foreground text-xs uppercase mb-1">Methods</div>
            <div className="space-y-0.5">
              {methods.map((m) => (
                <a
                  key={m.name}
                  href={`#${m.name}`}
                  className="block text-foreground hover:text-primary transition-colors font-mono"
                >
                  {moduleName}.{m.name}()
                </a>
              ))}
            </div>
          </div>
        )}
        {nested &&
          nested.length > 0 &&
          nested.map((n) => (
            <div key={n.name}>
              <div className="text-muted-foreground text-xs uppercase mb-1">
                {moduleName}.{n.name}
              </div>
              <div className="space-y-0.5">
                {n.methods.map((m) => (
                  <a
                    key={m.name}
                    href={`#${n.name}-${m.name}`}
                    className="block text-foreground hover:text-primary transition-colors font-mono"
                  >
                    {moduleName}.{n.name}.{m.name}()
                  </a>
                ))}
              </div>
            </div>
          ))}
      </div>
    </div>
  );
}

export function ModulePage() {
  const { moduleId } = useParams<{ moduleId: string }>();
  const decodedModuleId = moduleId ? decodeURIComponent(moduleId) : '';

  const { data: schemas, isLoading } = useQuery({
    queryKey: ['runtime-schemas'],
    queryFn: runtimeApi.getSchemas,
  });

  const module = schemas?.find((m) => m.name === decodedModuleId);

  useTitle(module ? `${module.name} - Docs` : 'Module - Docs');

  if (isLoading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-10 w-48" />
        <Skeleton className="h-4 w-96" />
        <Skeleton className="h-32 w-full" />
        <Skeleton className="h-32 w-full" />
      </div>
    );
  }

  if (!module) {
    return <Navigate to="/docs/getting-started" replace />;
  }

  return (
    <div className="space-y-8">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold tracking-tight font-mono mb-2">{module.name}</h1>
        <p className="text-muted-foreground text-lg">{module.description}</p>
      </div>

      {/* Table of Contents */}
      <TableOfContents
        types={module.types}
        methods={module.methods}
        nested={module.nested}
        rawTypes={module.rawTypes}
        moduleName={module.name}
      />

      {/* Types */}
      {module.types && module.types.length > 0 && (
        <section>
          <h2 className="text-xl font-semibold mb-4">Types</h2>
          <div className="space-y-4">
            {module.types.map((type) => (
              <TypeCard key={type.name} type={type} />
            ))}
          </div>
        </section>
      )}

      {/* Raw Types (TypeScript definitions) */}
      {module.rawTypes && (
        <section>
          <h2 className="text-xl font-semibold mb-4">Type Definitions</h2>
          <div className="space-y-4">
            {parseRawTypes(module.rawTypes).map((iface) => (
              <RawTypeCard key={iface.name} iface={iface} />
            ))}
          </div>
        </section>
      )}

      {/* Methods */}
      {module.methods.length > 0 && (
        <section>
          <h2 className="text-xl font-semibold mb-4">Methods</h2>
          <div className="space-y-4">
            {module.methods.map((method) => (
              <MethodCard key={method.name} method={method} moduleName={module.name} />
            ))}
          </div>
        </section>
      )}

      {/* Nested Modules */}
      {module.nested && module.nested.length > 0 && (
        <section>
          <h2 className="text-xl font-semibold mb-4">Namespaces</h2>
          {module.nested.map((nested) => (
            <div key={nested.name} className="mb-6">
              <h3 className="text-lg font-medium font-mono mb-3">
                {module.name}.{nested.name}
              </h3>
              <p className="text-muted-foreground mb-4">{nested.description}</p>
              <div className="space-y-4">
                {nested.methods.map((method) => (
                  <div
                    key={method.name}
                    id={`${nested.name}-${method.name}`}
                    className="scroll-mt-20"
                  >
                    <MethodCard method={method} moduleName={`${module.name}.${nested.name}`} />
                  </div>
                ))}
              </div>
            </div>
          ))}
        </section>
      )}
    </div>
  );
}
