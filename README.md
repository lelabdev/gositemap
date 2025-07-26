# GoSitemap [![GitHub release](https://img.shields.io/github/v/release/lelabdev/gositemap)](https://github.com/lelabdev/gositemap/releases/latest)

**GoSitemap** is a fast and minimal sitemap generator tailored for static **SvelteKit** sites.
It’s built with simplicity, performance, and automation in mind — perfect for CI/CD or local use.

---

## 🔥 What It Does

- Scans your SvelteKit `src/routes/` folder for all **static pages**
- Parses `.md` and `.svx` articles in `src/lib/content/` or any folder you define
- Builds a clean `sitemap.xml` with `<lastmod>` and optional `<changefreq>`
- Uses a single config file: `gositemap.toml` (can be auto-generated on first run)
- Outputs a ready-to-serve `static/sitemap.xml`
- 100% static, no server needed
- Built in Go — fast, lightweight, and dependency-free

---

## 🛠 How to Use

1. Put a `gositemap.toml` at your project root
   (or just run the tool and it’ll ask you interactively)

2. Run GoSitemap:

   ```sh
   go run .
   # or build:
   go build && ./gositemap

   ```

3. Your sitemap will be created at:

static/sitemap.xml

---

⚙️ CLI Options

`--help`, `-h` Show help and example config, then exit
`--dry-run` Output sitemap to stdout only
`--quiet` Suppress logs except errors
`preserve_existing` (in `gositemap.toml`) Controls how existing sitemap.xml files are handled.

---

## 🔄 Sitemap Overwrite Behavior (`preserve_existing`)

GoSitemap offers flexible control over how it updates your `sitemap.xml` file:

- **Default Behavior (or `preserve_existing = true` in `gositemap.toml`):**
  If a `sitemap.xml` file already exists, GoSitemap will **add only new entries** to it. Existing entries (and their `<lastmod>` dates) will be preserved and *not* updated. This is ideal for incremental updates where you want to maintain historical `<lastmod>` values.

- **Explicit Overwrite (`preserve_existing = false` in `gositemap.toml`):**
  If `preserve_existing` is explicitly set to `false`, GoSitemap will **regenerate the entire `sitemap.xml` file**. All entries, including existing ones, will have their `<lastmod>` dates updated based on the current scan. Use this when you want a fresh sitemap reflecting the latest modification times for all content.


---

🧠 Example gositemap.toml

```toml
base_url = "https://yoursite.com"
preserve_existing = true # Set to false to overwrite all entries and update lastmod dates

[content_types]
blog = "src/lib/content"
portfolio = "src/lib/portfolio"

# You can also use glob patterns to specify content directories.
# The slug will be the basename of the directory.
[[glob]]
paths = [
  "src/content/*"
]

[changefreq]
blog = "weekly"
portfolio = "monthly"
about = "yearly"

exclude = [
"/admin",
"/secret"
]
```

---

✨ Sample Output

```xml
<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>https://yoursite.com/</loc>
    <lastmod>2023-07-01</lastmod>
  </url>
  <url>
    <loc>https://yoursite.com/blog/article1</loc>
    <lastmod>2023-07-01</lastmod>
    <changefreq>weekly</changefreq>
  </url>
</urlset>
```

---

🧩 How It Works

Static pages: Finds all `+page.svelte` or `article.md` files in `src/routes/`

Articles: Includes `.md` / `.svx` from `content_types` folders

Exclusions: Ignores dynamic/param folders and anything in `exclude`.

  - If an exclusion starts with `/`, it's treated as a full path prefix (e.g., `/admin` excludes `/admin` and `/admin/users`).
  - Otherwise, it matches any directory segment in the URL (e.g., `(flow)` excludes `/blog/(flow)/post`).

lastmod: Uses `publishDate` (if found) or file mtime

changefreq: Defaults to never, customizable via [changefreq]

URL Order: Root → top-level pages → articles → subpages

---

## 📥 Installation

Download the binary for your OS from the [latest release](https://github.com/lelab/GoSitemap/releases/latest),
then make it executable (on Linux/macOS):

```sh
chmod +x gositemap
./gositemap
```

---

📦 Downloads

Grab the latest release:
➡️ [Latest Release](https://github.com/lelabdev/gositemap/releases/latest)

Each binary is compiled with optimizations for minimal size.
The version is embedded directly into the binary (-ldflags), no suffix needed.

---

✅ Requirements

Go 1.21 or newer

SvelteKit static site structure (uses /static, /routes, etc.)

---

🤔 Why GoSitemap?

Super light, no runtime dependencies

No YAML/JSON mess — config is TOML and explicit

CLI-first: works locally or in your CI

Clean, readable sitemap output

Works out of the box with SvelteKit projects

---

💬 FAQ

How do I change the <changefreq> per type?
Use the [changefreq] section in your TOML.

My page is missing from the sitemap!
Check if it’s dynamic (e.g. [slug]), in parentheses, or listed in exclude.

How to rename /blog/ to something else?
Just change the key in [content_types]:

articles = "src/lib/content"

→ This gives /articles/my-post.

My base_url is invalid!
It should start with http:// or https://, and no trailing slash.

---

🪪 License

MIT

---

Built for real-world static SvelteKit sites.
Scriptable. Predictable. No fluff.
