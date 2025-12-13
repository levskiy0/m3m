import { useRef, useEffect, useState } from 'react';
import Editor from '@monaco-editor/react';
import type { OnMount, Monaco } from '@monaco-editor/react';
import type { editor, IDisposable, MarkerSeverity } from 'monaco-editor';

interface KeyBinding {
  key: string;
  label: string;
  action: () => void;
}

export interface RuntimeError {
  line: number;
  column: number;
  message: string;
}

interface CodeEditorProps {
  value: string;
  onChange?: (value: string) => void;
  language?: string;
  readOnly?: boolean;
  height?: string | number;
  typeDefinitions?: string;
  keyBindings?: KeyBinding[];
  runtimeErrors?: RuntimeError[];
}


export function CodeEditor({
  value,
  onChange,
  language = 'javascript',
  readOnly = false,
  height = '100%',
  typeDefinitions,
  keyBindings,
  runtimeErrors = [],
}: CodeEditorProps) {
  const editorRef = useRef<editor.IStandaloneCodeEditor | null>(null);
  const monacoRef = useRef<Monaco | null>(null);
  const keyBindingsRef = useRef(keyBindings);
  const typeLibRef = useRef<IDisposable | null>(null);
  const runtimeErrorsRef = useRef(runtimeErrors);
  const [editorVersion, setEditorVersion] = useState(0);

  // Update refs to avoid stale closures
  useEffect(() => {
    keyBindingsRef.current = keyBindings;
  }, [keyBindings]);

  useEffect(() => {
    runtimeErrorsRef.current = runtimeErrors;
  }, [runtimeErrors]);

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

    // Register Ctrl+R for Find & Replace (standard Cmd+H often intercepted by browser)
    editor.addAction({
      id: 'custom-replace',
      label: 'Find and Replace',
      keybindings: [monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyR],
      run: (ed) => {
        ed.getAction('editor.action.startFindReplaceAction')?.run();
      },
    });

    // Trigger re-render to apply any pending markers
    setEditorVersion((v) => v + 1);
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

  // Update runtime error decorations (line highlight + markers)
  const decorationsRef = useRef<string[]>([]);

  useEffect(() => {
    const monaco = monacoRef.current;
    const editor = editorRef.current;
    const errors = runtimeErrorsRef.current;
    if (!monaco || !editor) return;

    const model = editor.getModel();
    if (!model) return;

    // Clear old decorations and markers
    if (errors.length === 0) {
      monaco.editor.setModelMarkers(model, 'runtime', []);
      decorationsRef.current = editor.deltaDecorations(decorationsRef.current, []);
      return;
    }

    // Filter out errors with invalid line numbers (Monaco requires 1 <= line <= lineCount)
    const lineCount = model.getLineCount();
    const validErrors = errors.filter(
      (e) => e.line >= 1 && e.line <= lineCount && Number.isFinite(e.line)
    );

    if (validErrors.length === 0) {
      monaco.editor.setModelMarkers(model, 'runtime', []);
      decorationsRef.current = editor.deltaDecorations(decorationsRef.current, []);
      return;
    }

    // Set markers (for Problems panel / hover tooltip)
    const markers = validErrors.map((error) => {
      const lineLength = model.getLineLength(error.line) || 100;
      return {
        severity: monaco.MarkerSeverity.Error as MarkerSeverity,
        message: error.message,
        startLineNumber: error.line,
        startColumn: 1,
        endLineNumber: error.line,
        endColumn: lineLength + 1,
      };
    });
    monaco.editor.setModelMarkers(model, 'runtime', markers);

    // Set decorations (for line highlight)
    const decorations = validErrors.map((error) => ({
      range: new monaco.Range(error.line, 1, error.line, 1),
      options: {
        isWholeLine: true,
        className: 'runtime-error-line',
        glyphMarginClassName: 'runtime-error-glyph',
        overviewRuler: {
          color: '#ff0000',
          position: monaco.editor.OverviewRulerLane.Full,
        },
      },
    }));

    decorationsRef.current = editor.deltaDecorations(decorationsRef.current, decorations);
  }, [runtimeErrors, editorVersion]);

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
        glyphMargin: true,
      }}
      theme="vs-dark"
    />
  );
}
