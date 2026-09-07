# Plain singular Markdown

Status: Accepted

Fieldless documents such as AGENTS.md must work without YAML metadata. Today init creates empty delimiters, writes add them, and show rejects existing plain documents. Read and write fieldless singular replacement documents verbatim; keep fielded and section behavior intact. Allow the schema body to seed init so a documentation harness can provide useful default instructions. Repeated init preserves content; force restores the seed. Tests cover plain and empty content, Markdown thematic breaks, default initialization, preservation, and fielded compatibility. No general YAML parser or migration is included.
