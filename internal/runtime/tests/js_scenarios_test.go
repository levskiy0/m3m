package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"m3m/internal/runtime/modules"
)

// toFloat64 converts various numeric types to float64
// Goja can return int64 or float64 depending on the value
func toFloat64(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int64:
		return float64(n)
	case int:
		return float64(n)
	default:
		return 0
	}
}

// ============== SCENARIO 1: URL SHORTENER SERVICE ==============
// URL shortener service with in-memory storage

func TestJS_Scenario_URLShortener(t *testing.T) {
	h := NewJSTestHelper(t)
	routerModule := h.SetupRouter()

	// Create URL shortener service
	h.MustRun(t, `
		var links = {};

		// Create short link
		$router.post("/shorten", function(ctx) {
			if (!ctx.body || !ctx.body.url) {
				return ctx.response(400, {error: "URL is required"});
			}

			var url = ctx.body.url;
			// URL validation
			if (!url.startsWith("http://") && !url.startsWith("https://")) {
				return ctx.response(400, {error: "Invalid URL format"});
			}

			// Generate short code
			var shortCode = $crypto.md5(url + $utils.timestamp()).substring(0, 8);
			links[shortCode] = {
				url: url,
				created: $utils.timestamp(),
				clicks: 0
			};

			return ctx.response(201, {
				shortCode: shortCode,
				shortUrl: "/r/" + shortCode
			});
		});

		// Redirect by short link
		$router.get("/r/:code", function(ctx) {
			var code = ctx.params.code;
			var link = links[code];

			if (!link) {
				return ctx.response(404, {error: "Link not found"});
			}

			link.clicks++;
			return ctx.response(302, {
				redirect: link.url,
				clicks: link.clicks
			});
		});

		// Link statistics
		$router.get("/stats/:code", function(ctx) {
			var link = links[ctx.params.code];
			if (!link) {
				return ctx.response(404, {error: "Link not found"});
			}
			return ctx.response(200, {
				url: link.url,
				clicks: link.clicks,
				created: link.created
			});
		});
	`)

	// Test 1: Create short link
	ctx := &modules.RequestContext{
		Method: "POST",
		Path:   "/shorten",
		Body:   map[string]interface{}{"url": "https://example.com/very/long/url"},
	}
	resp, err := routerModule.Handle("POST", "/shorten", ctx)
	if err != nil {
		t.Fatalf("Failed to shorten URL: %v", err)
	}
	if resp.Status != 201 {
		t.Errorf("Expected 201, got %d", resp.Status)
	}

	body := resp.Body.(map[string]interface{})
	shortCode := body["shortCode"].(string)
	if len(shortCode) != 8 {
		t.Errorf("Expected 8 char code, got %d", len(shortCode))
	}

	// Test 2: Redirect
	ctx = &modules.RequestContext{Method: "GET", Path: "/r/" + shortCode}
	resp, _ = routerModule.Handle("GET", "/r/"+shortCode, ctx)
	if resp.Status != 302 {
		t.Errorf("Expected 302 redirect, got %d", resp.Status)
	}

	// Test 3: Statistics
	ctx = &modules.RequestContext{Method: "GET", Path: "/stats/" + shortCode}
	resp, _ = routerModule.Handle("GET", "/stats/"+shortCode, ctx)
	if resp.Status != 200 {
		t.Errorf("Expected 200, got %d", resp.Status)
	}
	body = resp.Body.(map[string]interface{})
	if toFloat64(body["clicks"]) != 1 {
		t.Errorf("Expected 1 click, got %v", body["clicks"])
	}

	// Test 4: Invalid URL
	ctx = &modules.RequestContext{
		Method: "POST",
		Path:   "/shorten",
		Body:   map[string]interface{}{"url": "not-a-url"},
	}
	resp, _ = routerModule.Handle("POST", "/shorten", ctx)
	if resp.Status != 400 {
		t.Errorf("Expected 400 for invalid URL, got %d", resp.Status)
	}

	// Test 5: Non-existent code
	ctx = &modules.RequestContext{Method: "GET", Path: "/r/notexist"}
	resp, _ = routerModule.Handle("GET", "/r/notexist", ctx)
	if resp.Status != 404 {
		t.Errorf("Expected 404, got %d", resp.Status)
	}
}

// ============== SCENARIO 2: WEBHOOK PROCESSOR ==============
// Webhook processor with signature validation

