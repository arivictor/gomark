---
title: "Markdown Examples"
description: "GoMark supports a wide range of markdown features. Here are examples of the supported syntax."
order: 6
---

# Markdown Examples

GoMark supports several common Markdown and GitHub Flavored Markdown features.

## Tables

Tables are supported, including alignment and inline formatting.

```markdown
| Name | Role | Notes |
| :--- | :---: | ---: |
| Alice | Editor | `active` |
| Bob | Reviewer | **approved** |
| Charlie | Author | 3 |
```

| Name | Role | Notes |
| :--- | :---: | ---: |
| Alice | Editor | `active` |
| Bob | Reviewer | **approved** |
| Charlie | Author | 3 |

## Headings

Use headings to structure a page.

```markdown
# H1
## H2
### H3
```

# H1
## H2
### H3

## Emphasis

```markdown
This is **bold**, *italic*, and `inline code`.
```

This is **bold**, *italic*, and `inline code`.

## Links

```markdown
[GoMark documentation](/getting-started)
```

[GoMark documentation](/getting-started)

## Lists

```markdown
- First item
- Second item

1. First step
2. Second step
```

- First item
- Second item

1. First step
2. Second step

## Blockquotes

```markdown
> This is a quoted paragraph.
```

> This is a quoted paragraph.

## Code fences

````markdown
```go
fmt.Println("hello")
```
````

```go
fmt.Println("hello")
```

## Images

Images render with lazy loading.

```markdown
![A sample image](https://github.com/arivictor/gomark/blob/main/public/gomark-twitter-1200x628.png?raw=true)
```

![A sample image](https://github.com/arivictor/gomark/blob/main/public/gomark-twitter-1200x628.png?raw=true)

## Nested lists

```markdown
- First item
  - Nested item
- Second item
```

- First item
  - Nested item
- Second item

## Callouts

Admonition-style callouts are supported.

```markdown
> [!NOTE]
> This is a note callout.

> [!TIP] Use **bold** text inside a callout.

> [!IMPORTANT] This is an important callout.

> [!WARNING] This is a warning callout.

> [!CAUTION] This is a caution callout.
```

> [!NOTE]
> This is a note callout.

> [!TIP] Use **bold** text inside a callout.

> [!IMPORTANT] This is an important callout.

> [!WARNING] This is a warning callout.

> [!CAUTION] This is a caution callout.