# JSON Path Filter

The `json_path` assertion supports filter expressions to find elements within arrays based on field values. This is useful when dealing with dynamic responses where array positions may change.

## Filter Syntax

The filter syntax uses `{"key":"value"}` to match objects within an array:

```yaml
- type: "json_path"
  path: "$.array.{"field":"value"}.target"
  expected: "expected_value"
```

## How It Works

1. The filter `{"field":"value"}` searches an array for an object where `field` equals `value`
2. When found, navigation continues from that matched object
3. **The filter must match exactly one element** — multiple matches result in an error

## Examples

### Filter by Unique Field

Given:
```json
[
  {"name": "a", "description": "desc a"},
  {"name": "b", "description": "desc b"}
]
```

Path: `$.{"name":"a"}.description` → `"desc a"`

```yaml
assertions:
  - type: "json_path"
    path: "$.{\"name\":\"a\"}.description"
    expected: "desc a"
```

### Filter in Nested Array

Given:
```json
{
  "slideshow": {
    "slides": [
      {"title": "Wake up to WonderWidgets!", "type": "all"},
      {"title": "Overview", "items": [], "type": "all"}
    ]
  }
}
```

Path: `$.slideshow.slides.{"title":"Overview"}.items` → returns the items array

```yaml
assertions:
  - type: "json_path"
    path: "$.slideshow.slides.{\"title\":\"Overview\"}.items"
    expected: ["Why <em>WonderWidgets</em> are great","Who <em>buys</em> WonderWidgets"]
```

### Filter on Object

Filters can also be applied to objects (not just arrays):

Given:
```json
{
  "slideshow": {
    "author": "Yours Truly",
    "title": "Sample Slide Show"
  }
}
```

Path: `$.slideshow.{"author":"Yours Truly"}.title` → `"Sample Slide Show"`

```yaml
assertions:
  - type: "json_path"
    path: "$.slideshow.{\"author\":\"Yours Truly\"}.title"
    expected: "Sample Slide Show"
```

## Complete Example

```yaml
version: 1
metadata:
  name: "JSON Path Filter Test"
  tags:
    - jsonpath
    - filter
request:
  url: "https://api.example.com/users/1/orders"
  method: GET
  headers:
    Accept: "application/json"

assertions:
  - type: "status_code"
    expected: 200
  - type: "json_path"
    path: "$.orders.{\"status\":\"shipped\"}.tracking_id"
    expected: "TRACK123"
```

## Error Cases

### Filter on Non-Array

Filters can only be applied to arrays or objects:

```
✗ json_path: filter "{"key":"value"}" requires array but got map[string]interface{}
```

### No Matching Element

When no element matches the filter condition:

```
✗ json_path: no item found matching filter {"key":"value"}
```

### Multiple Matches (Ambiguous Filter)

When the filter matches more than one element, an error is returned since the result would be ambiguous:

```
✗ json_path: multiple matches (2) found for filter {"type":"all"}, expected exactly 1
```

**Tip:** Use a more specific filter that matches only one element, or use array indexing first to narrow down the array.

### Filter with String Value

The filter value must be a string:

```yaml
# Correct
path: "$.items.{\"status\":\"active\"}.id"

# The value "active" is a string
```

## Combining with Standard JSONPath

You can use both standard JSONPath and filter syntax in the same path:

```yaml
# Use array index first to narrow array, then filter
path: "$.data[0].items.{\"type\":\"premium\"}.name"
expected: "Gold Plan"
```

## When to Use Filters

Use filters when:
- Array order is unpredictable
- You need to find an item by a unique identifier
- The API returns multiple items and you need a specific one

Use array indexes when:
- The array order is fixed and predictable
- You always need the first, second, or last element