func TestJS_Scenario_WebhookProcessor(t *testing.T) {
	h := NewJSTestHelper(t)
	routerModule := h.SetupRouter()

	h.MustRun(t, `
		var SECRET = "webhook_secret_key";
		var processedEvents = [];

		$router.post("/webhook", function(ctx) {
			// Verify signature
			var signature = ctx.headers["X-Webhook-Signature"];
			if (!signature) {
				return ctx.response(401, {error: "Missing signature"});
			}

			// Calculate expected signature
			var payload = $encoding.jsonStringify(ctx.body);
			var expectedSig = $crypto.sha256(payload + SECRET);

			if (signature !== expectedSig) {
				return ctx.response(401, {error: "Invalid signature"});
			}

			// Process event
			var event = ctx.body;
			if (!event.type || !event.data) {
				return ctx.response(400, {error: "Invalid event format"});
			}

			// Save processed event
			processedEvents.push({
				id: $utils.uuid(),
				type: event.type,
				data: event.data,
				processedAt: $utils.timestamp()
			});

			return ctx.response(200, {
				status: "processed",
				eventId: processedEvents[processedEvents.length - 1].id
			});
		});

		$router.get("/events", function(ctx) {
			return ctx.response(200, {
				count: processedEvents.length,
				events: processedEvents
			});
		});
	`)

	// Test 1: Send webhook with correct signature
	payload := map[string]interface{}{
		"type": "user.created",
		"data": map[string]interface{}{"userId": "123", "email": "test@example.com"},
	}
	payloadJSON, _ := json.Marshal(payload)

	// Calculate signature
	cryptoModule := modules.NewCryptoModule()
	signature := cryptoModule.SHA256(string(payloadJSON) + "webhook_secret_key")

	ctx := &modules.RequestContext{
		Method:  "POST",
		Path:    "/webhook",
		Body:    payload,
		Headers: map[string]string{"X-Webhook-Signature": signature},
	}
	resp, err := routerModule.Handle("POST", "/webhook", ctx)
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}
	if resp.Status != 200 {
		t.Errorf("Expected 200, got %d", resp.Status)
	}

	// Test 2: Without signature
	ctx = &modules.RequestContext{
		Method:  "POST",
		Path:    "/webhook",
		Body:    payload,
		Headers: map[string]string{},
	}
	resp, _ = routerModule.Handle("POST", "/webhook", ctx)
	if resp.Status != 401 {
		t.Errorf("Expected 401 without signature, got %d", resp.Status)
	}

	// Test 3: Invalid signature
	ctx = &modules.RequestContext{
		Method:  "POST",
		Path:    "/webhook",
		Body:    payload,
		Headers: map[string]string{"X-Webhook-Signature": "invalid"},
	}
	resp, _ = routerModule.Handle("POST", "/webhook", ctx)
	if resp.Status != 401 {
		t.Errorf("Expected 401 for invalid signature, got %d", resp.Status)
	}

	// Test 4: Verify event was saved
	ctx = &modules.RequestContext{Method: "GET", Path: "/events"}
	resp, _ = routerModule.Handle("GET", "/events", ctx)
	body := resp.Body.(map[string]interface{})
	if toFloat64(body["count"]) != 1 {
		t.Errorf("Expected 1 event, got %v", body["count"])
	}
}

// ============== SCENARIO 3: API GATEWAY WITH RATE LIMITING ==============
// API Gateway with rate limiting

func TestJS_Scenario_APIGateway(t *testing.T) {
	h := NewJSTestHelper(t)
	routerModule := h.SetupRouter()

	h.MustRun(t, `
		var rateLimits = {}; // clientId -> {count, resetTime}
		var LIMIT = 5;
		var WINDOW = 60000; // 1 minute

		function checkRateLimit(clientId) {
			var now = $utils.timestamp();
			var client = rateLimits[clientId];

			if (!client || client.resetTime < now) {
				rateLimits[clientId] = {
					count: 1,
					resetTime: now + WINDOW
				};
				return {allowed: true, remaining: LIMIT - 1};
			}

			if (client.count >= LIMIT) {
				return {
					allowed: false,
					remaining: 0,
					retryAfter: client.resetTime - now
				};
			}

			client.count++;
			return {allowed: true, remaining: LIMIT - client.count};
		}

		$router.get("/api/:resource", function(ctx) {
			var clientId = ctx.headers["X-Client-ID"] || "anonymous";
			var rateCheck = checkRateLimit(clientId);

			if (!rateCheck.allowed) {
				return ctx.response(429, {
					error: "Rate limit exceeded",
					retryAfter: rateCheck.retryAfter
				});
			}

			// Process request
			var resource = ctx.params.resource;
			return ctx.response(200, {
				resource: resource,
				data: "Sample data for " + resource,
				rateLimitRemaining: rateCheck.remaining
			});
		});
	`)

	// Test: Send requests up to limit
	for i := 0; i < 5; i++ {
		ctx := &modules.RequestContext{
			Method:  "GET",
			Path:    "/api/users",
			Headers: map[string]string{"X-Client-ID": "client1"},
		}
		resp, _ := routerModule.Handle("GET", "/api/users", ctx)
		if resp.Status != 200 {
			t.Errorf("Request %d: Expected 200, got %d", i+1, resp.Status)
		}
	}

	// Next request should be blocked
	ctx := &modules.RequestContext{
		Method:  "GET",
		Path:    "/api/users",
		Headers: map[string]string{"X-Client-ID": "client1"},
	}
	resp, _ := routerModule.Handle("GET", "/api/users", ctx)
	if resp.Status != 429 {
		t.Errorf("Expected 429 after limit, got %d", resp.Status)
	}

	// Different client should work
	ctx = &modules.RequestContext{
		Method:  "GET",
		Path:    "/api/users",
		Headers: map[string]string{"X-Client-ID": "client2"},
	}
	resp, _ = routerModule.Handle("GET", "/api/users", ctx)
	if resp.Status != 200 {
		t.Errorf("Different client should not be limited, got %d", resp.Status)
	}
}

