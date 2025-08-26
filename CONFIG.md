# Configuration Guide

This document describes the required structure and location for the configuration file used by the application.

## File Location
Place the configuration file at: `config/config.json`

## File Format
- File type: JSON
- Structure: A top-level object with:
    - `search mode`: string — selects how items are searched; must be either `"category"` or `"supplier"`.
    - `category id`: string — the selected category ID (non-empty).
    - `path to folder`: string — absolute path to the local folder containing your images (non-empty). Each image must be named as follows: `{itemSKU}-{itemImageNumber}.jpg`
    - `activate sales channel`: bool - if `true`, the application will enable all sales channels on the parent item automatically.
    - `mappings`: array of entries (objects):
        - `category`: string — the category name (non-empty).
        - `shop website`: string — a valid URL (HTTPS recommended).

Notes:
- Keys are case-sensitive and must match exactly, including spaces (e.g., `search mode`, `category id`, `path to folder`, `shop website`).
- JSON does not allow comments or trailing commas.
- Use UTF-8 encoding.


## Example

```json
{
  "search mode": "category",
  "category id": "155",
  "path to folder": "C:\\Users\\your-username\\Pictures\\JTL-Wawi-Images",
  "activate sales channel": true,
  "mappings": 
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
}
```

### JTL-Wawi and Shop Link Requirements

- Categories must exactly match the category names defined in your JTL-Wawi system. This one-to-one match is required so the application can correctly reference and combine item images based on category.
- The `shop website` field must be the URL to the corresponding shop’s homepage for that category. Providing the correct homepage link per category ensures the application can resolve and fetch the appropriate item images to combine.

Fill in the real values for `category id`, and add more objects inside `mappings` as needed, separated by commas.




