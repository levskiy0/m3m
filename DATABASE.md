# $database Module

Module for working with model data. Uses MongoDB-like query syntax.

## Getting a Collection

```javascript
const users = $database.collection("users"); // model slug
```

## Collection Methods

### find(filter?)

Find documents matching filter.

```javascript
// All documents
const all = users.find({});

// By field
const active = users.find({ status: "active" });

// With operators
const adults = users.find({ age: { $gte: 18 } });
```

### findWithOptions(filter?, options?)

Find with pagination and sorting.

```javascript
const result = users.findWithOptions(
  { status: "active" },
  {
    page: 1,
    limit: 10,
    sort: "createdAt",
    order: "desc"
  }
);
```

**Options:**
- `page` - page number (default: 1)
- `limit` - results per page (default: 100)
- `sort` - field to sort by
- `order` - `"asc"` or `"desc"`

### findOne(filter?)

Find first matching document.

```javascript
const user = users.findOne({ email: "test@example.com" });
if (user) {
  $logger.info("Found:", user.name);
}
```

### insert(data)

Insert a document. Returns created document or `null`.

```javascript
const newUser = users.insert({
  name: "John",
  email: "john@example.com",
  age: 25
});

if (newUser) {
  $logger.info("Created with ID:", newUser._id);
}
```

### update(id, data)

Update document by ID. Returns `true`/`false`.

```javascript
const success = users.update("507f1f77bcf86cd799439011", {
  name: "John Updated",
  age: 26
});
```

### delete(id)

Delete document by ID. Returns `true`/`false`.

```javascript
const deleted = users.delete("507f1f77bcf86cd799439011");
```

### count(filter?)

Count matching documents.

```javascript
const total = users.count({});
const activeCount = users.count({ status: "active" });
```

### upsert(filter, data)

Update if found, otherwise insert.

```javascript
const user = users.upsert(
  { email: "john@example.com" },
  { name: "John", email: "john@example.com", visits: 1 }
);
```

### findOneAndUpdate(filter, update, options?)

Atomically find and update a document. Supports update operators.

```javascript
// Increment counter
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
);
```

**Options:**
- `returnNew` - return updated document instead of original (default: `false`)

## Filter Operators

| Operator | Description |
|----------|-------------|
| `$eq` | Equal |
| `$ne` | Not equal |
| `$gt` | Greater than |
| `$gte` | Greater than or equal |
| `$lt` | Less than |
| `$lte` | Less than or equal |
| `$contains` | Contains substring |
| `$startsWith` | Starts with |
| `$endsWith` | Ends with |
| `$in` | In array |
| `$nin` | Not in array |

```javascript
// Examples
users.find({ age: { $gt: 18, $lt: 65 } });
users.find({ status: { $in: ["active", "pending"] } });
users.find({ name: { $contains: "John" } });
users.find({ email: { $endsWith: "@gmail.com" } });
```

## Update Operators

For `findOneAndUpdate`:

| Operator | Description |
|----------|-------------|
| `$set` | Set field value |
| `$inc` | Increment numeric value |
| `$unset` | Remove field |
| `$push` | Add element to array |
| `$pull` | Remove element from array |
| `$addToSet` | Add unique element to array |

```javascript
// Increment counter by 1
users.findOneAndUpdate({ _id: id }, { $inc: { count: 1 } });

// Add tag
users.findOneAndUpdate({ _id: id }, { $push: { tags: "new-tag" } });

// Remove field
users.findOneAndUpdate({ _id: id }, { $unset: { tempField: "" } });
```

## Example: Simple CRUD API

```javascript
const posts = $database.collection("posts");

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
});
```