// ============== SCENARIO 4: USER REGISTRATION SERVICE ==============
// User registration service with validation

func TestJS_Scenario_UserRegistration(t *testing.T) {
	h := NewJSTestHelper(t)
	routerModule := h.SetupRouter()

	h.MustRun(t, `
		var users = {};

		function validateEmail(email) {
			return email && email.indexOf("@") > 0 && email.indexOf(".") > email.indexOf("@");
		}

		function validatePassword(password) {
			return password && password.length >= 8;
		}

		$router.post("/register", function(ctx) {
			var body = ctx.body;
			if (!body) {
				return ctx.response(400, {error: "Request body required"});
			}

			// Field validation
			var errors = [];

			if (!body.email) {
				errors.push("Email is required");
			} else if (!validateEmail(body.email)) {
				errors.push("Invalid email format");
			} else if (users[body.email]) {
				errors.push("Email already registered");
			}

			if (!body.password) {
				errors.push("Password is required");
			} else if (!validatePassword(body.password)) {
				errors.push("Password must be at least 8 characters");
			}

			if (!body.name) {
				errors.push("Name is required");
			}

			if (errors.length > 0) {
				return ctx.response(400, {errors: errors});
			}

			// Create user
			var userId = $utils.uuid();
			var passwordHash = $crypto.sha256(body.password);

			users[body.email] = {
				id: userId,
				email: body.email,
				name: body.name,
				passwordHash: passwordHash,
				createdAt: $utils.timestamp()
			};

			return ctx.response(201, {
				userId: userId,
				email: body.email,
				name: body.name
			});
		});

		$router.post("/login", function(ctx) {
			var body = ctx.body;
			if (!body || !body.email || !body.password) {
				return ctx.response(400, {error: "Email and password required"});
			}

			var user = users[body.email];
			if (!user) {
				return ctx.response(401, {error: "Invalid credentials"});
			}

			var passwordHash = $crypto.sha256(body.password);
			if (user.passwordHash !== passwordHash) {
				return ctx.response(401, {error: "Invalid credentials"});
			}

			// Generate token
			var token = $crypto.randomBytes(32);
			return ctx.response(200, {
				token: token,
				userId: user.id
			});
		});
	`)

	// Test 1: Successful registration
	ctx := &modules.RequestContext{
		Method: "POST",
		Path:   "/register",
		Body: map[string]interface{}{
			"email":    "test@example.com",
			"password": "password123",
			"name":     "Test User",
		},
	}
	resp, _ := routerModule.Handle("POST", "/register", ctx)
	if resp.Status != 201 {
		t.Errorf("Expected 201, got %d", resp.Status)
	}

	// Test 2: Duplicate email
	resp, _ = routerModule.Handle("POST", "/register", ctx)
	if resp.Status != 400 {
		t.Errorf("Expected 400 for duplicate email, got %d", resp.Status)
	}

	// Test 3: Invalid email
	ctx = &modules.RequestContext{
		Method: "POST",
		Path:   "/register",
		Body: map[string]interface{}{
			"email":    "invalid-email",
			"password": "password123",
			"name":     "Test User",
		},
	}
	resp, _ = routerModule.Handle("POST", "/register", ctx)
	if resp.Status != 400 {
		t.Errorf("Expected 400 for invalid email, got %d", resp.Status)
	}

	// Test 4: Short password
	ctx = &modules.RequestContext{
		Method: "POST",
		Path:   "/register",
		Body: map[string]interface{}{
			"email":    "test2@example.com",
			"password": "short",
			"name":     "Test User",
		},
	}
	resp, _ = routerModule.Handle("POST", "/register", ctx)
	if resp.Status != 400 {
		t.Errorf("Expected 400 for short password, got %d", resp.Status)
	}

	// Test 5: Successful login
	ctx = &modules.RequestContext{
		Method: "POST",
		Path:   "/login",
		Body: map[string]interface{}{
			"email":    "test@example.com",
			"password": "password123",
		},
	}
	resp, _ = routerModule.Handle("POST", "/login", ctx)
	if resp.Status != 200 {
		t.Errorf("Expected 200 for login, got %d", resp.Status)
	}

	// Test 6: Wrong password
	ctx = &modules.RequestContext{
		Method: "POST",
		Path:   "/login",
		Body: map[string]interface{}{
			"email":    "test@example.com",
			"password": "wrongpassword",
		},
	}
	resp, _ = routerModule.Handle("POST", "/login", ctx)
	if resp.Status != 401 {
		t.Errorf("Expected 401 for wrong password, got %d", resp.Status)
	}
}

