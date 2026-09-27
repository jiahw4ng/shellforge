# Markdown rendering debug page

This page is intentionally long. It exercises common Markdown syntax, terminal
word wrapping, vertical scrolling, nested indentation, and syntax highlighting.

Plain text can contain punctuation, symbols, and Unicode: `!@#$%^&*()`, →, ✓,
λ, café, 日本語, and emoji 🚀.

## Paragraphs and line breaks

This is a normal paragraph. It contains enough text to wrap when the lesson
pane is narrow. The quick brown fox jumps over the lazy dog while ShellForge
keeps the instructional text within the left side of the terminal.

This is a second paragraph, separated from the first by a blank line.

This line ends with two spaces.  
This text should begin on the next line without starting a new paragraph.

## Emphasis and inline text

- *Italic text*
- _Italic text using underscores_
- **Bold text**
- __Bold text using underscores__
- ***Bold and italic text***
- ~~Strikethrough text~~
- `inline code`
- Text containing **bold with `inline code` inside it**
- Escaped Markdown characters: \*not italic\*, \# not a heading, and \[not a link\]

Use `pwd`, `ls -la`, and `printf '%s\n' "hello"` as short inline command
examples. A deliberately long inline value tests wrapping:
`SHELLFORGE_DEBUG_VARIABLE=abcdefghijklmnopqrstuvwxyz-0123456789`.

## Headings

# Heading level 1

## Heading level 2

### Heading level 3

#### Heading level 4

##### Heading level 5

###### Heading level 6

Alternative heading level 1
===========================

Alternative heading level 2
---------------------------

## Links and images

- [Example link](https://example.com)
- [Link with a title](https://example.com "Example website")
- <https://example.com>
- <debug@example.com>
- [Reference-style link][shellforge-reference]
- ![Image alt text](https://example.com/debug-image.png "Debug image")

[shellforge-reference]: https://example.com/shellforge "Reference link title"

## Block quotes

> This is a block quote.
>
> It can contain multiple paragraphs and **formatted text**.
>
> > This is a nested block quote.
> >
> > - Nested quotes can contain lists.
> > - They can also contain `inline code`.

## Unordered lists

- First item
- Second item
  - Nested item
  - Another nested item
    - Third-level item
- Third item with a longer description that should wrap onto another terminal
  line while retaining its list indentation.

* Asterisk marker
* Another asterisk item

+ Plus marker
+ Another plus item

## Ordered lists

1. First step
2. Second step
   1. Nested numbered step
   2. Another nested numbered step
3. Third step

10. A list can start at another number.
11. Numbering should continue from there.

## Task lists

- [x] Completed task
- [ ] Incomplete task
- [x] Task containing **bold text** and `code`

## Mixed nested content

1. Inspect the current directory.

   ```bash
   pwd
   ls -la
   ```

2. Read the quoted explanation.

   > The shell expands variables before running most commands.

3. Review the nested checklist.
   - [x] Commands displayed
   - [ ] Commands executed

## Thematic breaks

Three hyphens follow this paragraph.

---

Three asterisks follow this paragraph.

***

Three underscores follow this paragraph.

___

## Fenced code blocks

```bash
#!/usr/bin/env bash

name="ShellForge"
for item in one two three; do
    printf 'hello from %s: %s\n' "$name" "$item"
done
```

```go
package main

import "fmt"

func main() {
	fmt.Println("syntax highlighting")
}
```

```json
{
  "lesson": 0,
  "page": "debug",
  "enabled": true,
  "features": ["scrolling", "wrapping", "highlighting"]
}
```

```
This fenced block has no language identifier.
Characters such as * _ # and [link] should remain literal code.
```

## Indented code block

    echo "This is an indented code block"
    printf '%s\n' "The indentation should be preserved"

## Tables

| Feature | Syntax | Expected result |
| :--- | :---: | ---: |
| Bold | `**text**` | **text** |
| Italic | `*text*` | *text* |
| Code | `` `text` `` | `text` |
| Link | `[text](url)` | clickable or styled |

| Left aligned | Center aligned | Right aligned |
| :-- | :-: | --: |
| alpha | beta | 100 |
| a longer value that may wrap | gamma | 2,000 |

## Footnotes

ShellForge teaches command-line concepts one page at a time.[^lesson]
Footnotes can also contain multiple sentences.[^long-note]

[^lesson]: This is a short footnote used to test footnote rendering.
[^long-note]: This is a longer footnote.

    Its second paragraph is indented beneath the same footnote.

## Inline HTML

<details>
<summary>Expandable debug content</summary>

This text is inside an HTML details element. Terminal renderers may display,
simplify, or omit the HTML tags.

</details>

Text with a manual HTML break follows.<br>
This should appear after the break.

## Special characters and escaping

Backslash escapes: \` \* \_ \{ \} \[ \] \( \) \# \+ \- \. \!  
HTML entities: &amp; &lt; &gt; &quot; &copy;  
Shell operators shown as text: `|`, `||`, `&&`, `>`, `>>`, `<`, and `2>&1`.

## Long wrapping samples

This sentence contains a verylongunbrokentoken-that-keeps-going-abcdefghijklmnopqrstuvwxyz-0123456789-ABCDEFGHIJKLMNOPQRSTUVWXYZ to show how the renderer handles content that cannot be wrapped naturally at spaces.

The following path is intentionally long:
`/home/student/projects/shellforge/debug/examples/a-very-long-directory-name/another-directory/example.txt`.

## Final scrolling checklist

- The page can scroll from beginning to end.
- The title and page label remain visible while the content scrolls.
- `PgUp` moves toward the beginning.
- `PgDn` moves toward the end.
- Code blocks retain indentation.
- Long prose wraps within the lesson pane.
- Lists, quotes, tables, and headings remain readable.
- The final line is reachable.

**End of Markdown debug page.**
