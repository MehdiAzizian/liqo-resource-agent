# Makefile for LaTeX thesis compilation

# Main LaTeX file (without .tex extension)
MAIN = main

# LaTeX compiler
LATEX = pdflatex
BIBTEX = bibtex

# Flags for pdflatex
LATEX_FLAGS = -interaction=nonstopmode -halt-on-error

# Output PDF
PDF = $(MAIN).pdf

# Source files
TEX_FILES = $(wildcard chapters/*.tex) $(MAIN).tex
BIB_FILES = references.bib

.PHONY: all clean cleanall view help

# Default target
all: $(PDF)

# Build the PDF (full compilation with bibliography)
$(PDF): $(TEX_FILES) $(BIB_FILES)
	@echo "==> First pass: compiling LaTeX..."
	$(LATEX) $(LATEX_FLAGS) $(MAIN).tex
	@echo "==> Processing bibliography..."
	-$(BIBTEX) $(MAIN)
	@echo "==> Second pass: resolving citations..."
	$(LATEX) $(LATEX_FLAGS) $(MAIN).tex
	@echo "==> Third pass: finalizing references..."
	$(LATEX) $(LATEX_FLAGS) $(MAIN).tex
	@echo "==> Done! Output: $(PDF)"

# Quick build (single pass, for draft viewing)
draft: $(TEX_FILES)
	@echo "==> Quick draft build (single pass)..."
	$(LATEX) $(LATEX_FLAGS) $(MAIN).tex
	@echo "==> Done! Output: $(PDF)"

# Clean auxiliary files
clean:
	@echo "==> Cleaning auxiliary files..."
	rm -f *.aux *.log *.out *.toc *.lof *.lot *.bbl *.blg *.bcf *.run.xml
	rm -f *.fls *.fdb_latexmk *.synctex.gz *.nav *.snm *.vrb
	rm -f chapters/*.aux
	@echo "==> Clean complete."

# Clean everything including PDF
cleanall: clean
	@echo "==> Removing PDF..."
	rm -f $(PDF)
	@echo "==> Complete clean done."

# View the PDF (Linux)
view: $(PDF)
	@echo "==> Opening PDF..."
	@if command -v xdg-open > /dev/null; then \
		xdg-open $(PDF); \
	elif command -v open > /dev/null; then \
		open $(PDF); \
	else \
		echo "Error: No PDF viewer found. Please open $(PDF) manually."; \
	fi

# Count words (approximate, using detex if available)
wordcount:
	@if command -v detex > /dev/null; then \
		echo "==> Approximate word count:"; \
		detex $(MAIN).tex | wc -w; \
	else \
		echo "Error: detex not found. Install it for word counting."; \
	fi

# Spell check (requires aspell)
spell:
	@if command -v aspell > /dev/null; then \
		echo "==> Running spell check on all chapters..."; \
		for file in chapters/*.tex; do \
			echo "Checking $$file..."; \
			aspell check $$file; \
		done; \
	else \
		echo "Error: aspell not found. Install it for spell checking."; \
		echo "Ubuntu/Debian: sudo apt-get install aspell"; \
	fi

# List TODO items
todo:
	@echo "==> Searching for TODO items..."
	@grep -rn "TODO" chapters/*.tex $(MAIN).tex references.bib || echo "No TODO items found!"

# Check for common LaTeX errors
check:
	@echo "==> Checking for common issues..."
	@echo "Checking for overfull hboxes..."
	@grep -i "overfull" $(MAIN).log || echo "  No overfull hboxes."
	@echo "Checking for undefined references..."
	@grep -i "undefined" $(MAIN).log || echo "  No undefined references."
	@echo "Checking for missing citations..."
	@grep -i "citation.*undefined" $(MAIN).log || echo "  No missing citations."
	@echo "Checking for multiply defined labels..."
	@grep -i "multiply defined" $(MAIN).log || echo "  No multiply defined labels."

# Help target
help:
	@echo "Available targets:"
	@echo "  make          - Full compilation (default)"
	@echo "  make draft    - Quick single-pass compilation"
	@echo "  make clean    - Remove auxiliary files"
	@echo "  make cleanall - Remove all generated files including PDF"
	@echo "  make view     - Open the PDF"
	@echo "  make wordcount- Count words (requires detex)"
	@echo "  make spell    - Run spell checker (requires aspell)"
	@echo "  make todo     - List all TODO items"
	@echo "  make check    - Check for common LaTeX errors"
	@echo "  make help     - Show this help message"