// ============== SCENARIO 5: DATA TRANSFORMATION PIPELINE ==============
// Data transformation pipeline

func TestJS_Scenario_DataPipeline(t *testing.T) {
	h := NewJSTestHelper(t)
	routerModule := h.SetupRouter()

	h.MustRun(t, `
		$router.post("/transform", function(ctx) {
			if (!ctx.body || !ctx.body.data) {
				return ctx.response(400, {error: "Data array required"});
			}

			var data = ctx.body.data;
			if (!Array.isArray(data)) {
				return ctx.response(400, {error: "Data must be an array"});
			}

			var operations = ctx.body.operations || [];
			var result = data;

			for (var i = 0; i < operations.length; i++) {
				var op = operations[i];

				switch (op.type) {
					case "filter":
						result = result.filter(function(item) {
							return item[op.field] === op.value;
						});
						break;

					case "map":
						result = result.map(function(item) {
							var newItem = {};
							for (var key in item) {
								newItem[key] = item[key];
							}
							if (op.addField) {
								newItem[op.addField] = op.addValue;
							}
							if (op.slugifyField) {
								newItem[op.slugifyField + "_slug"] = $utils.slugify(item[op.slugifyField] || "");
							}
							if (op.hashField) {
								newItem[op.hashField + "_hash"] = $crypto.md5(item[op.hashField] || "");
							}
							return newItem;
						});
						break;

					case "sort":
						result = result.sort(function(a, b) {
							if (a[op.field] < b[op.field]) return op.order === "desc" ? 1 : -1;
							if (a[op.field] > b[op.field]) return op.order === "desc" ? -1 : 1;
							return 0;
						});
						break;

					case "limit":
						result = result.slice(0, op.count);
						break;
				}
			}

			return ctx.response(200, {
				originalCount: data.length,
				resultCount: result.length,
				data: result
			});
		});
	`)

	// Test: Complex transformation pipeline
	ctx := &modules.RequestContext{
		Method: "POST",
		Path:   "/transform",
		Body: map[string]interface{}{
			"data": []interface{}{
				map[string]interface{}{"name": "Alice", "age": 25, "role": "admin"},
				map[string]interface{}{"name": "Bob", "age": 30, "role": "user"},
				map[string]interface{}{"name": "Charlie", "age": 35, "role": "admin"},
				map[string]interface{}{"name": "Diana", "age": 28, "role": "user"},
			},
			"operations": []interface{}{
				map[string]interface{}{"type": "filter", "field": "role", "value": "admin"},
				map[string]interface{}{"type": "map", "slugifyField": "name"},
				map[string]interface{}{"type": "sort", "field": "age", "order": "desc"},
			},
		},
	}

	resp, err := routerModule.Handle("POST", "/transform", ctx)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}
	if resp.Status != 200 {
		t.Errorf("Expected 200, got %d", resp.Status)
	}

	body := resp.Body.(map[string]interface{})
	if toFloat64(body["resultCount"]) != 2 {
		t.Errorf("Expected 2 results after filter, got %v", body["resultCount"])
	}

	data := body["data"].([]interface{})
	first := data[0].(map[string]interface{})
	if first["name"] != "Charlie" {
		t.Errorf("Expected Charlie first (oldest admin), got %v", first["name"])
	}
	if _, ok := first["name_slug"]; !ok {
		t.Error("Expected name_slug field after map operation")
	}
}

// ============== SCENARIO 6: HEALTH CHECK SERVICE ==============
// System health check service

