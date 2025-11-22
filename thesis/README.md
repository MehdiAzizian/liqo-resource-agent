# Master's Thesis: Automatic Cloud Resource Brokerage for Kubernetes

This directory contains the LaTeX source code for the master's thesis by Mehdi Azizian.

## Structure

```
thesis/
├── main.tex                  # Main thesis document
├── references.bib            # Bibliography database
├── chapters/                 # Individual chapters
│   ├── 00-titlepage.tex
│   ├── 01-abstract.tex
│   ├── 02-acknowledgments.tex
│   ├── 03-acronyms.tex
│   ├── 04-introduction.tex
│   ├── 05-state-of-art.tex
│   ├── 06-background.tex
│   ├── 07-system-design.tex
│   ├── 08-implementation.tex
│   ├── 09-evaluation.tex
│   ├── 10-conclusion.tex
│   └── 11-appendix.tex
├── figures/                  # Figures and images (TODO: create this)
├── Makefile                  # Build automation
└── README.md                 # This file
```

## Prerequisites

To compile the thesis, you need:

- **LaTeX Distribution**:
  - **Linux**: TeX Live (`sudo apt-get install texlive-full`)
  - **macOS**: MacTeX (`brew install --cask mactex`)
  - **Windows**: MiKTeX or TeX Live

- **LaTeX Editor** (optional but recommended):
  - TeXstudio
  - Overleaf (online)
  - VS Code with LaTeX Workshop extension

## Compilation

### Using Make (Recommended)

```bash
# Compile the thesis
make

# Clean auxiliary files
make clean

# Clean everything including PDF
make cleanall

# Open the PDF (Linux)
make view
```

### Manual Compilation

```bash
# First pass
pdflatex main.tex

# Process bibliography
bibtex main

# Process acronyms (if using)
makeindex main.nlo -s nomencl.ist -o main.nls

# Two more passes to resolve references
pdflatex main.tex
pdflatex main.tex
```

The output will be `main.pdf`.

## TODO List

Before submission, complete the following tasks:

### Content TODOs

- [ ] Fill in university/faculty information in titlepage
- [ ] Complete abstract with actual results
- [ ] Write personal acknowledgments
- [ ] Review and expand introduction
- [ ] Add actual references to `references.bib`
- [ ] Complete state-of-the-art review with real citations
- [ ] Add implementation details specific to your work
- [ ] **Run experiments and add actual results to evaluation chapter**
- [ ] Add experimental figures and graphs
- [ ] Review and refine conclusion
- [ ] Add appendix content as needed

### Figures TODOs

Create the `figures/` directory and add:

- [ ] `university-logo.png` - Your university logo
- [ ] `kubernetes-architecture.pdf` - K8s architecture diagram
- [ ] `rear-protocol.pdf` - REAR protocol phases diagram
- [ ] `system-architecture.pdf` - Overall system architecture
- [ ] `resource-agent-components.pdf` - RA component diagram
- [ ] `resource-broker-components.pdf` - RB component diagram
- [ ] `reservation-states.pdf` - Reservation state machine
- [ ] `deployment-topology.pdf` - Experimental setup topology
- [ ] `agent-broker-communication.pdf` - Sequence diagram
- [ ] `client-broker-communication.pdf` - Sequence diagram
- [ ] `resource-tracking-accuracy.pdf` - Accuracy graph
- [ ] `placement-latency.pdf` - Latency vs clusters graph
- [ ] `algorithm-comparison.pdf` - Algorithm comparison graph
- [ ] `agent-overhead.pdf` - Overhead vs cluster size
- [ ] `case-study-cost.pdf` - Cost savings over time

### References TODOs

In `references.bib`, replace all `TODO:` entries with actual references:

- [ ] Find and add Kubernetes papers and books
- [ ] Add Liqo paper (if published)
- [ ] Add REAR protocol paper (if exists)
- [ ] Add multi-cluster management papers
- [ ] Add resource allocation papers
- [ ] Add distributed systems papers
- [ ] Add any papers you read and referenced

### Formatting TODOs

- [ ] Ensure consistent citation style
- [ ] Check that all figures are referenced in text
- [ ] Verify all tables are properly formatted
- [ ] Ensure code listings are readable
- [ ] Check page breaks and spacing
- [ ] Verify table of contents is complete
- [ ] Check that acronyms are defined before use

## Tips

### Writing Tips

1. **Be Specific**: Replace all `TODO` markers with concrete information
2. **Use Active Voice**: Prefer "We implemented..." over "It was implemented..."
3. **Cite Sources**: Every claim should have a citation
4. **Explain Figures**: Each figure should be explained in the text
5. **Be Concise**: Academic writing should be clear and to the point

### LaTeX Tips

1. **References**: Use `\cite{key}` for citations
2. **Labels**: Use descriptive labels like `\label{fig:system-arch}`
3. **Cross-references**: Use `\ref{}` for figures, tables, chapters
4. **Equations**: Use `\eqref{}` for equation numbers
5. **Code**: Use `lstlisting` environment for code blocks

### Common LaTeX Commands

```latex
% Figures
\begin{figure}[ht]
    \centering
    \includegraphics[width=0.8\textwidth]{figures/example.pdf}
    \caption{Caption text}
    \label{fig:example}
\end{figure}

% Tables
\begin{table}[ht]
\centering
\caption{Table caption}
\label{tab:example}
\begin{tabular}{@{}lcc@{}}
\toprule
\textbf{Header1} & \textbf{Header2} & \textbf{Header3} \\ \midrule
Row1 & Value1 & Value2 \\
Row2 & Value3 & Value4 \\ \bottomrule
\end{tabular}
\end{table}

% Citations
According to~\cite{paper-key}, the approach...

% Cross-references
As shown in Figure~\ref{fig:example}, the system...
See Section~\ref{sec:background} for details.
```

## Troubleshooting

### Bibliography not showing

Run `bibtex main` after first `pdflatex` pass, then run `pdflatex` twice more.

### Missing figures

Ensure figures exist in `figures/` directory and paths in `\includegraphics{}` are correct.

### Undefined references

Run `pdflatex` multiple times (usually 2-3 passes needed).

### Package errors

Install missing packages:
```bash
# Ubuntu/Debian
sudo apt-get install texlive-latex-extra texlive-science

# macOS (if using MacTeX, usually has everything)
# Packages are typically included

# Manual package installation (if needed)
tlmgr install <package-name>
```

## Submission Checklist

Before submitting your thesis:

- [ ] All TODO items completed
- [ ] All figures added and referenced
- [ ] All references complete and properly formatted
- [ ] Spell-check completed
- [ ] Grammar check completed
- [ ] Consistent formatting throughout
- [ ] Page numbers correct
- [ ] Table of contents accurate
- [ ] Figures and tables listed correctly
- [ ] No LaTeX compilation warnings
- [ ] PDF bookmarks working (if required)
- [ ] File named according to submission requirements
- [ ] Printed version reviewed (if required)

## License

This thesis is © Mehdi Azizian, 2024-2025. All rights reserved.

The source code for the Resource Agent and Resource Broker is available under Apache License 2.0:
- https://github.com/MehdiAzizian/liqo-resource-agent
- https://github.com/MehdiAzizian/liqo-resource-broker

## Contact

For questions about this thesis:
- **Author**: Mehdi Azizian
- **GitHub**: https://github.com/MehdiAzizian

Good luck with your thesis! 🎓
