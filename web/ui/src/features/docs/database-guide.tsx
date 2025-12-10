import { useTitle } from '@/hooks';

export function DatabaseGuidePage() {
  useTitle('Database Guide - Docs');

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-3xl font-bold tracking-tight mb-2">$database Module</h1>
        <p className="text-muted-foreground text-lg">
          Module for working with model data. Uses MongoDB-like query syntax.
        </p>
      </div>

      {/* Getting a Collection */}
      <section className="space-y-4">
        <h2 className="text-xl font-semibold">Getting a Collection</h2>
        <pre className="bg-muted rounded-lg p-4 text-sm overflow-x-auto">
          <code>{`const users = $database.collection("users"); // model slug`}</code>
        </pre>
      </section>

      {/* Collection Methods */}
      <section className="space-y-4">
        <h2 className="text-xl font-semibold">Collection Methods</h2>

        {/* find */}
        <div className="border rounded-lg p-4">
          <h3 className="font-medium font-mono mb-2">find(filter?)</h3>
          <p className="text-sm text-muted-foreground mb-3">
            Find documents matching filter.
          </p>
          <pre className="bg-muted rounded-md p-3 text-sm overflow-x-auto">
            <code>{`// All documents
const all = users.find({});

// By field
const active = users.find({ status: "active" });

// With operators
const adults = users.find({ age: { $gte: 18 } });`}</code>
          </pre>
        </div>

        {/* findWithOptions */}
        <div className="border rounded-lg p-4">
          <h3 className="font-medium font-mono mb-2">findWithOptions(filter?, options?)</h3>
          <p className="text-sm text-muted-foreground mb-3">
            Find with pagination and sorting.
          </p>
          <pre className="bg-muted rounded-md p-3 text-sm overflow-x-auto">
            <code>{`const result = users.findWithOptions(
  { status: "active" },
  {
    page: 1,
    limit: 10,
    sort: "createdAt",
    order: "desc"
  }
);`}</code>
          </pre>
          <div className="mt-3 text-sm">
            <p className="font-medium mb-1">Options:</p>
            <ul className="text-muted-foreground space-y-1">
              <li><code className="font-mono bg-muted px-1 rounded">page</code> - page number (default: 1)</li>
              <li><code className="font-mono bg-muted px-1 rounded">limit</code> - results per page (default: 100)</li>
              <li><code className="font-mono bg-muted px-1 rounded">sort</code> - field to sort by</li>
              <li><code className="font-mono bg-muted px-1 rounded">order</code> - "asc" or "desc"</li>
            </ul>
          </div>
        </div>

        {/* findOne */}
        <div className="border rounded-lg p-4">
          <h3 className="font-medium font-mono mb-2">findOne(filter?)</h3>
          <p className="text-sm text-muted-foreground mb-3">
            Find first matching document.
          </p>
          <pre className="bg-muted rounded-md p-3 text-sm overflow-x-auto">
            <code>{`const user = users.findOne({ email: "test@example.com" });
if (user) {
  $logger.info("Found:", user.name);
}`}</code>
          </pre>
        </div>

        {/* insert */}
        <div className="border rounded-lg p-4">
          <h3 className="font-medium font-mono mb-2">insert(data)</h3>
          <p className="text-sm text-muted-foreground mb-3">
            Insert a document. Returns created document or <code className="font-mono bg-muted px-1 rounded">null</code>.
          </p>
          <pre className="bg-muted rounded-md p-3 text-sm overflow-x-auto">
            <code>{`const newUser = users.insert({
  name: "John",
  email: "john@example.com",
  age: 25
});

if (newUser) {
  $logger.info("Created with ID:", newUser._id);
}`}</code>
          </pre>
        </div>

        {/* update */}
        <div className="border rounded-lg p-4">
          <h3 className="font-medium font-mono mb-2">update(id, data)</h3>
          <p className="text-sm text-muted-foreground mb-3">
            Update document by ID. Returns <code className="font-mono bg-muted px-1 rounded">true</code>/<code className="font-mono bg-muted px-1 rounded">false</code>.
          </p>
          <pre className="bg-muted rounded-md p-3 text-sm overflow-x-auto">
            <code>{`const success = users.update("507f1f77bcf86cd799439011", {
  name: "John Updated",
  age: 26
});`}</code>
          </pre>
        </div>

        {/* delete */}
        <div className="border rounded-lg p-4">
          <h3 className="font-medium font-mono mb-2">delete(id)</h3>
          <p className="text-sm text-muted-foreground mb-3">
            Delete document by ID. Returns <code className="font-mono bg-muted px-1 rounded">true</code>/<code className="font-mono bg-muted px-1 rounded">false</code>.
          </p>
          <pre className="bg-muted rounded-md p-3 text-sm overflow-x-auto">
            <code>{`const deleted = users.delete("507f1f77bcf86cd799439011");`}</code>
          </pre>
        </div>

        {/* count */}
        <div className="border rounded-lg p-4">
          <h3 className="font-medium font-mono mb-2">count(filter?)</h3>
          <p className="text-sm text-muted-foreground mb-3">
            Count matching documents.
          </p>
          <pre className="bg-muted rounded-md p-3 text-sm overflow-x-auto">
            <code>{`const total = users.count({});
const activeCount = users.count({ status: "active" });`}</code>
          </pre>
        </div>

        {/* upsert */}
        <div className="border rounded-lg p-4">
          <h3 className="font-medium font-mono mb-2">upsert(filter, data)</h3>
          <p className="text-sm text-muted-foreground mb-3">
            Update if found, otherwise insert.
          </p>
          <pre className="bg-muted rounded-md p-3 text-sm overflow-x-auto">
            <code>{`const user = users.upsert(
  { email: "john@example.com" },
  { name: "John", email: "john@example.com", visits: 1 }
);`}</code>
          </pre>
        </div>

        {/* findOneAndUpdate */}
        <div className="border rounded-lg p-4">
          <h3 className="font-medium font-mono mb-2">findOneAndUpdate(filter, update, options?)</h3>
          <p className="text-sm text-muted-foreground mb-3">
            Atomically find and update a document. Supports update operators.
          </p>
          <pre className="bg-muted rounded-md p-3 text-sm overflow-x-auto">
            <code>{`// Increment counter
const user = users.findOneAndUpdate(
  { email: "john@example.com" },
  { $inc: { visits: 1 } },
  { returnNew: true }
);

// Update fields
const updated = users.findOneAndUpdate(
  { _id: "507f1f77bcf86cd799439011" },
  { $set: { lastLogin: new Date().toISOString() } },
  { returnNew: true }
);`}</code>
          </pre>
          <div className="mt-3 text-sm">
            <p className="font-medium mb-1">Options:</p>
            <ul className="text-muted-foreground">
              <li><code className="font-mono bg-muted px-1 rounded">returnNew</code> - return updated document instead of original (default: false)</li>
            </ul>
          </div>
        </div>
      </section>

      {/* Filter Operators */}
      <section className="space-y-4">
        <h2 className="text-xl font-semibold">Filter Operators</h2>
        <div className="border rounded-lg overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-muted">
              <tr>
                <th className="px-4 py-2 text-left font-medium">Operator</th>
                <th className="px-4 py-2 text-left font-medium">Description</th>
              </tr>
            </thead>
            <tbody className="divide-y">
              <tr><td className="px-4 py-2 font-mono">$eq</td><td className="px-4 py-2 text-muted-foreground">Equal</td></tr>
              <tr><td className="px-4 py-2 font-mono">$ne</td><td className="px-4 py-2 text-muted-foreground">Not equal</td></tr>
              <tr><td className="px-4 py-2 font-mono">$gt</td><td className="px-4 py-2 text-muted-foreground">Greater than</td></tr>
              <tr><td className="px-4 py-2 font-mono">$gte</td><td className="px-4 py-2 text-muted-foreground">Greater than or equal</td></tr>
              <tr><td className="px-4 py-2 font-mono">$lt</td><td className="px-4 py-2 text-muted-foreground">Less than</td></tr>
              <tr><td className="px-4 py-2 font-mono">$lte</td><td className="px-4 py-2 text-muted-foreground">Less than or equal</td></tr>
              <tr><td className="px-4 py-2 font-mono">$contains</td><td className="px-4 py-2 text-muted-foreground">Contains substring</td></tr>
              <tr><td className="px-4 py-2 font-mono">$startsWith</td><td className="px-4 py-2 text-muted-foreground">Starts with</td></tr>
              <tr><td className="px-4 py-2 font-mono">$endsWith</td><td className="px-4 py-2 text-muted-foreground">Ends with</td></tr>
              <tr><td className="px-4 py-2 font-mono">$in</td><td className="px-4 py-2 text-muted-foreground">In array</td></tr>
              <tr><td className="px-4 py-2 font-mono">$nin</td><td className="px-4 py-2 text-muted-foreground">Not in array</td></tr>
            </tbody>
          </table>
        </div>
        <pre className="bg-muted rounded-lg p-4 text-sm overflow-x-auto">
          <code>{`// Examples
users.find({ age: { $gt: 18, $lt: 65 } });
users.find({ status: { $in: ["active", "pending"] } });
users.find({ name: { $contains: "John" } });
users.find({ email: { $endsWith: "@gmail.com" } });`}</code>
        </pre>
      </section>

      {/* Update Operators */}
      <section className="space-y-4">
        <h2 className="text-xl font-semibold">Update Operators</h2>
        <p className="text-muted-foreground mb-2">
          For <code className="font-mono bg-muted px-1 rounded">findOneAndUpdate</code>:
        </p>
        <div className="border rounded-lg overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-muted">
              <tr>
                <th className="px-4 py-2 text-left font-medium">Operator</th>
                <th className="px-4 py-2 text-left font-medium">Description</th>
              </tr>
            </thead>
            <tbody className="divide-y">
              <tr><td className="px-4 py-2 font-mono">$set</td><td className="px-4 py-2 text-muted-foreground">Set field value</td></tr>
              <tr><td className="px-4 py-2 font-mono">$inc</td><td className="px-4 py-2 text-muted-foreground">Increment numeric value</td></tr>
              <tr><td className="px-4 py-2 font-mono">$unset</td><td className="px-4 py-2 text-muted-foreground">Remove field</td></tr>
              <tr><td className="px-4 py-2 font-mono">$push</td><td className="px-4 py-2 text-muted-foreground">Add element to array</td></tr>
              <tr><td className="px-4 py-2 font-mono">$pull</td><td className="px-4 py-2 text-muted-foreground">Remove element from array</td></tr>
              <tr><td className="px-4 py-2 font-mono">$addToSet</td><td className="px-4 py-2 text-muted-foreground">Add unique element to array</td></tr>
            </tbody>
          </table>
        </div>
        <pre className="bg-muted rounded-lg p-4 text-sm overflow-x-auto">
          <code>{`// Increment counter by 1
users.findOneAndUpdate({ _id: id }, { $inc: { count: 1 } });

// Add tag
users.findOneAndUpdate({ _id: id }, { $push: { tags: "new-tag" } });

// Remove field
users.findOneAndUpdate({ _id: id }, { $unset: { tempField: "" } });`}</code>
        </pre>
      </section>

      {/* Example: Simple CRUD API */}
      <section className="space-y-4">
        <h2 className="text-xl font-semibold">Example: Simple CRUD API</h2>
        <pre className="bg-muted rounded-lg p-4 text-sm overflow-x-auto">
          <code>{`const posts = $database.collection("posts");

// Create
$router.post("/posts", (ctx) => {
  const post = posts.insert({
    title: ctx.body.title,
    content: ctx.body.content,
    createdAt: new Date().toISOString()
  });
  ctx.json(post);
});

// Read
$router.get("/posts", (ctx) => {
  const result = posts.findWithOptions({}, {
    page: parseInt(ctx.query.page) || 1,
    limit: 10,
    sort: "createdAt",
    order: "desc"
  });
  ctx.json(result);
});

// Update
$router.put("/posts/:id", (ctx) => {
  const success = posts.update(ctx.params.id, {
    title: ctx.body.title,
    content: ctx.body.content
  });
  ctx.json({ success });
});

// Delete
$router.delete("/posts/:id", (ctx) => {
  const success = posts.delete(ctx.params.id);
  ctx.json({ success });
});`}</code>
        </pre>
      </section>
    </div>
  );
}