func TestJS_Scenario_HealthCheck(t *testing.T) {
	h := NewJSTestHelper(t)
	routerModule := h.SetupRouter()

	h.MustRun(t, `
		var startTime = $utils.timestamp();
		var requestCount = 0;

		$router.get("/health", function(ctx) {
			requestCount++;

			var uptime = $utils.timestamp() - startTime;

			return ctx.response(200, {
				status: "healthy",
				uptime: uptime,
				requestCount: requestCount,
				timestamp: $utils.timestamp(),
				version: "1.0.0"
			});
		});

		$router.get("/health/detailed", function(ctx) {
			requestCount++;

			var checks = {
				api: {status: "ok", latency: $utils.randomInt(1, 10)},
				memory: {status: "ok", used: $utils.randomInt(50, 80) + "%"},
				connections: {status: "ok", active: $utils.randomInt(1, 100)}
			};

			var allHealthy = true;
			for (var key in checks) {
				if (checks[key].status !== "ok") {
					allHealthy = false;
					break;
				}
			}

			return ctx.response(allHealthy ? 200 : 503, {
				status: allHealthy ? "healthy" : "degraded",
				checks: checks,
				timestamp: $utils.timestamp()
			});
		});
	`)

	// Test 1: Basic health check
	ctx := &modules.RequestContext{Method: "GET", Path: "/health"}
	resp, _ := routerModule.Handle("GET", "/health", ctx)
	if resp.Status != 200 {
		t.Errorf("Expected 200, got %d", resp.Status)
	}
	body := resp.Body.(map[string]interface{})
	if body["status"] != "healthy" {
		t.Errorf("Expected healthy status")
	}

	// Test 2: Detailed check
	ctx = &modules.RequestContext{Method: "GET", Path: "/health/detailed"}
	resp, _ = routerModule.Handle("GET", "/health/detailed", ctx)
	if resp.Status != 200 {
		t.Errorf("Expected 200, got %d", resp.Status)
	}
	body = resp.Body.(map[string]interface{})
	checks := body["checks"].(map[string]interface{})
	if _, ok := checks["api"]; !ok {
		t.Error("Expected api check")
	}

	// Test 3: Request counter increases
	ctx = &modules.RequestContext{Method: "GET", Path: "/health"}
	resp, _ = routerModule.Handle("GET", "/health", ctx)
	body = resp.Body.(map[string]interface{})
	if toFloat64(body["requestCount"]) < 2 {
		t.Error("Request count should increase")
	}
}

// ============== SCENARIO 7: CONTENT CACHE SERVICE ==============
// Content caching service with TTL

func TestJS_Scenario_ContentCache(t *testing.T) {
	h := NewJSTestHelper(t)
	routerModule := h.SetupRouter()

	h.MustRun(t, `
		var cache = {};
		var DEFAULT_TTL = 60000; // 1 minute

		$router.post("/cache", function(ctx) {
			if (!ctx.body || !ctx.body.key || !ctx.body.value) {
				return ctx.response(400, {error: "Key and value required"});
			}

			var key = ctx.body.key;
			var value = ctx.body.value;
			var ttl = ctx.body.ttl || DEFAULT_TTL;

			cache[key] = {
				value: value,
				expires: $utils.timestamp() + ttl,
				createdAt: $utils.timestamp()
			};

			return ctx.response(201, {
				key: key,
				expires: cache[key].expires
			});
		});

		$router.get("/cache/:key", function(ctx) {
			var key = ctx.params.key;
			var entry = cache[key];

			if (!entry) {
				return ctx.response(404, {error: "Key not found"});
			}

			if (entry.expires < $utils.timestamp()) {
				delete cache[key];
				return ctx.response(404, {error: "Key expired"});
			}

			return ctx.response(200, {
				key: key,
				value: entry.value,
				expires: entry.expires,
				ttlRemaining: entry.expires - $utils.timestamp()
			});
		});

		$router.delete("/cache/:key", function(ctx) {
			var key = ctx.params.key;
			if (!cache[key]) {
				return ctx.response(404, {error: "Key not found"});
			}
			delete cache[key];
			return ctx.response(204, {});
		});
	`)

	// Test 1: Save to cache
	ctx := &modules.RequestContext{
		Method: "POST",
		Path:   "/cache",
		Body: map[string]interface{}{
			"key":   "user:123",
			"value": map[string]interface{}{"name": "John", "age": 30},
			"ttl":   300000,
		},
	}
	resp, _ := routerModule.Handle("POST", "/cache", ctx)
	if resp.Status != 201 {
		t.Errorf("Expected 201, got %d", resp.Status)
	}

	// Test 2: Get from cache
	ctx = &modules.RequestContext{Method: "GET", Path: "/cache/user:123"}
	resp, _ = routerModule.Handle("GET", "/cache/user:123", ctx)
	if resp.Status != 200 {
		t.Errorf("Expected 200, got %d", resp.Status)
	}

	// Test 3: Non-existent key
	ctx = &modules.RequestContext{Method: "GET", Path: "/cache/nonexistent"}
	resp, _ = routerModule.Handle("GET", "/cache/nonexistent", ctx)
	if resp.Status != 404 {
		t.Errorf("Expected 404, got %d", resp.Status)
	}

	// Test 4: Delete
	ctx = &modules.RequestContext{Method: "DELETE", Path: "/cache/user:123"}
	resp, _ = routerModule.Handle("DELETE", "/cache/user:123", ctx)
	if resp.Status != 204 {
		t.Errorf("Expected 204, got %d", resp.Status)
	}

	// Test 5: Verify deleted
	ctx = &modules.RequestContext{Method: "GET", Path: "/cache/user:123"}
	resp, _ = routerModule.Handle("GET", "/cache/user:123", ctx)
	if resp.Status != 404 {
		t.Errorf("Expected 404 after delete, got %d", resp.Status)
	}
}

