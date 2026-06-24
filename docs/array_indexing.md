# Array Indexing

The `json_path` assertion supports full array indexing syntax to access elements within JSON arrays.

## Syntax

| Expression | Description |
|------------|-------------|
| `$.array[0]` | First element (0-indexed) |
| `$.array[n]` | nth element |
| `$.array[-1]` | Last element |
| `$.array[-2]` | Second-to-last element |
| `$.obj.array[0].field` | Array element with nested field access |

## Index Notation

- **Positive indexes**: `0` is the first element, `1` is second, etc.
- **Negative indexes**: `-1` is the last element, `-2` is second-to-last, etc.

## Examples

### First Element

```yaml
assertions:
  - type: "json_path"
    path: "$.users[0].name"
    expected: "Alice"
```

### Last Element

```yaml
assertions:
  - type: "json_path"
    path: "$.items[-1].id"
    expected: 42
```

### Specific Index

```yaml
assertions:
  - type: "json_path"
    path: "$.slideshow.slides[2].title"
    expected: "Index 2"
```

### Combined with Nested Fields

```yaml
assertions:
  - type: "json_path"
    path: "$.data.products[0].details.price"
    expected: 29.99
```

## Complete Example

```yaml
name: "Array Indexing Test"
request:
  url: "https://api.example.com/orders"
  method: GET
  headers:
    Accept: "application/json"

assertions:
  - type: "status_code"
    expected: 200
  - type: "json_path"
    path: "$.orders[0].customer"
    expected: "Alice"
  - type: "json_path"
    path: "$.orders[-1].status"
    expected: "completed"
  - type: "json_path"
    path: "$.orders[1].total"
    expected: 150.00
```

## Output

```
Array Indexing Test
===================
  ✓ PASS [200] 234ms
    ✓ status_code: status code: expected 200, got 200
    ✓ json_path: path "$.orders[0].customer" == "Alice"
    ✓ json_path: path "$.orders[-1].status" == "completed"
    ✓ json_path: path "$.orders[1].total" == "150"

SUMMARY
----------------------------------------
Total:  1
Passed: 1
Failed: 0
```

## Error Cases

### Index Out of Bounds

If the array has fewer elements than the specified index:

```
✗ json_path: path "$.array[100]": path "$.array[100]" not found in JSON
```

### Empty Array

```
✗ json_path: path "$.array[0].name": path "$.array[0].name" not found in JSON
```

### Mixed Array Types

JSON arrays may contain mixed types. Ensure the expected index contains the expected type:

```json
{
  "mixed": [1, "string", {"key": "value"}, null]
}
```

```yaml
assertions:
  - type: "json_path"
    path: "$.mixed[0]"
    expected: 1
  - type: "json_path"
    path: "$.mixed[2].key"
    expected: "value"
```
