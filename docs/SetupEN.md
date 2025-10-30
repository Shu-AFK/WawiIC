### Starting the JTL-Wawi API Server
> Inside the installation directory of JTL-Wawi there is a file called `JTL.Wawi.Rest.exe`.  
> This is the API server. The default installation path is:  
> `C:\Program Files (x86)\JTL-Software`  
> To start the server, run the following commands inside a terminal:

```sh
> cd C:\Program Files (x86)\JTL-Software
> JTL.Wawi.Rest.exe -w "Standard" -l 127.0.0.1
```

> Replace the path after `cd` with the correct installation path of JTL-Software.  
> If you have set up a different profile name for JTL-Wawi, replace `"Standard"` with that profile name.

---

### Requirements
#### 1. Wawi API Key
> If you don’t have an API key yet:  
> JTL-Wawi Admin → App registration → Add → Continue  
> Start WawiIC **after** the API server is running and accept the permissions.  
> The API key will then automatically be stored in the environment variables.  
> To use an already existing key, store it in the environment variable **`WAWIIC_APIKEY`**.

#### 2. OpenAI API Key
> You can create an OpenAI API key [here](https://platform.openai.com/api-keys).  
> Afterwards, store the key in the environment variable **`OPENAI_API_KEY`**.

#### 3. Images
> In order for WawiIC to merge article images or upload images to the parent item, the program needs an export of all item images in a single folder.  
> The images must follow this naming format:  
> **`[ItemSKU]-[ImageNumber].[jpg|png]`**  
> The easiest way to create this export is using JTL-Ameise.  
> After exporting, the path to this folder must be specified in the config file. More details [here](#config).

---

### Config

- Default path: `config/config.json`  
- Alternatively: provide a custom path using the `-c` flag:
  ```sh
  WawiIC.exe -c "D:\path\to\my\config.json"
  ```

---

#### Structure
- Type: JSON  
- Contains:
	- `api base url`: string — base URL of the JTL-Wawi API. Default: `"http://127.0.0.1:5883/api/eazybusiness/"`.  
      When the API server starts, it prints the exact URL.  
      Older versions may use: `"http://127.0.0.1:5883/api/eazybusiness/v1/"`.
	- `search mode`: string — controls how items are selected. Allowed values: `"category"`, `"supplier"`, `"none"`.  
	  However, category or supplier search currently may not work depending on the JTL-Wawi version.
	- `category id`: string — the category ID that will be assigned to the parent item.
	- `path to folder`: string — the path to the folder containing the images.
	- `activate sales channel`: bool — if `true`, the parent item is automatically activated in all sales channels.  
      This is required for automatic assignment of child items to the selected variations.

**Important:**
- Keys are **case-sensitive** and must match exactly (including spaces).
- JSON does **not** allow comments or trailing commas.
- File should be UTF-8 encoded.
- In JSON paths, `\` must be escaped with another `\`.

---

#### Example Config
```json
{
  "api base url": "http://127.0.0.1:5883/api/eazybusiness/",
  "search mode": "category",
  "category id": "155",
  "path to folder": "C:\\Users\\your-username\\Pictures\\JTL-Wawi-Images",
  "activate sales channel": true
}
```

---

### Starting the Application
> You can start the application by double-clicking it, or if you want to specify a custom config path, run:

```sh
WawiIC.exe -c "D:\\MyConfigs\\custom.json"
```

---

### Good to Know / Troubleshooting

- **Search mode does not work**:  
  Currently only search by item number or name works due to a bug in the JTL-Wawi API server (may be fixed in future versions).
- **Category/Supplier search mode is slow**:  
  With many categories, starting the app can take a long time, because the API first has to fetch everything.
- **Some items are not found**:  
  Sometimes searching by full name or full item number helps, or searching by only part of the name.
- **Program or API server freezes**:  
  Restart both WawiIC and the API server.
- **Review parent items**:  
  After merging items, always check item descriptions, images, and SEO description.  
  AI can make mistakes, so manual review is recommended.
- **Multiple searches**: When searching for items, selecting one or more of them, and then performing another search, the previously selected items remain selected.