// ============== SCENARIO 8: ANALYTICS TRACKER ==============
// Analytics tracking service

func TestJS_Scenario_AnalyticsTracker(t *testing.T) {
	h := NewJSTestHelper(t)
	routerModule := h.SetupRouter()

	h.MustRun(t, `
		var events = [];
		var pageViews = {};
		var userSessions = {};

		$router.post("/track/event", function(ctx) {
			var body = ctx.body;
			if (!body || !body.name) {
				return ctx.response(400, {error: "Event name required"});
			}

			var event = {
				id: $utils.uuid(),
				name: body.name,
				properties: body.properties || {},
				userId: body.userId || "anonymous",
				timestamp: $utils.timestamp(),
				userAgent: ctx.headers["User-Agent"] || "unknown"
			};

			events.push(event);

			return ctx.response(201, {eventId: event.id});
		});

		$router.post("/track/pageview", function(ctx) {
			var body = ctx.body;
			if (!body || !body.path) {
				return ctx.response(400, {error: "Path required"});
			}

			var path = body.path;
			pageViews[path] = (pageViews[path] || 0) + 1;

			return ctx.response(201, {
				path: path,
				totalViews: pageViews[path]
			});
		});

		$router.get("/analytics/summary", function(ctx) {
			var topPages = [];
			for (var path in pageViews) {
				topPages.push({path: path, views: pageViews[path]});
			}
			topPages.sort(function(a, b) { return b.views - a.views; });

			return ctx.response(200, {
				totalEvents: events.length,
				totalPageViews: topPages.reduce(function(sum, p) { return sum + p.views; }, 0),
				topPages: topPages.slice(0, 10),
				recentEvents: events.slice(-10)
			});
		});
	`)

	// Send events
	ctx := &modules.RequestContext{
		Method: "POST",
		Path:   "/track/event",
		Body: map[string]interface{}{
			"name":       "button_click",
			"properties": map[string]interface{}{"button": "signup"},
			"userId":     "user123",
		},
		Headers: map[string]string{"User-Agent": "TestBrowser/1.0"},
	}
	resp, _ := routerModule.Handle("POST", "/track/event", ctx)
	if resp.Status != 201 {
		t.Errorf("Expected 201, got %d", resp.Status)
	}

	// Send page views
	pages := []string{"/home", "/products", "/home", "/about", "/home"}
	for _, page := range pages {
		ctx = &modules.RequestContext{
			Method: "POST",
			Path:   "/track/pageview",
			Body:   map[string]interface{}{"path": page},
		}
		routerModule.Handle("POST", "/track/pageview", ctx)
	}

	// Check analytics
	ctx = &modules.RequestContext{Method: "GET", Path: "/analytics/summary"}
	resp, _ = routerModule.Handle("GET", "/analytics/summary", ctx)
	if resp.Status != 200 {
		t.Errorf("Expected 200, got %d", resp.Status)
	}

	body := resp.Body.(map[string]interface{})
	if toFloat64(body["totalEvents"]) != 1 {
		t.Errorf("Expected 1 event, got %v", body["totalEvents"])
	}
	if toFloat64(body["totalPageViews"]) != 5 {
		t.Errorf("Expected 5 page views, got %v", body["totalPageViews"])
	}

	topPages := body["topPages"].([]interface{})
	if len(topPages) > 0 {
		first := topPages[0].(map[string]interface{})
		if first["path"] != "/home" {
			t.Errorf("Expected /home as top page, got %v", first["path"])
		}
	}
}

// ============== SCENARIO 9: PAYMENT WEBHOOK HANDLER ==============
// Payment webhook handler

