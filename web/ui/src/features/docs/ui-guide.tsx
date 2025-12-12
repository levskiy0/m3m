import { useTitle } from '@/hooks';

export function UIGuidePage() {
  useTitle('UI Guide - Docs');

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-3xl font-bold tracking-tight mb-2">UI Guide</h1>
        <p className="text-muted-foreground text-lg">
          Interactive UI dialogs and action hooks for building dynamic user interfaces.
        </p>
      </div>

      {/* $ui Module */}
      <section className="space-y-6">
        <div>
          <h2 className="text-2xl font-semibold mb-2">$ui Module</h2>
          <p className="text-muted-foreground mb-4">
            Display interactive dialogs to users via WebSocket. Works inside action handlers registered with <code className="font-mono bg-muted px-1 rounded">$hook.onAction</code>.
          </p>
          <pre className="bg-muted rounded-md p-3 text-sm overflow-x-auto">
            <code>{`$hook.onAction('my-action', (a) => {
  $ui.toast({ text: 'Starting...', severity: 'info' });

  $ui.confirm({ title: 'Delete?', yes: 'Yes', no: 'No' }, (confirmed) => {
    if (confirmed) {
      a.loading(true);

      // Session context is preserved in delayed callbacks
      $schedule.delay(3000, () => {
        a.loading(false);
        $ui.alert({ title: 'Success', text: 'Done!', severity: 'success' });
      });
    }
  });
});`}</code>
          </pre>
        </div>

        {/* toast */}
        <div className="border rounded-lg p-4">
          <h3 className="font-medium font-mono mb-2">$ui.toast(options)</h3>
          <p className="text-sm text-muted-foreground mb-3">
            Show a toast notification. Fire-and-forget, auto-dismisses.
          </p>
          <pre className="bg-muted rounded-md p-3 text-sm overflow-x-auto">
            <code>{`$ui.toast({ text: "Operation completed" });

$ui.toast({
  text: "File uploaded successfully",
  severity: "success"
});

$ui.toast({
  text: "Something went wrong",
  severity: "error"
});`}</code>
          </pre>
          <div className="mt-3 text-sm">
            <p className="font-medium mb-1">Options:</p>
            <ul className="text-muted-foreground space-y-1">
              <li><code className="font-mono bg-muted px-1 rounded">text</code> - toast message</li>
              <li><code className="font-mono bg-muted px-1 rounded">severity</code> - "info" | "success" | "warning" | "error" (default: "info")</li>
            </ul>
          </div>
        </div>

        {/* alert */}
        <div className="border rounded-lg p-4">
          <h3 className="font-medium font-mono mb-2">$ui.alert(options)</h3>
          <p className="text-sm text-muted-foreground mb-3">
            Show an alert dialog. Fire-and-forget, no callback.
          </p>
          <pre className="bg-muted rounded-md p-3 text-sm overflow-x-auto">
            <code>{`$ui.alert({
  title: "Notice",
  text: "Your session will expire in 5 minutes",
  severity: "warning"
});`}</code>
          </pre>
          <div className="mt-3 text-sm">
            <p className="font-medium mb-1">Options:</p>
            <ul className="text-muted-foreground space-y-1">
              <li><code className="font-mono bg-muted px-1 rounded">title</code> - dialog title (optional)</li>
              <li><code className="font-mono bg-muted px-1 rounded">text</code> - message text (optional)</li>
              <li><code className="font-mono bg-muted px-1 rounded">severity</code> - "info" | "success" | "warning" | "error"</li>
            </ul>
          </div>
        </div>

        {/* confirm */}
        <div className="border rounded-lg p-4">
          <h3 className="font-medium font-mono mb-2">$ui.confirm(options, callback)</h3>
          <p className="text-sm text-muted-foreground mb-3">
            Show a confirmation dialog with Yes/No buttons.
          </p>
          <pre className="bg-muted rounded-md p-3 text-sm overflow-x-auto">
            <code>{`$ui.confirm(
  {
    title: "Delete item?",
    text: "This action cannot be undone",
    yes: "Delete",
    no: "Cancel"
  },
  (confirmed) => {
    if (confirmed) {
      // User clicked Yes
      $database.collection("items").delete(itemId);
      $ui.toast({ text: "Item deleted", severity: "success" });
    }
  }
);`}</code>
          </pre>
          <div className="mt-3 text-sm">
            <p className="font-medium mb-1">Options:</p>
            <ul className="text-muted-foreground space-y-1">
              <li><code className="font-mono bg-muted px-1 rounded">title</code> - dialog title (optional)</li>
              <li><code className="font-mono bg-muted px-1 rounded">text</code> - message text (optional)</li>
              <li><code className="font-mono bg-muted px-1 rounded">yes</code> - Yes button label (default: "Yes")</li>
              <li><code className="font-mono bg-muted px-1 rounded">no</code> - No button label (default: "No")</li>
            </ul>
            <p className="font-medium mb-1 mt-3">Callback:</p>
            <p className="text-muted-foreground">
              <code className="font-mono bg-muted px-1 rounded">(confirmed: boolean | null) =&gt; void</code> -
              true if Yes, false if No, null if no user context
            </p>
          </div>
        </div>

        {/* prompt */}
        <div className="border rounded-lg p-4">
          <h3 className="font-medium font-mono mb-2">$ui.prompt(options, callback)</h3>
          <p className="text-sm text-muted-foreground mb-3">
            Show a text input dialog.
          </p>
          <pre className="bg-muted rounded-md p-3 text-sm overflow-x-auto">
            <code>{`$ui.prompt(
  {
    title: "Rename file",
    text: "Enter new file name",
    placeholder: "file.txt",
    defaultValue: currentName
  },
  (value) => {
    if (value !== null) {
      // User entered a value
      $storage.move(oldPath, newPath + "/" + value);
      $ui.toast({ text: "File renamed", severity: "success" });
    }
  }
);`}</code>
          </pre>
          <div className="mt-3 text-sm">
            <p className="font-medium mb-1">Options:</p>
            <ul className="text-muted-foreground space-y-1">
              <li><code className="font-mono bg-muted px-1 rounded">title</code> - dialog title (optional)</li>
              <li><code className="font-mono bg-muted px-1 rounded">text</code> - message text (optional)</li>
              <li><code className="font-mono bg-muted px-1 rounded">placeholder</code> - input placeholder (optional)</li>
              <li><code className="font-mono bg-muted px-1 rounded">defaultValue</code> - default input value (optional)</li>
            </ul>
            <p className="font-medium mb-1 mt-3">Callback:</p>
            <p className="text-muted-foreground">
              <code className="font-mono bg-muted px-1 rounded">(value: string | null) =&gt; void</code> -
              entered string or null if cancelled
            </p>
          </div>
        </div>

        {/* form */}
        <div className="border rounded-lg p-4">
          <h3 className="font-medium font-mono mb-2">$ui.form(options, callback)</h3>
          <p className="text-sm text-muted-foreground mb-3">
            Show a form dialog with multiple fields. The callback receives a form controller for managing state.
          </p>
          <pre className="bg-muted rounded-md p-3 text-sm overflow-x-auto">
            <code>{`$ui.form(
  {
    title: "Create User",
    text: "Fill in the user details",
    schema: [
      { name: "name", type: "input", label: "Name", required: true },
      { name: "email", type: "input", label: "Email", placeholder: "user@example.com" },
      { name: "role", type: "select", label: "Role", options: ["admin", "user", "guest"] },
      { name: "active", type: "checkbox", label: "Active", defaultValue: true },
      { name: "notes", type: "textarea", label: "Notes", colspan: "full" }
    ],
    actions: [
      { label: "Cancel", action: "cancel", variant: "outline" },
      { label: "Create", action: "submit", variant: "default" }
    ]
  },
  (form, result) => {
    if (result === null || result.action === "cancel") {
      form.close();
      return;
    }

    form.loading(true);

    // Validate
    const errors = {};
    if (!result.data.name) errors.name = "Name is required";
    if (result.data.email && !result.data.email.includes("@")) {
      errors.email = "Invalid email format";
    }

    if (Object.keys(errors).length > 0) {
      form.error(errors);
      form.loading(false);
      return;
    }

    // Save to database
    $database.collection("users").insert(result.data);
    $ui.toast({ text: "User created", severity: "success" });
    form.close();
  }
);`}</code>
          </pre>
          <div className="mt-3 text-sm space-y-4">
            <div>
              <p className="font-medium mb-1">Field Types:</p>
              <ul className="text-muted-foreground space-y-1">
                <li><code className="font-mono bg-muted px-1 rounded">input</code> - text input</li>
                <li><code className="font-mono bg-muted px-1 rounded">textarea</code> - multi-line text</li>
                <li><code className="font-mono bg-muted px-1 rounded">checkbox</code> - boolean checkbox</li>
                <li><code className="font-mono bg-muted px-1 rounded">select</code> - dropdown select</li>
                <li><code className="font-mono bg-muted px-1 rounded">combobox</code> - searchable select</li>
                <li><code className="font-mono bg-muted px-1 rounded">radiogroup</code> - radio buttons</li>
                <li><code className="font-mono bg-muted px-1 rounded">date</code> - date picker</li>
                <li><code className="font-mono bg-muted px-1 rounded">datetime</code> - date and time picker</li>
              </ul>
            </div>
            <div>
              <p className="font-medium mb-1">Field Options:</p>
              <ul className="text-muted-foreground space-y-1">
                <li><code className="font-mono bg-muted px-1 rounded">name</code> - field key in result.data</li>
                <li><code className="font-mono bg-muted px-1 rounded">type</code> - field type</li>
                <li><code className="font-mono bg-muted px-1 rounded">label</code> - field label (optional)</li>
                <li><code className="font-mono bg-muted px-1 rounded">hint</code> - helper text (optional)</li>
                <li><code className="font-mono bg-muted px-1 rounded">colspan</code> - 1-6 or "full" (optional)</li>
                <li><code className="font-mono bg-muted px-1 rounded">required</code> - show required indicator (optional)</li>
                <li><code className="font-mono bg-muted px-1 rounded">placeholder</code> - placeholder text (optional)</li>
                <li><code className="font-mono bg-muted px-1 rounded">defaultValue</code> - initial value (optional)</li>
                <li><code className="font-mono bg-muted px-1 rounded">options</code> - for select/combobox/radiogroup (optional)</li>
              </ul>
            </div>
            <div>
              <p className="font-medium mb-1">Form Controller:</p>
              <ul className="text-muted-foreground space-y-1">
                <li><code className="font-mono bg-muted px-1 rounded">form.loading(bool)</code> - show/hide loading state</li>
                <li><code className="font-mono bg-muted px-1 rounded">form.error(&#123;field: "message"&#125;)</code> - show field errors</li>
                <li><code className="font-mono bg-muted px-1 rounded">form.close()</code> - close the form dialog</li>
              </ul>
            </div>
          </div>
        </div>
      </section>

      {/* $hook Module */}
      <section className="space-y-6 pt-4">
        <div>
          <h2 className="text-2xl font-semibold mb-2">$hook Module</h2>
          <p className="text-muted-foreground">
            Register handlers for action buttons and model events.
          </p>
        </div>

        {/* onAction */}
        <div className="border rounded-lg p-4">
          <h3 className="font-medium font-mono mb-2">$hook.onAction(slug, handler)</h3>
          <p className="text-sm text-muted-foreground mb-3">
            Register a handler for an action button. Actions are defined in Widget settings.
          </p>
          <pre className="bg-muted rounded-md p-3 text-sm overflow-x-auto">
            <code>{`// Simple action
$hook.onAction("send-report", (ctx) => {
  ctx.loading(true);

  const report = generateReport();
  $http.post("https://api.example.com/reports", { body: report });

  $ui.toast({ text: "Report sent!", severity: "success" });
  ctx.loading(false);
});

// Action with confirmation
$hook.onAction("clear-cache", (ctx) => {
  $ui.confirm(
    { title: "Clear cache?", text: "This will remove all cached data" },
    (confirmed) => {
      if (confirmed) {
        ctx.loading(true);
        clearAllCache();
        $ui.toast({ text: "Cache cleared", severity: "success" });
        ctx.loading(false);
      }
    }
  );
});

// Action with form
$hook.onAction("add-item", (ctx) => {
  $ui.form(
    {
      title: "Add Item",
      schema: [
        { name: "title", type: "input", label: "Title", required: true },
        { name: "price", type: "input", label: "Price" }
      ]
    },
    (form, result) => {
      if (!result || result.action === "cancel") {
        form.close();
        return;
      }

      form.loading(true);
      $database.collection("items").insert(result.data);
      $ui.toast({ text: "Item added", severity: "success" });
      form.close();
    }
  );
});`}</code>
          </pre>
          <div className="mt-3 text-sm">
            <p className="font-medium mb-1">Action Context (ctx):</p>
            <ul className="text-muted-foreground space-y-1">
              <li><code className="font-mono bg-muted px-1 rounded">ctx.name</code> - action name</li>
              <li><code className="font-mono bg-muted px-1 rounded">ctx.slug</code> - action slug identifier</li>
              <li><code className="font-mono bg-muted px-1 rounded">ctx.userId</code> - ID of user who triggered the action</li>
              <li><code className="font-mono bg-muted px-1 rounded">ctx.sessionId</code> - WebSocket session ID for UI targeting</li>
              <li><code className="font-mono bg-muted px-1 rounded">ctx.loading(bool)</code> - set button loading state</li>
              <li><code className="font-mono bg-muted px-1 rounded">ctx.active(bool)</code> - enable/disable button</li>
            </ul>
          </div>
        </div>

        {/* Model hooks */}
        <div className="border rounded-lg p-4">
          <h3 className="font-medium font-mono mb-2">Model Hooks</h3>
          <p className="text-sm text-muted-foreground mb-3">
            React to model data changes from the frontend.
          </p>
          <pre className="bg-muted rounded-md p-3 text-sm overflow-x-auto">
            <code>{`// When a new record is inserted via UI
$hook.onModelInsert("users", (data) => {
  $logger.info("New user created:", data.name);

  // Send welcome email
  $mail.send({
    to: data.email,
    subject: "Welcome!",
    body: "Thanks for signing up, " + data.name
  });
});

// When a record is updated via UI
$hook.onModelUpdate("orders", (data) => {
  $logger.info("Order updated:", data._id);

  if (data.status === "shipped") {
    notifyCustomer(data);
  }
});

// When a record is deleted via UI
$hook.onModelDelete("posts", (data) => {
  $logger.info("Post deleted:", data._id);

  // Clean up related files
  $storage.delete("posts/" + data._id);
});`}</code>
          </pre>
          <div className="mt-3 text-sm">
            <p className="font-medium mb-1">Methods:</p>
            <ul className="text-muted-foreground space-y-1">
              <li><code className="font-mono bg-muted px-1 rounded">$hook.onModelInsert(modelSlug, handler)</code> - triggered after insert</li>
              <li><code className="font-mono bg-muted px-1 rounded">$hook.onModelUpdate(modelSlug, handler)</code> - triggered after update</li>
              <li><code className="font-mono bg-muted px-1 rounded">$hook.onModelDelete(modelSlug, handler)</code> - triggered after delete</li>
            </ul>
            <p className="text-muted-foreground mt-2">
              <strong>Note:</strong> Model hooks are triggered only from frontend operations, not from <code className="font-mono bg-muted px-1 rounded">$database</code> calls in code.
            </p>
          </div>
        </div>
      </section>
    </div>
  );
}
