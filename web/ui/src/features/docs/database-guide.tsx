import { useTitle } from '@/hooks';

export function DatabaseGuidePage() {
  useTitle('Database & Models - Docs');

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-3xl font-bold tracking-tight mb-2">Database & Models</h1>
        <p className="text-muted-foreground text-lg">
          Learn how to work with the database module and define data models.
        </p>
      </div>

      {/* Overview */}
      <section className="space-y-4">
        <h2 className="text-xl font-semibold">Overview</h2>
        <p className="text-muted-foreground">
          M3M provides a MongoDB-like database interface through the{' '}
          <code className="font-mono text-sm bg-muted px-1.5 py-0.5 rounded">$database</code> module.
          Each project has isolated data storage with optional schema validation through Models.
        </p>
      </section>

      {/* Models */}
      <section className="space-y-4">
        <h2 className="text-xl font-semibold">Defining Models</h2>
        <p className="text-muted-foreground mb-4">
          Models define the structure and validation rules for your collections. Create models in the
          <strong> Data Storage </strong> section of your project.
        </p>

        <div className="border rounded-lg p-4">
          <h3 className="font-medium mb-2">Schema Field Types</h3>
          <div className="grid gap-2 text-sm">
            {[
              { type: 'string', desc: 'Text data' },
              { type: 'number', desc: 'Integer or float values' },
              { type: 'boolean', desc: 'True/false values' },
              { type: 'date', desc: 'Date/time values' },
              { type: 'array', desc: 'List of values' },
              { type: 'object', desc: 'Nested object' },
            ].map((t) => (
              <div key={t.type} className="flex items-center gap-3">
                <code className="font-mono bg-muted px-2 py-0.5 rounded w-20">{t.type}</code>
                <span className="text-muted-foreground">{t.desc}</span>
              </div>
            ))}
          </div>
        </div>

        <div className="border rounded-lg p-4">
          <h3 className="font-medium mb-2">Field Options</h3>
          <div className="space-y-2 text-sm">
            <div className="flex items-start gap-3">
              <code className="font-mono bg-muted px-2 py-0.5 rounded">required</code>
              <span className="text-muted-foreground">Field must have a value</span>
            </div>
            <div className="flex items-start gap-3">
              <code className="font-mono bg-muted px-2 py-0.5 rounded">default</code>
              <span className="text-muted-foreground">Default value if not provided</span>
            </div>
            <div className="flex items-start gap-3">
              <code className="font-mono bg-muted px-2 py-0.5 rounded">unique</code>
              <span className="text-muted-foreground">Value must be unique in collection</span>
            </div>
          </div>
        </div>
      </section>

      {/* CRUD Operations */}
      <section className="space-y-4">
        <h2 className="text-xl font-semibold">CRUD Operations</h2>

        <div className="space-y-4">
          <div className="border rounded-lg p-4">
            <h3 className="font-medium mb-2">Insert Documents</h3>
            <pre className="bg-muted rounded-md p-3 text-sm overflow-x-auto">
              <code>{`// Get collection reference
const users = $database.collection('users');

// Insert single document
const user = users.insert({
  name: 'John',
  email: 'john@example.com',
  age: 30
});
// Returns: { _id: '...', name: 'John', ... }`}</code>
            </pre>
          </div>

          <div className="border rounded-lg p-4">
            <h3 className="font-medium mb-2">Find Documents</h3>
            <pre className="bg-muted rounded-md p-3 text-sm overflow-x-auto">
              <code>{`const users = $database.collection('users');

// Find all documents
const allUsers = users.find({});

// Find with filter
const adults = users.find({ age: { $gte: 18 } });

// Find with options (pagination, sorting)
const topUsers = users.findWithOptions({}, {
  sort: 'score',
  order: 'desc',
  limit: 10,
  page: 1
});

// Find single document
const user = users.findOne({ email: 'john@example.com' });`}</code>
            </pre>
          </div>

          <div className="border rounded-lg p-4">
            <h3 className="font-medium mb-2">Update Documents</h3>
            <pre className="bg-muted rounded-md p-3 text-sm overflow-x-auto">
              <code>{`const users = $database.collection('users');

// Update by document ID
const success = users.update('507f1f77bcf86cd799439011', {
  name: 'John Doe',
  age: 31
});
// Returns: boolean

// Find and update atomically
const updated = users.findOneAndUpdate(
  { email: 'john@example.com' },
  { $set: { verified: true }, $inc: { loginCount: 1 } },
  { returnNew: true }
);

// Upsert - insert if not found, update if exists
const result = users.upsert(
  { email: 'john@example.com' },
  { name: 'John', email: 'john@example.com', role: 'user' }
);`}</code>
            </pre>
          </div>

          <div className="border rounded-lg p-4">
            <h3 className="font-medium mb-2">Delete Documents</h3>
            <pre className="bg-muted rounded-md p-3 text-sm overflow-x-auto">
              <code>{`const users = $database.collection('users');

// Delete by document ID
const success = users.delete('507f1f77bcf86cd799439011');
// Returns: boolean`}</code>
            </pre>
          </div>
        </div>
      </section>

      {/* Query Operators */}
      <section className="space-y-4">
        <h2 className="text-xl font-semibold">Query Operators</h2>

        <div className="grid gap-3 sm:grid-cols-2">
          <div className="border rounded-lg p-3">
            <h3 className="font-medium text-sm mb-2">Comparison</h3>
            <div className="space-y-1 text-xs font-mono">
              <div><span className="text-primary">$eq</span> - Equal</div>
              <div><span className="text-primary">$ne</span> - Not equal</div>
              <div><span className="text-primary">$gt</span> - Greater than</div>
              <div><span className="text-primary">$gte</span> - Greater or equal</div>
              <div><span className="text-primary">$lt</span> - Less than</div>
              <div><span className="text-primary">$lte</span> - Less or equal</div>
              <div><span className="text-primary">$in</span> - In array</div>
              <div><span className="text-primary">$nin</span> - Not in array</div>
            </div>
          </div>

          <div className="border rounded-lg p-3">
            <h3 className="font-medium text-sm mb-2">String</h3>
            <div className="space-y-1 text-xs font-mono">
              <div><span className="text-primary">$contains</span> - Contains substring</div>
              <div><span className="text-primary">$startsWith</span> - Starts with</div>
              <div><span className="text-primary">$endsWith</span> - Ends with</div>
            </div>
          </div>
        </div>

        <div className="border rounded-lg p-4">
          <h3 className="font-medium mb-2">Query Examples</h3>
          <pre className="bg-muted rounded-md p-3 text-sm overflow-x-auto">
            <code>{`const products = $database.collection('products');

// Filter with comparison operators
const affordable = products.find({
  price: { $gte: 10, $lte: 100 },
  category: { $in: ['electronics', 'books'] },
  inStock: true
});

// String search
const articles = $database.collection('articles');
const found = articles.find({
  title: { $contains: 'javascript' }
});

// With pagination and sorting
const topProducts = products.findWithOptions(
  { inStock: true },
  { sort: 'price', order: 'asc', limit: 20, page: 1 }
);`}</code>
          </pre>
        </div>
      </section>

      {/* Update Operators */}
      <section className="space-y-4">
        <h2 className="text-xl font-semibold">Update Operators</h2>
        <p className="text-muted-foreground">
          Update operators are used with <code className="font-mono text-sm bg-muted px-1.5 py-0.5 rounded">findOneAndUpdate</code> method.
        </p>

        <div className="border rounded-lg p-4">
          <div className="grid gap-2 text-sm mb-4">
            {[
              { op: '$set', desc: 'Set field value' },
              { op: '$unset', desc: 'Remove field' },
              { op: '$inc', desc: 'Increment numeric field' },
              { op: '$push', desc: 'Add to array' },
              { op: '$pull', desc: 'Remove from array' },
              { op: '$addToSet', desc: 'Add unique to array' },
            ].map((o) => (
              <div key={o.op} className="flex items-center gap-3">
                <code className="font-mono bg-muted px-2 py-0.5 rounded text-primary w-24">{o.op}</code>
                <span className="text-muted-foreground">{o.desc}</span>
              </div>
            ))}
          </div>

          <pre className="bg-muted rounded-md p-3 text-sm overflow-x-auto">
            <code>{`const stats = $database.collection('stats');
const posts = $database.collection('posts');

// Increment counter
stats.findOneAndUpdate(
  { type: 'views' },
  { $inc: { count: 1 } }
);

// Add tag to array and update timestamp
posts.findOneAndUpdate(
  { slug: 'my-post' },
  {
    $push: { tags: 'featured' },
    $set: { updatedAt: Date.now() }
  },
  { returnNew: true }
);`}</code>
          </pre>
        </div>
      </section>

      {/* Counting */}
      <section className="space-y-4">
        <h2 className="text-xl font-semibold">Counting Documents</h2>

        <div className="border rounded-lg p-4">
          <pre className="bg-muted rounded-md p-3 text-sm overflow-x-auto">
            <code>{`const users = $database.collection('users');

// Count all documents
const total = users.count({});

// Count with filter
const active = users.count({ status: 'active' });

// Count with complex filter
const recentActive = users.count({
  status: 'active',
  lastLogin: { $gte: Date.now() - 7 * 24 * 60 * 60 * 1000 }
});`}</code>
          </pre>
        </div>
      </section>

      {/* Best Practices */}
      <section className="space-y-4">
        <h2 className="text-xl font-semibold">Best Practices</h2>

        <div className="space-y-3">
          <div className="border rounded-lg p-4">
            <h3 className="font-medium mb-2">Use Models for Validation</h3>
            <p className="text-sm text-muted-foreground">
              Define models with schemas to ensure data consistency and catch errors early.
              Set required fields and types to prevent invalid data.
            </p>
          </div>

          <div className="border rounded-lg p-4">
            <h3 className="font-medium mb-2">Check Return Values</h3>
            <pre className="bg-muted rounded-md p-3 text-sm overflow-x-auto">
              <code>{`const users = $database.collection('users');

// insert returns null on failure
const user = users.insert(data);
if (!user) {
  $logger.error('Failed to create user');
  return ctx.response(500, { error: 'Failed to create user' });
}
return { user };

// update/delete return boolean
const success = users.update(id, newData);
if (!success) {
  return ctx.response(404, { error: 'User not found' });
}`}</code>
            </pre>
          </div>

          <div className="border rounded-lg p-4">
            <h3 className="font-medium mb-2">Use Indexes for Performance</h3>
            <p className="text-sm text-muted-foreground">
              Mark frequently queried fields as unique or create indexes through the Models UI
              for better query performance.
            </p>
          </div>
        </div>
      </section>
    </div>
  );
}