func TestJS_Scenario_PaymentWebhook(t *testing.T) {
	h := NewJSTestHelper(t)
	routerModule := h.SetupRouter()

	h.MustRun(t, `
		var payments = {};
		var orders = {
			"order-001": {status: "pending", amount: 100},
			"order-002": {status: "pending", amount: 250}
		};

		$router.post("/payment/webhook", function(ctx) {
			var body = ctx.body;
			if (!body || !body.event) {
				return ctx.response(400, {error: "Invalid webhook payload"});
			}

			var event = body.event;
			var data = body.data || {};

			switch (event) {
				case "payment.created":
					payments[data.paymentId] = {
						status: "pending",
						amount: data.amount,
						orderId: data.orderId,
						createdAt: $utils.timestamp()
					};
					break;

				case "payment.completed":
					if (!payments[data.paymentId]) {
						return ctx.response(404, {error: "Payment not found"});
					}
					payments[data.paymentId].status = "completed";
					payments[data.paymentId].completedAt = $utils.timestamp();

					// Update order
					if (data.orderId && orders[data.orderId]) {
						orders[data.orderId].status = "paid";
					}
					break;

				case "payment.failed":
					if (!payments[data.paymentId]) {
						return ctx.response(404, {error: "Payment not found"});
					}
					payments[data.paymentId].status = "failed";
					payments[data.paymentId].failReason = data.reason || "Unknown";
					break;

				case "payment.refunded":
					if (!payments[data.paymentId]) {
						return ctx.response(404, {error: "Payment not found"});
					}
					payments[data.paymentId].status = "refunded";
					payments[data.paymentId].refundedAt = $utils.timestamp();
					break;

				default:
					return ctx.response(400, {error: "Unknown event type"});
			}

			return ctx.response(200, {
				received: true,
				event: event,
				timestamp: $utils.timestamp()
			});
		});

		$router.get("/payment/:id", function(ctx) {
			var payment = payments[ctx.params.id];
			if (!payment) {
				return ctx.response(404, {error: "Payment not found"});
			}
			return ctx.response(200, payment);
		});

		$router.get("/order/:id", function(ctx) {
			var order = orders[ctx.params.id];
			if (!order) {
				return ctx.response(404, {error: "Order not found"});
			}
			return ctx.response(200, order);
		});
	`)

	// Create payment
	ctx := &modules.RequestContext{
		Method: "POST",
		Path:   "/payment/webhook",
		Body: map[string]interface{}{
			"event": "payment.created",
			"data": map[string]interface{}{
				"paymentId": "pay-001",
				"amount":    100,
				"orderId":   "order-001",
			},
		},
	}
	resp, _ := routerModule.Handle("POST", "/payment/webhook", ctx)
	if resp.Status != 200 {
		t.Errorf("Expected 200, got %d", resp.Status)
	}

	// Complete payment
	ctx = &modules.RequestContext{
		Method: "POST",
		Path:   "/payment/webhook",
		Body: map[string]interface{}{
			"event": "payment.completed",
			"data": map[string]interface{}{
				"paymentId": "pay-001",
				"orderId":   "order-001",
			},
		},
	}
	resp, _ = routerModule.Handle("POST", "/payment/webhook", ctx)
	if resp.Status != 200 {
		t.Errorf("Expected 200, got %d", resp.Status)
	}

	// Check payment status
	ctx = &modules.RequestContext{Method: "GET", Path: "/payment/pay-001"}
	resp, _ = routerModule.Handle("GET", "/payment/pay-001", ctx)
	body := resp.Body.(map[string]interface{})
	if body["status"] != "completed" {
		t.Errorf("Expected completed status, got %v", body["status"])
	}

	// Check order status
	ctx = &modules.RequestContext{Method: "GET", Path: "/order/order-001"}
	resp, _ = routerModule.Handle("GET", "/order/order-001", ctx)
	body = resp.Body.(map[string]interface{})
	if body["status"] != "paid" {
		t.Errorf("Expected paid status, got %v", body["status"])
	}

	// Unknown event type
	ctx = &modules.RequestContext{
		Method: "POST",
		Path:   "/payment/webhook",
		Body:   map[string]interface{}{"event": "unknown.event"},
	}
	resp, _ = routerModule.Handle("POST", "/payment/webhook", ctx)
	if resp.Status != 400 {
		t.Errorf("Expected 400 for unknown event, got %d", resp.Status)
	}
}

// ============== SCENARIO 10: HTTP PROXY/AGGREGATOR SERVICE ==============
// Data aggregation service from external API

