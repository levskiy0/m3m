import { useRef, useEffect } from 'react';
import Editor from '@monaco-editor/react';
import type { OnMount, Monaco } from '@monaco-editor/react';
import type { editor, IDisposable } from 'monaco-editor';

interface KeyBinding {
  key: string;
  label: string;
  action: () => void;
}

interface CodeEditorProps {
  value: string;
  onChange?: (value: string) => void;
  language?: string;
  readOnly?: boolean;
  height?: string | number;
  typeDefinitions?: string;
  keyBindings?: KeyBinding[];
}


export function CodeEditor({
  value,
  onChange,
  language = 'javascript',
  readOnly = false,
  height = '100%',
  typeDefinitions,
  keyBindings,
}: CodeEditorProps) {
  const editorRef = useRef<editor.IStandaloneCodeEditor | null>(null);
  const monacoRef = useRef<Monaco | null>(null);
  const keyBindingsRef = useRef(keyBindings);
  const typeLibRef = useRef<IDisposable | null>(null);

  // Update ref in effect to avoid updating during render
  useEffect(() => {
    keyBindingsRef.current = keyBindings;
  }, [keyBindings]);

  const handleMount: OnMount = (editor, monaco) => {
    editorRef.current = editor;
    monacoRef.current = monaco;

    // Configure JavaScript/TypeScript defaults
    monaco.languages.typescript.javascriptDefaults.setDiagnosticsOptions({
      noSemanticValidation: false,
      noSyntaxValidation: false,
    });

    monaco.languages.typescript.javascriptDefaults.setCompilerOptions({
      target: monaco.languages.typescript.ScriptTarget.ES2020,
      allowNonTsExtensions: true,
      moduleResolution: monaco.languages.typescript.ModuleResolutionKind.NodeJs,
      module: monaco.languages.typescript.ModuleKind.CommonJS,
      noEmit: true,
      allowJs: true,
      checkJs: true,
    });

    // Add type definitions if provided
    if (typeDefinitions) {
      typeLibRef.current?.dispose();
      typeLibRef.current = monaco.languages.typescript.javascriptDefaults.addExtraLib(
        typeDefinitions,
        'ts:runtime.d.ts'
      );
    }

    // Register Ctrl+S for save
    editor.addAction({
      id: 'custom-save',
      label: 'Save',
      keybindings: [monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS],
      run: () => {
        const bindings = keyBindingsRef.current;
        const saveBinding = bindings?.find(b => b.key === 'ctrl+s');
        saveBinding?.action();
      },
    });

    // Register Ctrl+, for run/restart
    editor.addAction({
      id: 'custom-run',
      label: 'Run / Restart',
      keybindings: [monaco.KeyMod.CtrlCmd | monaco.KeyCode.Comma],
      run: () => {
        const bindings = keyBindingsRef.current;
        const runBinding = bindings?.find(b => b.key === 'ctrl+,');
        runBinding?.action();
      },
    });

    // Register Ctrl+. for stop
    editor.addAction({
      id: 'custom-stop',
      label: 'Stop',
      keybindings: [monaco.KeyMod.CtrlCmd | monaco.KeyCode.Period],
      run: () => {
        const bindings = keyBindingsRef.current;
        const stopBinding = bindings?.find(b => b.key === 'ctrl+.');
        stopBinding?.action();
      },
    });
  };

  useEffect(() => {
    // Update type definitions when they change
    if (monacoRef.current && typeDefinitions) {
      typeLibRef.current?.dispose();
      typeLibRef.current = monacoRef.current.languages.typescript.javascriptDefaults.addExtraLib(
        typeDefinitions,
        'ts:runtime.d.ts'
      );
    }
  }, [typeDefinitions]);

  return (
    <Editor
      height={height}
      language={language}
      value={value}
      onChange={(v) => onChange?.(v || '')}
      onMount={handleMount}
      options={{
        readOnly,
        minimap: { enabled: false },
        fontSize: 14,
        lineNumbers: 'on',
        scrollBeyondLastLine: false,
        wordWrap: 'on',
        automaticLayout: true,
        tabSize: 2,
        insertSpaces: true,
        folding: true,
        renderLineHighlight: 'line',
        suggestOnTriggerCharacters: true,
        acceptSuggestionOnEnter: 'on',
        quickSuggestions: true,
        padding: { top: 16, bottom: 16 },
      }}
      theme="vs-dark"
    />
  );
}
