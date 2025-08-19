# Configuration Guide

This document describes the required structure and location for the configuration file used by the application.

## File Location

Place the configuration file at: config/config.json

## File Format

- File type: JSON
- Structure: An array of entries (objects)
- Each entry must include the following fields:
    - `category`: string — the category name (non-empty)
    - `shop website`: string — a valid URL (HTTPS recommended)

Notes:
- Keys are case-sensitive and must match exactly (including the space in `shop website`).
- JSON does not allow comments or trailing commas.
- Use UTF-8 encoding.

## Example
```json
[
    {
        "category": "Electronics",
        "shop website": "https://www.your-electronics-shop.com"
    },
    {
        "category": "Books",
        "shop website": "https://www.your-books-shop.com"
    }
]
```

### JTL-Wawi and Shop Link Requirements

- Categories must exactly match the category names defined in your JTL-Wawi system. This one-to-one match is required so the application can correctly reference and combine item images based on category.
- The `shop website` field must be the URL to the corresponding shop’s homepage for that category. Providing the correct homepage link per category ensures the application can resolve and fetch the appropriate item images to combine.

Fill in the empty strings with real values and add more objects as needed, separated by commas within the array.