func TestJS_Scenario_HTTPAggregator(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/users":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`[{"id": 1, "name": "Alice"}, {"id": 2, "name": "Bob"}]`))
		case "/posts":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`[{"id": 1, "title": "Hello"}, {"id": 2, "title": "World"}]`))
		default:
			w.WriteHeader(404)
		}
	}))
	defer server.Close()

	h := NewJSTestHelper(t)

	// Register HTTP module
	httpModule := modules.NewHTTPModule(30 * time.Second)
	h.VM.Set("$http", map[string]interface{}{
		"get":    httpModule.Get,
		"post":   httpModule.Post,
		"put":    httpModule.Put,
		"delete": httpModule.Delete,
	})

	routerModule := h.SetupRouter()

	code := fmt.Sprintf(`
		var API_BASE = "%s";

		$router.get("/aggregate", function(ctx) {
			// Make requests to external API
			var usersResp = $http.get(API_BASE + "/users");
			var postsResp = $http.get(API_BASE + "/posts");

			if (usersResp.status !== 200 || postsResp.status !== 200) {
				return ctx.response(502, {error: "Failed to fetch data from upstream"});
			}

			var users = $encoding.jsonParse(usersResp.body) || [];
			var posts = $encoding.jsonParse(postsResp.body) || [];

			// Aggregate data
			var result = {
				users: users,
				posts: posts,
				summary: {
					userCount: users.length,
					postCount: posts.length,
					fetchedAt: $utils.timestamp()
				}
			};

			return ctx.response(200, result);
		});

		$router.get("/users/:id/posts", function(ctx) {
			var userId = parseInt(ctx.params.id);

			var usersResp = $http.get(API_BASE + "/users");
			var postsResp = $http.get(API_BASE + "/posts");

			if (usersResp.status !== 200) {
				return ctx.response(502, {error: "Failed to fetch users"});
			}

			var users = $encoding.jsonParse(usersResp.body) || [];
			var user = null;
			for (var i = 0; i < users.length; i++) {
				if (users[i].id === userId) {
					user = users[i];
					break;
				}
			}

			if (!user) {
				return ctx.response(404, {error: "User not found"});
			}

			return ctx.response(200, {
				user: user,
				posts: $encoding.jsonParse(postsResp.body) || []
			});
		});
	`, server.URL)

	h.MustRun(t, code)

	// Test 1: Data aggregation
	ctx := &modules.RequestContext{Method: "GET", Path: "/aggregate"}
	resp, err := routerModule.Handle("GET", "/aggregate", ctx)
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}
	if resp.Status != 200 {
		t.Errorf("Expected 200, got %d", resp.Status)
	}

	body := resp.Body.(map[string]interface{})
	summary := body["summary"].(map[string]interface{})
	if toFloat64(summary["userCount"]) != 2 {
		t.Errorf("Expected 2 users, got %v", summary["userCount"])
	}
	if toFloat64(summary["postCount"]) != 2 {
		t.Errorf("Expected 2 posts, got %v", summary["postCount"])
	}

	// Test 2: Get user with posts
	ctx = &modules.RequestContext{Method: "GET", Path: "/users/1/posts"}
	resp, _ = routerModule.Handle("GET", "/users/1/posts", ctx)
	if resp.Status != 200 {
		t.Errorf("Expected 200, got %d", resp.Status)
	}

	body = resp.Body.(map[string]interface{})
	user := body["user"].(map[string]interface{})
	if user["name"] != "Alice" {
		t.Errorf("Expected Alice, got %v", user["name"])
	}

	// Test 3: Non-existent user
	ctx = &modules.RequestContext{Method: "GET", Path: "/users/999/posts"}
	resp, _ = routerModule.Handle("GET", "/users/999/posts", ctx)
	if resp.Status != 404 {
		t.Errorf("Expected 404, got %d", resp.Status)
	}
}

// ============== ERROR HANDLING TESTS ==============

func TestJS_Error_SyntaxError(t *testing.T) {
	h := NewJSTestHelper(t)

	_, err := h.Run(`function broken( { }`)
	if err == nil {
		t.Error("Should return error for syntax error")
	}
}

func TestJS_Error_RuntimeError(t *testing.T) {
	h := NewJSTestHelper(t)

	_, err := h.Run(`nonExistentFunction()`)
	if err == nil {
		t.Error("Should return error for undefined function call")
	}
}

func TestJS_Error_TypeError(t *testing.T) {
	h := NewJSTestHelper(t)

	_, err := h.Run(`null.property`)
	if err == nil {
		t.Error("Should return error for null property access")
	}
}

func TestJS_Error_HandlerWithTryCatch(t *testing.T) {
	h := NewJSTestHelper(t)
	routerModule := h.SetupRouter()

	h.MustRun(t, `
		$router.get("/safe", function(ctx) {
			try {
				// Potentially failing code
				var result = someUndefinedVar;
				return ctx.response(200, {result: result});
			} catch (e) {
				return ctx.response(500, {error: e.message});
			}
		});
	`)

	ctx := &modules.RequestContext{Method: "GET", Path: "/safe"}
	resp, err := routerModule.Handle("GET", "/safe", ctx)
	if err != nil {
		t.Fatalf("Handler should not fail with try/catch: %v", err)
	}
	if resp.Status != 500 {
		t.Errorf("Expected 500, got %d", resp.Status)
	}
}
