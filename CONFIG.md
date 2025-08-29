# Configuration Guide

This document describes the required structure and location for the configuration file used by the application.

## File Location
Place the configuration file at: `config/config.json`
Or supply the path to the file using the `-c` flag when running the application from the command line.

## File Format
- File type: JSON
- Structure: A top-level object with:
    - `api base url`: string - the base URL of the JTL-Wawi API. The default is `"http://127.0.0.1:5883/api/eazybusiness/"`
    - `search mode`: string — selects how items are searched; must be either `"category"`, `"supplier"` or `none`.
    - `category id`: string — the selected category ID (non-empty).
    - `path to folder`: string — absolute path to the local folder containing your images (non-empty). Each image must be named as follows: `{itemSKU}-{itemImageNumber}.jpg`
    - `activate sales channel`: bool - if `true`, the application will enable all sales channels on the parent item automatically.

Notes:
- Keys are case-sensitive and must match exactly, including spaces (e.g., `search mode`, `category id`, `path to folder`, `activate sales channel`).
- JSON does not allow comments or trailing commas.
- Use UTF-8 encoding.


## Example

```json
{
  "api base url": "http://127.0.0.1:5883/api/eazybusiness/",
  "search mode": "category",
  "category id": "155",
  "path to folder": "C:\\Users\\your-username\\Pictures\\JTL-Wawi-Images",
  "activate sales channel": true
}
```


